package database

import (
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

// CreateMetric stores a new metric in the database.
func CreateMetric(deviceID uuid.UUID, metric *devicev1.Metric) (*MetricRecord, error) {
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
	if result := DB.Create(metricRecord); errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
		return nil, status.Errorf(codes.FailedPrecondition, "device %q does not exist", deviceID)
	} else if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to store metric for device %q: %v", deviceID, result.Error)
	}

	err := checkAndMaybeCreateAlerts(metricRecord)
	if err != nil {
		log.Printf("Failed to process alerts for metric: %v", err)
		// continue processing even if alert creation has failed.
	}

	return metricRecord, nil
}

// checkAndMaybeCreateAlerts checks if a metric has met alert conditions and creates alerts if needed.
func checkAndMaybeCreateAlerts(metricRecord *MetricRecord) error {
	configRecord, err := GetDeviceConfig(metricRecord.DeviceID)
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
		err := DB.Create(&alerts).Error
		if err != nil {
			return err
		}
	}

	return nil
}
