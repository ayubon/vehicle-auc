"""Flask application factory."""
import os
import logging
import sentry_sdk
from flask import Flask, request, g
from sentry_sdk.integrations.flask import FlaskIntegration
from sentry_sdk.integrations.sqlalchemy import SqlalchemyIntegration
from sentry_sdk.integrations.redis import RedisIntegration
from .extensions import db, migrate, login_manager, socketio, csrf, metrics, redis_client
from .config import config
from .custom_metrics import init_app_info


def create_app(config_name=None):
    """Create and configure the Flask application."""
    if config_name is None:
        config_name = os.environ.get('FLASK_ENV', 'development')
    
    # Initialize Sentry before anything else
    init_sentry(config_name)
    
    app = Flask(__name__)
    app.config.from_object(config[config_name])
    
    # Configure logging first
    configure_logging(app)
    
    # Initialize extensions
    register_extensions(app)
    
    # Register blueprints
    register_blueprints(app)
    
    # Register error handlers
    register_error_handlers(app)
    
    # Register request hooks for observability
    register_request_hooks(app)
    
    # Initialize custom metrics
    init_app_info(
        version='0.1.0',
        environment=config_name
    )
    
    return app


def register_extensions(app):
    """Register Flask extensions."""
    db.init_app(app)
    migrate.init_app(app, db)
    login_manager.init_app(app)
    csrf.init_app(app)
    socketio.init_app(app, cors_allowed_origins="*")
    metrics.init_app(app)
    
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


def register_request_hooks(app):
    """Register request hooks for observability."""
    import time
    import uuid
    import structlog
    
    logger = structlog.get_logger(__name__)
    
    @app.before_request
    def before_request():
        """Log request start and set correlation ID."""
        g.request_id = str(uuid.uuid4())[:8]
        g.start_time = time.time()
        
        logger.info(
            "request_started",
            request_id=g.request_id,
            method=request.method,
            path=request.path,
            remote_addr=request.remote_addr,
        )
    
    @app.after_request
    def after_request(response):
        """Log request completion with timing."""
        duration_ms = (time.time() - g.start_time) * 1000
        
        logger.info(
            "request_completed",
            request_id=g.request_id,
            method=request.method,
            path=request.path,
            status_code=response.status_code,
            duration_ms=round(duration_ms, 2),
        )
        
        # Add request ID to response headers for debugging
        response.headers['X-Request-ID'] = g.request_id
        return response


def configure_logging(app):
    """Configure structured logging."""
    import structlog
    
    # Configure stdlib logging
    logging.basicConfig(
        format="%(message)s",
        level=logging.INFO if not app.debug else logging.DEBUG,
    )
    
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
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
    
    app.logger.info("Logging configured")


def init_sentry(environment: str):
    """Initialize Sentry error tracking."""
    sentry_dsn = os.environ.get('SENTRY_DSN')
    
    if not sentry_dsn:
        logging.info("Sentry DSN not configured, skipping Sentry initialization")
        return
    
    sentry_sdk.init(
        dsn=sentry_dsn,
        environment=os.environ.get('SENTRY_ENVIRONMENT', environment),
        integrations=[
            FlaskIntegration(transaction_style="url"),
            SqlalchemyIntegration(),
            RedisIntegration(),
        ],
        # Performance monitoring
        traces_sample_rate=0.1 if environment == 'production' else 1.0,
        # Profile 10% of sampled transactions
        profiles_sample_rate=0.1,
        # Send user info with errors
        send_default_pii=True,
        # Release tracking
        release=os.environ.get('APP_VERSION', '0.1.0'),
        # Attach request data
        request_bodies="medium",
        # Filter out health checks from transactions
        traces_sampler=traces_sampler,
    )
    logging.info(f"Sentry initialized for environment: {environment}")


def traces_sampler(sampling_context):
    """Custom sampler to filter out noisy transactions."""
    # Don't trace health checks or metrics
    transaction_context = sampling_context.get("transaction_context", {})
    name = transaction_context.get("name", "")
    
    if name in ["/health", "/health/detailed", "/metrics"]:
        return 0  # Don't sample these
    
    # Sample everything else based on environment
    return 1.0  # Will be overridden by traces_sample_rate in production
