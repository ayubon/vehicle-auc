"""Auction and bidding models.

SICP: Data abstraction - Auction knows its state and how to represent itself.
"""
from datetime import datetime
from ..extensions import db
from ..constants import AuctionStatus


class Auction(db.Model):
    """Auction model."""
    __tablename__ = 'auctions'
    
    id = db.Column(db.BigInteger, primary_key=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), unique=True, nullable=False)
    
    auction_type = db.Column(
        db.Enum('timed', 'live', 'buy_now_only', 'make_offer', name='auction_type'),
        default='timed',
        nullable=False
    )
    status = db.Column(
        db.Enum('scheduled', 'active', 'ended', 'cancelled', name='auction_status'),
        default='scheduled',
        nullable=False,
        index=True
    )
    
    # Timing
    starts_at = db.Column(db.DateTime, nullable=False)
    ends_at = db.Column(db.DateTime, nullable=False, index=True)
    extended_count = db.Column(db.SmallInteger, default=0)
    
    # Current State
    current_bid = db.Column(db.Numeric(10, 2), default=0)
    bid_count = db.Column(db.Integer, default=0)
    winner_id = db.Column(db.BigInteger, db.ForeignKey('users.id'))
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    vehicle = db.relationship('Vehicle', back_populates='auction')
    winner = db.relationship('User', foreign_keys=[winner_id])
    bids = db.relationship('Bid', back_populates='auction', lazy='dynamic', order_by='Bid.amount.desc()')
    
    @property
    def is_active(self):
        """Check if auction is currently active."""
        now = datetime.utcnow()
        return self.status == AuctionStatus.ACTIVE.value and self.starts_at <= now <= self.ends_at
    
    @property
    def time_remaining(self):
        """Get time remaining in seconds."""
        if not self.is_active:
            return 0
        delta = self.ends_at - datetime.utcnow()
        return max(0, int(delta.total_seconds()))
    
    @property
    def highest_bid(self):
        """Get the highest bid."""
        return self.bids.first()
    
    # ─────────────────────────────────────────────────────────────────────────
    # SICP: Message passing - ask the object for its representation
    # ─────────────────────────────────────────────────────────────────────────
    
    def to_summary_dict(self) -> dict:
        """Serialize for embedding in vehicle detail."""
        return {
            'id': self.id,
            'status': self.status,
            'current_bid': float(self.current_bid) if self.current_bid else 0,
            'bid_count': self.bid_count,
            'ends_at': self.ends_at.isoformat() if self.ends_at else None,
            'time_remaining': self.time_remaining,
        }
    
    def to_detail_dict(self) -> dict:
        """Serialize for auction detail view."""
        return {
            **self.to_summary_dict(),
            'auction_type': self.auction_type,
            'starts_at': self.starts_at.isoformat() if self.starts_at else None,
            'extended_count': self.extended_count,
            'vehicle': self.vehicle.to_summary_dict() if self.vehicle else None,
        }
    
    def __repr__(self):
        return f'<Auction {self.id} for Vehicle {self.vehicle_id}>'


class Bid(db.Model):
    """Bid model."""
    __tablename__ = 'bids'
    
    id = db.Column(db.BigInteger, primary_key=True)
    auction_id = db.Column(db.BigInteger, db.ForeignKey('auctions.id'), nullable=False, index=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    max_bid = db.Column(db.Numeric(10, 2))  # For auto-bidding
    is_auto_bid = db.Column(db.Boolean, default=False)
    
    ip_address = db.Column(db.String(45))
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow, index=True)
    
    # Relationships
    auction = db.relationship('Auction', back_populates='bids')
    user = db.relationship('User', back_populates='bids')
    
    __table_args__ = (
        db.Index('ix_bids_auction_amount', 'auction_id', 'amount'),
    )
    
    def to_dict(self) -> dict:
        """Serialize bid."""
        return {
            'id': self.id,
            'amount': float(self.amount),
            'user_display': f'user***{str(self.user_id)[-2:]}',
            'is_auto_bid': self.is_auto_bid,
            'created_at': self.created_at.isoformat(),
        }
    
    def __repr__(self):
        return f'<Bid ${self.amount} by User {self.user_id}>'


class Offer(db.Model):
    """Offer model for 'Make an Offer' flow."""
    __tablename__ = 'offers'
    
    id = db.Column(db.BigInteger, primary_key=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), nullable=False, index=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    status = db.Column(
        db.Enum('pending', 'accepted', 'rejected', 'expired', 'countered', name='offer_status'),
        default='pending',
        nullable=False
    )
    counter_amount = db.Column(db.Numeric(10, 2))
    message = db.Column(db.Text)
    
    expires_at = db.Column(db.DateTime, nullable=False)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    vehicle = db.relationship('Vehicle', back_populates='offers')
    user = db.relationship('User', back_populates='offers')
    
    @property
    def is_expired(self):
        """Check if offer has expired."""
        return datetime.utcnow() > self.expires_at and self.status == 'pending'
    
    def __repr__(self):
        return f'<Offer ${self.amount} for Vehicle {self.vehicle_id}>'
