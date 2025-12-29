"""Fulfillment models for title transfer and transport."""
from datetime import datetime
from ..extensions import db


class TitleTransfer(db.Model):
    """Title transfer model for DMV processing."""
    __tablename__ = 'title_transfers'
    
    id = db.Column(db.BigInteger, primary_key=True)
    order_id = db.Column(db.BigInteger, db.ForeignKey('orders.id'), unique=True, nullable=False)
    
    status = db.Column(
        db.Enum('pending', 'submitted_to_dmv', 'processing', 'completed', 'rejected', name='title_transfer_status'),
        default='pending',
        nullable=False
    )
    
    # DLR DMV Integration
    dlr_submission_id = db.Column(db.String(255))
    dlr_status = db.Column(db.String(100))
    new_title_number = db.Column(db.String(100))
    rejection_reason = db.Column(db.Text)
    
    # Documents
    title_document_s3_key = db.Column(db.String(500))
    bill_of_sale_s3_key = db.Column(db.String(500))
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    submitted_at = db.Column(db.DateTime)
    completed_at = db.Column(db.DateTime)
    
    # Relationships
    order = db.relationship('Order', back_populates='title_transfer')
    
    def __repr__(self):
        return f'<TitleTransfer {self.id} - {self.status}>'


class TransportOrder(db.Model):
    """Transport order model for vehicle shipping."""
    __tablename__ = 'transport_orders'
    
    id = db.Column(db.BigInteger, primary_key=True)
    order_id = db.Column(db.BigInteger, db.ForeignKey('orders.id'), unique=True, nullable=False)
    
    status = db.Column(
        db.Enum(
            'quote_requested', 'quoted', 'booked', 'assigned',
            'picked_up', 'in_transit', 'delivered', 'cancelled',
            name='transport_status'
        ),
        default='quote_requested',
        nullable=False
    )
    
    # Super Dispatch Integration
    super_dispatch_order_id = db.Column(db.String(255))
    carrier_name = db.Column(db.String(255))
    driver_name = db.Column(db.String(255))
    driver_phone = db.Column(db.String(20))
    
    # Pickup Address
    pickup_address = db.Column(db.String(255))
    pickup_city = db.Column(db.String(100))
    pickup_state = db.Column(db.String(2))
    pickup_zip = db.Column(db.String(10))
    pickup_contact_name = db.Column(db.String(255))
    pickup_contact_phone = db.Column(db.String(20))
    
    # Delivery Address
    delivery_address = db.Column(db.String(255))
    delivery_city = db.Column(db.String(100))
    delivery_state = db.Column(db.String(2))
    delivery_zip = db.Column(db.String(10))
    delivery_contact_name = db.Column(db.String(255))
    delivery_contact_phone = db.Column(db.String(20))
    
    # Pricing & Dates
    quoted_price = db.Column(db.Numeric(10, 2))
    final_price = db.Column(db.Numeric(10, 2))
    estimated_pickup_date = db.Column(db.Date)
    actual_pickup_date = db.Column(db.Date)
    estimated_delivery_date = db.Column(db.Date)
    actual_delivery_date = db.Column(db.Date)
    
    # Notes
    special_instructions = db.Column(db.Text)
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    
    # Relationships
    order = db.relationship('Order', back_populates='transport_order')
    
    def __repr__(self):
        return f'<TransportOrder {self.id} - {self.status}>'
