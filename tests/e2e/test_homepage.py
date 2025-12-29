"""E2E tests for homepage."""
import pytest
from playwright.sync_api import Page, expect


@pytest.mark.e2e
class TestHomepage:
    """E2E tests for homepage."""
    
    def test_homepage_loads(self, page: Page, live_server):
        """Test homepage loads successfully."""
        page.goto(live_server.url('/'))
        
        # Check title
        expect(page).to_have_title('Vehicle Auction - Marketplace, Made Simple')
        
        # Check hero section
        expect(page.locator('h1')).to_contain_text('Marketplace, Made Simple')
    
    def test_navigation_links(self, page: Page, live_server):
        """Test navigation links work."""
        page.goto(live_server.url('/'))
        
        # Click Inventory link
        page.click('text=Inventory')
        expect(page).to_have_url(live_server.url('/vehicles/'))
        
        # Go back and click Auctions
        page.goto(live_server.url('/'))
        page.click('text=Auctions')
        expect(page).to_have_url(live_server.url('/auctions/'))
    
    def test_search_form(self, page: Page, live_server):
        """Test search form submission."""
        page.goto(live_server.url('/'))
        
        # Fill search form
        page.select_option('select[name="make"]', 'Toyota')
        page.select_option('select[name="price_max"]', '20000')
        
        # Submit form
        page.click('button:has-text("Search")')
        
        # Should redirect to vehicles page with filters
        expect(page).to_have_url(lambda url: '/vehicles/' in url and 'make=Toyota' in url)
    
    def test_login_link_visible_when_logged_out(self, page: Page, live_server):
        """Test login link is visible when not logged in."""
        page.goto(live_server.url('/'))
        
        expect(page.locator('text=Log In')).to_be_visible()
        expect(page.locator('text=Sign Up')).to_be_visible()


@pytest.fixture
def live_server(app):
    """Create a live server for E2E testing."""
    from werkzeug.serving import make_server
    import threading
    
    class LiveServer:
        def __init__(self, app, host='127.0.0.1', port=5001):
            self.app = app
            self.host = host
            self.port = port
            self.server = make_server(host, port, app, threaded=True)
            self.thread = threading.Thread(target=self.server.serve_forever)
            self.thread.daemon = True
            self.thread.start()
        
        def url(self, path=''):
            return f'http://{self.host}:{self.port}{path}'
        
        def shutdown(self):
            self.server.shutdown()
    
    server = LiveServer(app)
    yield server
    server.shutdown()
