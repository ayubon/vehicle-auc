"""Custom Prometheus metrics for the vehicle auction platform."""
from prometheus_client import Counter, Histogram, Gauge, Info

# =============================================================================
# HTTP / Request Metrics (handled by prometheus-flask-exporter automatically)
# =============================================================================

# =============================================================================
# Database Metrics
# =============================================================================
db_query_duration = Histogram(
    'db_query_duration_seconds',
    'Database query duration in seconds',
    ['query_type', 'table'],
    buckets=[0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5]
)

db_connection_pool = Gauge(
    'db_connection_pool_size',
    'Number of connections in the database pool',
    ['state']  # active, idle
)

# =============================================================================
# Redis Metrics
# =============================================================================
redis_operation_duration = Histogram(
    'redis_operation_duration_seconds',
    'Redis operation duration in seconds',
    ['operation'],  # get, set, delete, publish, etc.
    buckets=[0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1]
)

redis_operations_total = Counter(
    'redis_operations_total',
    'Total number of Redis operations',
    ['operation', 'status']  # status: success, error
)

# =============================================================================
# Business Metrics - Auctions
# =============================================================================
bids_total = Counter(
    'auction_bids_total',
    'Total number of bids placed',
    ['auction_type', 'result']  # result: accepted, rejected, outbid
)

bid_amount = Histogram(
    'auction_bid_amount_dollars',
    'Bid amounts in dollars',
    ['auction_type'],
    buckets=[100, 500, 1000, 2500, 5000, 10000, 25000, 50000, 100000, 250000]
)

active_auctions = Gauge(
    'auction_active_count',
    'Number of currently active auctions'
)

auctions_total = Counter(
    'auctions_total',
    'Total number of auctions',
    ['status']  # created, started, ended, cancelled
)

websocket_connections = Gauge(
    'websocket_connections_active',
    'Number of active WebSocket connections',
    ['room']  # auction room ID or 'global'
)

# =============================================================================
# Business Metrics - Users
# =============================================================================
user_registrations_total = Counter(
    'user_registrations_total',
    'Total number of user registrations',
    ['role']  # buyer, seller
)

user_verifications_total = Counter(
    'user_verifications_total',
    'Total number of user verifications',
    ['type', 'status']  # type: id, payment; status: success, failed
)

active_users = Gauge(
    'users_active_count',
    'Number of users active in the last 15 minutes'
)

# =============================================================================
# Business Metrics - Orders & Payments
# =============================================================================
orders_total = Counter(
    'orders_total',
    'Total number of orders',
    ['status']  # created, paid, completed, cancelled, refunded
)

order_value = Histogram(
    'order_value_dollars',
    'Order values in dollars',
    buckets=[1000, 2500, 5000, 10000, 25000, 50000, 100000, 250000, 500000]
)

payments_total = Counter(
    'payments_total',
    'Total number of payment attempts',
    ['type', 'status']  # type: card, ach; status: success, failed
)

payment_processing_duration = Histogram(
    'payment_processing_duration_seconds',
    'Payment processing duration in seconds',
    ['provider'],  # authorize_net
    buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0]
)

# =============================================================================
# Business Metrics - Vehicles
# =============================================================================
vehicles_total = Counter(
    'vehicles_total',
    'Total number of vehicles',
    ['status']  # listed, sold, archived
)

vin_decode_duration = Histogram(
    'vin_decode_duration_seconds',
    'VIN decode API call duration',
    ['provider', 'status'],  # provider: clearvin; status: success, error
    buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0]
)

# =============================================================================
# External API Metrics
# =============================================================================
external_api_duration = Histogram(
    'external_api_duration_seconds',
    'External API call duration in seconds',
    ['service', 'endpoint', 'status'],
    buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0]
)

external_api_errors_total = Counter(
    'external_api_errors_total',
    'Total number of external API errors',
    ['service', 'error_type']
)

# =============================================================================
# Application Info
# =============================================================================
app_info = Info(
    'vehicle_auction_app',
    'Vehicle Auction application information'
)


def init_app_info(version='0.1.0', environment='development'):
    """Initialize application info metric."""
    app_info.info({
        'version': version,
        'environment': environment,
    })
