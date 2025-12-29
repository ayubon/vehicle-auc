"""Flask application factory."""
import os
from flask import Flask
from .extensions import db, migrate, login_manager, socketio, csrf
from .config import config


def create_app(config_name=None):
    """Create and configure the Flask application."""
    if config_name is None:
        config_name = os.environ.get('FLASK_ENV', 'development')
    
    app = Flask(__name__)
    app.config.from_object(config[config_name])
    
    # Initialize extensions
    register_extensions(app)
    
    # Register blueprints
    register_blueprints(app)
    
    # Register error handlers
    register_error_handlers(app)
    
    # Configure logging
    configure_logging(app)
    
    return app


def register_extensions(app):
    """Register Flask extensions."""
    db.init_app(app)
    migrate.init_app(app, db)
    login_manager.init_app(app)
    csrf.init_app(app)
    socketio.init_app(app, cors_allowed_origins="*")
    
    # Import models to ensure they're registered with SQLAlchemy
    from . import models  # noqa: F401


def register_blueprints(app):
    """Register Flask blueprints."""
    from .routes import main_bp, auth_bp, vehicles_bp, auctions_bp, orders_bp, api_bp
    
    app.register_blueprint(main_bp)
    app.register_blueprint(auth_bp, url_prefix='/auth')
    app.register_blueprint(vehicles_bp, url_prefix='/vehicles')
    app.register_blueprint(auctions_bp, url_prefix='/auctions')
    app.register_blueprint(orders_bp, url_prefix='/orders')
    app.register_blueprint(api_bp, url_prefix='/api')


def register_error_handlers(app):
    """Register error handlers."""
    @app.errorhandler(404)
    def not_found_error(error):
        return {'error': 'Not found'}, 404
    
    @app.errorhandler(500)
    def internal_error(error):
        db.session.rollback()
        return {'error': 'Internal server error'}, 500


def configure_logging(app):
    """Configure structured logging."""
    import structlog
    
    structlog.configure(
        processors=[
            structlog.stdlib.filter_by_level,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            structlog.processors.TimeStamper(fmt="iso"),
            structlog.processors.StackInfoRenderer(),
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            structlog.processors.JSONRenderer()
        ],
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        logger_factory=structlog.stdlib.LoggerFactory(),
        cache_logger_on_first_use=True,
    )
