package main

import (
	"github.com/greenvine/go-metrics/internal/server/service"
	"github.com/greenvine/go-metrics/internal/server/serving"
)

var rateLimitConfig = []serving.RateLimitConfig{
	serving.NewRateLimiterConfig(
		"greenvine.gometrics.device.v1.CreateMetricRequest").
		WithQPSLimit(10).
		WithKeyExtractor(resourceNameExtractor),

	serving.NewRateLimiterConfig(
		"greenvine.gometrics.device.v1.UpsertConfigRequest").
		WithQPSLimit(5).
		WithKeyExtractor(resourceNameExtractor),

	serving.NewRateLimiterConfig(
		"greenvine.gometrics.device.v1.ListAlertsRequest").
		WithQPSLimit(10),

	serving.NewRateLimiterConfig(
		"greenvine.gometrics.core.v1.GetHealthInfo").
		WithQPSLimit(10),
}

// resourceNameExtractor extracts the resource in requests.
func resourceNameExtractor(fullMethodName string, req any) string {
	switch r := req.(type) {
	case interface{ GetParent() string }:
		if service.ExtractDeviceID(r.GetParent()) != nil {
			return fullMethodName + "@" + r.GetParent()
		}
	}

	return fullMethodName
}
