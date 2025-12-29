"""Instrumentation helpers for metrics and tracing."""
import time
import functools
from contextlib import contextmanager
from . import custom_metrics as m


@contextmanager
def track_db_query(query_type: str, table: str):
    """Context manager to track database query duration.
    
    Usage:
        with track_db_query('select', 'users'):
            result = db.session.execute(query)
    """
    start = time.perf_counter()
    try:
        yield
    finally:
        duration = time.perf_counter() - start
        m.db_query_duration.labels(query_type=query_type, table=table).observe(duration)


@contextmanager
def track_redis_operation(operation: str):
    """Context manager to track Redis operation duration.
    
    Usage:
        with track_redis_operation('get'):
            value = redis_client.get(key)
    """
    start = time.perf_counter()
    status = 'success'
    try:
        yield
    except Exception:
        status = 'error'
        raise
    finally:
        duration = time.perf_counter() - start
        m.redis_operation_duration.labels(operation=operation).observe(duration)
        m.redis_operations_total.labels(operation=operation, status=status).inc()


@contextmanager
def track_external_api(service: str, endpoint: str):
    """Context manager to track external API call duration.
    
    Usage:
        with track_external_api('clearvin', 'decode'):
            response = httpx.get(url)
    """
    start = time.perf_counter()
    status = 'success'
    try:
        yield
    except Exception as e:
        status = 'error'
        m.external_api_errors_total.labels(service=service, error_type=type(e).__name__).inc()
        raise
    finally:
        duration = time.perf_counter() - start
        m.external_api_duration.labels(service=service, endpoint=endpoint, status=status).observe(duration)


def track_bid(auction_type: str, amount: float, result: str):
    """Record a bid metric.
    
    Args:
        auction_type: 'timed', 'live', etc.
        amount: Bid amount in dollars
        result: 'accepted', 'rejected', 'outbid'
    """
    m.bids_total.labels(auction_type=auction_type, result=result).inc()
    if result == 'accepted':
        m.bid_amount.labels(auction_type=auction_type).observe(amount)


def track_user_registration(role: str):
    """Record a user registration."""
    m.user_registrations_total.labels(role=role).inc()


def track_user_verification(verification_type: str, status: str):
    """Record a user verification attempt.
    
    Args:
        verification_type: 'id' or 'payment'
        status: 'success' or 'failed'
    """
    m.user_verifications_total.labels(type=verification_type, status=status).inc()


def track_order(status: str, value: float = None):
    """Record an order status change.
    
    Args:
        status: 'created', 'paid', 'completed', 'cancelled', 'refunded'
        value: Order value in dollars (only for 'created')
    """
    m.orders_total.labels(status=status).inc()
    if value is not None and status == 'created':
        m.order_value.observe(value)


def track_payment(payment_type: str, status: str, duration: float = None):
    """Record a payment attempt.
    
    Args:
        payment_type: 'card' or 'ach'
        status: 'success' or 'failed'
        duration: Processing duration in seconds
    """
    m.payments_total.labels(type=payment_type, status=status).inc()
    if duration is not None:
        m.payment_processing_duration.labels(provider='authorize_net').observe(duration)


def track_vehicle(status: str):
    """Record a vehicle status change."""
    m.vehicles_total.labels(status=status).inc()


def set_active_auctions(count: int):
    """Set the current number of active auctions."""
    m.active_auctions.set(count)


def set_active_users(count: int):
    """Set the current number of active users."""
    m.active_users.set(count)


def set_websocket_connections(room: str, count: int):
    """Set the current number of WebSocket connections for a room."""
    m.websocket_connections.labels(room=room).set(count)


def increment_websocket_connections(room: str):
    """Increment WebSocket connections for a room."""
    m.websocket_connections.labels(room=room).inc()


def decrement_websocket_connections(room: str):
    """Decrement WebSocket connections for a room."""
    m.websocket_connections.labels(room=room).dec()
