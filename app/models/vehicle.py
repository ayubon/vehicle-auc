"""Vehicle models."""
from datetime import datetime
from ..extensions import db


class Vehicle(db.Model):
    """Vehicle model."""
    __tablename__ = 'vehicles'
    
    id = db.Column(db.BigInteger, primary_key=True)
    seller_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    # Status
    status = db.Column(
        db.Enum('draft', 'pending_review', 'active', 'sold', 'archived', name='vehicle_status'),
        default='draft',
        nullable=False,
        index=True
    )
    
    # VIN & Decoded Data (from ClearVIN)
    vin = db.Column(db.String(17), unique=True, nullable=False, index=True)
    year = db.Column(db.SmallInteger)
    make = db.Column(db.String(100))
    model = db.Column(db.String(100))
    trim = db.Column(db.String(100))
    body_type = db.Column(db.String(100))
    engine = db.Column(db.String(255))
    transmission = db.Column(db.String(100))
    drivetrain = db.Column(db.String(50))
    exterior_color = db.Column(db.String(50))
    interior_color = db.Column(db.String(50))
    
    # Condition
    mileage = db.Column(db.Integer)
    condition = db.Column(
        db.Enum('runs_drives', 'starts_not_drive', 'does_not_start', name='vehicle_condition'),
        default='runs_drives'
    )
    title_type = db.Column(
        db.Enum('clean', 'salvage', 'rebuilt', 'junk', 'prior_salvage', 'flood', 'lemon', 
                'total_loss_history', 'not_actual_mileage', name='title_type'),
        default='clean'
    )
    title_state = db.Column(db.String(2))
    has_keys = db.Column(db.Boolean, default=True)
    description = db.Column(db.Text)
    
    # Pricing
    starting_price = db.Column(db.Numeric(10, 2), nullable=False)
    reserve_price = db.Column(db.Numeric(10, 2))
    buy_now_price = db.Column(db.Numeric(10, 2))
    
    # Location
    location_address = db.Column(db.String(255))
    location_city = db.Column(db.String(100))
    location_state = db.Column(db.String(2))
    location_zip = db.Column(db.String(10))
    latitude = db.Column(db.Numeric(10, 8))
    longitude = db.Column(db.Numeric(11, 8))
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    seller = db.relationship('User', back_populates='vehicles')
    images = db.relationship('VehicleImage', back_populates='vehicle', lazy='dynamic', cascade='all, delete-orphan')
    documents = db.relationship('VehicleDocument', back_populates='vehicle', lazy='dynamic', cascade='all, delete-orphan')
    auction = db.relationship('Auction', back_populates='vehicle', uselist=False)
    offers = db.relationship('Offer', back_populates='vehicle', lazy='dynamic')
    watchlist_entries = db.relationship('Watchlist', back_populates='vehicle', lazy='dynamic')
    
    @property
    def primary_image(self):
        """Get the primary image for this vehicle."""
        return self.images.filter_by(is_primary=True).first()
    
    @property
    def display_title(self):
        """Get display title for vehicle."""
        return f'{self.year} {self.make} {self.model}'
    
    def __repr__(self):
        return f'<Vehicle {self.vin}>'


class VehicleImage(db.Model):
    """Vehicle image model."""
    __tablename__ = 'vehicle_images'
    
    id = db.Column(db.BigInteger, primary_key=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), nullable=False, index=True)
    
    s3_key = db.Column(db.String(500), nullable=False)
    url = db.Column(db.String(1000))
    is_primary = db.Column(db.Boolean, default=False)
    sort_order = db.Column(db.SmallInteger, default=0)
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    # Relationships
    vehicle = db.relationship('Vehicle', back_populates='images')
    
    def __repr__(self):
        return f'<VehicleImage {self.id}>'


class VehicleDocument(db.Model):
    """Vehicle document model."""
    __tablename__ = 'vehicle_documents'
    
    id = db.Column(db.BigInteger, primary_key=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), nullable=False, index=True)
    
    doc_type = db.Column(
        db.Enum('title', 'registration', 'inspection', 'other', name='document_type'),
        nullable=False
    )
    s3_key = db.Column(db.String(500), nullable=False)
    filename = db.Column(db.String(255))
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    # Relationships
    vehicle = db.relationship('Vehicle', back_populates='documents')
    
    def __repr__(self):
        return f'<VehicleDocument {self.doc_type}>'
