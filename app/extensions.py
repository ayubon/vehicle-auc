"""Flask extensions initialization."""
from flask_sqlalchemy import SQLAlchemy
from flask_migrate import Migrate
from flask_login import LoginManager
from flask_wtf.csrf import CSRFProtect
from flask_socketio import SocketIO

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
