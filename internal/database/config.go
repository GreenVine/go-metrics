package database

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

// UpsertConfig creates or updates a device configuration.
func UpsertConfig(db *gorm.DB, deviceID uuid.UUID, config *devicev1.Config) (*ConfigRecord, error) {
	// Creates the device if it doesn't exist yet.
	if _, err := getOrCreateDevice(db, deviceID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create device %q: %v", deviceID, err)
	}

	configRecord, err := getDeviceConfig(db, deviceID)

	var upsertResult *gorm.DB

	if err == nil {
		// Config already exists, update it.
		configRecord.TemperatureThreshold = config.GetTemperatureThreshold()
		configRecord.BatteryThreshold = config.GetBatteryThreshold()
		upsertResult = db.Save(&configRecord)
	} else {
		// Create a new device config.
		configRecord = &ConfigRecord{
			DeviceID:             deviceID,
			TemperatureThreshold: config.GetTemperatureThreshold(),
			BatteryThreshold:     config.GetBatteryThreshold(),
		}
		upsertResult = db.Create(&configRecord)
	}

	if upsertResult.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to upsert config for device %q: %v", deviceID, upsertResult.Error)
	}

	return configRecord, nil
}

// getDeviceConfig retrieves the config for a given device.
func getDeviceConfig(db *gorm.DB, deviceID uuid.UUID) (*ConfigRecord, error) {
	var configRecord ConfigRecord

	result := db.First(&configRecord, "device_id = ?", deviceID)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve config for device %q: %v", deviceID, result.Error)
	}

	return &configRecord, nil
}
