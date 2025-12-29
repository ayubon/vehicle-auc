"""Vehicle models.

SICP: Data abstraction - Vehicle knows how to represent itself.
Ask the object for its representation, don't reach into its internals.
"""
from datetime import datetime
from ..extensions import db
from ..constants import VehicleStatus


class Vehicle(db.Model):
    """Vehicle model - the core domain object for vehicle listings."""
    __tablename__ = 'vehicles'
    
    id = db.Column(db.BigInteger, primary_key=True)
    seller_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    
    # Status
    status = db.Column(
        db.Enum('draft', 'pending_review', 'active', 'sold', 'archived', name='vehicle_status'),
        default=VehicleStatus.DRAFT.value,
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
    
    # ─────────────────────────────────────────────────────────────────────────
    # SICP: Message passing - ask the object for its representation
    # ─────────────────────────────────────────────────────────────────────────
    
    def to_summary_dict(self) -> dict:
        """Serialize for list views. Minimal data for browsing."""
        return {
            'id': self.id,
            'vin': self.vin,
            'year': self.year,
            'make': self.make,
            'model': self.model,
            'trim': self.trim,
            'mileage': self.mileage,
            'condition': self.condition,
            'title_type': self.title_type,
            'starting_price': float(self.starting_price) if self.starting_price else None,
            'buy_now_price': float(self.buy_now_price) if self.buy_now_price else None,
            'location_city': self.location_city,
            'location_state': self.location_state,
            'primary_image_url': self.primary_image.url if self.primary_image else None,
        }
    
    def to_detail_dict(self) -> dict:
        """Serialize for detail views. Full vehicle data."""
        return {
            'id': self.id,
            'vin': self.vin,
            'year': self.year,
            'make': self.make,
            'model': self.model,
            'trim': self.trim,
            'body_type': self.body_type,
            'engine': self.engine,
            'transmission': self.transmission,
            'drivetrain': self.drivetrain,
            'exterior_color': self.exterior_color,
            'interior_color': self.interior_color,
            'mileage': self.mileage,
            'condition': self.condition,
            'title_type': self.title_type,
            'title_state': self.title_state,
            'has_keys': self.has_keys,
            'description': self.description,
            'starting_price': float(self.starting_price) if self.starting_price else None,
            'reserve_price': float(self.reserve_price) if self.reserve_price else None,
            'buy_now_price': float(self.buy_now_price) if self.buy_now_price else None,
            'location': {
                'address': self.location_address,
                'city': self.location_city,
                'state': self.location_state,
                'zip': self.location_zip,
            },
            'images': [img.to_dict() for img in self.images.order_by('sort_order').all()],
            'auction': self.auction.to_summary_dict() if self.auction else None,
        }
    
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
    
    def to_dict(self) -> dict:
        """Serialize image."""
        return {
            'id': self.id,
            'url': self.url,
            'is_primary': self.is_primary,
        }
    
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
