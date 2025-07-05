package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// getOrCreateDevice returns an existing device or creates one if a device with the given ID doesn't exist.
func getOrCreateDevice(db *gorm.DB, deviceID uuid.UUID) (*DeviceRecord, error) {
	var device DeviceRecord

	result := db.FirstOrCreate(&device, DeviceRecord{
		BaseRecord{
			ID: deviceID,
		},
	})

	return &device, result.Error
}
