package serving

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor creates a UnaryServerInterceptor for logging requests and responses.
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	method := strings.TrimLeft(info.FullMethod, "/")
	log.Printf("[%s] req = %s", method, req)

	resp, err := handler(ctx, req)

	if err != nil {
		st, _ := status.FromError(err)
		log.Printf("[%s] err (code = %s) = %v",
			method, st.Code(), st.Message(),
		)
	} else {
		log.Printf("[%s] res = %s", method, resp)
	}

	return resp, err
}
