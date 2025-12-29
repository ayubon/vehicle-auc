"""Flask blueprints."""
from flask import Blueprint

# Main blueprint for homepage and static pages
main_bp = Blueprint('main', __name__)

# Auth blueprint
auth_bp = Blueprint('auth', __name__)

# Vehicles blueprint
vehicles_bp = Blueprint('vehicles', __name__)

# Auctions blueprint
auctions_bp = Blueprint('auctions', __name__)

# Orders blueprint
orders_bp = Blueprint('orders', __name__)

# API blueprint
api_bp = Blueprint('api', __name__)

# Import routes to register them
from . import main, auth, vehicles, auctions, orders, api  # noqa: F401, E402
