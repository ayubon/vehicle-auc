"""Order and payment models."""
from datetime import datetime
from ..extensions import db


class Order(db.Model):
    """Order model."""
    __tablename__ = 'orders'
    
    id = db.Column(db.BigInteger, primary_key=True)
    order_number = db.Column(db.String(50), unique=True, nullable=False, index=True)
    
    # Source (either auction or offer)
    auction_id = db.Column(db.BigInteger, db.ForeignKey('auctions.id'))
    offer_id = db.Column(db.BigInteger, db.ForeignKey('offers.id'))
    
    # Parties
    buyer_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    seller_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False, index=True)
    vehicle_id = db.Column(db.BigInteger, db.ForeignKey('vehicles.id'), nullable=False)
    
    # Amounts
    vehicle_price = db.Column(db.Numeric(10, 2), nullable=False)
    buyer_fee = db.Column(db.Numeric(10, 2), default=0)
    transport_fee = db.Column(db.Numeric(10, 2))
    title_fee = db.Column(db.Numeric(10, 2), default=0)
    tax = db.Column(db.Numeric(10, 2), default=0)
    total = db.Column(db.Numeric(10, 2), nullable=False)
    
    # Status
    status = db.Column(
        db.Enum(
            'pending_payment', 'paid', 'title_processing', 'transport_scheduled',
            'in_transit', 'delivered', 'completed', 'cancelled', 'refunded',
            name='order_status'
        ),
        default='pending_payment',
        nullable=False,
        index=True
    )
    
    # Timestamps
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    updated_at = db.Column(db.DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    paid_at = db.Column(db.DateTime)
    completed_at = db.Column(db.DateTime)
    
    # Relationships
    auction = db.relationship('Auction')
    offer = db.relationship('Offer')
    buyer = db.relationship('User', foreign_keys=[buyer_id], back_populates='orders_as_buyer')
    seller = db.relationship('User', foreign_keys=[seller_id], back_populates='orders_as_seller')
    vehicle = db.relationship('Vehicle')
    invoice = db.relationship('Invoice', back_populates='order', uselist=False)
    payments = db.relationship('Payment', back_populates='order', lazy='dynamic')
    title_transfer = db.relationship('TitleTransfer', back_populates='order', uselist=False)
    transport_order = db.relationship('TransportOrder', back_populates='order', uselist=False)
    
    def __repr__(self):
        return f'<Order {self.order_number}>'


class Invoice(db.Model):
    """Invoice model."""
    __tablename__ = 'invoices'
    
    id = db.Column(db.BigInteger, primary_key=True)
    order_id = db.Column(db.BigInteger, db.ForeignKey('orders.id'), unique=True, nullable=False)
    
    invoice_number = db.Column(db.String(50), unique=True, nullable=False, index=True)
    due_date = db.Column(db.Date, nullable=False)
    
    paid_at = db.Column(db.DateTime)
    pdf_s3_key = db.Column(db.String(500))
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    # Relationships
    order = db.relationship('Order', back_populates='invoice')
    
    @property
    def is_paid(self):
        """Check if invoice is paid."""
        return self.paid_at is not None
    
    @property
    def is_overdue(self):
        """Check if invoice is overdue."""
        from datetime import date
        return not self.is_paid and date.today() > self.due_date
    
    def __repr__(self):
        return f'<Invoice {self.invoice_number}>'


class Payment(db.Model):
    """Payment model."""
    __tablename__ = 'payments'
    
    id = db.Column(db.BigInteger, primary_key=True)
    order_id = db.Column(db.BigInteger, db.ForeignKey('orders.id'), nullable=False, index=True)
    user_id = db.Column(db.BigInteger, db.ForeignKey('users.id'), nullable=False)
    
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    payment_type = db.Column(
        db.Enum('card', 'ach', 'refund', name='payment_type'),
        nullable=False
    )
    status = db.Column(
        db.Enum('pending', 'authorized', 'captured', 'failed', 'refunded', name='payment_status'),
        default='pending',
        nullable=False
    )
    
    # Authorize.Net fields
    authorize_transaction_id = db.Column(db.String(255))
    authorize_response_code = db.Column(db.String(10))
    last_four = db.Column(db.String(4))
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    
    # Relationships
    order = db.relationship('Order', back_populates='payments')
    user = db.relationship('User')
    refunds = db.relationship('Refund', back_populates='payment', lazy='dynamic')
    
    def __repr__(self):
        return f'<Payment ${self.amount} - {self.status}>'


class Refund(db.Model):
    """Refund model."""
    __tablename__ = 'refunds'
    
    id = db.Column(db.BigInteger, primary_key=True)
    payment_id = db.Column(db.BigInteger, db.ForeignKey('payments.id'), nullable=False, index=True)
    
    amount = db.Column(db.Numeric(10, 2), nullable=False)
    reason = db.Column(db.Text)
    
    authorize_transaction_id = db.Column(db.String(255))
    status = db.Column(
        db.Enum('pending', 'processed', 'failed', name='refund_status'),
        default='pending',
        nullable=False
    )
    
    created_at = db.Column(db.DateTime, default=datetime.utcnow)
    processed_at = db.Column(db.DateTime)
    
    # Relationships
    payment = db.relationship('Payment', back_populates='refunds')
    
    def __repr__(self):
        return f'<Refund ${self.amount} - {self.status}>'
