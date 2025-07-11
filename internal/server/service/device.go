// Package service contains implementations for RPC services.
package service

import (
	"context"
	"regexp"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/greenvine/go-metrics/internal/database"
	"github.com/greenvine/go-metrics/proto/gen/device/v1"
)

// deviceIDRegex extracts the device ID from the resource name.
var deviceIDRegex = regexp.MustCompile(`^devices/([^/]+)`)

// DeviceMgmtV1Service implements the DeviceMgmtService V1 gRPC service.
type DeviceMgmtV1Service struct {
	devicev1.UnimplementedDeviceMgmtServiceServer
}

// NewDeviceMgmtV1Service creates a new instance of the device management V1 service.
func NewDeviceMgmtV1Service() *DeviceMgmtV1Service {
	return &DeviceMgmtV1Service{}
}

// CreateMetric accepts metrics data from a device.
func (s *DeviceMgmtV1Service) CreateMetric(ctx context.Context, req *devicev1.CreateMetricRequest) (*devicev1.Metric, error) {
	deviceID := ExtractDeviceID(req.GetParent())
	if deviceID == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device resource name: %q", req.GetParent())
	}

	newCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var metricRecord *database.MetricRecord
	var err error

	err = database.WithContext(newCtx).Transaction(func(tx *gorm.DB) error {
		metricRecord, err = database.CreateMetric(tx, *deviceID, req.GetMetric())
		return err
	})

	if err != nil {
		return nil, err
	}

	return metricRecord.Proto(), nil
}

// UpsertConfig upserts configuration for a specific device.
func (s *DeviceMgmtV1Service) UpsertConfig(ctx context.Context, req *devicev1.UpsertConfigRequest) (*devicev1.Config, error) {
	deviceID := ExtractDeviceID(req.GetParent())
	if deviceID == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device resource name: %q", req.GetParent())
	}

	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var configRecord *database.ConfigRecord
	var err error

	err = database.WithContext(newCtx).Transaction(func(tx *gorm.DB) error {
		configRecord, err = database.UpsertConfig(tx, *deviceID, req.GetConfig())
		return err
	})

	if err != nil {
		return nil, err
	}

	return configRecord.Proto(), nil
}

// ListAlerts retrieves alerts for a device.
func (s *DeviceMgmtV1Service) ListAlerts(ctx context.Context, req *devicev1.ListAlertsRequest) (*devicev1.ListAlertsResponse, error) {
	deviceID := ExtractDeviceID(req.GetParent())
	if deviceID == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid device resource name: %q", req.GetParent())
	}

	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50 // default page size
	}

	newCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	var alerts []database.AlertRecord
	var err error

	err = database.WithContext(newCtx).Transaction(func(tx *gorm.DB) error {
		alerts, err = database.ListAlerts(tx, *deviceID, int(pageSize))
		return err
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve alerts for device %q: %v", deviceID, err)
	}

	return &devicev1.ListAlertsResponse{
		Alerts: database.AlertRecordsToProto(alerts),
	}, nil
}

func ExtractDeviceID(resourceName string) *uuid.UUID {
	if matches := deviceIDRegex.FindStringSubmatch(resourceName); len(matches) == 2 {
		if parsedUUID, err := uuid.Parse(matches[1]); err == nil {
			return &parsedUUID
		}
	}

	return nil
}
