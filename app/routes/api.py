"""API routes for AJAX and external integrations."""
from flask import request, jsonify
from flask_jwt_extended import jwt_required, current_user
from . import api_bp
from ..models import Vehicle, Auction
from ..extensions import db, csrf

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


@api_bp.route('/vehicles', methods=['POST'])
@jwt_required()
def create_vehicle():
    """Create a new vehicle listing."""
    from decimal import Decimal
    
    data = request.get_json()
    
    # Validate required fields
    vin = data.get('vin', '').strip().upper()
    if not vin or len(vin) != 17:
        return jsonify({'error': 'Valid 17-character VIN is required'}), 400
    
    starting_price = data.get('starting_price')
    if not starting_price or float(starting_price) <= 0:
        return jsonify({'error': 'Starting price is required'}), 400
    
    # Check for duplicate VIN
    if Vehicle.query.filter_by(vin=vin).first():
        return jsonify({'error': 'Vehicle with this VIN already exists'}), 409
    
    vehicle = Vehicle(
        seller_id=current_user.id,
        vin=vin,
        year=data.get('year'),
        make=data.get('make'),
        model=data.get('model'),
        trim=data.get('trim'),
        body_type=data.get('body_type'),
        engine=data.get('engine'),
        transmission=data.get('transmission'),
        drivetrain=data.get('drivetrain'),
        exterior_color=data.get('exterior_color'),
        interior_color=data.get('interior_color'),
        mileage=data.get('mileage'),
        condition=data.get('condition', 'runs_drives'),
        title_type=data.get('title_type', 'clean'),
        title_state=data.get('title_state'),
        has_keys=data.get('has_keys', True),
        description=data.get('description'),
        starting_price=Decimal(str(starting_price)),
        reserve_price=Decimal(str(data['reserve_price'])) if data.get('reserve_price') else None,
        buy_now_price=Decimal(str(data['buy_now_price'])) if data.get('buy_now_price') else None,
        location_address=data.get('location_address'),
        location_city=data.get('location_city'),
        location_state=data.get('location_state'),
        location_zip=data.get('location_zip'),
        status='draft',
    )
    
    db.session.add(vehicle)
    db.session.commit()
    
    return jsonify({
        'message': 'Vehicle created',
        'vehicle_id': vehicle.id,
        'vin': vehicle.vin,
    }), 201


@api_bp.route('/vehicles/<int:vehicle_id>', methods=['PUT'])
@jwt_required()
def update_vehicle(vehicle_id):
    """Update a vehicle listing."""
    from decimal import Decimal
    
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    # Check ownership
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized to edit this vehicle'}), 403
    
    # Can't edit sold vehicles
    if vehicle.status == 'sold':
        return jsonify({'error': 'Cannot edit sold vehicles'}), 400
    
    data = request.get_json()
    
    # Update allowed fields
    updatable_fields = [
        'year', 'make', 'model', 'trim', 'body_type', 'engine', 'transmission',
        'drivetrain', 'exterior_color', 'interior_color', 'mileage', 'condition',
        'title_type', 'title_state', 'has_keys', 'description',
        'location_address', 'location_city', 'location_state', 'location_zip',
    ]
    
    for field in updatable_fields:
        if field in data:
            setattr(vehicle, field, data[field])
    
    # Handle price fields separately (need Decimal conversion)
    if 'starting_price' in data:
        vehicle.starting_price = Decimal(str(data['starting_price']))
    if 'reserve_price' in data:
        vehicle.reserve_price = Decimal(str(data['reserve_price'])) if data['reserve_price'] else None
    if 'buy_now_price' in data:
        vehicle.buy_now_price = Decimal(str(data['buy_now_price'])) if data['buy_now_price'] else None
    
    db.session.commit()
    
    return jsonify({
        'message': 'Vehicle updated',
        'vehicle_id': vehicle.id,
    })


@api_bp.route('/vehicles/<int:vehicle_id>', methods=['DELETE'])
@jwt_required()
def delete_vehicle(vehicle_id):
    """Delete a vehicle listing."""
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    # Check ownership
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized to delete this vehicle'}), 403
    
    # Can't delete sold vehicles or those with active auctions
    if vehicle.status == 'sold':
        return jsonify({'error': 'Cannot delete sold vehicles'}), 400
    
    if vehicle.auction and vehicle.auction.status == 'active':
        return jsonify({'error': 'Cannot delete vehicle with active auction'}), 400
    
    db.session.delete(vehicle)
    db.session.commit()
    
    return jsonify({'message': 'Vehicle deleted'})


@api_bp.route('/vehicles/<int:vehicle_id>/submit', methods=['POST'])
@jwt_required()
def submit_vehicle(vehicle_id):
    """Submit a vehicle for review."""
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    if vehicle.status != 'draft':
        return jsonify({'error': 'Only draft vehicles can be submitted'}), 400
    
    # Validate required fields before submission
    if not all([vehicle.year, vehicle.make, vehicle.model, vehicle.starting_price]):
        return jsonify({'error': 'Missing required fields (year, make, model, starting_price)'}), 400
    
    vehicle.status = 'pending_review'
    db.session.commit()
    
    return jsonify({
        'message': 'Vehicle submitted for review',
        'status': vehicle.status,
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


@api_bp.route('/vehicles/<int:vehicle_id>/upload-url', methods=['POST'])
@jwt_required()
def get_upload_url(vehicle_id):
    """Get a presigned URL for uploading an image."""
    from ..services import s3_service
    from ..models import VehicleImage
    
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    data = request.get_json()
    filename = data.get('filename', 'image.jpg')
    content_type = data.get('content_type', 'image/jpeg')
    
    result = s3_service.generate_upload_url(
        vehicle_id=vehicle_id,
        filename=filename,
        content_type=content_type,
    )
    
    return jsonify(result)


@api_bp.route('/vehicles/<int:vehicle_id>/images', methods=['POST'])
@jwt_required()
def add_vehicle_image(vehicle_id):
    """Register an uploaded image with a vehicle."""
    from ..models import VehicleImage
    
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    data = request.get_json()
    s3_key = data.get('s3_key')
    url = data.get('url')
    is_primary = data.get('is_primary', False)
    
    if not s3_key or not url:
        return jsonify({'error': 's3_key and url are required'}), 400
    
    # If this is primary, unset other primary images
    if is_primary:
        VehicleImage.query.filter_by(vehicle_id=vehicle_id, is_primary=True).update({'is_primary': False})
    
    # Get next sort order
    max_order = db.session.query(db.func.max(VehicleImage.sort_order)).filter_by(vehicle_id=vehicle_id).scalar() or 0
    
    image = VehicleImage(
        vehicle_id=vehicle_id,
        s3_key=s3_key,
        url=url,
        is_primary=is_primary,
        sort_order=max_order + 1,
    )
    db.session.add(image)
    db.session.commit()
    
    return jsonify({
        'message': 'Image added',
        'image_id': image.id,
        'is_primary': image.is_primary,
    }), 201


@api_bp.route('/vehicles/<int:vehicle_id>/images/<int:image_id>', methods=['DELETE'])
@jwt_required()
def delete_vehicle_image(vehicle_id, image_id):
    """Delete a vehicle image."""
    from ..services import s3_service
    from ..models import VehicleImage
    
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    image = db.session.get(VehicleImage, image_id)
    if not image or image.vehicle_id != vehicle_id:
        return jsonify({'error': 'Image not found'}), 404
    
    # Delete from S3
    s3_service.delete_file(image.s3_key)
    
    # Delete from database
    db.session.delete(image)
    db.session.commit()
    
    return jsonify({'message': 'Image deleted'})


@api_bp.route('/decode-vin', methods=['POST'])
@jwt_required()
def decode_vin():
    """API endpoint to decode a VIN via ClearVIN."""
    from ..services import clearvin_service, ClearVINError
    
    data = request.get_json()
    vin = data.get('vin', '').strip()
    
    if not vin or len(vin) != 17:
        return jsonify({
            'success': False,
            'error': 'Invalid VIN. Must be 17 characters.'
        }), 400
    
    try:
        vehicle_data = clearvin_service.decode_vin(vin)
        return jsonify({
            'success': True,
            'data': vehicle_data
        })
    except ClearVINError as e:
        return jsonify({
            'success': False,
            'error': str(e)
        }), 400
