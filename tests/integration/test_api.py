"""Integration tests for API endpoints."""
import pytest
import json


class TestHealthEndpoint:
    """Tests for health check endpoint."""
    
    def test_health_check(self, client):
        """Test health endpoint returns healthy status."""
        response = client.get('/health')
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['status'] == 'healthy'


class TestVehiclesAPI:
    """Tests for vehicles API."""
    
    def test_list_vehicles_empty(self, client):
        """Test listing vehicles when none exist."""
        response = client.get('/api/vehicles')
        assert response.status_code == 200
        data = json.loads(response.data)
        assert 'vehicles' in data
        assert isinstance(data['vehicles'], list)
    
    def test_list_vehicles_with_data(self, client, test_vehicle):
        """Test listing vehicles with data."""
        response = client.get('/api/vehicles')
        assert response.status_code == 200
        data = json.loads(response.data)
        assert len(data['vehicles']) >= 1
    
    def test_list_vehicles_filter_by_make(self, client, test_vehicle):
        """Test filtering vehicles by make."""
        response = client.get('/api/vehicles?make=Honda')
        assert response.status_code == 200
        data = json.loads(response.data)
        for vehicle in data['vehicles']:
            assert vehicle['make'] == 'Honda'
    
    def test_list_vehicles_filter_by_year(self, client, test_vehicle):
        """Test filtering vehicles by year range."""
        response = client.get('/api/vehicles?year_min=2020&year_max=2022')
        assert response.status_code == 200
        data = json.loads(response.data)
        for vehicle in data['vehicles']:
            assert 2020 <= vehicle['year'] <= 2022
    
    def test_get_vehicle_detail(self, client, test_vehicle):
        """Test getting vehicle detail."""
        response = client.get(f'/api/vehicles/{test_vehicle.id}')
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['vin'] == test_vehicle.vin
        assert data['make'] == 'Honda'
        assert data['model'] == 'Accord'
    
    def test_get_vehicle_not_found(self, client):
        """Test getting non-existent vehicle."""
        response = client.get('/api/vehicles/99999')
        assert response.status_code == 404


class TestAuctionsAPI:
    """Tests for auctions API."""
    
    def test_get_auction_bids_empty(self, client, test_auction):
        """Test getting bids for auction with no bids."""
        response = client.get(f'/api/auctions/{test_auction.id}/bids')
        assert response.status_code == 200
        data = json.loads(response.data)
        assert data['bid_count'] == 0
        assert data['bids'] == []
    
    def test_get_auction_bids_not_found(self, client):
        """Test getting bids for non-existent auction."""
        response = client.get('/api/auctions/99999/bids')
        assert response.status_code == 404


class TestVINDecodeAPI:
    """Tests for VIN decode API."""
    
    def test_decode_vin_valid(self, client, test_user):
        """Test decoding a valid VIN."""
        # Login required - skip for now
        pass
    
    def test_decode_vin_invalid_length(self, client):
        """Test decoding invalid VIN length."""
        # Login required - skip for now
        pass
