"""Vehicle image management API endpoints."""
from flask import request, jsonify
from flask_jwt_extended import jwt_required, current_user
from . import api_bp
from ...models import Vehicle, VehicleImage
from ...extensions import db
from ...services import s3_service


@api_bp.route('/vehicles/<int:vehicle_id>/upload-url', methods=['POST'])
@jwt_required()
def get_upload_url(vehicle_id):
    """Get a presigned URL for uploading an image."""
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
    vehicle = db.session.get(Vehicle, vehicle_id)
    if not vehicle:
        return jsonify({'error': 'Vehicle not found'}), 404
    if vehicle.seller_id != current_user.id:
        return jsonify({'error': 'Not authorized'}), 403
    
    image = db.session.get(VehicleImage, image_id)
    if not image or image.vehicle_id != vehicle_id:
        return jsonify({'error': 'Image not found'}), 404
    
    s3_service.delete_file(image.s3_key)
    db.session.delete(image)
    db.session.commit()
    
    return jsonify({'message': 'Image deleted'})
