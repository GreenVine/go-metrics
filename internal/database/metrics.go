package database

import (
	"context"
	"errors"
	"log"
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

	telemetry.Add(metricRecord.Proto())

	// Run alert checks asynchronously in the background.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		if err := WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return checkAndMaybeCreateAlerts(tx, metricRecord)
		}); err != nil {
			log.Printf("Failed to check or create alerts for metric record %q: %v", metricRecord.ID, err)
		}
	}()

	return metricRecord, nil
}

// checkAndMaybeCreateAlerts checks if a metric has met alert conditions and creates alerts if needed.
func checkAndMaybeCreateAlerts(db *gorm.DB, metricRecord *MetricRecord) error {
	configRecord, err := getDeviceConfig(db, metricRecord.DeviceID)
	if err != nil {
		// Skip alert creation since the config cannot be found.
		return nil
	}

	// Alerts to be inserted into the database.
	var alerts []*AlertRecord

	if metricRecord.Temperature > configRecord.TemperatureThreshold {
		alerts = append(alerts, &AlertRecord{
			DeviceID:  metricRecord.DeviceID,
			Reason:    AlertReason(devicev1.AlertReason_ALERT_REASON_TEMPERATURE),
			Value:     metricRecord.Temperature,
			Threshold: configRecord.TemperatureThreshold,

			BaseRecord: BaseRecord{
				CreatedAt: metricRecord.CreatedAt,
			},
		})
	}

	if metricRecord.Battery < configRecord.BatteryThreshold {
		alerts = append(alerts, &AlertRecord{
			DeviceID:  metricRecord.DeviceID,
			Reason:    AlertReason(devicev1.AlertReason_ALERT_REASON_BATTERY),
			Value:     float32(metricRecord.Battery),
			Threshold: float32(configRecord.BatteryThreshold),

			BaseRecord: BaseRecord{
				CreatedAt: metricRecord.CreatedAt,
			},
		})
	}

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
