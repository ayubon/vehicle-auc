package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayubfarah/vehicle-auc/internal/bidengine"
	"github.com/ayubfarah/vehicle-auc/internal/config"
	"github.com/ayubfarah/vehicle-auc/internal/handler"
	"github.com/ayubfarah/vehicle-auc/internal/middleware"
	"github.com/ayubfarah/vehicle-auc/internal/realtime"
	"github.com/ayubfarah/vehicle-auc/internal/tracing"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize Sentry
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.Environment,
			TracesSampleRate: 0.1,
		}); err != nil {
			logger.Error("failed to init sentry", slog.String("error", err.Error()))
		} else {
			defer sentry.Flush(2 * time.Second)
		}
	}

	// Initialize tracing
	ctx := context.Background()
	tracingShutdown, err := tracing.Init(ctx, "vehicle-auc", cfg.OTLPEndpoint, cfg.Environment)
	if err != nil {
		logger.Warn("failed to init tracing", slog.String("error", err.Error()))
	} else {
		defer tracingShutdown(ctx)
	}

	// Connect to database
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	dbConfig.MaxConns = int32(cfg.DBMaxConns)
	dbConfig.MinConns = int32(cfg.DBMinConns)
	dbConfig.MaxConnLifetime = cfg.DBMaxConnLife

	db, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		logger.Error("failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("database_connected")

	// Initialize SSE broker
	broker := realtime.NewBroker(logger)
	broker.Start()
	defer broker.Stop()

	// Initialize bid engine
	engine := bidengine.NewEngine(
		db, logger, broker,
		bidengine.WithQueueSize(cfg.BidQueueSize),
		bidengine.WithMaxRetries(cfg.BidMaxRetries),
		bidengine.WithRetryBackoff(cfg.BidRetryBackoff),
		bidengine.WithSyncMode(cfg.SyncBidMode),
	)
	engine.Start()
	defer engine.Stop()

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(db)
	vehicleHandler := handler.NewVehicleHandler(db, logger)
	auctionHandler := handler.NewAuctionHandler(db, logger)
	bidHandler := handler.NewBidHandler(engine, logger)
	sseHandler := handler.NewSSEHandler(broker, logger, cfg)
	debugHandler := handler.NewDebugHandler(engine, broker)
	authHandler := handler.NewAuthHandler(db, logger)
	imageHandler := handler.NewImageHandler(db, logger, cfg, nil) // S3 client nil for now
	watchlistHandler := handler.NewWatchlistHandler(db, logger)
	notificationHandler := handler.NewNotificationHandler(db, logger)
	vinHandler := handler.NewVINHandler(logger, nil) // VIN decoder nil for now

	// Initialize auth middleware
	clerkAuth := middleware.NewClerkAuth(logger, cfg.ClerkJWKSURL, cfg.ClerkSecretKey)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Tracing)
	r.Use(middleware.Logging(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health endpoints (no auth)
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)
	r.Get("/live", healthHandler.Live)

	// Metrics endpoint
	r.Handle(cfg.MetricsPath, promhttp.Handler())

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public endpoints
		r.Get("/vehicles", vehicleHandler.ListVehicles)
		r.Get("/vehicles/{id}", vehicleHandler.GetVehicle)
		r.Get("/vehicles/{id}/images", vehicleHandler.GetVehicleImages)
		r.Get("/auctions", auctionHandler.ListAuctions)
		r.Get("/auctions/{id}", auctionHandler.GetAuction)
		r.Get("/auctions/{id}/bids", auctionHandler.GetBidHistory)

		// SSE endpoint (optional auth)
		r.With(clerkAuth.OptionalAuth).Get("/auctions/{id}/stream", sseHandler.StreamAuction)

		// Auth - Clerk sync (no auth required - creates user)
		r.Post("/auth/clerk-sync", authHandler.ClerkSync)

		// Protected endpoints
		r.Group(func(r chi.Router) {
			r.Use(clerkAuth.Middleware)

			// Auth / User
			r.Get("/auth/me", authHandler.Me)
			r.Put("/auth/me", authHandler.UpdateProfile)

			// Vehicles
			r.Post("/vehicles", vehicleHandler.CreateVehicle)
			r.Put("/vehicles/{id}", vehicleHandler.UpdateVehicle)
			r.Delete("/vehicles/{id}", vehicleHandler.DeleteVehicle)
			r.Post("/vehicles/{id}/submit", vehicleHandler.SubmitVehicle)

			// Vehicle Images
			r.Post("/vehicles/{id}/upload-url", imageHandler.GetUploadURL)
			r.Post("/vehicles/{id}/images", imageHandler.AddImage)
			r.Delete("/vehicles/{id}/images/{imageId}", imageHandler.DeleteImage)

			// VIN Decode
			r.Post("/decode-vin", vinHandler.DecodeVIN)

			// Auctions
			r.Post("/auctions", auctionHandler.CreateAuction)

			// Bids (support both /bid and /bids for backwards compatibility)
			r.Post("/auctions/{id}/bid", bidHandler.PlaceBid)
			r.Post("/auctions/{id}/bids", bidHandler.PlaceBid)
			r.Get("/bids/{ticketId}/status", bidHandler.GetBidStatus)

			// Watchlist
			r.Get("/watchlist", watchlistHandler.GetWatchlist)
			r.Post("/auctions/{id}/watch", watchlistHandler.AddToWatchlist)
			r.Delete("/auctions/{id}/watch", watchlistHandler.RemoveFromWatchlist)
			r.Get("/auctions/{id}/watching", watchlistHandler.IsWatching)

			// Notifications
			r.Get("/notifications", notificationHandler.GetNotifications)
			r.Get("/notifications/unread-count", notificationHandler.GetUnreadCount)
			r.Post("/notifications/{id}/read", notificationHandler.MarkRead)
			r.Post("/notifications/read-all", notificationHandler.MarkAllRead)
			r.Delete("/notifications/{id}", notificationHandler.DeleteNotification)
		})
	})

	// Debug endpoints (development only)
	if cfg.DebugEndpointsEnabled {
		r.Route("/debug", func(r chi.Router) {
			r.Get("/bidengine", debugHandler.BidEngineStats)
			r.Get("/sse", debugHandler.SSEStats)
			r.Get("/stats", debugHandler.AllStats)
		})
	}

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("server_starting",
			slog.Int("port", cfg.Port),
			slog.String("environment", cfg.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server_error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server_shutting_down")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_error", slog.String("error", err.Error()))
	}

	logger.Info("server_stopped")
}

