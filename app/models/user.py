"""User and authentication models."""
from datetime import datetime
from flask_security import UserMixin, RoleMixin
from ..extensions import db

# Association table for users and roles
roles_users = db.Table(
    'roles_users',
    db.Column('user_id', db.BigInteger, db.ForeignKey('users.id')),
    db.Column('role_id', db.Integer, db.ForeignKey('roles.id'))
)


class Role(db.Model, RoleMixin):
    """User role model."""
    __tablename__ = 'roles'
    
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(80), unique=True, nullable=False)
    description = db.Column(db.String(255))
    
    def __repr__(self):
        return f'<Role {self.name}>'


class User(db.Model, UserMixin):
    """User model."""
    __tablename__ = 'users'
    
    id = db.Column(db.BigInteger, primary_key=True)
    email = db.Column(db.String(255), unique=True, nullable=False, index=True)
    password = db.Column(db.String(255), nullable=False)
    
    # Profile
    first_name = db.Column(db.String(100))
    last_name = db.Column(db.String(100))
    phone = db.Column(db.String(20))
    
    # Address
    address_line1 = db.Column(db.String(255))
    address_line2 = db.Column(db.String(255))
    city = db.Column(db.String(100))
    state = db.Column(db.String(2))
    zip_code = db.Column(db.String(10))
    
    # Flask-Security fields
    active = db.Column(db.Boolean, default=True)
    fs_uniquifier = db.Column(db.String(64), unique=True, nullable=False)
    confirmed_at = db.Column(db.DateTime)
    
    # Tracking (Flask-Security)
    last_login_at = db.Column(db.DateTime)
    current_login_at = db.Column(db.DateTime)
    last_login_ip = db.Column(db.String(45))
    current_login_ip = db.Column(db.String(45))
    login_count = db.Column(db.Integer, default=0)
    
    # Persona ID Verification
    persona_inquiry_id = db.Column(db.String(255))
    id_verified_at = db.Column(db.DateTime)
    id_document_type = db.Column(db.Enum('drivers_license', 'passport', 'state_id', name='id_document_type'))
    
    # Authorize.Net Payment Profile
    authorize_customer_profile_id = db.Column(db.String(255))
    authorize_payment_profile_id = db.Column(db.String(255))
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    roles = db.relationship('Role', secondary=roles_users, backref=db.backref('users', lazy='dynamic'))
    seller_profile = db.relationship('SellerProfile', back_populates='user', uselist=False)
    vehicles = db.relationship('Vehicle', back_populates='seller', lazy='dynamic')
    bids = db.relationship('Bid', back_populates='user', lazy='dynamic')
    offers = db.relationship('Offer', back_populates='user', lazy='dynamic')
    orders_as_buyer = db.relationship('Order', foreign_keys='Order.buyer_id', back_populates='buyer', lazy='dynamic')
    orders_as_seller = db.relationship('Order', foreign_keys='Order.seller_id', back_populates='seller', lazy='dynamic')
    notifications = db.relationship('Notification', back_populates='user', lazy='dynamic')
    saved_searches = db.relationship('SavedSearch', back_populates='user', lazy='dynamic')
    watchlist = db.relationship('Watchlist', back_populates='user', lazy='dynamic')
    
    @property
    def full_name(self):
        """Return user's full name."""
        if self.first_name and self.last_name:
            return f'{self.first_name} {self.last_name}'
        return self.email
    
    @property
    def is_id_verified(self):
        """Check if user has verified their ID."""
        return self.id_verified_at is not None
    
    @property
    def has_payment_method(self):
        """Check if user has a payment method on file."""
        return self.authorize_payment_profile_id is not None
    
    @property
    def can_bid(self):
        """Check if user can place bids."""
        return self.is_id_verified and self.has_payment_method
    
    def __repr__(self):
        return f'<User {self.email}>'


class SellerProfile(db.Model):
    """Seller profile for dealers/businesses."""
    __tablename__ = 'seller_profiles'
    
    id = db.Column(db.BigInteger, primary_key=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), unique=True, nullable=False)
    
    business_name = db.Column(db.String(255))
    dealer_license_number = db.Column(db.String(100))
    tax_id = db.Column(db.String(50))
    
    approved_at = db.Column(db.DateTime)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    user = db.relationship('User', back_populates='seller_profile')
    
    @property
    def is_approved(self):
        """Check if seller is approved."""
        return self.approved_at is not None
    
    def __repr__(self):
        return f'<SellerProfile {self.business_name}>'
