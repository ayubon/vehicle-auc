"""Auction API endpoints."""
from flask import jsonify
from . import api_bp
from ...models import Auction


@api_bp.route('/auctions/<int:auction_id>/bids')
def get_auction_bids(auction_id):
    """Get bid history for an auction."""
    auction = Auction.query.get_or_404(auction_id)
    
    bids = [{
        'id': bid.id,
        'amount': float(bid.amount),
        'user_display': f'user***{str(bid.user_id)[-2:]}',
        'is_auto_bid': bid.is_auto_bid,
        'created_at': bid.created_at.isoformat(),
    } for bid in auction.bids.limit(50).all()]
    
    return jsonify({
        'auction_id': auction.id,
        'current_bid': float(auction.current_bid) if auction.current_bid else 0,
        'bid_count': auction.bid_count,
        'bids': bids,
    })
