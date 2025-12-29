"""Order routes."""
from flask import render_template, request, jsonify, flash, redirect, url_for, send_file
from flask_login import login_required, current_user
from . import orders_bp
from ..models import Order, Invoice
from ..extensions import db


@orders_bp.route('/')
@login_required
def list_orders():
    """List user's orders."""
    orders = current_user.orders_as_buyer.order_by(Order.created_at.desc()).all()
    return render_template('orders/list.html', orders=orders)


@orders_bp.route('/<int:order_id>')
@login_required
def detail(order_id):
    """Order detail page."""
    order = Order.query.get_or_404(order_id)
    
    # Check access
    if order.buyer_id != current_user.id and order.seller_id != current_user.id:
        flash('You do not have permission to view this order.', 'error')
        return redirect(url_for('orders.list_orders'))
    
    return render_template('orders/detail.html', order=order)


@orders_bp.route('/<int:order_id>/pay', methods=['GET', 'POST'])
@login_required
def pay(order_id):
    """Payment page for an order."""
    order = Order.query.get_or_404(order_id)
    
    # Check access
    if order.buyer_id != current_user.id:
        flash('You do not have permission to pay for this order.', 'error')
        return redirect(url_for('orders.list_orders'))
    
    # Check status
    if order.status != 'pending_payment':
        flash('This order has already been paid or is not available for payment.', 'info')
        return redirect(url_for('orders.detail', order_id=order_id))
    
    if request.method == 'POST':
        # TODO: Process payment via Authorize.Net
        flash('Payment processed successfully.', 'success')
        return redirect(url_for('orders.detail', order_id=order_id))
    
    return render_template('orders/pay.html', order=order)


@orders_bp.route('/<int:order_id>/invoice')
@login_required
def invoice(order_id):
    """View or download invoice."""
    order = Order.query.get_or_404(order_id)
    
    # Check access
    if order.buyer_id != current_user.id and order.seller_id != current_user.id:
        flash('You do not have permission to view this invoice.', 'error')
        return redirect(url_for('orders.list_orders'))
    
    if not order.invoice:
        flash('Invoice not found.', 'error')
        return redirect(url_for('orders.detail', order_id=order_id))
    
    # TODO: Generate/retrieve PDF from S3
    return render_template('orders/invoice.html', order=order, invoice=order.invoice)
