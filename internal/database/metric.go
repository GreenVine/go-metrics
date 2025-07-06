package database

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/greenvine/go-metrics/internal/telemetry"
	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

// CreateMetric stores a new metric in the database.
func CreateMetric(db *gorm.DB, deviceID uuid.UUID, metric *devicev1.Metric) (*MetricRecord, error) {
	now := time.Now()
	metricRecord := &MetricRecord{
		DeviceID:    deviceID,
		Temperature: metric.GetTemperature(),
		Battery:     metric.GetBattery(),

		BaseRecord: BaseRecord{
			CreatedAt: now,
		},
	}

	// Inserts the metric to the database.
	if result := db.Create(metricRecord); errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return nil, status.Errorf(codes.FailedPrecondition, "device %q does not exist", deviceID)
	} else if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to store metric for device %q: %v", deviceID, result.Error)
	}

	alertGen.QueueMetric(metricRecord)
	telemetry.Add(metricRecord.Proto())

	return metricRecord, nil
}

// checkAndMaybeCreateAlerts checks if metrics have met alert conditions and creates alerts if needed.
func checkAndMaybeCreateAlerts(db *gorm.DB, metricRecords []*MetricRecord) error {
	if len(metricRecords) == 0 {
		return nil
	}

	// Extract unique device IDs from metrics.
	deviceIDs := make(map[uuid.UUID]bool)
	for _, metric := range metricRecords {
		deviceIDs[metric.DeviceID] = true
	}

	var configs []ConfigRecord
	deviceIDList := make([]uuid.UUID, 0, len(deviceIDs))
	for deviceID := range deviceIDs {
		deviceIDList = append(deviceIDList, deviceID)
	}

	if err := db.Where("device_id IN ?", deviceIDList).Find(&configs).Error; err != nil {
		return err
	}

	// Create a map of device ID to device config.
	configMap := make(map[uuid.UUID]*ConfigRecord, len(configs))
	for i := range configs {
		configMap[configs[i].DeviceID] = &configs[i]
	}

	// List of alerts to be created, across all devices.
	var alerts []*AlertRecord

	// Process each metric and gather alerts
	for _, metric := range metricRecords {
		config, ok := configMap[metric.DeviceID]
		if !ok {
			continue
		}

		if metric.Temperature > config.TemperatureThreshold {
			alerts = append(alerts, &AlertRecord{
				DeviceID:  metric.DeviceID,
				Reason:    AlertReason(devicev1.AlertReason_ALERT_REASON_TEMPERATURE),
				Value:     metric.Temperature,
				Threshold: config.TemperatureThreshold,
				BaseRecord: BaseRecord{
					CreatedAt: metric.CreatedAt,
				},
			})
		}

		if metric.Battery < config.BatteryThreshold {
			alerts = append(alerts, &AlertRecord{
				DeviceID:  metric.DeviceID,
				Reason:    AlertReason(devicev1.AlertReason_ALERT_REASON_BATTERY),
				Value:     float32(metric.Battery),
				Threshold: float32(config.BatteryThreshold),
				BaseRecord: BaseRecord{
					CreatedAt: metric.CreatedAt,
				},
			})
		}
	}

	// Insert all alerts in a single database operation
	if len(alerts) > 0 {
		if err := db.Create(&alerts).Error; err != nil {
			return err
		}

		// Log alerts to telemetry
		for _, alert := range alerts {
			telemetry.Add(alert.Proto())
		}
	}

	return nil
}
