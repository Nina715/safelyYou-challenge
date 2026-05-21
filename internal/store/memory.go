package store

import (
	"fleetmetrics/internal/model"
	"sync"
)

type MemoryStore struct {
	mu      sync.RWMutex
	devices map[string]*model.DeviceData
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		devices: make(map[string]*model.DeviceData),
	}
}

func (s *MemoryStore) Register(deviceID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.devices[deviceID]; ok {
		return
	}
	s.devices[deviceID] = model.NewDeviceData()
}

func (s *MemoryStore) Exists(deviceID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.devices[deviceID]
	return ok
}

func (s *MemoryStore) Get(deviceID string) (*model.DeviceData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.devices[deviceID]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	return d, nil
}

