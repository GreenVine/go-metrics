package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/greenvine/go-metrics/internal/server/serving"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const serverLogPrefix = "[metrics-server] "

func main() {
	host := flag.String("host", "", "Binding address")
	port := flag.Int("port", 3000, "Listening port")
	flag.Parse()

	log.SetPrefix(serverLogPrefix)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}

	// Create the server with reflection support.
	server := grpc.NewServer(
		grpc.UnaryInterceptor(serving.LoggingInterceptor),
	)
	reflection.Register(server)
	serving.RegisterServices(server)

	log.Printf("Starting metrics server on %s", listener.Addr().String())
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Failed to start the metrics server: %v", err)
		}
	}()

	// Wait for the interrupt signal, then stop the server.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Metrics server is shutting down...")
	server.GracefulStop()

	log.Println("Bye.")
}
