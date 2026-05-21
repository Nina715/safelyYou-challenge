package service

import (
	"errors"
	"fleetmetrics/internal/store"
	"time"
)

var ErrDeviceNotFound = errors.New("device not found")

type Fleet struct {
	store store.Store
}

func New(s store.Store) *Fleet {
	return &Fleet{store: s}
}

func (f *Fleet) RegisterDevices(deviceIDs []string) {
	for _, id := range deviceIDs {
		f.store.Register(id)
	}
}

func (f *Fleet) RecordHeartbeat(deviceID string, sentAt time.Time) error {
	d, err := f.store.Get(deviceID)
	if err != nil {
		return ErrDeviceNotFound
	}
	d.RecordHeartbeat(sentAt)
	return nil
}

func (f *Fleet) RecordUploadTime(deviceID string, _ time.Time, uploadTimeNs int64) error {
	d, err := f.store.Get(deviceID)
	if err != nil {
		return ErrDeviceNotFound
	}
	d.RecordUploadTime(uploadTimeNs)
	return nil
}

type Stats struct {
	Uptime          float64
	AvgUploadTimeNs int64
}

func (f *Fleet) GetStats(deviceID string) (Stats, error) {
	d, err := f.store.Get(deviceID)
	if err != nil {
		return Stats{}, ErrDeviceNotFound
	}
	s := d.ComputeStats()
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
