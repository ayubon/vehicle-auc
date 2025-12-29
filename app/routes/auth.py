"""Authentication API routes."""
import uuid
from datetime import datetime
from flask import request, jsonify
from flask_jwt_extended import (
    create_access_token, create_refresh_token, 
    jwt_required, current_user, get_jwt_identity
)
from werkzeug.security import generate_password_hash, check_password_hash
from . import auth_bp
from ..extensions import db, csrf
from ..models import User

# Exempt auth routes from CSRF (using JWT instead)
csrf.exempt(auth_bp)


@auth_bp.route('/register', methods=['POST'])
def register():
    """Register a new user."""
    data = request.get_json()
    
    # Validate required fields
    email = data.get('email', '').strip().lower()
    password = data.get('password', '')
    
    if not email or not password:
        return jsonify({'error': 'Email and password are required'}), 400
    
    if len(password) < 8:
        return jsonify({'error': 'Password must be at least 8 characters'}), 400
    
    # Check if user exists
    if User.query.filter_by(email=email).first():
        return jsonify({'error': 'Email already registered'}), 409
    
    # Create user
    user = User(
        email=email,
        password=generate_password_hash(password),
        fs_uniquifier=str(uuid.uuid4()),
        first_name=data.get('first_name', ''),
        last_name=data.get('last_name', ''),
        active=True,
    )
    
    db.session.add(user)
    db.session.commit()
    
    # Generate tokens
    access_token = create_access_token(identity=user)
    refresh_token = create_refresh_token(identity=user)
    
    return jsonify({
        'message': 'Registration successful',
        'access_token': access_token,
        'refresh_token': refresh_token,
        'user': {
            'id': user.id,
            'email': user.email,
            'first_name': user.first_name,
            'last_name': user.last_name,
        }
    }), 201


@auth_bp.route('/login', methods=['POST'])
def login():
    """Login and get JWT tokens."""
    data = request.get_json()
    
    email = data.get('email', '').strip().lower()
    password = data.get('password', '')
    
    if not email or not password:
        return jsonify({'error': 'Email and password are required'}), 400
    
    user = User.query.filter_by(email=email).first()
    
    if not user or not check_password_hash(user.password, password):
        return jsonify({'error': 'Invalid email or password'}), 401
    
    if not user.active:
        return jsonify({'error': 'Account is deactivated'}), 401
    
    # Update login tracking
    user.last_login_at = user.current_login_at
    user.current_login_at = datetime.utcnow()
    user.last_login_ip = user.current_login_ip
    user.current_login_ip = request.remote_addr
    user.login_count = (user.login_count or 0) + 1
    db.session.commit()
    
    # Generate tokens
    access_token = create_access_token(identity=user)
    refresh_token = create_refresh_token(identity=user)
    
    return jsonify({
        'access_token': access_token,
        'refresh_token': refresh_token,
        'user': {
            'id': user.id,
            'email': user.email,
            'first_name': user.first_name,
            'last_name': user.last_name,
            'is_id_verified': user.is_id_verified,
            'has_payment_method': user.has_payment_method,
            'can_bid': user.can_bid,
        }
    })


@auth_bp.route('/refresh', methods=['POST'])
@jwt_required(refresh=True)
def refresh():
    """Refresh access token."""
    identity = get_jwt_identity()
    access_token = create_access_token(identity=identity)
    return jsonify({'access_token': access_token})


@auth_bp.route('/me', methods=['GET'])
@jwt_required()
def me():
    """Get current user info."""
    return jsonify({
        'id': current_user.id,
        'email': current_user.email,
        'first_name': current_user.first_name,
        'last_name': current_user.last_name,
        'phone': current_user.phone,
        'address': {
            'line1': current_user.address_line1,
            'line2': current_user.address_line2,
            'city': current_user.city,
            'state': current_user.state,
            'zip_code': current_user.zip_code,
        },
        'is_id_verified': current_user.is_id_verified,
        'has_payment_method': current_user.has_payment_method,
        'can_bid': current_user.can_bid,
        'created_at': current_user.created_at.isoformat() if current_user.created_at else None,
    })


@auth_bp.route('/me', methods=['PUT'])
@jwt_required()
def update_profile():
    """Update current user profile."""
    data = request.get_json()
    
    # Update allowed fields
    allowed_fields = ['first_name', 'last_name', 'phone', 
                      'address_line1', 'address_line2', 'city', 'state', 'zip_code']
    
    for field in allowed_fields:
        if field in data:
            setattr(current_user, field, data[field])
    
    db.session.commit()
    
    return jsonify({
        'message': 'Profile updated',
        'user': {
            'id': current_user.id,
            'email': current_user.email,
            'first_name': current_user.first_name,
            'last_name': current_user.last_name,
        }
    })


@auth_bp.route('/change-password', methods=['POST'])
@jwt_required()
def change_password():
    """Change user password."""
    data = request.get_json()
    
    current_password = data.get('current_password', '')
    new_password = data.get('new_password', '')
    
    if not current_password or not new_password:
        return jsonify({'error': 'Current and new password are required'}), 400
    
    if not check_password_hash(current_user.password, current_password):
        return jsonify({'error': 'Current password is incorrect'}), 401
    
    if len(new_password) < 8:
        return jsonify({'error': 'New password must be at least 8 characters'}), 400
    
    current_user.password = generate_password_hash(new_password)
    db.session.commit()
    
    return jsonify({'message': 'Password changed successfully'})


@auth_bp.route('/clerk-sync', methods=['POST'])
def clerk_sync():
    """
    Sync Clerk user with local database and issue Flask JWT token.
    Called from frontend after Clerk sign-in.
    """
    data = request.get_json()
    
    clerk_user_id = data.get('clerk_user_id')
    email = data.get('email', '').strip().lower()
    first_name = data.get('first_name', '')
    last_name = data.get('last_name', '')
    
    if not clerk_user_id or not email:
        return jsonify({'error': 'clerk_user_id and email are required'}), 400
    
    # Find or create user
    user = User.query.filter_by(email=email).first()
    
    if not user:
        # Create new user from Clerk
        user = User(
            email=email,
            password=generate_password_hash(str(uuid.uuid4())),  # Random password (won't be used)
            fs_uniquifier=str(uuid.uuid4()),
            first_name=first_name,
            last_name=last_name,
            clerk_user_id=clerk_user_id,
            active=True,
        )
        db.session.add(user)
    else:
        # Update existing user with Clerk ID if not set
        if not user.clerk_user_id:
            user.clerk_user_id = clerk_user_id
        user.first_name = first_name or user.first_name
        user.last_name = last_name or user.last_name
    
    # Update login tracking
    user.last_login_at = user.current_login_at
    user.current_login_at = datetime.utcnow()
    user.login_count = (user.login_count or 0) + 1
    db.session.commit()
    
    # Generate Flask JWT tokens
    access_token = create_access_token(identity=user)
    refresh_token = create_refresh_token(identity=user)
    
    return jsonify({
        'access_token': access_token,
        'refresh_token': refresh_token,
        'user': {
            'id': user.id,
            'email': user.email,
            'first_name': user.first_name,
            'last_name': user.last_name,
        }
    })
