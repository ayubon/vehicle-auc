package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metric names match Flask custom_metrics.py for consistency

var (
	// ==========================================================================
	// HTTP Metrics
	// ==========================================================================
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// ==========================================================================
	// Database Metrics
	// ==========================================================================
	DBQueryTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_total",
			Help: "Total number of database queries",
		},
		[]string{"query_type", "table"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"query_type", "table"},
	)

	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// ==========================================================================
	// Auction Metrics
	// ==========================================================================
	AuctionBidsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auction_bids_total",
			Help: "Total number of bids placed",
		},
		[]string{"status"}, // accepted, rejected, error
	)

	AuctionBidAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auction_bid_amount",
			Help:    "Distribution of bid amounts",
			Buckets: []float64{100, 500, 1000, 2500, 5000, 10000, 25000, 50000, 100000},
		},
		[]string{"auction_id"},
	)

	AuctionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auctions_active_total",
			Help: "Number of currently active auctions",
		},
	)

	AuctionExtensions = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auction_extensions_total",
			Help: "Total number of auction extensions (anti-snipe)",
		},
	)

	// ==========================================================================
	// Bid Engine Metrics
	// ==========================================================================
	BidEngineQueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bid_engine_queue_depth",
			Help: "Current depth of the bid processing queue",
		},
	)

	BidEngineWorkersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "bid_engine_workers_active",
			Help: "Number of active bid engine workers",
		},
	)

	BidProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "bid_processing_duration_seconds",
			Help:    "Time to process a bid (from queue to completion)",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
	)

	BidOCCRetries = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "bid_occ_retries",
			Help:    "Number of OCC retries per bid",
			Buckets: []float64{0, 1, 2, 3, 4, 5},
		},
	)

	BidOCCConflictsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bid_occ_conflicts_total",
			Help: "Total number of OCC version conflicts",
		},
	)

	// ==========================================================================
	// SSE Metrics
	// ==========================================================================
	SSEConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sse_connections_active",
			Help: "Number of active SSE connections",
		},
	)

	SSEMessagesSent = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sse_messages_sent_total",
			Help: "Total SSE messages sent",
		},
		[]string{"event_type"},
	)

	SSESubscribersPerAuction = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sse_subscribers_per_auction",
			Help:    "Number of SSE subscribers per auction when broadcasting",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
	)

	// ==========================================================================
	// User Metrics
	// ==========================================================================
	UsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_total",
			Help: "Total number of registered users",
		},
	)

	UsersVerified = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "users_verified_total",
			Help: "Total number of verified users (can bid)",
		},
	)

	// ==========================================================================
	// Vehicle Metrics
	// ==========================================================================
	VehiclesTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vehicles_total",
			Help: "Total number of vehicles by status",
		},
		[]string{"status"},
	)

	// ==========================================================================
	// Order Metrics
	// ==========================================================================
	OrdersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	OrderValue = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "order_value",
			Help:    "Distribution of order values",
			Buckets: []float64{1000, 5000, 10000, 25000, 50000, 100000, 250000},
		},
	)

	// ==========================================================================
	// External API Metrics
	// ==========================================================================
	ExternalAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_api_calls_total",
			Help: "Total external API calls",
		},
		[]string{"service", "endpoint", "status"},
	)

	ExternalAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_api_latency_seconds",
			Help:    "External API call latency",
			Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "endpoint"},
	)
)

