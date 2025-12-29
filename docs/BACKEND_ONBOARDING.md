# Backend Onboarding Guide

Welcome to the Vehicle Auction Platform backend! This guide will get you up to speed quickly.

---

## Quick Start

```bash
# 1. Start dependencies (MySQL + Redis)
docker compose up -d

# 2. Set up Python environment
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# 3. Copy environment variables
cp .env.example .env
# Edit .env with your credentials

# 4. Run database migrations
flask db upgrade

# 5. Start the server
flask run --port 5001
```

**Verify it's working:**
```bash
curl http://localhost:5001/health
# {"status": "healthy"}
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Flask Application                        │
├─────────────────────────────────────────────────────────────┤
│  routes/           │  models/           │  services/         │
│  (API endpoints)   │  (Database)        │  (External APIs)   │
├─────────────────────────────────────────────────────────────┤
│                     extensions.py                            │
│            (db, jwt, socketio, redis, metrics)              │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
         MySQL 8.x        Redis           AWS S3
```

---

## Directory Structure

```
app/
├── __init__.py          # App factory - creates Flask app
├── config.py            # Environment configs (dev/test/prod)
├── constants.py         # Enums for status values
├── extensions.py        # Flask extensions initialization
├── custom_metrics.py    # Prometheus metrics
│
├── models/              # SQLAlchemy ORM models
│   ├── __init__.py      # Exports all models
│   ├── user.py          # User, Role, SellerProfile
│   ├── vehicle.py       # Vehicle, VehicleImage, VehicleDocument
│   ├── auction.py       # Auction, Bid, Offer
│   ├── order.py         # Order, Invoice, Payment, Refund
│   ├── fulfillment.py   # TitleTransfer, TransportOrder
│   └── misc.py          # Notification, SavedSearch, Watchlist
│
├── routes/              # API endpoints
│   ├── __init__.py      # Blueprint registration
│   ├── auth.py          # /api/auth/* - JWT authentication
│   └── api/             # /api/* - Domain routes
│       ├── __init__.py  # API blueprint
│       ├── vehicles.py  # Vehicle CRUD
│       ├── images.py    # S3 upload URLs
│       ├── auctions.py  # Bid history
│       └── vin.py       # VIN decoding
│
└── services/            # External integrations
    ├── __init__.py
    ├── s3.py            # AWS S3 presigned URLs
    └── clearvin.py      # VIN decoding API
```

---

## Key Concepts

### 1. App Factory Pattern

The app is created via `create_app()` in `app/__init__.py`:

```python
from app import create_app

app = create_app()  # Uses FLASK_ENV to pick config
app = create_app('testing')  # Or specify explicitly
```

This allows different configs for dev/test/prod and makes testing easier.

### 2. Blueprints

Routes are organized into blueprints:

| Blueprint | URL Prefix | Purpose |
|-----------|------------|---------|
| `main_bp` | `/` | Health checks, metrics |
| `auth_bp` | `/api/auth` | Authentication |
| `api_bp` | `/api` | All API endpoints |

### 3. Models & Serialization

Each model has a `to_dict()` method for JSON serialization:

```python
# In routes, return model data like this:
vehicle = db.session.get(Vehicle, vehicle_id)
return jsonify(vehicle.to_dict())

# Or for lists:
vehicles = Vehicle.query.filter_by(status='active').all()
return jsonify([v.to_dict() for v in vehicles])
```

### 4. Status Constants

Use enums from `constants.py` instead of magic strings:

```python
from app.constants import VehicleStatus, AuctionStatus

# Good ✅
vehicle.status = VehicleStatus.ACTIVE.value

# Bad ❌
vehicle.status = 'active'
```

### 5. Authentication

We use **Clerk** for SSO on the frontend, synced to **Flask-JWT-Extended**:

```python
from flask_jwt_extended import jwt_required, current_user

@api_bp.route('/vehicles', methods=['POST'])
@jwt_required()  # Requires valid JWT token
def create_vehicle():
    # current_user is the authenticated User object
    vehicle = Vehicle(seller_id=current_user.id, ...)
```

**Auth flow:**
1. User signs in via Clerk (frontend)
2. Frontend calls `POST /api/auth/clerk-sync` with Clerk user data
3. Backend creates/finds user, returns Flask JWT
4. Frontend includes JWT in `Authorization: Bearer <token>` header

---

## Common Tasks

### Adding a New API Endpoint

1. **Choose the right file** in `routes/api/`:
   - Vehicle-related → `vehicles.py`
   - New domain → create new file

2. **Add the route:**
```python
# routes/api/vehicles.py
from . import api_bp

@api_bp.route('/vehicles/<int:id>/archive', methods=['POST'])
@jwt_required()
def archive_vehicle(id):
    vehicle = db.session.get(Vehicle, id)
    if not vehicle:
        return jsonify({'error': 'Not found'}), 404
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    vehicle.status = VehicleStatus.ARCHIVED.value
    db.session.commit()
    return jsonify({'message': 'Vehicle archived'})
```

### Adding a New Model

1. **Create model** in `models/`:
```python
# models/example.py
from ..extensions import db

class Example(db.Model):
    __tablename__ = 'examples'
    
    id = db.Column(db.BigInteger, primary_key=True)
    name = db.Column(db.String(255), nullable=False)
    created_at = db.Column(db.DateTime, default=db.func.now())
    
    def to_dict(self):
        return {
            'id': self.id,
            'name': self.name,
            'created_at': self.created_at.isoformat() if self.created_at else None,
        }
```

2. **Export it** in `models/__init__.py`:
```python
from .example import Example
__all__ = [..., 'Example']
```

3. **Create migration:**
```bash
flask db migrate -m "Add examples table"
flask db upgrade
```

### Adding a New Service

1. **Create service** in `services/`:
```python
# services/example_api.py
import os
import structlog

logger = structlog.get_logger(__name__)

class ExampleService:
    def __init__(self):
        self.api_key = os.environ.get('EXAMPLE_API_KEY')
        self.enabled = bool(self.api_key)
        
        if not self.enabled:
            logger.warning("Example API key not configured")
    
    def do_something(self, data):
        if not self.enabled:
            return {'mock': True, 'data': data}
        
        # Real API call here
        ...

example_service = ExampleService()
```

2. **Export it** in `services/__init__.py`:
```python
from .example_api import example_service
```

---

## Database

### Running Migrations

```bash
# Create a new migration after model changes
flask db migrate -m "Description of changes"

# Apply migrations
flask db upgrade

# Rollback one migration
flask db downgrade
```

### Useful Queries

```python
# Get by ID
vehicle = db.session.get(Vehicle, 123)

# Filter
active_vehicles = Vehicle.query.filter_by(status='active').all()

# Complex filter
from sqlalchemy import and_
vehicles = Vehicle.query.filter(
    and_(
        Vehicle.year >= 2020,
        Vehicle.starting_price <= 50000
    )
).all()

# Join
from app.models import Vehicle, VehicleImage
vehicles_with_images = db.session.query(Vehicle).join(VehicleImage).all()
```

---

## Testing

```bash
# Run all tests
pytest

# Run with verbose output
pytest -v

# Run specific test file
pytest tests/unit/test_models.py

# Run with coverage
pytest --cov=app --cov-report=html
```

### Writing Tests

```python
# tests/unit/test_example.py
import pytest
from app.models import Vehicle

def test_vehicle_to_dict(app):
    """Test Vehicle serialization."""
    with app.app_context():
        vehicle = Vehicle(
            vin='12345678901234567',
            year=2020,
            make='Toyota',
            model='Camry',
        )
        
        data = vehicle.to_dict()
        
        assert data['year'] == 2020
        assert data['make'] == 'Toyota'
```

---

## Environment Variables

Key variables in `.env`:

```bash
# Flask
FLASK_APP=app
FLASK_ENV=development
SECRET_KEY=your-secret-key

# Database
DATABASE_URL=mysql+pymysql://user:pass@localhost:3306/vehicle_auc

# Redis
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_SECRET_KEY=your-jwt-secret

# AWS S3
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx
AWS_S3_BUCKET=vehicle-auc-images
AWS_REGION=us-east-1

# ClearVIN (optional - has mock fallback)
CLEARVIN_API_KEY=xxx
```

---

## Debugging

### Logs

Structured JSON logs via `structlog`:

```python
import structlog
logger = structlog.get_logger(__name__)

logger.info("Something happened", vehicle_id=123, user_id=456)
# Output: {"event": "Something happened", "vehicle_id": 123, "user_id": 456, ...}
```

### Request Tracing

Every request gets a `request_id` in logs and response headers:

```bash
curl -i http://localhost:5001/api/vehicles
# X-Request-ID: a1b2c3d4
```

Search logs by this ID to trace a request.

### Common Issues

| Issue | Solution |
|-------|----------|
| `401 Unauthorized` | Check JWT token in `Authorization: Bearer <token>` header |
| `403 Forbidden` | User doesn't own the resource |
| `500 Internal Server Error` | Check Flask logs in terminal |
| Database connection error | Is MySQL running? `docker compose up -d` |

---

## API Reference

### Authentication

```bash
# Sync Clerk user → get JWT
POST /api/auth/clerk-sync
Body: {"clerk_user_id": "...", "email": "...", "first_name": "...", "last_name": "..."}
Response: {"access_token": "...", "refresh_token": "...", "user": {...}}

# Get current user
GET /api/auth/me
Headers: Authorization: Bearer <token>
```

### Vehicles

```bash
# List vehicles
GET /api/vehicles?make=Toyota&year_min=2020&max_price=50000

# Get vehicle
GET /api/vehicles/123

# Create vehicle (requires auth)
POST /api/vehicles
Headers: Authorization: Bearer <token>
Body: {"vin": "...", "year": 2020, "make": "Toyota", ...}

# Submit for review → active
POST /api/vehicles/123/submit
Headers: Authorization: Bearer <token>

# Get S3 upload URL
POST /api/vehicles/123/upload-url
Headers: Authorization: Bearer <token>
Body: {"filename": "photo.jpg", "content_type": "image/jpeg"}
```

### Health

```bash
GET /health           # Basic health
GET /health/detailed  # DB + Redis status
GET /metrics          # Prometheus metrics
```

---

## Questions?

- Check the main `README.md` for project overview
- Look at existing code in `routes/api/vehicles.py` for patterns
- Run tests to understand expected behavior
