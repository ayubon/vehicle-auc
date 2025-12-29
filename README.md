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

### Frontend
| Technology | Purpose |
|------------|---------|
| **Bootstrap 5** | Responsive UI framework |
| **Vanilla JavaScript + AJAX** | Client-side interactivity |
| **Jinja2** | Server-side HTML templating |
| **Socket.IO Client** | Real-time bid updates |

### Backend
| Technology | Purpose |
|------------|---------|
| **Python 3.11+** | Runtime |
| **Flask 3.x** | Web framework |
| **Flask-SQLAlchemy** | ORM for MySQL |
| **Flask-Security-Too** | Authentication (registration, login, roles, 2FA) |
| **Flask-SocketIO** | WebSocket server for real-time bidding |
| **Flask-Admin** | Admin panel UI |
| **Flask-WTF** | Forms + CSRF protection |
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

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND                                        │
│  Bootstrap 5 + JavaScript/AJAX + Socket.IO Client + Jinja2 Templates        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         BACKEND (Flask + Python 3.11)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Auth         │  │ Auction      │  │ Payments     │  │ Titles/DMV   │     │
│  │ (Flask-      │  │ Engine       │  │ (Authorize.  │  │ (DLR DMV)    │     │
│  │ Security-Too)│  │ (WebSocket)  │  │ Net)         │  │              │     │
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

| Service | Purpose | Sandbox Available |
|---------|---------|-------------------|
| **Authorize.Net** | Payment processing ($0 auth, charges, refunds) | ✅ Yes |
| **Persona** | ID verification (driver's license, passport) | ✅ Yes |
| **ClearVIN** | VIN decoding (year, make, model, specs) | ⚠️ Check |
| **Super Dispatch** | Towing/transport dispatch and tracking | ⚠️ Check |
| **DLR DMV** | Title transfer processing (state-specific) | ⚠️ Check |
| **SendGrid** | Transactional emails | ✅ Yes |
| **OpenAI** | AI assistant (vehicle descriptions, support) | ✅ Yes |
| **AWS S3** | File storage (images, documents) | ✅ Yes |

---

## Development Phases

### Phase 1: Foundation ⏳
- [ ] Test harness (pytest + Playwright)
- [ ] Docker setup (MySQL + Redis + Flask)
- [ ] Flask app factory with blueprints
- [ ] SQLAlchemy models for all tables
- [ ] Database migrations (Alembic)
- [ ] Observability (structlog, OpenTelemetry, Prometheus)

### Phase 2: Authentication
- [ ] Flask-Security-Too configuration
- [ ] User registration + email verification
- [ ] Login/logout + password reset
- [ ] Persona ID verification integration
- [ ] Authorize.Net payment profile setup ($0 auth)
- [ ] Role-based access (buyer, seller, admin)

### Phase 3: Vehicles
- [ ] ClearVIN API integration
- [ ] Vehicle CRUD (create, read, update, delete)
- [ ] S3 image uploads (presigned URLs)
- [ ] Inventory listing page with filters
- [ ] Vehicle detail page

### Phase 4: Auctions (Core Feature)
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
├── app/
│   ├── __init__.py              # Flask app factory
│   ├── config.py                # Configuration classes
│   ├── extensions.py            # Flask extensions init
│   ├── models/                  # SQLAlchemy models
│   │   ├── __init__.py
│   │   ├── user.py
│   │   ├── vehicle.py
│   │   ├── auction.py
│   │   ├── order.py
│   │   └── ...
│   ├── routes/                  # Blueprints
│   │   ├── __init__.py
│   │   ├── auth.py
│   │   ├── vehicles.py
│   │   ├── auctions.py
│   │   ├── orders.py
│   │   └── admin.py
│   ├── services/                # Business logic & integrations
│   │   ├── __init__.py
│   │   ├── authorize_net.py
│   │   ├── persona.py
│   │   ├── clearvin.py
│   │   ├── super_dispatch.py
│   │   ├── dlr_dmv.py
│   │   ├── sendgrid.py
│   │   └── openai.py
│   ├── templates/               # Jinja2 templates
│   │   ├── base.html
│   │   ├── auth/
│   │   ├── vehicles/
│   │   ├── auctions/
│   │   └── ...
│   ├── static/                  # CSS, JS, images
│   │   ├── css/
│   │   ├── js/
│   │   └── img/
│   └── websocket.py             # Socket.IO event handlers
├── migrations/                  # Alembic migrations
├── tests/
│   ├── conftest.py              # Shared fixtures
│   ├── e2e/                     # Playwright E2E tests
│   │   ├── test_registration.py
│   │   ├── test_bidding.py
│   │   └── ...
│   ├── integration/             # API integration tests
│   │   ├── test_auth_api.py
│   │   ├── test_auction_api.py
│   │   └── ...
│   ├── unit/                    # Unit tests
│   │   ├── test_models.py
│   │   └── test_services.py
│   └── load/                    # Load tests (Locust)
│       └── locustfile.py
├── observability/
│   ├── prometheus.yml
│   └── grafana/
├── docker-compose.yml           # Local development
├── docker-compose.test.yml      # Test environment
├── docker-compose.prod.yml      # Production
├── Dockerfile
├── requirements.txt
├── pytest.ini
├── pyproject.toml
├── .env.example
└── README.md
```

---

## Getting Started

### Prerequisites

- Python 3.11+
- Docker & Docker Compose
- Node.js (for Playwright)
- GitHub CLI (`gh`)

### Local Development Setup

```bash
# Clone the repository
git clone https://github.com/ayubon/vehicle-auc.git
cd vehicle-auc

# Copy environment variables
cp .env.example .env
# Edit .env with your API keys

# Start services (MySQL, Redis)
docker-compose up -d

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Run database migrations
flask db upgrade

# Start the development server
flask run --debug

# In another terminal, start Celery worker
celery -A app.celery worker --loglevel=info
```

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

## API Endpoints (Planned)

### Auth
- `POST /auth/register` — User registration
- `POST /auth/login` — User login
- `POST /auth/logout` — User logout
- `POST /auth/reset-password` — Password reset

### Vehicles
- `GET /api/vehicles` — List vehicles (with filters)
- `GET /api/vehicles/<id>` — Vehicle detail
- `POST /api/vehicles` — Create vehicle (seller)
- `PUT /api/vehicles/<id>` — Update vehicle
- `DELETE /api/vehicles/<id>` — Delete vehicle
- `POST /api/vehicles/decode-vin` — Decode VIN via ClearVIN

### Auctions
- `GET /api/auctions` — List active auctions
- `GET /api/auctions/<id>` — Auction detail
- `POST /api/auctions/<id>/bid` — Place bid
- `POST /api/auctions/<id>/auto-bid` — Set auto-bid
- `WebSocket /socket.io` — Real-time bid updates

### Orders
- `GET /api/orders` — User's orders
- `GET /api/orders/<id>` — Order detail
- `POST /api/orders/<id>/pay` — Process payment
- `GET /api/orders/<id>/invoice` — Download invoice PDF

### Admin
- `GET /admin/` — Admin dashboard (Flask-Admin)

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
