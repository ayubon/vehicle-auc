"""Main routes for homepage and static pages."""
import structlog
from flask import render_template, jsonify
from . import main_bp
from ..extensions import db, redis_client

logger = structlog.get_logger(__name__)


@main_bp.route('/')
def index():
    """Homepage with vehicle listings."""
    logger.info("homepage_accessed")
    return render_template('index.html')


@main_bp.route('/how-it-works')
def how_it_works():
    """How it works page."""
    return render_template('how_it_works.html')


@main_bp.route('/pricing')
def pricing():
    """Buyer premium fees page."""
    return render_template('pricing.html')


@main_bp.route('/contact')
def contact():
    """Contact us page."""
    return render_template('contact.html')


@main_bp.route('/health')
def health():
    """Basic health check endpoint."""
    return {'status': 'healthy'}


@main_bp.route('/health/detailed')
def health_detailed():
    """Detailed health check with all service statuses."""
    health_status = {
        'status': 'healthy',
        'services': {}
    }
    
    # Check MySQL
    try:
        db.session.execute(db.text('SELECT 1'))
        health_status['services']['mysql'] = {'status': 'healthy'}
    except Exception as e:
        health_status['services']['mysql'] = {'status': 'unhealthy', 'error': str(e)}
        health_status['status'] = 'degraded'
    
    # Check Redis
    try:
        redis_client.ping()
        health_status['services']['redis'] = {'status': 'healthy'}
    except Exception as e:
        health_status['services']['redis'] = {'status': 'unhealthy', 'error': str(e)}
        health_status['status'] = 'degraded'
    
    logger.info("health_check", **health_status)
    
    status_code = 200 if health_status['status'] == 'healthy' else 503
    return jsonify(health_status), status_code


@main_bp.route('/debug/redis')
def debug_redis():
    """Debug endpoint to test Redis operations."""
    try:
        # Test basic operations
        redis_client.set('test_key', 'test_value', ex=60)
        value = redis_client.get('test_key')
        redis_client.delete('test_key')
        
        # Get Redis info
        info = redis_client.info()
        
        return jsonify({
            'status': 'connected',
            'test_write_read': value == 'test_value',
            'redis_version': info.get('redis_version'),
            'connected_clients': info.get('connected_clients'),
            'used_memory_human': info.get('used_memory_human'),
        })
    except Exception as e:
        logger.error("redis_error", error=str(e))
        return jsonify({'status': 'error', 'error': str(e)}), 500


@main_bp.route('/debug/log-test')
def debug_log_test():
    """Debug endpoint to test structured logging."""
    logger.debug("debug_message", extra_data="debug level")
    logger.info("info_message", user_action="test", extra_data="info level")
    logger.warning("warning_message", potential_issue="test warning")
    logger.error("error_message", error_type="test_error")
    
    return jsonify({
        'status': 'ok',
        'message': 'Check server logs for structured JSON output'
    })


@main_bp.route('/debug/sentry-test')
def debug_sentry_test():
    """Debug endpoint to test Sentry error tracking."""
    import sentry_sdk
    
    # Capture a test message
    sentry_sdk.capture_message("Test message from /debug/sentry-test")
    
    # Optionally trigger a real error (commented out by default)
    # raise ValueError("Test error from /debug/sentry-test")
    
    return jsonify({
        'status': 'ok',
        'message': 'Test message sent to Sentry. Check your Sentry dashboard.',
        'sentry_configured': bool(sentry_sdk.Hub.current.client)
    })


@main_bp.route('/debug/metrics-test')
def debug_metrics_test():
    """Debug endpoint to test custom Prometheus metrics."""
    from ..instrumentation import (
        track_bid, track_user_registration, track_order,
        track_redis_operation, set_active_auctions
    )
    
    # Simulate some business events
    track_bid(auction_type='timed', amount=15000, result='accepted')
    track_bid(auction_type='timed', amount=14500, result='outbid')
    track_user_registration(role='buyer')
    track_order(status='created', value=16500)
    set_active_auctions(42)
    
    # Track a Redis operation
    with track_redis_operation('set'):
        redis_client.set('metrics_test', 'value', ex=60)
    
    with track_redis_operation('get'):
        redis_client.get('metrics_test')
    
    return jsonify({
        'status': 'ok',
        'message': 'Custom metrics recorded. Check /metrics endpoint.',
        'metrics_fired': [
            'auction_bids_total',
            'auction_bid_amount_dollars',
            'user_registrations_total',
            'orders_total',
            'order_value_dollars',
            'auction_active_count',
            'redis_operation_duration_seconds',
            'redis_operations_total'
        ]
    })
