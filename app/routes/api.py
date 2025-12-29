"""API routes for AJAX and external integrations."""
from flask import request, jsonify
from flask_login import login_required, current_user
from . import api_bp
from ..models import Vehicle, Auction
from ..extensions import db, csrf


@api_bp.route('/vehicles')
def list_vehicles():
    """API endpoint to list vehicles with filters."""
    page = request.args.get('page', 1, type=int)
    per_page = request.args.get('per_page', 20, type=int)
    
    query = Vehicle.query.filter_by(status='active')
    
    # Apply filters
    make = request.args.get('make')
    if make:
        query = query.filter(Vehicle.make == make)
    
    model = request.args.get('model')
    if model:
        query = query.filter(Vehicle.model == model)
    
    year_min = request.args.get('year_min', type=int)
    if year_min:
        query = query.filter(Vehicle.year >= year_min)
    
    year_max = request.args.get('year_max', type=int)
    if year_max:
        query = query.filter(Vehicle.year <= year_max)
    
    price_max = request.args.get('price_max', type=float)
    if price_max:
        query = query.filter(Vehicle.starting_price <= price_max)
    
    # Pagination
    pagination = query.order_by(Vehicle.created_at.desc()).paginate(
        page=page, per_page=per_page, error_out=False
    )
    
    vehicles = [{
        'id': v.id,
        'vin': v.vin,
        'year': v.year,
        'make': v.make,
        'model': v.model,
        'trim': v.trim,
        'mileage': v.mileage,
        'condition': v.condition,
        'title_type': v.title_type,
        'starting_price': float(v.starting_price) if v.starting_price else None,
        'buy_now_price': float(v.buy_now_price) if v.buy_now_price else None,
        'location_city': v.location_city,
        'location_state': v.location_state,
        'primary_image_url': v.primary_image.url if v.primary_image else None,
    } for v in pagination.items]
    
    return jsonify({
        'vehicles': vehicles,
        'total': pagination.total,
        'pages': pagination.pages,
        'current_page': pagination.page,
    })


@api_bp.route('/vehicles/<int:vehicle_id>')
def get_vehicle(vehicle_id):
    """API endpoint to get vehicle details."""
    vehicle = Vehicle.query.get_or_404(vehicle_id)
    
    return jsonify({
        'id': vehicle.id,
        'vin': vehicle.vin,
        'year': vehicle.year,
        'make': vehicle.make,
        'model': vehicle.model,
        'trim': vehicle.trim,
        'body_type': vehicle.body_type,
        'engine': vehicle.engine,
        'transmission': vehicle.transmission,
        'drivetrain': vehicle.drivetrain,
        'exterior_color': vehicle.exterior_color,
        'interior_color': vehicle.interior_color,
        'mileage': vehicle.mileage,
        'condition': vehicle.condition,
        'title_type': vehicle.title_type,
        'title_state': vehicle.title_state,
        'has_keys': vehicle.has_keys,
        'description': vehicle.description,
        'starting_price': float(vehicle.starting_price) if vehicle.starting_price else None,
        'reserve_price': float(vehicle.reserve_price) if vehicle.reserve_price else None,
        'buy_now_price': float(vehicle.buy_now_price) if vehicle.buy_now_price else None,
        'location': {
            'address': vehicle.location_address,
            'city': vehicle.location_city,
            'state': vehicle.location_state,
            'zip': vehicle.location_zip,
        },
        'images': [{
            'url': img.url,
            'is_primary': img.is_primary,
        } for img in vehicle.images.order_by('sort_order').all()],
        'auction': {
            'id': vehicle.auction.id,
            'status': vehicle.auction.status,
            'current_bid': float(vehicle.auction.current_bid) if vehicle.auction.current_bid else 0,
            'bid_count': vehicle.auction.bid_count,
            'ends_at': vehicle.auction.ends_at.isoformat() if vehicle.auction.ends_at else None,
            'time_remaining': vehicle.auction.time_remaining,
        } if vehicle.auction else None,
    })


@api_bp.route('/auctions/<int:auction_id>/bids')
def get_auction_bids(auction_id):
    """API endpoint to get bid history for an auction."""
    auction = Auction.query.get_or_404(auction_id)
    
    bids = [{
        'id': bid.id,
        'amount': float(bid.amount),
        'user_display': f'user***{str(bid.user_id)[-2:]}',
        'is_auto_bid': bid.is_auto_bid,
        'created_at': bid.created_at.isoformat(),
    } for bid in auction.bids.limit(50).all()]
    
    return jsonify({
        'auction_id': auction.id,
        'current_bid': float(auction.current_bid) if auction.current_bid else 0,
        'bid_count': auction.bid_count,
        'bids': bids,
    })


@api_bp.route('/decode-vin', methods=['POST'])
@login_required
def decode_vin():
    """API endpoint to decode a VIN via ClearVIN."""
    data = request.get_json()
    vin = data.get('vin')
    
    if not vin or len(vin) != 17:
        return jsonify({
            'success': False,
            'error': 'Invalid VIN. Must be 17 characters.'
        }), 400
    
    # TODO: Call ClearVIN API
    # For now, return mock data
    return jsonify({
        'success': True,
        'data': {
            'vin': vin,
            'year': 2021,
            'make': 'Toyota',
            'model': 'Camry',
            'trim': 'SE',
            'body_type': 'Sedan',
            'engine': '2.5L 4-Cylinder',
            'transmission': 'Automatic',
            'drivetrain': 'FWD',
        }
    })
