package database

import (
	"time"

	"database/sql/driver"
	"github.com/google/uuid"
	"github.com/greenvine/go-metrics/proto/gen/device/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// BaseRecord contains common columns for all tables with a primary key.
type BaseRecord struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	CreatedAt time.Time
}

// BeforeCreate will generate a UUID automatically if not specified.
func (b *BaseRecord) BeforeCreate(txn *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	return nil
}

// DeviceRecord represents a device in the database.
type DeviceRecord struct {
	BaseRecord
}

func (*DeviceRecord) TableName() string {
	return "devices"
}

// MetricRecord represents a metric in the database.
type MetricRecord struct {
	BaseRecord

	DeviceID    uuid.UUID    `gorm:"type:uuid"`
	Device      DeviceRecord `gorm:"foreignKey:DeviceID"`
	Temperature float32
	Battery     int32
}

func (*MetricRecord) TableName() string {
	return "metrics"
}

// Proto converts a database MetricRecord to the Protobuf representation.
func (m *MetricRecord) Proto() *devicev1.Metric {
	resourceName := "devices/" + m.DeviceID.String() + "/metrics/" + m.ID.String()

	return &devicev1.Metric{
		Name:        &resourceName,
		CreateTime:  timestamppb.New(m.CreatedAt),
		Temperature: &m.Temperature,
		Battery:     &m.Battery,
	}
}

// ConfigRecord represents a device configuration in the database.
type ConfigRecord struct {
	DeviceID             uuid.UUID    `gorm:"primaryKey;type:uuid"`
	Device               DeviceRecord `gorm:"foreignKey:DeviceID"`
	TemperatureThreshold float32
	BatteryThreshold     int32
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (*ConfigRecord) TableName() string {
	return "configurations"
}

// Proto converts a database ConfigRecord to the Protobuf representation.
func (c *ConfigRecord) Proto() *devicev1.Config {
	resourceName := "devices/" + c.DeviceID.String() + "/config"

	return &devicev1.Config{
		Name:                 &resourceName,
		TemperatureThreshold: &c.TemperatureThreshold,
		BatteryThreshold:     &c.BatteryThreshold,
	}
}

type AlertReason devicev1.AlertReason

// Scan converts the integer to the Protobuf enum.
func (r *AlertReason) Scan(value devicev1.AlertReason) error {
	*r = AlertReason(*value.Enum())
	return nil
}

// Value converts the Protobuf enum to the integer for storage.
func (r *AlertReason) Value() (driver.Value, error) {
	return int64(*r), nil
}

// AlertRecord represents an alert in the database.
type AlertRecord struct {
	BaseRecord

	DeviceID  uuid.UUID    `gorm:"type:uuid"`
	Device    DeviceRecord `gorm:"foreignKey:DeviceID"`
	Reason    AlertReason
	Value     float32
	Threshold float32
}

func (*AlertRecord) TableName() string {
	return "alerts"
}

// Proto converts a database AlertRecord to the Protobuf representation.
func (a *AlertRecord) Proto() *devicev1.Alert {
	resourceName := "devices/" + a.DeviceID.String() + "/alerts/" + a.ID.String()

	return &devicev1.Alert{
		Name:       &resourceName,
		Reason:     devicev1.AlertReason(a.Reason).Enum(),
		Value:      &a.Value,
		Threshold:  &a.Threshold,
		CreateTime: timestamppb.New(a.CreatedAt),
	}
}

// AlertRecordsToProto converts multiple database alerts to their Protobuf representation.
func AlertRecordsToProto(alerts []AlertRecord) []*devicev1.Alert {
	protoAlerts := make([]*devicev1.Alert, len(alerts))
	for i, alert := range alerts {
		protoAlerts[i] = alert.Proto()
	}

	return protoAlerts
}
