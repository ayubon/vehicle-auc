# Vehicle Auction Platform

A real-time vehicle auction platform built with Go (Chi router) and React, featuring optimistic concurrency control for high-throughput bid processing and Server-Sent Events for live updates.

---

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
  - [Functional Requirements](#functional-requirements)
  - [Non-Functional Requirements](#non-functional-requirements)
- [System Architecture](#system-architecture)
- [Bid Processing Deep Dive](#bid-processing-deep-dive)
- [Real-Time Updates (SSE)](#real-time-updates-sse)
- [Database Schema](#database-schema)
- [API Reference](#api-reference)
- [Frontend User Journeys](#frontend-user-journeys)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Testing](#testing)
- [Observability](#observability)
- [Project Structure](#project-structure)

---

## Overview

This platform enables dealers and private sellers to list vehicles for auction, while buyers can browse, watch, and bid on vehicles in real-time. The system is designed to handle high-concurrency bidding scenarios with strong consistency guarantees.

### Key Features

- **Real-time bidding** with sub-second updates via SSE
- **Optimistic Concurrency Control (OCC)** for lock-free bid processing
- **Anti-snipe protection** with automatic auction extensions
- **VIN decoding** for automatic vehicle details
- **Clerk SSO** for authentication
- **S3 image uploads** with presigned URLs
- **Full observability** with Prometheus, Jaeger, and Sentry

---

## Requirements

### Functional Requirements

#### Core Features

| Feature | Description |
|---------|-------------|
| **Vehicle Listing** | Sellers can create vehicle listings with VIN decode, photos, and details |
| **Auction Creation** | Sellers can schedule auctions with start/end times and starting price |
| **Real-time Bidding** | Buyers can place bids; all users see updates in real-time |
| **Bid Validation** | Bids must exceed current highest bid |
| **Anti-Snipe** | Auctions extend by 2 minutes if bid placed in final 2 minutes |
| **Watchlist** | Users can watch auctions and receive notifications |
| **Notifications** | System notifies users of outbids, auction endings, wins |
| **User Profiles** | Clerk SSO with synced user profiles |

#### Below the Line (Future)

- Search with filters (make, model, year, price range)
- Auto-bidding / proxy bidding
- Buy-it-now pricing
- Seller verification / KYC
- Payment processing
- Order fulfillment / shipping coordination

### Non-Functional Requirements

| Requirement | Target | Implementation |
|-------------|--------|----------------|
| **Bid Consistency** | Strong | OCC with version column, retry on conflict |
| **Real-time Latency** | < 500ms | SSE push, in-memory broker |
| **Bid Throughput** | 10K bids/sec | Goroutine workers, channel queue |
| **Availability** | 99.9% | Health checks, graceful shutdown |
| **Durability** | Zero bid loss | PostgreSQL with WAL, bid history table |
| **Observability** | Full stack | Prometheus, Jaeger, Sentry, structured logs |

---

## System Architecture

```mermaid
flowchart TB
    subgraph clients [Clients]
        ReactApp[React SPA<br/>Port 3000]
    end

    subgraph goserver [Go Server - Single Binary :8080]
        subgraph api [API Layer]
            Router[Chi Router]
            AuthMW[Auth Middleware<br/>Clerk JWT]
            LogMW[Logging Middleware<br/>slog JSON]
            TraceMW[OpenTelemetry<br/>Middleware]
        end

        subgraph handlers [HTTP Handlers]
            VehicleH[Vehicle Handler]
            AuctionH[Auction Handler]
            BidH[Bid Handler]
            AuthH[Auth Handler]
            SSEH[SSE Handler]
            WatchH[Watchlist Handler]
            NotifH[Notification Handler]
            DebugH[Debug Handler]
        end

        subgraph bidengine [Bid Engine]
            BidQueue[Bid Queue<br/>Buffered Channel<br/>10K capacity]
            Dispatcher[Dispatcher<br/>Goroutine]
            Workers[Auction Workers<br/>1 per active auction]
            ResultMap[Result Channels<br/>Per ticket ID]
        end

        subgraph realtime [Real-time SSE]
            SSEBroker[SSE Broker<br/>Goroutine]
            Subscribers[Subscribers Map<br/>Per auction ID]
            Broadcast[Broadcast Channel]
        end

        subgraph observability [Observability]
            Metrics[Prometheus<br/>Metrics]
            Tracer[OpenTelemetry<br/>Tracer]
            Logger[slog JSON<br/>Logger]
        end
    end

    subgraph external [External Services]
        PG[(PostgreSQL<br/>Port 5432)]
        Redis[(Redis<br/>Port 6379)]
        S3[AWS S3<br/>Images]
        Clerk[Clerk<br/>Auth]
        Sentry[Sentry<br/>Errors]
        Jaeger[Jaeger<br/>Traces]
    end

    ReactApp -->|HTTP/SSE| Router
    Router --> AuthMW --> LogMW --> TraceMW --> handlers
    
    VehicleH --> PG
    AuctionH --> PG
    BidH -->|Submit| BidQueue
    SSEH --> SSEBroker
    WatchH --> PG
    NotifH --> PG
    
    BidQueue --> Dispatcher --> Workers
    Workers -->|OCC Write| PG
    Workers -->|Broadcast| SSEBroker
    Workers -->|Result| ResultMap
    SSEBroker -->|SSE Events| ReactApp
    
    VehicleH -->|Presigned URL| S3
    AuthMW -->|JWKS| Clerk
    Logger --> Sentry
    Tracer --> Jaeger
```

### Component Breakdown

| Component | Purpose | Technology |
|-----------|---------|------------|
| **Chi Router** | HTTP routing, middleware chain | `go-chi/chi/v5` |
| **Auth Middleware** | JWT validation via Clerk JWKS | `golang-jwt/jwt/v5` |
| **Bid Engine** | Async bid processing with OCC | Goroutines + Channels |
| **SSE Broker** | Real-time event fan-out | Pure Go channels |
| **PostgreSQL** | Primary data store | `jackc/pgx/v5` |
| **Prometheus** | Metrics collection | `prometheus/client_golang` |
| **OpenTelemetry** | Distributed tracing | `go.opentelemetry.io/otel` |

---

## Bid Processing Deep Dive

The bid system uses **Optimistic Concurrency Control (OCC)** to handle concurrent bids without pessimistic locking, enabling high throughput while maintaining consistency.

### Why OCC Over Pessimistic Locking?

| Approach | Pros | Cons |
|----------|------|------|
| **Pessimistic (SELECT FOR UPDATE)** | Simple, guaranteed consistency | Blocks concurrent requests, deadlock risk |
| **Optimistic (Version column)** | Non-blocking, high throughput | Retry logic needed, wasted work on conflict |

For auctions, OCC is preferred because:
- Most bids don't conflict (different auctions)
- Conflicts are cheap to retry (< 10ms)
- Throughput matters more than latency

### Bid Flow Diagram

```mermaid
sequenceDiagram
    participant Client
    participant BidHandler
    participant BidQueue
    participant Dispatcher
    participant Worker
    participant PostgreSQL
    participant SSEBroker
    participant Subscribers

    Client->>BidHandler: POST /auctions/123/bids<br/>{amount: 150.00}
    BidHandler->>BidHandler: Generate ticket_id (UUID)
    BidHandler->>BidQueue: Submit bid to channel
    BidHandler-->>Client: 202 Accepted<br/>{ticket_id, status: "queued"}
    
    Note over Client: Client can poll<br/>GET /bids/{ticket}/status<br/>or wait for SSE

    BidQueue->>Dispatcher: Receive bid
    Dispatcher->>Dispatcher: Find/create worker for auction 123
    Dispatcher->>Worker: Route bid to auction worker

    loop OCC Retry (max 3)
        Worker->>PostgreSQL: SELECT current_bid, version<br/>FROM auctions WHERE id=123
        PostgreSQL-->>Worker: current_bid=100, version=5
        
        Worker->>Worker: Validate: 150 > 100 ✓
        
        Worker->>PostgreSQL: UPDATE auctions<br/>SET current_bid=150, version=6<br/>WHERE id=123 AND version=5
        
        alt Version Match (Success)
            PostgreSQL-->>Worker: 1 row updated
            Worker->>PostgreSQL: INSERT INTO bids<br/>(auction_id, amount, status='accepted')
        else Version Mismatch (Conflict)
            PostgreSQL-->>Worker: 0 rows updated
            Worker->>Worker: Backoff 10ms, retry
        end
    end

    Worker->>SSEBroker: Broadcast bid event
    SSEBroker->>Subscribers: event: bid_accepted<br/>data: {auction_id, amount, user}
    Subscribers-->>Client: SSE event received
```

### OCC Implementation Details

```sql
-- Auction table with OCC version column
CREATE TABLE auctions (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL,
    status auction_status NOT NULL DEFAULT 'scheduled',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    
    -- Denormalized for fast reads (no JOIN needed)
    current_bid NUMERIC(10, 2) NOT NULL DEFAULT 0,
    current_bid_user_id BIGINT REFERENCES users(id),
    bid_count INT NOT NULL DEFAULT 0,
    
    -- OCC version - incremented on every bid
    version INT NOT NULL DEFAULT 0,
    
    -- Anti-snipe
    extension_count SMALLINT NOT NULL DEFAULT 0,
    max_extensions SMALLINT NOT NULL DEFAULT 10,
    snipe_threshold_minutes SMALLINT NOT NULL DEFAULT 2
);

-- Bid history (never lose a bid)
CREATE TABLE bids (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL REFERENCES auctions(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    amount NUMERIC(10, 2) NOT NULL,
    status bid_status NOT NULL,  -- 'accepted', 'rejected', 'outbid'
    previous_high_bid NUMERIC(10, 2),  -- Audit trail
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Bid Processing States

```mermaid
stateDiagram-v2
    [*] --> Queued: POST /bids
    Queued --> Processing: Worker picks up
    Processing --> Validating: Read auction state
    
    Validating --> Accepted: bid > current
    Validating --> Rejected: bid <= current
    Validating --> Rejected: auction ended
    Validating --> Rejected: user not verified
    
    Accepted --> Writing: OCC UPDATE
    Writing --> Accepted: version match
    Writing --> Retrying: version conflict
    Retrying --> Validating: retry < 3
    Retrying --> Failed: retry >= 3
    
    Accepted --> Broadcasting: INSERT bid record
    Broadcasting --> [*]: SSE event sent
    
    Rejected --> [*]: Return error
    Failed --> [*]: Return error
```

---

## Real-Time Updates (SSE)

Server-Sent Events provide real-time bid updates without WebSocket complexity.

### SSE Architecture

```mermaid
flowchart LR
    subgraph clients [Browser Clients]
        C1[Client 1<br/>Watching Auction 123]
        C2[Client 2<br/>Watching Auction 123]
        C3[Client 3<br/>Watching Auction 456]
    end

    subgraph broker [SSE Broker]
        BC[Broadcast Channel]
        
        subgraph subs [Subscribers Map]
            A123[Auction 123<br/>Subscribers]
            A456[Auction 456<br/>Subscribers]
        end
    end

    subgraph workers [Bid Workers]
        W1[Worker 123]
        W2[Worker 456]
    end

    W1 -->|BidEvent| BC
    W2 -->|BidEvent| BC
    
    BC --> A123
    BC --> A456
    
    A123 -->|SSE| C1
    A123 -->|SSE| C2
    A456 -->|SSE| C3
```

### SSE Event Types

| Event | Payload | When |
|-------|---------|------|
| `bid_accepted` | `{auction_id, amount, user_id, bid_count}` | New high bid |
| `bid_rejected` | `{auction_id, reason}` | Bid too low |
| `auction_extended` | `{auction_id, new_ends_at}` | Anti-snipe triggered |
| `auction_ended` | `{auction_id, winner_id, final_bid}` | Auction closed |
| `keepalive` | `{}` | Every 30s to prevent timeout |

### Client Connection

```javascript
// Frontend SSE connection
const eventSource = new EventSource('/api/auctions/123/stream');

eventSource.addEventListener('bid_accepted', (e) => {
  const data = JSON.parse(e.data);
  updateCurrentBid(data.amount);
  updateBidCount(data.bid_count);
});

eventSource.addEventListener('auction_extended', (e) => {
  const data = JSON.parse(e.data);
  updateEndTime(data.new_ends_at);
  showSnipeAlert();
});
```

---

## Database Schema

### Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ VEHICLES : owns
    USERS ||--o{ BIDS : places
    USERS ||--o{ WATCHLIST : watches
    USERS ||--o{ NOTIFICATIONS : receives
    
    VEHICLES ||--o| AUCTIONS : listed_in
    VEHICLES ||--o{ VEHICLE_IMAGES : has
    
    AUCTIONS ||--o{ BIDS : receives
    AUCTIONS ||--o{ WATCHLIST : watched_by
    
    USERS {
        bigint id PK
        string clerk_user_id
        string email
        string first_name
        string last_name
        string phone
        string role
        boolean email_verified
        boolean phone_verified
        boolean payment_verified
        timestamp created_at
    }
    
    VEHICLES {
        bigint id PK
        bigint seller_id FK
        string vin
        int year
        string make
        string model
        string trim
        int mileage
        string exterior_color
        string interior_color
        string transmission
        string fuel_type
        string body_type
        text description
        numeric starting_price
        string status
        timestamp created_at
    }
    
    VEHICLE_IMAGES {
        bigint id PK
        bigint vehicle_id FK
        string s3_key
        string url
        int display_order
        boolean is_primary
    }
    
    AUCTIONS {
        bigint id PK
        bigint vehicle_id FK
        string status
        timestamp starts_at
        timestamp ends_at
        numeric current_bid
        bigint current_bid_user_id FK
        int bid_count
        int version
        smallint extension_count
    }
    
    BIDS {
        bigint id PK
        bigint auction_id FK
        bigint user_id FK
        numeric amount
        string status
        numeric previous_high_bid
        timestamp created_at
    }
    
    WATCHLIST {
        bigint id PK
        bigint user_id FK
        bigint auction_id FK
        timestamp created_at
    }
    
    NOTIFICATIONS {
        bigint id PK
        bigint user_id FK
        string type
        string title
        text message
        string data
        boolean is_read
        timestamp created_at
    }
```

### Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Denormalized `current_bid` on auctions** | Fast reads without JOIN; updated atomically with OCC |
| **Separate `bids` table** | Full audit trail; never lose bid history |
| **`version` column for OCC** | Detect concurrent modifications |
| **`bid_status` enum** | Track accepted/rejected/outbid for transparency |
| **`previous_high_bid` in bids** | Audit: know what bid was beaten |

---

## API Reference

### Authentication

All authenticated endpoints require `Authorization: Bearer <clerk_jwt>` header.

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/ready` | Readiness probe |
| `GET` | `/metrics` | Prometheus metrics |
| `GET` | `/api/vehicles` | List vehicles with pagination |
| `GET` | `/api/vehicles/:id` | Get vehicle details |
| `GET` | `/api/vehicles/:id/images` | Get vehicle images |
| `GET` | `/api/auctions` | List active auctions |
| `GET` | `/api/auctions/:id` | Get auction details |
| `GET` | `/api/auctions/:id/bids` | Get bid history |
| `GET` | `/api/auctions/:id/stream` | SSE real-time stream |

### Authenticated Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/clerk-sync` | Sync Clerk user to DB |
| `GET` | `/api/auth/me` | Get current user profile |
| `PUT` | `/api/auth/me` | Update profile |
| `POST` | `/api/vehicles` | Create vehicle listing |
| `PUT` | `/api/vehicles/:id` | Update vehicle |
| `DELETE` | `/api/vehicles/:id` | Delete vehicle |
| `POST` | `/api/vehicles/:id/submit` | Submit for auction |
| `POST` | `/api/vehicles/:id/upload-url` | Get S3 presigned URL |
| `POST` | `/api/vehicles/:id/images` | Add image record |
| `DELETE` | `/api/vehicles/:id/images/:imgId` | Delete image |
| `POST` | `/api/decode-vin` | Decode VIN |
| `POST` | `/api/auctions` | Create auction |
| `POST` | `/api/auctions/:id/bids` | Place bid |
| `POST` | `/api/auctions/:id/bid` | Place bid (alias) |
| `GET` | `/api/bids/:ticketId/status` | Check bid status |
| `GET` | `/api/watchlist` | Get user's watchlist |
| `POST` | `/api/auctions/:id/watch` | Add to watchlist |
| `DELETE` | `/api/auctions/:id/watch` | Remove from watchlist |
| `GET` | `/api/auctions/:id/watching` | Check if watching |
| `GET` | `/api/notifications` | Get notifications |
| `GET` | `/api/notifications/unread-count` | Get unread count |
| `POST` | `/api/notifications/:id/read` | Mark as read |
| `POST` | `/api/notifications/read-all` | Mark all as read |
| `DELETE` | `/api/notifications/:id` | Delete notification |

### Debug Endpoints (Development Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/debug/bidengine` | Bid engine stats |
| `GET` | `/debug/sse` | SSE broker stats |
| `GET` | `/debug/stats` | All internal stats |

### Bid Request/Response

**Request:**
```json
POST /api/auctions/123/bids
{
  "amount": 15000.00
}
```

**Response (202 Accepted):**
```json
{
  "ticket_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "queued",
  "message": "Bid submitted for processing"
}
```

**Check Status:**
```json
GET /api/bids/550e8400-e29b-41d4-a716-446655440000/status

{
  "ticket_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "accepted",
  "auction_id": 123,
  "amount": "15000.00",
  "new_high_bid": true,
  "processed_at": "2025-01-01T12:00:00Z"
}
```

---

## Frontend User Journeys

### Buyer Journey

```mermaid
journey
    title Buyer: Finding and Winning a Vehicle
    section Discovery
      Browse auctions: 5: Buyer
      Filter by make/model: 4: Buyer
      View vehicle details: 5: Buyer
      Check vehicle photos: 5: Buyer
    section Engagement
      Sign in with Clerk: 4: Buyer
      Add to watchlist: 5: Buyer
      Receive outbid notification: 3: Buyer
    section Bidding
      Place first bid: 5: Buyer
      See real-time bid updates: 5: Buyer
      Get outbid, place higher: 4: Buyer
      Win auction: 5: Buyer
    section Post-Auction
      Receive winner notification: 5: Buyer
      Complete payment: 3: Buyer
      Arrange pickup/delivery: 3: Buyer
```

### Seller Journey

```mermaid
journey
    title Seller: Listing and Selling a Vehicle
    section Listing
      Sign in with Clerk: 4: Seller
      Enter VIN for auto-fill: 5: Seller
      Add vehicle photos: 4: Seller
      Set starting price: 5: Seller
      Submit for review: 4: Seller
    section Auction
      Schedule auction dates: 5: Seller
      Monitor bid activity: 5: Seller
      See real-time bids: 5: Seller
    section Post-Auction
      Receive sale notification: 5: Seller
      Coordinate with buyer: 4: Seller
      Complete title transfer: 3: Seller
      Receive payment: 5: Seller
```

### User Flow Diagram

```mermaid
flowchart TD
    Start([User Visits Site]) --> Browse[Browse Auctions]
    Browse --> ViewDetail[View Auction Details]
    ViewDetail --> SignedIn{Signed In?}
    
    SignedIn -->|No| SignIn[Sign In with Clerk]
    SignIn --> ClerkModal[Clerk Modal<br/>Google/GitHub SSO]
    ClerkModal --> SyncUser[POST /api/auth/clerk-sync]
    SyncUser --> ViewDetail
    
    SignedIn -->|Yes| Actions{User Action}
    
    Actions -->|Watch| AddWatch[POST /auctions/:id/watch]
    AddWatch --> ViewDetail
    
    Actions -->|Bid| PlaceBid[Enter Bid Amount]
    PlaceBid --> ValidBid{Amount > Current?}
    ValidBid -->|No| ShowError[Show Error]
    ShowError --> PlaceBid
    ValidBid -->|Yes| SubmitBid[POST /auctions/:id/bids]
    SubmitBid --> Queued[202 Accepted + Ticket]
    Queued --> WaitSSE[Wait for SSE Event]
    
    WaitSSE --> BidResult{SSE Event}
    BidResult -->|bid_accepted| Success[Show Success Toast]
    BidResult -->|bid_rejected| ShowError
    
    Success --> ViewDetail
    
    Actions -->|Sell| CreateVehicle[Create Vehicle Listing]
    CreateVehicle --> EnterVIN[Enter VIN]
    EnterVIN --> DecodeVIN[POST /api/decode-vin]
    DecodeVIN --> AutoFill[Auto-fill Make/Model/Year]
    AutoFill --> UploadPhotos[Upload Photos to S3]
    UploadPhotos --> SetPrice[Set Starting Price]
    SetPrice --> SubmitVehicle[POST /api/vehicles/:id/submit]
    SubmitVehicle --> ScheduleAuction[Schedule Auction Dates]
    ScheduleAuction --> AuctionLive[Auction Goes Live]
```

---

## Tech Stack

### Backend

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.23 | Performance, concurrency |
| Router | Chi | Lightweight, middleware |
| Database | PostgreSQL 16 | ACID, JSON support |
| DB Driver | pgx/v5 | Native PostgreSQL |
| Auth | Clerk + JWT | SSO, JWKS validation |
| Real-time | SSE | Server-Sent Events |
| Metrics | Prometheus | Monitoring |
| Tracing | OpenTelemetry | Distributed tracing |
| Errors | Sentry | Error tracking |
| Storage | AWS S3 | Image uploads |
| Cache | Redis | Future: caching |

### Frontend

| Component | Technology | Purpose |
|-----------|------------|---------|
| Framework | React 18 | UI library |
| Language | TypeScript | Type safety |
| Build | Vite | Fast dev server |
| Styling | Tailwind CSS v4 | Utility-first CSS |
| Components | shadcn/ui | Accessible components |
| Routing | React Router | Client routing |
| Data | TanStack Query | Server state |
| Forms | React Hook Form + Zod | Validation |
| Auth | Clerk React | SSO integration |

### Infrastructure

| Component | Technology | Purpose |
|-----------|------------|---------|
| Container | Docker | Containerization |
| Orchestration | Docker Compose | Local dev |
| Tracing UI | Jaeger | Trace visualization |
| CI/CD | (TBD) | Automation |

---

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Node.js 18+ (for frontend)
- Make

### Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/ayubfarah/vehicle-auc.git
cd vehicle-auc

# 2. Start infrastructure (PostgreSQL, Redis, Jaeger)
make docker-up

# 3. Run database migrations
make migrate

# 4. Start the Go backend (port 8080)
make run

# 5. In another terminal, start the frontend (port 3000)
cd frontend
npm install
npm run dev
```

### Environment Variables

Copy `.env.example` to `.env` and configure:

```env
# Server
PORT=8080
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vehicle_auc?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# Auth (Clerk) - Required for protected routes
CLERK_SECRET_KEY=sk_test_...
CLERK_JWKS_URL=https://your-instance.clerk.accounts.dev/.well-known/jwks.json

# AWS S3 - Required for image uploads
AWS_ACCESS_KEY_ID=...
AWS_SECRET_ACCESS_KEY=...
AWS_S3_BUCKET=vehicle-auc-images
AWS_S3_REGION=us-east-1

# Observability
SENTRY_DSN=https://...
OTLP_ENDPOINT=localhost:4317

# Features
DEBUG_ENDPOINTS_ENABLED=true
SYNC_BID_MODE=false
```

### Available Make Commands

```bash
make help           # Show all commands
make build          # Build server binary
make run            # Run server
make dev            # Run with hot reload (air)
make test           # Run all tests
make test-int       # Run integration tests
make test-cover     # Run tests with coverage
make docker-up      # Start PostgreSQL, Redis, Jaeger
make docker-down    # Stop all containers
make migrate        # Apply migrations to dev DB
make migrate-test   # Apply migrations to test DB
make sqlc           # Generate sqlc code
make lint           # Run linters
make fmt            # Format code
```

---

## Testing

### Test Categories

| Type | Location | Purpose |
|------|----------|---------|
| Unit | `internal/*/..._test.go` | Component logic |
| Integration | `tests/integration/` | API + database |
| E2E | (TBD) | Full user flows |

### Running Tests

```bash
# All tests
make test

# Integration tests only (requires test DB)
make test-int

# With coverage report
make test-cover

# Specific package
go test -v ./internal/bidengine/...
```

### Test Database

Integration tests use a separate database on port 5433:

```bash
# Ensure test DB is running
docker compose up -d postgres-test

# Apply migrations
make migrate-test

# Run integration tests
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/vehicle_auc_test?sslmode=disable" \
  go test ./tests/integration/...
```

### Test Coverage

Current coverage: **81 tests passing**

| Package | Coverage |
|---------|----------|
| `internal/bidengine` | OCC logic, retries |
| `internal/realtime` | SSE broker |
| `internal/middleware` | Auth, logging |
| `tests/integration` | All API endpoints |

---

## Observability

### Metrics (Prometheus)

Available at `http://localhost:8080/metrics`:

```
# Bid engine metrics
bidengine_bids_submitted_total
bidengine_bids_processed_total{status="accepted|rejected"}
bidengine_bid_processing_duration_seconds
bidengine_occ_retries_total
bidengine_queue_depth

# SSE metrics
sse_connections_active
sse_events_broadcast_total

# HTTP metrics
http_requests_total{method, path, status}
http_request_duration_seconds
```

### Tracing (Jaeger)

View traces at `http://localhost:16686`:

- Request flow through middleware
- Database queries with timing
- Bid processing spans
- SSE event publishing

### Logging (slog)

Structured JSON logs to stdout:

```json
{
  "time": "2025-01-01T12:00:00Z",
  "level": "INFO",
  "msg": "bid_submitted",
  "request_id": "abc-123",
  "trace_id": "def-456",
  "auction_id": 123,
  "user_id": 456,
  "amount": "15000.00"
}
```

### Debug Endpoints

In development (`DEBUG_ENDPOINTS_ENABLED=true`):

```bash
# Bid engine internal state
curl http://localhost:8080/debug/bidengine

# SSE broker connections
curl http://localhost:8080/debug/sse

# All stats combined
curl http://localhost:8080/debug/stats
```

---

## Project Structure

```
vehicle-auc/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── bidengine/
│   │   ├── engine.go            # Bid queue + dispatcher
│   │   ├── worker.go            # Per-auction workers
│   │   ├── processor.go         # OCC bid processing
│   │   ├── errors.go            # Custom errors
│   │   └── engine_test.go       # Unit tests
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── domain/
│   │   └── types.go             # Shared domain types
│   ├── handler/
│   │   ├── auctions.go          # Auction endpoints
│   │   ├── auth.go              # Auth endpoints
│   │   ├── bids.go              # Bid endpoints
│   │   ├── debug.go             # Debug endpoints
│   │   ├── health.go            # Health checks
│   │   ├── images.go            # Image upload
│   │   ├── notifications.go     # Notifications
│   │   ├── sse.go               # SSE streaming
│   │   ├── vehicles.go          # Vehicle CRUD
│   │   ├── vin.go               # VIN decode
│   │   └── watchlist.go         # Watchlist
│   ├── metrics/
│   │   └── metrics.go           # Prometheus metrics
│   ├── middleware/
│   │   ├── auth.go              # JWT validation
│   │   ├── logging.go           # Request logging
│   │   ├── requestid.go         # Request ID
│   │   ├── tracing.go           # OpenTelemetry
│   │   └── middleware_test.go   # Tests
│   ├── realtime/
│   │   ├── broker.go            # SSE broker
│   │   └── broker_test.go       # Tests
│   ├── repository/
│   │   └── queries/             # SQL files for sqlc
│   └── tracing/
│       └── tracing.go           # OpenTelemetry setup
├── migrations-go/
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── tests/
│   ├── fixtures/
│   │   ├── db.go                # Test DB setup
│   │   └── fixtures.go          # Test data helpers
│   └── integration/
│       ├── auctions_test.go
│       ├── auth_test.go
│       ├── bids_test.go
│       ├── health_test.go
│       ├── images_test.go
│       ├── notifications_test.go
│       ├── vehicles_test.go
│       ├── vin_test.go
│       └── watchlist_test.go
├── frontend/                    # React application
├── docs/
│   ├── BACKEND_ONBOARDING.md
│   └── FRONTEND_ONBOARDING.md
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── sqlc.yaml
```

---

## License

MIT
