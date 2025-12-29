"""Miscellaneous models for notifications, saved searches, etc."""
from datetime import datetime
from ..extensions import db


class Notification(db.Model):
    """Notification model."""
    __tablename__ = 'notifications'
    
    id = db.Column(db.BigInteger, primary_key=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    type = db.Column(
        db.Enum(
            'bid_placed', 'outbid', 'auction_won', 'auction_lost',
            'payment_received', 'title_update', 'transport_update',
            'offer_received', 'offer_accepted', 'offer_rejected',
            'vehicle_approved', 'vehicle_rejected', 'general',
            name='notification_type'
        ),
        nullable=False
    )
    title = db.Column(db.String(255), nullable=False)
    message = db.Column(db.Text)
    data = db.Column(db.JSON)  # Contextual payload
    
    read_at = db.Column(db.DateTime)
    email_sent_at = db.Column(db.DateTime)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow, index=True)
    
    # Relationships
    user = db.relationship('User', back_populates='notifications')
    
    @property
    def is_read(self):
        """Check if notification has been read."""
        return self.read_at is not None
    
    def mark_as_read(self):
        """Mark notification as read."""
        if not self.read_at:
            self.read_at = datetime.utcnow()
    
    def __repr__(self):
        return f'<Notification {self.type} for User {self.user_id}>'


class SavedSearch(db.Model):
    """Saved search model for Vehicle Finder feature."""
    __tablename__ = 'saved_searches'
    
    id = db.Column(db.BigInteger, primary_key=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    name = db.Column(db.String(255), nullable=False)
    filters = db.Column(db.JSON, nullable=False)  # {make, model, year_min, year_max, price_max, etc.}
    notify_email = db.Column(db.Boolean, default=True)
    
    last_notified_at = db.Column(db.DateTime)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    user = db.relationship('User', back_populates='saved_searches')
    
    def __repr__(self):
        return f'<SavedSearch "{self.name}">'


class Watchlist(db.Model):
    """Watchlist model for tracking vehicles."""
    __tablename__ = 'watchlist'
    
    id = db.Column(db.BigInteger, primary_key=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), nullable=False, index=True)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    # Relationships
    user = db.relationship('User', back_populates='watchlist')
    vehicle = db.relationship('Vehicle', back_populates='watchlist_entries')
    
    __table_args__ = (
        db.UniqueConstraint('user_id', 'vehicle_id', name='uq_watchlist_user_vehicle'),
    )
    
    def __repr__(self):
        return f'<Watchlist User {self.user_id} -> Vehicle {self.vehicle_id}>'


class AuditLog(db.Model):
    """Audit log for tracking changes."""
    __tablename__ = 'audit_log'
    
    id = db.Column(db.BigInteger, primary_key=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), index=True)
    
    action = db.Column(db.String(100), nullable=False, index=True)  # e.g., 'bid.placed', 'payment.captured'
    entity_type = db.Column(db.String(100), nullable=False)  # e.g., 'Bid', 'Payment'
    entity_id = db.Column(db.BigInteger)
    
    old_data = db.Column(db.JSON)
    new_data = db.Column(db.JSON)
    
    ip_address = db.Column(db.String(45))
    user_agent = db.Column(db.String(500))
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow, index=True)
    
    # Relationships
    user = db.relationship('User')
    
    def __repr__(self):
        return f'<AuditLog {self.action} on {self.entity_type}>'
