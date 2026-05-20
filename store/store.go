package store

import (
	"errors"
	"fleetmetrics/model"
)

var ErrDeviceNotFound = errors.New("device not found")
type Store interface {

	Register(deviceID string)

	Exists(deviceID string) bool

	Get(deviceID string) (*model.DeviceData, error)
}