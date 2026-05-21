package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"fleetmetrics/internal/service"
	"fleetmetrics/internal/store"
)

func newTestServer(t *testing.T, deviceIDs ...string) *Server {
	t.Helper()
	memStore := store.NewMemoryStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fleet := service.New(memStore, logger)
	fleet.RegisterDevices(deviceIDs)
	return NewServer(fleet, logger)
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewReader(b)
}

func TestPostHeartbeat_Valid(t *testing.T) {
	s := newTestServer(t, "dev-1")
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"sent_at": time.Now(),
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdHeartbeat(w, r, "dev-1")
	if w.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestPostHeartbeat_MissingBody(t *testing.T) {
	s := newTestServer(t, "dev-1")
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(nil))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdHeartbeat(w, r, "dev-1")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostHeartbeat_ZeroSentAt(t *testing.T) {
	s := newTestServer(t, "dev-1")
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"sent_at": time.Time{},
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdHeartbeat(w, r, "dev-1")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostHeartbeat_UnknownDevice(t *testing.T) {
	s := newTestServer(t)
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"sent_at": time.Now(),
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdHeartbeat(w, r, "unknown")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPostStats_Valid(t *testing.T) {
	s := newTestServer(t, "dev-1")
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"upload_time": 1_000_000_000,
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdStats(w, r, "dev-1")
	if w.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestPostStats_NegativeUploadTime(t *testing.T) {
	s := newTestServer(t, "dev-1")
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"upload_time": -1,
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdStats(w, r, "dev-1")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostStats_UnknownDevice(t *testing.T) {
	s := newTestServer(t)
	r := httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{
		"upload_time": 1_000_000_000,
	}))
	w := httptest.NewRecorder()
	s.PostDevicesDeviceIdStats(w, r, "unknown")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetStats_Valid(t *testing.T) {
	s := newTestServer(t, "dev-1")
	now := time.Now()

	s.PostDevicesDeviceIdHeartbeat(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{"sent_at": now})),
		"dev-1",
	)
	s.PostDevicesDeviceIdStats(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/", jsonBody(t, map[string]any{"upload_time": 60_000_000_000})),
		"dev-1",
	)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.GetDevicesDeviceIdStats(w, r, "dev-1")

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var resp StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Uptime != 100 {
		t.Errorf("uptime: got %f, want 100", resp.Uptime)
	}
	if resp.AvgUploadTime != "1m0s" {
		t.Errorf("avg_upload_time: got %q, want %q", resp.AvgUploadTime, "1m0s")
	}
}

func TestGetStats_UnknownDevice(t *testing.T) {
	s := newTestServer(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.GetDevicesDeviceIdStats(w, r, "unknown")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}
