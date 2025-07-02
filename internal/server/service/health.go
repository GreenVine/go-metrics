package service

import (
	"context"
	"time"

	"github.com/greenvine/go-metrics/proto/gen/core/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HealthService implements the HealthService gRPC service
type HealthService struct {
	corev1.UnimplementedHealthServiceServer

	startTime time.Time
}

// NewHealthService creates a new instance of the health service
func NewHealthService() *HealthService {
	return &HealthService{
		startTime: time.Now(),
	}
}

// Get returns the health status of the service
func (s *HealthService) Get(ctx context.Context, _ *emptypb.Empty) (*corev1.Healthz, error) {
	return &corev1.Healthz{}, nil
}
