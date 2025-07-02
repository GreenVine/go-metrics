package registry

import (
	"github.com/greenvine/go-metrics/internal/server/service"
	"github.com/greenvine/go-metrics/proto/gen/core/v1"
	"google.golang.org/grpc"
)

func RegisterServices(s *grpc.Server) {
	corev1.RegisterObservabilityServiceServer(s, service.NewObservabilityService())
}
