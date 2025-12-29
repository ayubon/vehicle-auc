# Vehicle Auction Platform

> A full-featured vehicle auction marketplace clone of [RemarketSpace.com](https://www.remarketspace.com/) — enabling buyers to bid on vehicles, sellers to list inventory, and admins to manage the entire transaction lifecycle from listing to title transfer and delivery.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Data Model](#data-model)
- [External Integrations](#external-integrations)
- [Development Phases](#development-phases)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Testing Strategy](#testing-strategy)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
- [User Journeys](#user-journeys)
- [Contributing](#contributing)

---

## Project Overview

### What We're Building

A Minnesota-style vehicle auction platform that handles:

| Feature | Description |
|---------|-------------|
| **User Registration** | Sign up, email verification, ID verification (Persona), payment method setup (Authorize.Net) |
| **Vehicle Listings** | VIN decoding (ClearVIN), multi-image uploads (S3), filtering, search |
| **Live Auctions** | Real-time WebSocket bidding, auto-bid (proxy), anti-snipe timer extension |
| **Make an Offer** | Alternative to auction — direct negotiation |
| **Payments** | Invoice generation, card capture, refunds via Authorize.Net |
| **Title Transfer** | DMV integration (DLR DMV) for legal title processing |
| **Transport** | Towing dispatch via Super Dispatch API |
| **Notifications** | Transactional emails via SendGrid |
| **Admin Panel** | User management, auction oversight, refunds, reporting |

### Target Reference

We are cloning the functionality of **[RemarketSpace.com](https://www.remarketspace.com/)** — a vehicle marketplace serving Minnesota with:
- Inventory browsing with filters
- Timed auctions with real-time bidding
- "Make an Offer" flow
- Buyer verification (ID + payment)
- Seller onboarding
- Title and transport coordination

---

## Tech Stack

### Frontend (React SPA)
| Technology | Purpose |
|------------|---------|
| **React 18 + TypeScript** | UI framework |
| **Vite** | Build tool & dev server |
| **Tailwind CSS v4** | Utility-first styling |
| **shadcn/ui** | Pre-built accessible components |
| **React Router** | Client-side routing |
| **TanStack Query** | Server state management |
| **React Hook Form + Zod** | Form handling & validation |
| **Socket.IO Client** | Real-time bid updates |
| **Lucide React** | Icons |

### Backend (Flask JSON API)
| Technology | Purpose |
|------------|---------|
| **Python 3.11+** | Runtime |
| **Flask 3.x** | REST API framework |
| **Flask-SQLAlchemy** | ORM for MySQL |
| **Flask-JWT-Extended** | JWT authentication |
| **Flask-SocketIO** | WebSocket server for real-time bidding |
| **Flask-Admin** | Admin panel UI |
| **Celery + Redis** | Background job processing |
| **APScheduler** | Cron jobs (auction timers) |

### Database & Cache
| Technology | Purpose |
|------------|---------|
| **MySQL 8.x** | Primary datastore |
| **Redis** | Session cache, Celery broker, WebSocket pub/sub |

### Infrastructure
| Technology | Purpose |
|------------|---------|
| **Docker + Docker Compose** | Local dev & deployment |
| **AWS EC2** | Application server |
| **AWS S3** | Vehicle images, documents |
| **AWS CloudFront** | CDN for static assets |
| **Cloudflare** | DNS, SSL, DDoS protection |
| **Apache** | Reverse proxy |

### Observability
| Technology | Purpose |
|------------|---------|
| **structlog** | Structured JSON logging |
| **OpenTelemetry** | Distributed tracing |
| **Prometheus** | Metrics collection |
| **Grafana** | Dashboards & alerting |

---

## Architecture

This project uses a **decoupled SPA architecture**:

- **Frontend**: React SPA served by Vite (dev) or static hosting (prod)
- **Backend**: Flask JSON API (no server-side rendering)
- **Communication**: REST API + WebSocket for real-time features

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         FRONTEND (React SPA)                                 │
│  React 18 + TypeScript + Vite + Tailwind CSS + shadcn/ui                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Pages        │  │ Components   │  │ Services     │  │ Hooks        │     │
│  │ (Routes)     │  │ (shadcn/ui)  │  │ (API calls)  │  │ (React Query)│     │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘     │
│                         + Socket.IO Client (real-time bids)                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │ REST API + WebSocket
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         BACKEND (Flask JSON API)                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Auth         │  │ Auction      │  │ Payments     │  │ Titles/DMV   │     │
│  │ (JWT)        │  │ Engine       │  │ (Authorize.  │  │ (DLR DMV)    │     │
│  │              │  │ (WebSocket)  │  │ Net)         │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Transport    │  │ VIN Decode   │  │ Notifications│  │ AI Assistant │     │
│  │ (Super       │  │ (ClearVIN)   │  │ (SendGrid)   │  │ (OpenAI)     │     │
│  │ Dispatch)    │  │              │  │              │  │              │     │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘     │
│                         + Celery Workers + APScheduler                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATABASE                                        │
│                    MySQL 8.x + Redis (cache/broker)                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            INFRASTRUCTURE                                    │
│  AWS EC2 (Docker) │ AWS S3 (files) │ CloudFront (CDN) │ Cloudflare (DNS)   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Why This Architecture?

| Benefit | Description |
|---------|-------------|
| **Better UX** | SPA provides instant navigation, no page reloads |
| **Reusable Components** | shadcn/ui provides consistent, accessible UI |
| **Type Safety** | TypeScript catches errors at compile time |
| **Scalability** | Frontend can be deployed to CDN, backend scales independently |
| **Modern Tooling** | Hot reload, tree shaking, code splitting out of the box |

---

## Data Model

### Core Entities (13+ tables)

#### Users & Auth
```
users
├── id, email, password_hash, role (buyer/seller/admin), status
├── first_name, last_name, phone, address fields
├── persona_inquiry_id, id_verified_at (Persona)
└── authorize_customer_profile_id, authorize_payment_profile_id (Authorize.Net)

seller_profiles
├── user_id (FK), business_name, dealer_license_number, tax_id
└── approved_at
```

#### Vehicles
```
vehicles
├── id, seller_id (FK), status (draft/active/sold/archived)
├── vin, year, make, model, trim, body_type, engine, transmission
├── mileage, condition, title_type, title_state, description
├── starting_price, reserve_price, buy_now_price
└── location fields, lat/lng

vehicle_images
├── vehicle_id (FK), s3_key, url, is_primary, sort_order

vehicle_documents
├── vehicle_id (FK), doc_type, s3_key
```

#### Auctions & Bids
```
auctions
├── id, vehicle_id (FK), auction_type, status
├── starts_at, ends_at, extended_count
└── current_bid, bid_count, winner_id (FK)

bids
├── auction_id (FK), user_id (FK), amount, max_bid
├── is_auto_bid, ip_address, created_at

offers
├── vehicle_id (FK), user_id (FK), amount, status
├── counter_amount, expires_at
```

#### Orders & Payments
```
orders
├── id, order_number, auction_id/offer_id (FK), buyer_id, seller_id, vehicle_id
├── vehicle_price, buyer_fee, transport_fee, title_fee, tax, total
└── status (pending_payment → paid → title_processing → delivered → completed)

invoices
├── order_id (FK), invoice_number, due_date, paid_at, pdf_s3_key

payments
├── order_id (FK), user_id (FK), amount, payment_type, status
└── authorize_transaction_id, authorize_response_code, last_four

refunds
├── payment_id (FK), amount, reason, authorize_transaction_id, status
```

#### Fulfillment
```
title_transfers
├── order_id (FK), status, dlr_submission_id, new_title_number
└── title_document_s3_key, bill_of_sale_s3_key

transport_orders
├── order_id (FK), status, super_dispatch_order_id
├── carrier_name, driver_name, driver_phone
├── pickup/delivery addresses, quoted_price, dates
```

#### Supporting
```
notifications
├── user_id (FK), type, title, message, data (JSON), read_at

saved_searches
├── user_id (FK), name, filters (JSON), notify_email

watchlist
├── user_id (FK), vehicle_id (FK)

audit_log
├── user_id (FK), action, entity_type, entity_id, old_data, new_data, ip_address
```

---

## External Integrations

| Service | Purpose | Status |
|---------|---------|--------|
| **Clerk** | SSO authentication (Google, GitHub) | ✅ Configured |
| **AWS S3** | Vehicle image storage | ✅ Working |
| **ClearVIN** | VIN decoding (year, make, model, specs) | ✅ Working (mock fallback) |
| **Authorize.Net** | Payment processing | ❌ Not started |
| **Persona** | ID verification | ❌ Not started |
| **SendGrid** | Transactional emails | ❌ Not started |
| **Super Dispatch** | Towing/transport dispatch | ❌ Not started |
| **DLR DMV** | Title transfer processing | ❌ Not started |
| **OpenAI** | AI assistant | ❌ Not started |

---

## Development Phases

### Phase 1: Foundation ✅ Complete
- [x] Test harness (pytest + Playwright)
- [x] Docker setup (MySQL + Redis + Flask)
- [x] Flask app factory with blueprints
- [x] SQLAlchemy models for all tables (16 models)
- [x] Database migrations (Alembic)
- [x] Observability (structlog, Prometheus metrics, Sentry)
- [x] Health check endpoints (`/health`, `/health/detailed`)

### Phase 2: Authentication ✅ Complete
- [x] Clerk SSO integration (Google, GitHub sign-in)
- [x] Flask-JWT-Extended for API auth
- [x] Clerk → Flask JWT sync endpoint (`/api/auth/clerk-sync`)
- [x] User registration + profile management
- [x] Role-based access (buyer, seller, admin)
- [ ] Persona ID verification integration (not started)
- [ ] Authorize.Net payment profile setup (not started)

### Phase 3: Vehicles ✅ Complete
- [x] ClearVIN API integration (with mock fallback)
- [x] Vehicle CRUD (create, read, update, delete)
- [x] S3 image uploads (presigned URLs for upload + view)
- [x] Drag-and-drop image upload UI
- [x] Inventory listing page with filters (make, year, price)
- [x] Vehicle detail page
- [x] Vehicle create form with VIN decode
- [x] Submit for review → active (admin approval skipped for now)

### Phase 4: Auctions ⏳ Next
- [ ] Auction creation and management
- [ ] Real-time bidding via WebSocket (Flask-SocketIO)
- [ ] Auto-bid (proxy bidding)
- [ ] Auction timer with anti-snipe extension
- [ ] "Make an Offer" flow

### Phase 5: Orders & Payments
- [ ] Order creation on auction win / offer accept
- [ ] Invoice generation (PDF via WeasyPrint)
- [ ] Payment capture via Authorize.Net
- [ ] Refund workflow

### Phase 6: Fulfillment
- [ ] DLR DMV title transfer integration
- [ ] Super Dispatch transport integration
- [ ] Order status tracking

### Phase 7: Notifications & Polish
- [ ] SendGrid transactional emails
- [ ] Saved searches (Vehicle Finder)
- [ ] Admin panel (Flask-Admin)
- [ ] OpenAI integration

### Phase 8: Deployment
- [ ] Docker production configuration
- [ ] AWS EC2 + S3 + RDS setup
- [ ] CloudFront CDN configuration
- [ ] Cloudflare DNS + SSL
- [ ] GitHub Actions CI/CD

---

## Project Structure

```
vehicle-auc/
├── frontend/                    # React SPA (Vite + TypeScript)
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/              # shadcn/ui (Button, Card, Input, etc.)
│   │   │   ├── Layout.tsx       # App layout with nav
│   │   │   └── ImageUpload.tsx  # Drag-and-drop S3 upload
│   │   ├── pages/
│   │   │   ├── HomePage.tsx
│   │   │   ├── VehiclesPage.tsx      # Grid + filters
│   │   │   ├── VehicleDetailPage.tsx
│   │   │   ├── VehicleCreatePage.tsx # Full form + image upload
│   │   │   ├── AuctionsPage.tsx
│   │   │   ├── AuctionDetailPage.tsx
│   │   │   └── DashboardPage.tsx
│   │   ├── hooks/
│   │   │   ├── useVehicles.ts   # List + filter vehicles
│   │   │   ├── useVehicle.ts    # Single vehicle fetch
│   │   │   └── useAuth.ts       # Clerk → Flask JWT sync
│   │   ├── services/
│   │   │   └── api.ts           # Axios client + API functions
│   │   ├── types/
│   │   │   ├── vehicle.ts       # Vehicle interfaces
│   │   │   └── form.ts          # Zod schemas
│   │   ├── App.tsx              # Routes + Clerk provider
│   │   └── main.tsx
│   └── package.json
│
├── app/                         # Flask JSON API
│   ├── __init__.py              # App factory + setup
│   ├── config.py                # Environment configs
│   ├── constants.py             # Status enums (VehicleStatus, etc.)
│   ├── extensions.py            # Flask extensions (db, jwt, socketio)
│   ├── custom_metrics.py        # Prometheus metrics
│   ├── models/
│   │   ├── user.py              # User, Role, SellerProfile
│   │   ├── vehicle.py           # Vehicle, VehicleImage, VehicleDocument
│   │   ├── auction.py           # Auction, Bid, Offer
│   │   ├── order.py             # Order, Invoice, Payment, Refund
│   │   ├── fulfillment.py       # TitleTransfer, TransportOrder
│   │   └── misc.py              # Notification, SavedSearch, Watchlist
│   ├── routes/
│   │   ├── api/                 # Domain-organized API routes
│   │   │   ├── vehicles.py      # CRUD, submit, filters
│   │   │   ├── images.py        # S3 upload URLs
│   │   │   ├── auctions.py      # Bid history
│   │   │   └── vin.py           # VIN decoding
│   │   └── auth.py              # JWT auth + Clerk sync
│   └── services/
│       ├── s3.py                # Presigned URLs
│       └── clearvin.py          # VIN decoding (mock available)
│
├── migrations/                  # Alembic migrations
├── tests/
│   ├── e2e/                     # Playwright E2E tests
│   └── unit/                    # Unit tests (14 passing)
├── docker-compose.yml
├── requirements.txt
└── .env.example
```

---

## Getting Started

### Prerequisites

- Python 3.11+
- Node.js 18+
- Docker & Docker Compose

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/ayubon/vehicle-auc.git
cd vehicle-auc

# Copy environment variables
cp .env.example .env
# Edit .env with your API keys

# Start services (MySQL, Redis)
docker compose up -d

# ─────────────────────────────────────────────────────────────
# BACKEND (Flask API)
# ─────────────────────────────────────────────────────────────

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run database migrations
flask db upgrade

# Start the API server (port 5001)
flask run --port 5001

# ─────────────────────────────────────────────────────────────
# FRONTEND (React SPA)
# ─────────────────────────────────────────────────────────────

# In a new terminal
cd frontend

# Install dependencies
npm install

# Start dev server (port 3000, proxies /api to Flask)
npm run dev
```

### Development URLs

| Service | URL |
|---------|-----|
| **Frontend** | http://localhost:3000 |
| **Backend API** | http://localhost:5001/api |
| **Health Check** | http://localhost:5001/health |
| **Prometheus Metrics** | http://localhost:5001/metrics |

### Running Tests

```bash
# Install test dependencies
pip install -r requirements-dev.txt
npx playwright install

# Run all tests
pytest

# Run E2E tests only
pytest tests/e2e/

# Run with coverage
pytest --cov=app --cov-report=html

# Run Playwright tests with UI
pytest tests/e2e/ --headed
```

---

## Testing Strategy

### Layers

| Layer | Tool | Purpose |
|-------|------|---------|
| **E2E / Black-box** | Playwright | User journey testing in real browser |
| **Integration** | pytest + httpx | API endpoint testing |
| **Unit** | pytest | Isolated function/model testing |
| **Load** | Locust | Performance under stress |

### Observability for Debugging

- **Structured logging** with correlation IDs
- **OpenTelemetry tracing** across Flask → MySQL → external APIs
- **Prometheus metrics** for request latency, bid counts, errors
- **Playwright traces** (screenshots, video on failure)

---

## Environment Variables

```bash
# Flask
FLASK_APP=app
FLASK_ENV=development
SECRET_KEY=your-secret-key

# Database
DATABASE_URL=mysql+pymysql://user:pass@localhost:3306/vehicle_auc
REDIS_URL=redis://localhost:6379/0

# Authorize.Net
AUTHORIZE_NET_API_LOGIN_ID=your-login-id
AUTHORIZE_NET_TRANSACTION_KEY=your-transaction-key
AUTHORIZE_NET_SANDBOX=true

# Persona
PERSONA_API_KEY=your-api-key
PERSONA_TEMPLATE_ID=your-template-id

# ClearVIN
CLEARVIN_API_KEY=your-api-key

# Super Dispatch
SUPER_DISPATCH_API_KEY=your-api-key

# DLR DMV
DLR_DMV_API_KEY=your-api-key

# SendGrid
SENDGRID_API_KEY=your-api-key
SENDGRID_FROM_EMAIL=noreply@yourdomain.com

# OpenAI
OPENAI_API_KEY=your-api-key

# AWS
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_S3_BUCKET=your-bucket-name
AWS_REGION=us-east-1
```

---

## User Journeys

### 1. Buyer Registration → First Bid
```
Register → Verify Email → Upload ID (Persona) → Add Payment Card → 
Browse Inventory → View Vehicle → Place Bid → Win Auction → 
Pay Invoice → Track Title → Track Transport → Receive Vehicle
```

### 2. Seller Listing → Sale Complete
```
Register as Seller → Verify Business License → 
Enter VIN → Auto-decode → Upload Photos → Set Pricing → 
Submit for Review → Approved → Auction Goes Live → 
Auction Ends → Receive Payment → Ship Title → Coordinate Transport
```

### 3. Admin Workflow
```
Review New Listings → Approve/Reject → 
Monitor Live Auctions → Handle Disputes → 
Process Refunds → Manage Users → View Reports
```

---

## API Endpoints

### Auth (Implemented ✅)
- `POST /api/auth/clerk-sync` — Sync Clerk user → get Flask JWT
- `POST /api/auth/register` — User registration
- `POST /api/auth/login` — User login (email/password)
- `POST /api/auth/refresh` — Refresh JWT token
- `GET /api/auth/me` — Get current user
- `PUT /api/auth/profile` — Update profile

### Vehicles (Implemented ✅)
- `GET /api/vehicles` — List vehicles (with filters: make, year, price)
- `GET /api/vehicles/<id>` — Vehicle detail
- `POST /api/vehicles` — Create vehicle (seller, JWT required)
- `PUT /api/vehicles/<id>` — Update vehicle
- `DELETE /api/vehicles/<id>` — Delete vehicle
- `POST /api/vehicles/<id>/submit` — Submit for review → active
- `POST /api/vehicles/<id>/upload-url` — Get S3 presigned upload URL
- `POST /api/vehicles/<id>/images` — Register uploaded image
- `DELETE /api/vehicles/<id>/images/<image_id>` — Delete image

### VIN (Implemented ✅)
- `GET /api/vin/decode/<vin>` — Decode VIN via ClearVIN

### Auctions (Partial)
- `GET /api/auctions/<id>/bids` — Get bid history

### Health (Implemented ✅)
- `GET /health` — Basic health check
- `GET /health/detailed` — Detailed health (DB, Redis)
- `GET /metrics` — Prometheus metrics

### Planned
- `POST /api/auctions/<id>/bid` — Place bid (WebSocket)
- `POST /api/auctions/<id>/auto-bid` — Set auto-bid
- `GET /api/orders` — User's orders
- `POST /api/orders/<id>/pay` — Process payment

---

## Contributing

1. Create a feature branch from `main`
2. Write tests first (TDD)
3. Implement until tests pass
4. Submit PR with description of changes
5. Ensure CI passes

---

## License

[MIT](LICENSE)

---

## Contact

- **Repository**: https://github.com/ayubon/vehicle-auc
- **Reference Site**: https://www.remarketspace.com/
