package model

import (
	"sync"
	"time"
)

type DeviceData struct {
	mu sync.RWMutex
	firstHeartbeat time.Time
	lastHeartbeat time.Time
	hasHeartbeats bool
	uploadTimeSum   int64
	uploadTimeCount int64 
	heartbeats map[time.Time]struct{}
}

type DeviceStats struct {
	Uptime        float64
	AvgUploadTime int64
}

func NewDeviceData() *DeviceData {
	return &DeviceData{
		heartbeats: make(map[time.Time]struct{}),
	}
}

func (d *DeviceData) RecordHeartbeat(sentAt time.Time) {
	minute := sentAt.Truncate(time.Minute)

	d.mu.Lock()
	defer d.mu.Unlock()

	d.heartbeats[minute] = struct{}{}

	if !d.hasHeartbeats {
		d.firstHeartbeat = minute
		d.lastHeartbeat = minute
		d.hasHeartbeats = true
		return
	}

	if minute.Before(d.firstHeartbeat) {
		d.firstHeartbeat = minute
	}
	if minute.After(d.lastHeartbeat) {
		d.lastHeartbeat = minute
	}
}

func (d *DeviceData) RecordUploadTime(uploadTimeMs int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.uploadTimeSum += uploadTimeMs
	d.uploadTimeCount++
}


func (d *DeviceData) ComputeStats() DeviceStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var stats DeviceStats
	switch {
	case !d.hasHeartbeats:
		stats.Uptime = 0
	case d.firstHeartbeat.Equal(d.lastHeartbeat):
		stats.Uptime = 100
	default:
		windowMinutes := int64(d.lastHeartbeat.Sub(d.firstHeartbeat)/time.Minute) + 1
		stats.Uptime = (float64(len(d.heartbeats)) / float64(windowMinutes)) * 100
	}

	if d.uploadTimeCount > 0 {
		stats.AvgUploadTime = d.uploadTimeSum / d.uploadTimeCount
	}

	return stats
}