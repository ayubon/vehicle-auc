"""Auction API endpoints.

SICP: Routes are thin - delegate serialization to models.
"""
from flask import jsonify
from . import api_bp
from ...models import Auction


@api_bp.route('/auctions/<int:auction_id>/bids')
def get_auction_bids(auction_id):
    """Get bid history for an auction."""
    auction = Auction.query.get_or_404(auction_id)
    
    return jsonify({
        'auction_id': auction.id,
        'current_bid': float(auction.current_bid) if auction.current_bid else 0,
        'bid_count': auction.bid_count,
        'bids': [bid.to_dict() for bid in auction.bids.limit(50).all()],
    })
