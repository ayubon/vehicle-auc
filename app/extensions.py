"""Flask extensions initialization."""
import os
import redis
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
from flask_login import LoginManager
from flask_wtf.csrf import CSRFProtect
from flask_socketio import SocketIO
from prometheus_flask_exporter import PrometheusMetrics

# Database
db = SQLAlchemy()

# Migrations
migrate = Migrate()

# Login manager
login_manager = LoginManager()
login_manager.login_view = 'auth.login'
login_manager.login_message_category = 'info'


@login_manager.user_loader
def load_user(user_id):
    """Load user by ID for Flask-Login."""
    from .models import User
    return User.query.get(int(user_id))


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
