package database

import "github.com/google/uuid"

// GetOrCreateDevice returns an existing device or creates one if a device with the given ID doesn't exist.
func GetOrCreateDevice(deviceID uuid.UUID) (*DeviceRecord, error) {
	var device DeviceRecord

	result := DB.FirstOrCreate(&device, DeviceRecord{
		BaseRecord{
			ID: deviceID,
		},
	})

	return &device, result.Error
}
