"""Main routes for homepage and static pages."""
from flask import render_template
from . import main_bp


@main_bp.route('/')
def index():
    """Homepage with vehicle listings."""
    return render_template('index.html')


@main_bp.route('/how-it-works')
def how_it_works():
    """How it works page."""
    return render_template('how_it_works.html')


@main_bp.route('/pricing')
def pricing():
    """Buyer premium fees page."""
    return render_template('pricing.html')


@main_bp.route('/contact')
def contact():
    """Contact us page."""
    return render_template('contact.html')


@main_bp.route('/health')
def health():
    """Health check endpoint."""
    return {'status': 'healthy'}
