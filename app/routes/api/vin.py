"""VIN decoding API endpoint."""
from flask import request, jsonify
from flask_jwt_extended import jwt_required
from . import api_bp
from ...services import clearvin_service, ClearVINError


@api_bp.route('/decode-vin', methods=['POST'])
@jwt_required()
def decode_vin():
    """Decode a VIN via ClearVIN."""
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
