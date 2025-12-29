"""Vehicle CRUD API endpoints.

SICP: Routes are thin - they delegate serialization to models.
This keeps routes focused on HTTP concerns (request/response, auth, validation).
"""
from decimal import Decimal
from flask import request, jsonify
from flask_jwt_extended import jwt_required, current_user
from . import api_bp
from ...models import Vehicle
from ...extensions import db
from ...constants import VehicleStatus


@api_bp.route('/vehicles')
def list_vehicles():
    """List vehicles with filters."""
    page = request.args.get('page', 1, type=int)
    per_page = request.args.get('per_page', 20, type=int)
    
    query = Vehicle.query.filter_by(status=VehicleStatus.ACTIVE.value)
    
    # Apply filters
    if make := request.args.get('make'):
        query = query.filter(Vehicle.make == make)
    if model := request.args.get('model'):
        query = query.filter(Vehicle.model == model)
    if year_min := request.args.get('year_min', type=int):
        query = query.filter(Vehicle.year >= year_min)
    if year_max := request.args.get('year_max', type=int):
        query = query.filter(Vehicle.year <= year_max)
    if price_max := request.args.get('price_max', type=float):
        query = query.filter(Vehicle.starting_price <= price_max)
    
    pagination = query.order_by(Vehicle.created_at.desc()).paginate(
        page=page, per_page=per_page, error_out=False
    )
    
    return jsonify({
        'vehicles': [v.to_summary_dict() for v in pagination.items],
        'total': pagination.total,
        'pages': pagination.pages,
        'current_page': pagination.page,
    })


@api_bp.route('/vehicles/<int:vehicle_id>')
def get_vehicle(vehicle_id):
    """Get vehicle details."""
    vehicle = Vehicle.query.get_or_404(vehicle_id)
    return jsonify(vehicle.to_detail_dict())


@api_bp.route('/vehicles', methods=['POST'])
@jwt_required()
def create_vehicle():
    """Create a new vehicle listing."""
    data = request.get_json()
    
    vin = data.get('vin', '').strip().upper()
    if not vin or len(vin) != 17:
        return jsonify({'error': 'Valid 17-character VIN is required'}), 400
    
    starting_price = data.get('starting_price')
    if not starting_price or float(starting_price) <= 0:
        return jsonify({'error': 'Starting price is required'}), 400
    
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
        status=VehicleStatus.DRAFT.value,
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
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized to edit this vehicle'}), 403
    if vehicle.status == VehicleStatus.SOLD.value:
        return jsonify({'error': 'Cannot edit sold vehicles'}), 400
    
    data = request.get_json()
    
    updatable_fields = [
        'year', 'make', 'model', 'trim', 'body_type', 'engine', 'transmission',
        'drivetrain', 'exterior_color', 'interior_color', 'mileage', 'condition',
        'title_type', 'title_state', 'has_keys', 'description',
        'location_address', 'location_city', 'location_state', 'location_zip',
    ]
    
    for field in updatable_fields:
        if field in data:
            setattr(vehicle, field, data[field])
    
    if 'starting_price' in data:
        vehicle.starting_price = Decimal(str(data['starting_price']))
    if 'reserve_price' in data:
        vehicle.reserve_price = Decimal(str(data['reserve_price'])) if data['reserve_price'] else None
    if 'buy_now_price' in data:
        vehicle.buy_now_price = Decimal(str(data['buy_now_price'])) if data['buy_now_price'] else None
    
    db.session.commit()
    
    return jsonify({'message': 'Vehicle updated', 'vehicle_id': vehicle.id})


@api_bp.route('/vehicles/<int:vehicle_id>', methods=['DELETE'])
@jwt_required()
def delete_vehicle(vehicle_id):
    """Delete a vehicle listing."""
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized to delete this vehicle'}), 403
    if vehicle.status == VehicleStatus.SOLD.value:
        return jsonify({'error': 'Cannot delete sold vehicles'}), 400
    from ...constants import AuctionStatus
    if vehicle.auction and vehicle.auction.status == AuctionStatus.ACTIVE.value:
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
    if vehicle.status != VehicleStatus.DRAFT.value:
        return jsonify({'error': 'Only draft vehicles can be submitted'}), 400
    if not all([vehicle.year, vehicle.make, vehicle.model, vehicle.starting_price]):
        return jsonify({'error': 'Missing required fields (year, make, model, starting_price)'}), 400
    
    vehicle.status = VehicleStatus.PENDING_REVIEW.value
    db.session.commit()
    
    return jsonify({'message': 'Vehicle submitted for review', 'status': vehicle.status})
