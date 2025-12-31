package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ayubfarah/vehicle-auc/internal/bidengine"
	"github.com/ayubfarah/vehicle-auc/internal/realtime"
)

type DebugHandler struct {
	engine *bidengine.Engine
	broker *realtime.Broker
}

func NewDebugHandler(engine *bidengine.Engine, broker *realtime.Broker) *DebugHandler {
	return &DebugHandler{
		engine: engine,
		broker: broker,
	}
}

// BidEngineStats returns current bid engine statistics
func (h *DebugHandler) BidEngineStats(w http.ResponseWriter, r *http.Request) {
	stats := h.engine.Stats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "running",
		"queue_depth":     stats.QueueDepth,
		"active_workers":  stats.ActiveWorkers,
		"total_processed": stats.TotalProcessed,
		"total_retries":   stats.TotalRetries,
		"workers":         stats.Workers,
	})
}

// SSEStats returns current SSE broker statistics
func (h *DebugHandler) SSEStats(w http.ResponseWriter, r *http.Request) {
	stats := h.broker.Stats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_connections": stats.TotalConnections,
		"auctions":          stats.Auctions,
	})
}

// AllStats returns combined debug information
func (h *DebugHandler) AllStats(w http.ResponseWriter, r *http.Request) {
	engineStats := h.engine.Stats()
	sseStats := h.broker.Stats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bid_engine": map[string]interface{}{
			"status":          "running",
			"queue_depth":     engineStats.QueueDepth,
			"active_workers":  engineStats.ActiveWorkers,
			"total_processed": engineStats.TotalProcessed,
			"total_retries":   engineStats.TotalRetries,
		},
		"sse": map[string]interface{}{
			"total_connections": sseStats.TotalConnections,
			"auction_count":     len(sseStats.Auctions),
		},
	})
}

