package api

import (
	"encoding/json"
	"errors"
	"fleetmetrics/internal/service"
	"log/slog"
	"net/http"
)

type Server struct {
	fleet  *service.Fleet
	logger *slog.Logger
}

func NewServer(fleet *service.Fleet, logger *slog.Logger) *Server {
	return &Server{fleet: fleet, logger: logger}
}

var _ ServerInterface = (*Server)(nil)

type StatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}
func (s *Server) PostDevicesDeviceIdHeartbeat(w http.ResponseWriter, r *http.Request, deviceID DeviceIDPathParam) {
	var req PostDevicesDeviceIdHeartbeatJSONRequestBody
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.SentAt.IsZero() {
		s.writeError(w, http.StatusBadRequest, "sent_at is required")
		return
	}

	if err := s.fleet.RecordHeartbeat(deviceID, req.SentAt); err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			s.writeNotFound(w, "device not found")
			return
		}
		s.logger.Error("record heartbeat failed", "device_id", deviceID, "err", err)
		s.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) PostDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceID DeviceIDPathParam) {
	var req PostDevicesDeviceIdStatsJSONRequestBody
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.SentAt.IsZero() {
		s.writeError(w, http.StatusBadRequest, "sent_at is required")
		return
	}
	if req.UploadTime < 0 {
		s.writeError(w, http.StatusBadRequest, "upload_time must be non-negative")
		return
	}


	if err := s.fleet.RecordUploadTime(deviceID, req.SentAt, int64(req.UploadTime)); err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			s.writeNotFound(w, "device not found")
			return
		}
		s.logger.Error("record upload time failed", "device_id", deviceID, "err", err)
		s.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) GetDevicesDeviceIdStats(w http.ResponseWriter, r *http.Request, deviceID DeviceIDPathParam) {
	stats, err := s.fleet.GetStats(deviceID)
	if err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			s.writeNotFound(w, "device not found")
			return
		}
		s.logger.Error("get stats failed", "device_id", deviceID, "err", err)
		s.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := StatsResponse{
		Uptime:        stats.Uptime,
		AvgUploadTime: service.FormatUploadTime(stats.AvgUploadTimeNs),
	}
	writeJSON(w, http.StatusOK, resp)
}


func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("unexpected data after JSON value")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, Error{Msg: msg})
}

func (s *Server) writeNotFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, NotFound{Msg: msg})
}
