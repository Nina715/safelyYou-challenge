package model

import (
	"testing"
	"time"
)

var base = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestComputeStats_NoHeartbeats(t *testing.T) {
	d := NewDeviceData()
	stats := d.ComputeStats()
	if stats.Uptime != 0 {
		t.Errorf("uptime: got %.2f, want 0", stats.Uptime)
	}
	if stats.AvgUploadTime != 0 {
		t.Errorf("avg_upload_time: got %d, want 0", stats.AvgUploadTime)
	}
}

func TestComputeStats_SingleHeartbeat(t *testing.T) {
	d := NewDeviceData()
	d.RecordHeartbeat(base)
	stats := d.ComputeStats()
	if stats.Uptime != 100 {
		t.Errorf("uptime: got %.2f, want 100", stats.Uptime)
	}
}

func TestComputeStats_UptimePercentage(t *testing.T) {
	d := NewDeviceData()
	// Active at minutes 0, 2, 4 out of a 4-minute window → 3/4 = 75%
	for _, m := range []int{0, 2, 4} {
		d.RecordHeartbeat(base.Add(time.Duration(m) * time.Minute))
	}
	stats := d.ComputeStats()
	want := 75.0
	if stats.Uptime != want {
		t.Errorf("uptime: got %.2f, want %.2f", stats.Uptime, want)
	}
}

func TestComputeStats_DeduplicatesHeartbeats(t *testing.T) {
	d := NewDeviceData()
	// Two events in the same minute count as one heartbeat
	d.RecordHeartbeat(base)
	d.RecordHeartbeat(base.Add(30 * time.Second))
	d.RecordHeartbeat(base.Add(2 * time.Minute))
	stats := d.ComputeStats()
	// 2 unique minutes, window = 2 → 2/2 * 100 = 100%
	want := 100.0
	if stats.Uptime != want {
		t.Errorf("uptime: got %.2f, want %.2f", stats.Uptime, want)
	}
}

func TestRecordUploadTime_Average(t *testing.T) {
	d := NewDeviceData()
	d.RecordUploadTime(100)
	d.RecordUploadTime(200)
	d.RecordUploadTime(300)
	stats := d.ComputeStats()
	want := int64(200)
	if stats.AvgUploadTime != want {
		t.Errorf("avg_upload_time: got %d, want %d", stats.AvgUploadTime, want)
	}
}

func TestRecordUploadTime_NoUploads(t *testing.T) {
	d := NewDeviceData()
	stats := d.ComputeStats()
	if stats.AvgUploadTime != 0 {
		t.Errorf("avg_upload_time: got %d, want 0", stats.AvgUploadTime)
	}
}
