package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Server
	Port            int           `env:"PORT" envDefault:"8080"`
	Environment     string        `env:"ENVIRONMENT" envDefault:"development"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// Database
	DatabaseURL     string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/vehicle_auc?sslmode=disable"`
	DBMaxConns      int    `env:"DB_MAX_CONNS" envDefault:"25"`
	DBMinConns      int    `env:"DB_MIN_CONNS" envDefault:"5"`
	DBMaxConnLife   time.Duration `env:"DB_MAX_CONN_LIFE" envDefault:"1h"`

	// Redis (for future use)
	RedisURL string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`

	// Auth (Clerk)
	ClerkSecretKey  string `env:"CLERK_SECRET_KEY"`
	ClerkPublishableKey string `env:"CLERK_PUBLISHABLE_KEY"`
	ClerkJWKSURL    string `env:"CLERK_JWKS_URL"`

	// AWS S3
	AWSS3Bucket     string `env:"AWS_S3_BUCKET" envDefault:"vehicle-auc-images"`
	AWSS3Region     string `env:"AWS_S3_REGION" envDefault:"us-east-1"`
	AWSAccessKeyID  string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretKey    string `env:"AWS_SECRET_ACCESS_KEY"`

	// Observability
	SentryDSN       string `env:"SENTRY_DSN"`
	OTLPEndpoint    string `env:"OTLP_ENDPOINT" envDefault:"localhost:4317"`
	MetricsPath     string `env:"METRICS_PATH" envDefault:"/metrics"`

	// Bid Engine
	BidQueueSize    int           `env:"BID_QUEUE_SIZE" envDefault:"10000"`
	BidWorkerCount  int           `env:"BID_WORKER_COUNT" envDefault:"100"`
	BidMaxRetries   int           `env:"BID_MAX_RETRIES" envDefault:"3"`
	BidRetryBackoff time.Duration `env:"BID_RETRY_BACKOFF" envDefault:"10ms"`

	// SSE
	SSEKeepaliveInterval time.Duration `env:"SSE_KEEPALIVE_INTERVAL" envDefault:"30s"`

	// CORS
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:5173,http://localhost:3000"`

	// Feature flags
	DebugEndpointsEnabled bool `env:"DEBUG_ENDPOINTS_ENABLED" envDefault:"true"`
	SyncBidMode           bool `env:"SYNC_BID_MODE" envDefault:"false"` // For testing
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) Validate() error {
	if c.IsProduction() {
		if c.ClerkSecretKey == "" {
			return fmt.Errorf("CLERK_SECRET_KEY is required in production")
		}
		if c.SentryDSN == "" {
			return fmt.Errorf("SENTRY_DSN is required in production")
		}
	}
	return nil
}

