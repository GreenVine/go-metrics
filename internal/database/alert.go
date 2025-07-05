package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListAlerts retrieves the most recent alerts for a device.
func ListAlerts(db *gorm.DB, deviceID uuid.UUID, limit int) ([]AlertRecord, error) {
	var alertRecords []AlertRecord

	result := db.Where("device_id = ?", deviceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&alertRecords)

	return alertRecords, result.Error
}
