package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/config"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/internal/realtime"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SSEHandler struct {
	broker *realtime.Broker
	logger *slog.Logger
	cfg    *config.Config
}

func NewSSEHandler(broker *realtime.Broker, logger *slog.Logger, cfg *config.Config) *SSEHandler {
	return &SSEHandler{
		broker: broker,
		logger: logger,
		cfg:    cfg,
	}
}

// StreamAuction handles SSE connections for auction updates
func (h *SSEHandler) StreamAuction(w http.ResponseWriter, r *http.Request) {
	auctionIDStr := chi.URLParam(r, "id")
	auctionID, err := strconv.ParseInt(auctionIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid auction id", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create subscriber
	sub := &realtime.Subscriber{
		ID:       uuid.New().String(),
		UserID:   middleware.GetUserID(r.Context()),
		Messages: make(chan []byte, 100),
		Done:     make(chan struct{}),
	}

	// Subscribe to auction
	h.broker.Subscribe(auctionID, sub)
	defer h.broker.Unsubscribe(auctionID, sub)

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	h.logger.Info("sse_connection_opened",
		slog.String("subscriber_id", sub.ID),
		slog.Int64("auction_id", auctionID),
		slog.String("request_id", middleware.GetRequestID(r.Context())),
	)

	// Send initial connection message
	w.Write([]byte("event: connected\ndata: {\"auction_id\":" + auctionIDStr + "}\n\n"))
	flusher.Flush()

	// Keepalive ticker
	keepalive := time.NewTicker(h.cfg.SSEKeepaliveInterval)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			h.logger.Info("sse_connection_closed",
				slog.String("subscriber_id", sub.ID),
				slog.Int64("auction_id", auctionID),
			)
			return

		case msg := <-sub.Messages:
			_, err := w.Write(msg)
			if err != nil {
				return
			}
			flusher.Flush()

		case <-keepalive.C:
			_, err := w.Write([]byte(": keepalive\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

