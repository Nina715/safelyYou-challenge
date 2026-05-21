package service

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"fleetmetrics/internal/store"
)

func newTestFleet(t *testing.T, deviceIDs ...string) *Fleet {
	t.Helper()
	s := store.NewMemoryStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	f := New(s, logger)
	f.RegisterDevices(deviceIDs)
	return f
}

func TestRecordHeartbeat_KnownDevice(t *testing.T) {
	f := newTestFleet(t, "dev-1")
	if err := f.RecordHeartbeat("dev-1", time.Now()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRecordHeartbeat_UnknownDevice(t *testing.T) {
	f := newTestFleet(t)
	if err := f.RecordHeartbeat("unknown", time.Now()); err != ErrDeviceNotFound {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestRecordUploadTime_KnownDevice(t *testing.T) {
	f := newTestFleet(t, "dev-1")
	if err := f.RecordUploadTime("dev-1", time.Time{}, 1_000_000_000); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRecordUploadTime_UnknownDevice(t *testing.T) {
	f := newTestFleet(t)
	if err := f.RecordUploadTime("unknown", time.Time{}, 1_000_000_000); err != ErrDeviceNotFound {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestGetStats_KnownDevice(t *testing.T) {
	f := newTestFleet(t, "dev-1")
	_ = f.RecordHeartbeat("dev-1", time.Now())
	_ = f.RecordUploadTime("dev-1", time.Time{}, 5_000_000_000)

	stats, err := f.GetStats("dev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Uptime != 100 {
		t.Errorf("uptime: got %f, want 100", stats.Uptime)
	}
	if stats.AvgUploadTimeNs != 5_000_000_000 {
		t.Errorf("avg_upload_time_ns: got %d, want 5000000000", stats.AvgUploadTimeNs)
	}
}

func TestGetStats_UnknownDevice(t *testing.T) {
	f := newTestFleet(t)
	if _, err := f.GetStats("unknown"); err != ErrDeviceNotFound {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestFormatUploadTime(t *testing.T) {
	tests := []struct {
		ns   int64
		want string
	}{
		{0, "0s"},
		{1_000_000_000, "1s"},
		{90_000_000_000, "1m30s"},
		{209_226_522_788, "3m29.226522788s"},
	}
	for _, tt := range tests {
		got := FormatUploadTime(tt.ns)
		if got != tt.want {
			t.Errorf("FormatUploadTime(%d): got %q, want %q", tt.ns, got, tt.want)
		}
	}
}
