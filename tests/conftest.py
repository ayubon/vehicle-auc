"""Pytest configuration and fixtures."""
import os
import pytest
from flask import Flask
from app import create_app
from app.extensions import db
from app.models import User, Role, Vehicle, Auction


@pytest.fixture(scope='session')
def app():
    """Create application for testing."""
    os.environ['FLASK_ENV'] = 'testing'
    app = create_app('testing')
    
    with app.app_context():
        db.create_all()
        
        # Create default roles
        buyer_role = Role(name='buyer', description='Can browse and bid on vehicles')
        seller_role = Role(name='seller', description='Can list vehicles for sale')
        admin_role = Role(name='admin', description='Full administrative access')
        
        db.session.add_all([buyer_role, seller_role, admin_role])
        db.session.commit()
        
        yield app
        
        db.drop_all()


@pytest.fixture(scope='function')
def client(app):
    """Create test client."""
    return app.test_client()


@pytest.fixture(scope='function')
def db_session(app):
    """Create database session for testing - cleans up after each test."""
    with app.app_context():
        yield db.session
        
        # Clean up all test data after each test
        db.session.rollback()


@pytest.fixture
def test_user(db_session):
    """Create a test user."""
    import uuid
    unique_id = str(uuid.uuid4())[:8]
    user = User(
        email=f'testuser-{unique_id}@example.com',
        password='hashedpassword',
        first_name='Test',
        last_name='User',
        fs_uniquifier=str(uuid.uuid4()),
        active=True
    )
    db_session.add(user)
    db_session.commit()
    return user


@pytest.fixture
def verified_user(db_session):
    """Create a verified user who can bid."""
    import uuid
    from datetime import datetime
    
    unique_id = str(uuid.uuid4())[:8]
    user = User(
        email=f'verified-{unique_id}@example.com',
        password='hashedpassword',
        first_name='Verified',
        last_name='User',
        fs_uniquifier=str(uuid.uuid4()),
        active=True,
        id_verified_at=datetime.utcnow(),
        authorize_payment_profile_id='test-payment-profile'
    )
    db_session.add(user)
    db_session.commit()
    return user


@pytest.fixture
def test_vehicle(db_session, test_user):
    """Create a test vehicle."""
    import uuid
    from decimal import Decimal
    
    unique_id = str(uuid.uuid4())[:8].upper()
    vehicle = Vehicle(
        seller_id=test_user.id,
        vin=f'1HGBH41JX{unique_id}',  # Unique VIN per test
        year=2021,
        make='Honda',
        model='Accord',
        trim='Sport',
        mileage=25000,
        condition='runs_drives',
        title_type='clean',
        title_state='MN',
        starting_price=Decimal('15000.00'),
        location_city='Minneapolis',
        location_state='MN',
        location_zip='55401',
        status='active'
    )
    db_session.add(vehicle)
    db_session.commit()
    return vehicle


@pytest.fixture
def test_auction(db_session, test_vehicle):
    """Create a test auction."""
    from datetime import datetime, timedelta
    from decimal import Decimal
    
    auction = Auction(
        vehicle_id=test_vehicle.id,
        auction_type='timed',
        status='active',
        starts_at=datetime.utcnow() - timedelta(hours=1),
        ends_at=datetime.utcnow() + timedelta(hours=23),
        current_bid=Decimal('0'),
        bid_count=0
    )
    db_session.add(auction)
    db_session.commit()
    return auction


# Playwright fixtures for E2E testing
@pytest.fixture(scope='session')
def browser_context_args(browser_context_args):
    """Configure browser context for Playwright."""
    return {
        **browser_context_args,
        'viewport': {'width': 1280, 'height': 720},
        'record_video_dir': 'test-results/videos/',
    }


@pytest.fixture(scope='session')
def browser_type_launch_args(browser_type_launch_args):
    """Configure browser launch args for Playwright."""
    return {
        **browser_type_launch_args,
        'headless': True,
    }
