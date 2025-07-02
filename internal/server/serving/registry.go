package serving

import (
	"github.com/greenvine/go-metrics/internal/server/service"
	"github.com/greenvine/go-metrics/proto/gen/core/v1"
	"github.com/greenvine/go-metrics/proto/gen/device/v1"
	"google.golang.org/grpc"
)

func RegisterServices(s *grpc.Server) {
	corev1.RegisterObservabilityServiceServer(s, service.NewObservabilityV1Service())
	devicev1.RegisterDeviceMgmtServiceServer(s, service.NewDeviceMgmtV1Service())
}
