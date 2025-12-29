"""API root endpoint."""
from flask import jsonify
from . import api_bp
from ...extensions import csrf

# Exempt API routes from CSRF (using JWT instead)
csrf.exempt(api_bp)


@api_bp.route('/', strict_slashes=False)
def api_index():
    """API root endpoint."""
    return jsonify({
        'name': 'Vehicle Auction API',
        'version': '0.1.0',
        'endpoints': {
            'vehicles': '/api/vehicles',
            'auctions': '/api/auctions',
            'auth': '/api/auth',
            'health': '/health',
        }
    })
