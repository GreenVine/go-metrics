package service

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/greenvine/go-metrics/proto/gen/core/v1"
)

// ObservabilityV1Service implements the ObservabilityService V1 gRPC service.
type ObservabilityV1Service struct {
	corev1.UnimplementedObservabilityServiceServer

	startTime time.Time
}

// NewObservabilityV1Service creates a new instance of the health V1 service.
func NewObservabilityV1Service() *ObservabilityV1Service {
	return &ObservabilityV1Service{
		startTime: time.Now(),
	}
}

// GetHealthInfo returns the health status of the service.
func (s *ObservabilityV1Service) GetHealthInfo(ctx context.Context, _ *emptypb.Empty) (*corev1.Healthz, error) {
	return &corev1.Healthz{}, nil
}
