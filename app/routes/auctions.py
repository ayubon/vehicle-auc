"""Auction routes."""
from flask import render_template, request, jsonify, flash, redirect, url_for
from flask_login import login_required, current_user
from . import auctions_bp
from ..models import Auction, Bid
from ..extensions import db


@auctions_bp.route('/')
def list_auctions():
    """List active auctions."""
    page = request.args.get('page', 1, type=int)
    per_page = request.args.get('per_page', 20, type=int)
    
    auctions = Auction.query.filter_by(status='active').order_by(
        Auction.ends_at.asc()
    ).paginate(page=page, per_page=per_page, error_out=False)
    
    return render_template('auctions/list.html', auctions=auctions)


@auctions_bp.route('/<int:auction_id>')
def detail(auction_id):
    """Auction detail page with bidding interface."""
    auction = Auction.query.get_or_404(auction_id)
    return render_template('auctions/detail.html', auction=auction)


@auctions_bp.route('/<int:auction_id>/bid', methods=['POST'])
@login_required
def place_bid(auction_id):
    """Place a bid on an auction."""
    auction = Auction.query.get_or_404(auction_id)
    
    # Check if user can bid
    if not current_user.can_bid:
        return jsonify({
            'success': False,
            'error': 'You must verify your ID and add a payment method before bidding.'
        }), 403
    
    # Check if auction is active
    if not auction.is_active:
        return jsonify({
            'success': False,
            'error': 'This auction is not currently active.'
        }), 400
    
    # Get bid amount
    data = request.get_json()
    amount = data.get('amount')
    
    if not amount:
        return jsonify({
            'success': False,
            'error': 'Bid amount is required.'
        }), 400
    
    try:
        amount = float(amount)
    except (TypeError, ValueError):
        return jsonify({
            'success': False,
            'error': 'Invalid bid amount.'
        }), 400
    
    # Validate bid amount
    min_bid = float(auction.current_bid) + 100 if auction.current_bid else float(auction.vehicle.starting_price)
    if amount < min_bid:
        return jsonify({
            'success': False,
            'error': f'Minimum bid is ${min_bid:,.2f}'
        }), 400
    
    # Create bid
    bid = Bid(
        auction_id=auction.id,
        user_id=current_user.id,
        amount=amount,
        ip_address=request.remote_addr
    )
    
    # Update auction
    auction.current_bid = amount
    auction.bid_count += 1
    
    db.session.add(bid)
    db.session.commit()
    
    # TODO: Emit WebSocket event for real-time updates
    # TODO: Check for auto-bids and process them
    # TODO: Send outbid notifications
    
    return jsonify({
        'success': True,
        'bid_id': bid.id,
        'current_bid': float(auction.current_bid),
        'bid_count': auction.bid_count
    })


@auctions_bp.route('/<int:auction_id>/auto-bid', methods=['POST'])
@login_required
def set_auto_bid(auction_id):
    """Set an auto-bid (proxy bid) on an auction."""
    auction = Auction.query.get_or_404(auction_id)
    
    if not current_user.can_bid:
        return jsonify({
            'success': False,
            'error': 'You must verify your ID and add a payment method before bidding.'
        }), 403
    
    data = request.get_json()
    max_bid = data.get('max_bid')
    
    if not max_bid:
        return jsonify({
            'success': False,
            'error': 'Maximum bid amount is required.'
        }), 400
    
    # TODO: Implement auto-bid logic
    
    return jsonify({
        'success': True,
        'message': 'Auto-bid set successfully.'
    })
