/**
 * Vehicle Auction - Main JavaScript
 */

// Socket.IO connection for real-time updates
let socket = null;

document.addEventListener('DOMContentLoaded', function() {
    // Initialize Socket.IO if on auction page
    if (document.querySelector('[data-auction-id]')) {
        initializeSocket();
    }
    
    // Initialize tooltips
    const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
    tooltipTriggerList.forEach(el => new bootstrap.Tooltip(el));
});

/**
 * Initialize Socket.IO connection for real-time bidding
 */
function initializeSocket() {
    socket = io();
    
    socket.on('connect', function() {
        console.log('Connected to auction server');
        
        // Join auction room
        const auctionId = document.querySelector('[data-auction-id]')?.dataset.auctionId;
        if (auctionId) {
            socket.emit('join_auction', { auction_id: auctionId });
        }
    });
    
    socket.on('bid_update', function(data) {
        console.log('Bid update received:', data);
        updateBidDisplay(data);
    });
    
    socket.on('auction_ended', function(data) {
        console.log('Auction ended:', data);
        handleAuctionEnded(data);
    });
    
    socket.on('disconnect', function() {
        console.log('Disconnected from auction server');
    });
}

/**
 * Update bid display when new bid is received
 */
function updateBidDisplay(data) {
    // Update current bid
    const currentBidEl = document.getElementById('current-bid');
    if (currentBidEl) {
        currentBidEl.textContent = '$' + data.current_bid.toLocaleString();
        currentBidEl.classList.add('text-success');
        setTimeout(() => currentBidEl.classList.remove('text-success'), 1000);
    }
    
    // Update bid count
    const bidCountEl = document.getElementById('bid-count');
    if (bidCountEl) {
        bidCountEl.textContent = data.bid_count + ' bids';
    }
    
    // Add to bid history
    const bidHistoryEl = document.getElementById('bid-history');
    if (bidHistoryEl) {
        const bidItem = document.createElement('div');
        bidItem.className = 'bid-history-item';
        bidItem.innerHTML = `
            <div class="d-flex justify-content-between">
                <span>${data.user_display}</span>
                <span class="fw-bold">$${data.amount.toLocaleString()}</span>
            </div>
            <small class="text-muted">Just now</small>
        `;
        bidHistoryEl.prepend(bidItem);
    }
    
    // Update minimum bid
    const minBidEl = document.getElementById('min-bid');
    if (minBidEl) {
        const minBid = data.current_bid + 100;
        minBidEl.textContent = '$' + minBid.toLocaleString();
        
        const bidInput = document.getElementById('bid-amount');
        if (bidInput) {
            bidInput.min = minBid;
            bidInput.placeholder = minBid;
        }
    }
}

/**
 * Handle auction ended event
 */
function handleAuctionEnded(data) {
    const auctionStatusEl = document.getElementById('auction-status');
    if (auctionStatusEl) {
        auctionStatusEl.innerHTML = `
            <div class="alert alert-info">
                <i class="bi bi-flag-fill me-2"></i>
                <strong>Auction Ended!</strong>
                ${data.winner_id ? `Winner: ${data.winner_display}` : 'No winner (reserve not met)'}
            </div>
        `;
    }
    
    // Disable bid form
    const bidForm = document.getElementById('bid-form');
    if (bidForm) {
        bidForm.querySelectorAll('input, button').forEach(el => el.disabled = true);
    }
}

/**
 * Place a bid
 */
async function placeBid(auctionId, amount) {
    try {
        const response = await fetch(`/auctions/${auctionId}/bid`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRFToken': getCSRFToken()
            },
            body: JSON.stringify({ amount: amount })
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast('Bid placed successfully!', 'success');
        } else {
            showToast(data.error || 'Failed to place bid', 'danger');
        }
        
        return data;
    } catch (error) {
        console.error('Error placing bid:', error);
        showToast('An error occurred. Please try again.', 'danger');
        return { success: false, error: error.message };
    }
}

/**
 * Set auto-bid
 */
async function setAutoBid(auctionId, maxBid) {
    try {
        const response = await fetch(`/auctions/${auctionId}/auto-bid`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRFToken': getCSRFToken()
            },
            body: JSON.stringify({ max_bid: maxBid })
        });
        
        const data = await response.json();
        
        if (data.success) {
            showToast('Auto-bid set successfully!', 'success');
        } else {
            showToast(data.error || 'Failed to set auto-bid', 'danger');
        }
        
        return data;
    } catch (error) {
        console.error('Error setting auto-bid:', error);
        showToast('An error occurred. Please try again.', 'danger');
        return { success: false, error: error.message };
    }
}

/**
 * Countdown timer for auctions
 */
function startCountdown(endTime, elementId) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    const endDate = new Date(endTime).getTime();
    
    const timer = setInterval(function() {
        const now = new Date().getTime();
        const distance = endDate - now;
        
        if (distance < 0) {
            clearInterval(timer);
            element.textContent = 'Ended';
            element.classList.add('text-danger');
            return;
        }
        
        const days = Math.floor(distance / (1000 * 60 * 60 * 24));
        const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
        const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
        const seconds = Math.floor((distance % (1000 * 60)) / 1000);
        
        let display = '';
        if (days > 0) display += days + 'd ';
        if (hours > 0 || days > 0) display += hours + 'h ';
        display += minutes + 'm ' + seconds + 's';
        
        element.textContent = display;
        
        // Add warning class if less than 5 minutes
        if (distance < 5 * 60 * 1000) {
            element.classList.add('ending-soon');
        }
    }, 1000);
}

/**
 * Get CSRF token from meta tag or cookie
 */
function getCSRFToken() {
    const meta = document.querySelector('meta[name="csrf-token"]');
    if (meta) return meta.content;
    
    const cookies = document.cookie.split(';');
    for (let cookie of cookies) {
        const [name, value] = cookie.trim().split('=');
        if (name === 'csrf_token') return value;
    }
    
    return '';
}

/**
 * Show toast notification
 */
function showToast(message, type = 'info') {
    const toastContainer = document.getElementById('toast-container') || createToastContainer();
    
    const toast = document.createElement('div');
    toast.className = `toast align-items-center text-white bg-${type} border-0`;
    toast.setAttribute('role', 'alert');
    toast.innerHTML = `
        <div class="d-flex">
            <div class="toast-body">${message}</div>
            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
    `;
    
    toastContainer.appendChild(toast);
    
    const bsToast = new bootstrap.Toast(toast, { delay: 5000 });
    bsToast.show();
    
    toast.addEventListener('hidden.bs.toast', () => toast.remove());
}

/**
 * Create toast container if it doesn't exist
 */
function createToastContainer() {
    const container = document.createElement('div');
    container.id = 'toast-container';
    container.className = 'toast-container position-fixed bottom-0 end-0 p-3';
    document.body.appendChild(container);
    return container;
}

/**
 * Format currency
 */
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 0,
        maximumFractionDigits: 0
    }).format(amount);
}

/**
 * Format date/time
 */
function formatDateTime(dateString) {
    return new Intl.DateTimeFormat('en-US', {
        dateStyle: 'medium',
        timeStyle: 'short'
    }).format(new Date(dateString));
}
