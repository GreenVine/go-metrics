package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"buf.build/go/protovalidate"
	pbvmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/greenvine/go-metrics/internal/database"
	"github.com/greenvine/go-metrics/internal/server/serving"
)

const logPrefix = "[metrics-server] "

func main() {
	host := flag.String("host", "", "Binding address")
	port := flag.Int("port", 3000, "Listening port")
	dbPath := flag.String("dbPath", "go-metrics.db", "Path to SQLite database file")
	alertGenInterval := flag.Duration("alertGenInterval", 2*time.Second, "Interval between alert generation attempts")

	flag.Parse()

	log.SetPrefix(logPrefix)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	// Init request proto validator
	validator, err := protovalidate.New()
	if err != nil {
		log.Fatalf("Failed to initialise proto validator: %v", err)
	}

	// Init database
	err = database.Init(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialise database: %v", err)
	}

	// Init async alert generator with the root context
	database.InitAlertGenerator(rootCtx, *alertGenInterval)

	addr := fmt.Sprintf("%s:%d", *host, *port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	rateLimiter := serving.NewTokenBucketRateLimiter(rateLimitConfig)

	// Create the server with reflection support.
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			serving.LoggingInterceptor,
			serving.RateLimitInterceptor(rateLimiter),
			pbvmiddleware.UnaryServerInterceptor(validator),
		),
	)
	reflection.Register(server)
	serving.RegisterServices(server)

	log.Printf("Starting metrics server on %s", listener.Addr().String())

	go func() {
		err := server.Serve(listener)
		if err != nil {
			log.Fatalf("Failed to start the metrics server: %v", err)
		}
	}()

	// Wait for the interrupt signal, then stop the server.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Metrics server is shutting down...")
	server.GracefulStop()

	rootCancel()
}
