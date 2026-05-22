package model

import (
	"sync"
	"time"
)

// WindowMinutes is the size of the rolling window used for uptime calculation.
// Heartbeats older than this are evicted, bounding memory to WindowMinutes bytes per device.
const WindowMinutes = 1440 // 24 hours

type DeviceData struct {
	mu         sync.RWMutex
	slots      []bool    // ring buffer: slot i tracks whether minute (origin + i) had a heartbeat
	head       int       // ring index of the oldest minute (= origin)
	count      int       // number of active (true) slots in the window
	windowSize int
	origin     time.Time // wall-clock minute that ring index head corresponds to
	hasData    bool

	// Welford's online algorithm: avoids int64 overflow and catastrophic
	// cancellation that affect the naive sum/count approach at large sample counts.
	uploadCount int64
	uploadMean  float64 // running mean in nanoseconds
}

type DeviceStats struct {
	Uptime        float64
	AvgUploadTime int64
}

func NewDeviceData() *DeviceData {
	return &DeviceData{
		slots:      make([]bool, WindowMinutes),
		windowSize: WindowMinutes,
	}
}

func (d *DeviceData) RecordHeartbeat(sentAt time.Time) {
	minute := sentAt.Truncate(time.Minute)

	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.hasData {
		d.origin = minute
		d.hasData = true
		d.slots[0] = true
		d.count = 1
		return
	}

	offset := int(minute.Sub(d.origin) / time.Minute)
	if offset < 0 {
		return // older than the window start; ignore
	}

	if offset >= d.windowSize {
		// Slide the window forward, clearing evicted slots.
		advance := offset - d.windowSize + 1
		if advance > d.windowSize {
			advance = d.windowSize
		}
		for i := 0; i < advance; i++ {
			idx := (d.head + i) % d.windowSize
			if d.slots[idx] {
				d.slots[idx] = false
				d.count--
			}
		}
		d.head = (d.head + advance) % d.windowSize
		d.origin = d.origin.Add(time.Duration(advance) * time.Minute)
		offset = d.windowSize - 1
	}

	idx := (d.head + offset) % d.windowSize
	if !d.slots[idx] {
		d.slots[idx] = true
		d.count++
	}
}

func (d *DeviceData) RecordUploadTime(ns int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.uploadCount++
	d.uploadMean += (float64(ns) - d.uploadMean) / float64(d.uploadCount)
}

func (d *DeviceData) ComputeStats() DeviceStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var stats DeviceStats

	// Uptime and avg upload time are independent: compute each only when data exists.
	if d.hasData && d.count > 0 {
		// Scan the ring buffer in logical order (head = oldest) to find the
		// first and last active minute offsets.
		first, last := -1, -1
		for i := 0; i < d.windowSize; i++ {
			if d.slots[(d.head+i)%d.windowSize] {
				if first == -1 {
					first = i
				}
				last = i
			}
		}

		if first == last {
			stats.Uptime = 100
		} else {
			uptime := (float64(d.count) / float64(last-first)) * 100
			if uptime > 100 {
				uptime = 100
			}
			stats.Uptime = uptime
		}
	}

	if d.uploadCount > 0 {
		stats.AvgUploadTime = int64(d.uploadMean)
	}

	return stats
}
