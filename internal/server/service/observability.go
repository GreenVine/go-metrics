package service

import (
  "context"
  "time"

  "github.com/greenvine/go-metrics/proto/gen/core/v1"
  "google.golang.org/protobuf/types/known/emptypb"
)

// ObservabilityService implements the ObservabilityService gRPC service.
type ObservabilityService struct {
  corev1.UnimplementedObservabilityServiceServer

  startTime time.Time
}

// NewObservabilityService creates a new instance of the health service.
func NewObservabilityService() *ObservabilityService {
  return &ObservabilityService{
    startTime: time.Now(),
  }
}

// GetHealthInfo returns the health status of the service.
func (s *ObservabilityService) GetHealthInfo(ctx context.Context, _ *emptypb.Empty) (*corev1.Healthz, error) {
  return &corev1.Healthz{}, nil
}
