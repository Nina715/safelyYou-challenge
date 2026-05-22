package store

import (
	"hash/fnv"
	"sync"

	"fleetmetrics/internal/model"
)

// numShards controls how many independent locks protect the device map.
// Distributing devices across shards reduces write contention by ~numShards×
// compared to a single global mutex.
const numShards = 16

type shard struct {
	sync.RWMutex
	devices map[string]*model.DeviceData
}

type MemoryStore struct {
	shards [numShards]shard
}

func NewMemoryStore() *MemoryStore {
	ms := &MemoryStore{}
	for i := range ms.shards {
		ms.shards[i].devices = make(map[string]*model.DeviceData)
	}
	return ms
}

func (s *MemoryStore) shardFor(deviceID string) *shard {
	h := fnv.New32a()
	h.Write([]byte(deviceID))
	return &s.shards[h.Sum32()%numShards]
}

func (s *MemoryStore) Register(deviceID string) {
	sh := s.shardFor(deviceID)
	sh.Lock()
	defer sh.Unlock()
	if _, ok := sh.devices[deviceID]; !ok {
		sh.devices[deviceID] = model.NewDeviceData()
	}
}

func (s *MemoryStore) Exists(deviceID string) bool {
	sh := s.shardFor(deviceID)
	sh.RLock()
	defer sh.RUnlock()
	_, ok := sh.devices[deviceID]
	return ok
}

func (s *MemoryStore) Get(deviceID string) (*model.DeviceData, error) {
	sh := s.shardFor(deviceID)
	sh.RLock()
	defer sh.RUnlock()
	d, ok := sh.devices[deviceID]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	return d, nil
}
