package service

import (
	"errors"
	"fleetmetrics/internal/store"
	"log/slog"
	"time"
)

var ErrDeviceNotFound = errors.New("device not found")

type Fleet struct {
	store  store.Store
	logger *slog.Logger
}

func New(s store.Store, logger *slog.Logger) *Fleet {
	return &Fleet{store: s, logger: logger}
}

func (f *Fleet) RegisterDevices(deviceIDs []string) {
	for _, id := range deviceIDs {
		f.store.Register(id)
	}
	f.logger.Debug("devices registered", "count", len(deviceIDs))
}

func (f *Fleet) RecordHeartbeat(deviceID string, sentAt time.Time) error {
	d, err := f.store.Get(deviceID)
	if err != nil {
		f.logger.Warn("heartbeat for unknown device", "device_id", deviceID)
		return ErrDeviceNotFound
	}
	d.RecordHeartbeat(sentAt)
	f.logger.Debug("heartbeat recorded", "device_id", deviceID, "sent_at", sentAt)
	return nil
}

func (f *Fleet) RecordUploadTime(deviceID string, _ time.Time, uploadTimeNs int64) error {
	d, err := f.store.Get(deviceID)
	if err != nil {
		f.logger.Warn("upload time for unknown device", "device_id", deviceID)
		return ErrDeviceNotFound
	}
	d.RecordUploadTime(uploadTimeNs)
	f.logger.Debug("upload time recorded", "device_id", deviceID, "upload_time_ns", uploadTimeNs)
	return nil
}

type Stats struct {
	Uptime          float64
	AvgUploadTimeNs int64
}

func (f *Fleet) GetStats(deviceID string) (Stats, error) {
	d, err := f.store.Get(deviceID)
	if err != nil {
		f.logger.Warn("stats requested for unknown device", "device_id", deviceID)
		return Stats{}, ErrDeviceNotFound
	}
	s := d.ComputeStats()
	f.logger.Debug("stats computed", "device_id", deviceID, "uptime", s.Uptime)
	return Stats{
		Uptime:          s.Uptime,
		AvgUploadTimeNs: s.AvgUploadTime,
	}, nil
}

func FormatUploadTime(ns int64) string {
	if ns == 0 {
		return "0s"
	}
	return time.Duration(ns).String()
}
