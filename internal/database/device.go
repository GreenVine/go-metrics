package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// getOrCreateDevice returns an existing device or creates one if a device with the given ID doesn't exist.
func getOrCreateDevice(db *gorm.DB, deviceID uuid.UUID) (*DeviceRecord, error) {
	var device DeviceRecord

	result := db.Clauses(clause.Locking{
		Strength: clause.LockingOptionsSkipLocked,
	}).FirstOrCreate(&device, DeviceRecord{
		BaseRecord{
			ID: deviceID,
		},
	})

	return &device, result.Error
}
