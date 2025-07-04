package database

import "github.com/google/uuid"

// ListAlerts retrieves the most recent alerts for a device.
func ListAlerts(deviceID uuid.UUID, limit int) ([]AlertRecord, error) {
	var alertRecords []AlertRecord

	result := DB.Where("device_id = ?", deviceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&alertRecords)

	return alertRecords, result.Error
}
