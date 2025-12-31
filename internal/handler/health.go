package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db        *pgxpool.Pool
	startTime time.Time
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	status := "healthy"

	// Check database
	ctx := r.Context()
	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	resp := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Readiness check - can the service accept traffic?
	ctx := r.Context()
	if err := h.db.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	// Liveness check - is the service alive?
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("alive"))
}

