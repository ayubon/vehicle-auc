"""Vehicle routes."""
from flask import render_template, request, redirect, url_for, flash
from flask_login import login_required, current_user
from . import vehicles_bp
from ..models import Vehicle
from ..extensions import db


@vehicles_bp.route('/')
def list_vehicles():
    """List all active vehicles with filters."""
    page = request.args.get('page', 1, type=int)
    per_page = request.args.get('per_page', 20, type=int)
    
    query = Vehicle.query.filter_by(status='active')
    
    # Apply filters
    make = request.args.get('make')
    if make:
        query = query.filter(Vehicle.make == make)
    
    year_min = request.args.get('year_min', type=int)
    if year_min:
        query = query.filter(Vehicle.year >= year_min)
    
    year_max = request.args.get('year_max', type=int)
    if year_max:
        query = query.filter(Vehicle.year <= year_max)
    
    price_max = request.args.get('price_max', type=float)
    if price_max:
        query = query.filter(Vehicle.starting_price <= price_max)
    
    # Pagination
    vehicles = query.order_by(Vehicle.created_at.desc()).paginate(
        page=page, per_page=per_page, error_out=False
    )
    
    return render_template('vehicles/list.html', vehicles=vehicles)


@vehicles_bp.route('/<int:vehicle_id>')
def detail(vehicle_id):
    """Vehicle detail page."""
    vehicle = Vehicle.query.get_or_404(vehicle_id)
    return render_template('vehicles/detail.html', vehicle=vehicle)


@vehicles_bp.route('/create', methods=['GET', 'POST'])
@login_required
def create():
    """Create a new vehicle listing."""
    if request.method == 'POST':
        # TODO: Implement vehicle creation with VIN decoding
        flash('Vehicle listing created successfully.', 'success')
        return redirect(url_for('vehicles.list_vehicles'))
    return render_template('vehicles/create.html')


@vehicles_bp.route('/<int:vehicle_id>/edit', methods=['GET', 'POST'])
@login_required
def edit(vehicle_id):
    """Edit a vehicle listing."""
    vehicle = Vehicle.query.get_or_404(vehicle_id)
    
    # Check ownership
    if vehicle.seller_id != current_user.id:
        flash('You do not have permission to edit this vehicle.', 'error')
        return redirect(url_for('vehicles.detail', vehicle_id=vehicle_id))
    
    if request.method == 'POST':
        # TODO: Implement vehicle update
        flash('Vehicle updated successfully.', 'success')
        return redirect(url_for('vehicles.detail', vehicle_id=vehicle_id))
    
    return render_template('vehicles/edit.html', vehicle=vehicle)
