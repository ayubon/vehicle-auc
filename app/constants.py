"""
Domain constants - single source of truth for status values and enums.

SICP: These are the "agreed-upon interfaces" between components.
Changing a status value here propagates everywhere automatically.
"""
from enum import Enum


class VehicleStatus(str, Enum):
    """Vehicle lifecycle states."""
    DRAFT = 'draft'
    PENDING_REVIEW = 'pending_review'
    ACTIVE = 'active'
    SOLD = 'sold'
    EXPIRED = 'expired'
    CANCELLED = 'cancelled'


class AuctionStatus(str, Enum):
    """Auction lifecycle states."""
    SCHEDULED = 'scheduled'
    ACTIVE = 'active'
    ENDED = 'ended'
    CANCELLED = 'cancelled'


class OrderStatus(str, Enum):
    """Order lifecycle states."""
    PENDING_PAYMENT = 'pending_payment'
    PAID = 'paid'
    PENDING_PICKUP = 'pending_pickup'
    IN_TRANSIT = 'in_transit'
    DELIVERED = 'delivered'
    COMPLETED = 'completed'
    CANCELLED = 'cancelled'
    REFUNDED = 'refunded'


class TitleType(str, Enum):
    """Vehicle title types."""
    CLEAN = 'clean'
    SALVAGE = 'salvage'
    REBUILT = 'rebuilt'
    FLOOD = 'flood'
    LEMON = 'lemon'


class VehicleCondition(str, Enum):
    """Vehicle condition ratings."""
    RUNS_DRIVES = 'runs_drives'
    STARTS = 'starts'
    NON_RUNNING = 'non_running'
    PARTS_ONLY = 'parts_only'
