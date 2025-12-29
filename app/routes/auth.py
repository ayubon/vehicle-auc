"""Authentication routes (placeholder - Flask-Security-Too handles most of this)."""
from flask import render_template, redirect, url_for, flash, request
from flask_login import login_required, current_user, login_user, logout_user
from . import auth_bp


@auth_bp.route('/login', methods=['GET', 'POST'])
def login():
    """Login page (placeholder until Flask-Security-Too is configured)."""
    if current_user.is_authenticated:
        return redirect(url_for('main.index'))
    return render_template('auth/login.html')


@auth_bp.route('/register', methods=['GET', 'POST'])
def register():
    """Registration page (placeholder until Flask-Security-Too is configured)."""
    if current_user.is_authenticated:
        return redirect(url_for('main.index'))
    return render_template('auth/register.html')


@auth_bp.route('/logout')
@login_required
def logout():
    """Logout user."""
    logout_user()
    flash('You have been logged out.', 'info')
    return redirect(url_for('main.index'))


@auth_bp.route('/profile')
@login_required
def profile():
    """User profile page."""
    return render_template('auth/profile.html')


@auth_bp.route('/verify-id')
@login_required
def verify_id():
    """ID verification page (Persona integration)."""
    if current_user.is_id_verified:
        flash('Your ID is already verified.', 'info')
        return redirect(url_for('auth.profile'))
    return render_template('auth/verify_id.html')


@auth_bp.route('/payment-method')
@login_required
def payment_method():
    """Payment method setup page (Authorize.Net integration)."""
    return render_template('auth/payment_method.html')


@auth_bp.route('/dashboard')
@login_required
def dashboard():
    """User dashboard."""
    return render_template('auth/dashboard.html')
