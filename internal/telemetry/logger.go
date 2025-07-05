// Package telemetry provides utilities for tracking service operations.
package telemetry

import (
	"sort"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/greenvine/go-metrics/proto/gen/core/v1"
	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

var logger = NewLogger()

// Logger stores global telemetry data.
type Logger struct {
	sync.RWMutex

	metrics []*devicev1.Metric
	alerts  []*devicev1.Alert
	configs []*devicev1.Config
}

// NewLogger returns a new logger instance.
func NewLogger() *Logger {
	return &Logger{
		metrics: make([]*devicev1.Metric, 0),
		alerts:  make([]*devicev1.Alert, 0),
		configs: make([]*devicev1.Config, 0),
	}
}

// Add appends a telemetry entry to the appropriate history list.
func Add(entry proto.Message) {
	if entry == nil {
		return
	}

	logger.Lock()
	defer logger.Unlock()

	// Clone the proto to avoid concurrent modification and append it to the corresponding slice.
	switch v := entry.(type) {
	case *devicev1.Metric:
		logger.metrics = append(logger.metrics, proto.Clone(v).(*devicev1.Metric))
	case *devicev1.Alert:
		logger.alerts = append(logger.alerts, proto.Clone(v).(*devicev1.Alert))
	case *devicev1.Config:
		logger.configs = append(logger.configs, proto.Clone(v).(*devicev1.Config))
	}
}

// Logs returns all logged data.
func Logs() *corev1.Logs {
	logger.RLock()
	defer logger.RUnlock()

	metrics := make([]*devicev1.Metric, len(logger.metrics))
	copy(metrics, logger.metrics)
	sort.Slice(metrics, func(i, j int) bool {
		time1 := metrics[i].GetCreateTime().AsTime()
		time2 := metrics[j].GetCreateTime().AsTime()
		return time1.After(time2)
	})

	alerts := make([]*devicev1.Alert, len(logger.alerts))
	copy(alerts, logger.alerts)
	sort.Slice(alerts, func(i, j int) bool {
		time1 := alerts[i].GetCreateTime().AsTime()
		time2 := alerts[j].GetCreateTime().AsTime()
		return time1.After(time2)
	})

	configs := make([]*devicev1.Config, len(logger.configs))
	copy(configs, logger.configs)
	sort.Slice(configs, func(i, j int) bool {
		time1 := configs[i].GetUpdateTime().AsTime()
		time2 := configs[j].GetUpdateTime().AsTime()
		return time1.After(time2)
	})

	return &corev1.Logs{
		MetricHistory:     metrics,
		ThresholdBreaches: alerts,
		ConfigUpdates:     configs,
	}
}
