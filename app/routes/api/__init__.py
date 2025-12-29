"""API routes package - domain-organized endpoints."""
from flask import Blueprint

api_bp = Blueprint('api', __name__, url_prefix='/api')

# Import and register sub-modules (they attach routes to api_bp)
from . import root, vehicles, images, auctions, vin
