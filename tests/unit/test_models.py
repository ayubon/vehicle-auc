"""Unit tests for models."""
import pytest
from decimal import Decimal
from datetime import datetime, timedelta


class TestUserModel:
    """Tests for User model."""
    
    def test_user_creation(self, test_user):
        """Test user can be created."""
        assert test_user.id is not None
        assert test_user.email == 'testuser@example.com'
        assert test_user.first_name == 'Test'
        assert test_user.last_name == 'User'
    
    def test_user_full_name(self, test_user):
        """Test full_name property."""
        assert test_user.full_name == 'Test User'
    
    def test_user_full_name_fallback(self, db_session):
        """Test full_name falls back to email."""
        import uuid
        from app.models import User
        
        user = User(
            email='noname@example.com',
            password='hashedpassword',
            fs_uniquifier=str(uuid.uuid4()),
            active=True
        )
        db_session.add(user)
        db_session.commit()
        
        assert user.full_name == 'noname@example.com'
    
    def test_user_can_bid_unverified(self, test_user):
        """Test unverified user cannot bid."""
        assert test_user.can_bid is False
    
    def test_user_can_bid_verified(self, verified_user):
        """Test verified user can bid."""
        assert verified_user.can_bid is True


class TestVehicleModel:
    """Tests for Vehicle model."""
    
    def test_vehicle_creation(self, test_vehicle):
        """Test vehicle can be created."""
        assert test_vehicle.id is not None
        assert test_vehicle.vin == '1HGBH41JXMN109186'
        assert test_vehicle.year == 2021
        assert test_vehicle.make == 'Honda'
        assert test_vehicle.model == 'Accord'
    
    def test_vehicle_display_title(self, test_vehicle):
        """Test display_title property."""
        assert test_vehicle.display_title == '2021 Honda Accord'
    
    def test_vehicle_status_default(self, db_session, test_user):
        """Test vehicle status defaults to draft."""
        from app.models import Vehicle
        
        vehicle = Vehicle(
            seller_id=test_user.id,
            vin='2HGBH41JXMN109187',
            starting_price=Decimal('10000.00')
        )
        db_session.add(vehicle)
        db_session.commit()
        
        assert vehicle.status == 'draft'


class TestAuctionModel:
    """Tests for Auction model."""
    
    def test_auction_creation(self, test_auction):
        """Test auction can be created."""
        assert test_auction.id is not None
        assert test_auction.status == 'active'
        assert test_auction.auction_type == 'timed'
    
    def test_auction_is_active(self, test_auction):
        """Test is_active property."""
        assert test_auction.is_active is True
    
    def test_auction_time_remaining(self, test_auction):
        """Test time_remaining property."""
        # Should have roughly 23 hours remaining
        assert test_auction.time_remaining > 22 * 60 * 60
        assert test_auction.time_remaining < 24 * 60 * 60
    
    def test_auction_ended(self, db_session, test_vehicle):
        """Test ended auction is not active."""
        from app.models import Auction
        
        auction = Auction(
            vehicle_id=test_vehicle.id,
            auction_type='timed',
            status='active',
            starts_at=datetime.utcnow() - timedelta(hours=25),
            ends_at=datetime.utcnow() - timedelta(hours=1),
        )
        db_session.add(auction)
        db_session.commit()
        
        assert auction.is_active is False
        assert auction.time_remaining == 0


class TestBidModel:
    """Tests for Bid model."""
    
    def test_bid_creation(self, db_session, test_auction, verified_user):
        """Test bid can be created."""
        from app.models import Bid
        
        bid = Bid(
            auction_id=test_auction.id,
            user_id=verified_user.id,
            amount=Decimal('16000.00'),
            ip_address='127.0.0.1'
        )
        db_session.add(bid)
        db_session.commit()
        
        assert bid.id is not None
        assert bid.amount == Decimal('16000.00')
        assert bid.is_auto_bid is False


class TestOrderModel:
    """Tests for Order model."""
    
    def test_order_creation(self, db_session, test_auction, test_vehicle, verified_user, test_user):
        """Test order can be created."""
        from app.models import Order
        
        order = Order(
            order_number='ORD-2024-00001',
            auction_id=test_auction.id,
            buyer_id=verified_user.id,
            seller_id=test_user.id,
            vehicle_id=test_vehicle.id,
            vehicle_price=Decimal('16000.00'),
            buyer_fee=Decimal('800.00'),
            title_fee=Decimal('75.00'),
            total=Decimal('16875.00')
        )
        db_session.add(order)
        db_session.commit()
        
        assert order.id is not None
        assert order.status == 'pending_payment'
        assert order.total == Decimal('16875.00')
