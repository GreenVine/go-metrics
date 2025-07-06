// Package serving contains implementations for request serving such as middlewares and interceptors.
package serving

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// LoggingInterceptor creates a UnaryServerInterceptor for logging requests and responses.
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	method := extractRequestMethod(req)
	log.Printf("[%s] req = \"%s\"", method, req)

	resp, err := handler(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		log.Printf("[%s] err (code = %s) = %v",
			method, st.Code(), st.Message(),
		)
	} else {
		log.Printf("[%s] res = \"%s\"", method, resp)
	}

	return resp, err
}

func extractRequestMethod(maybeProto any) string {
	if p, ok := maybeProto.(proto.Message); ok {
		return string(p.ProtoReflect().Descriptor().FullName())
	}

	return ""
}
