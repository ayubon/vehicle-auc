"""Flask extensions initialization."""
import os
import redis
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
from flask_wtf.csrf import CSRFProtect
from flask_socketio import SocketIO
from flask_jwt_extended import JWTManager
from prometheus_flask_exporter import PrometheusMetrics

# Database
db = SQLAlchemy()

# Migrations
migrate = Migrate()

# JWT Authentication
jwt = JWTManager()

# CSRF protection
csrf = CSRFProtect()

# WebSocket
socketio = SocketIO()

# Prometheus metrics
metrics = PrometheusMetrics.for_app_factory()

# Redis client
redis_client = redis.from_url(
    os.environ.get('REDIS_URL', 'redis://localhost:6379/0'),
    decode_responses=True
)
