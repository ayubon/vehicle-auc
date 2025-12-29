"""ClearVIN API integration for VIN decoding."""
import os
import httpx
from typing import Optional
import structlog

logger = structlog.get_logger(__name__)

CLEARVIN_API_URL = "https://www.clearvin.com/rest/vendor/report"


class ClearVINError(Exception):
    """ClearVIN API error."""
    pass


class ClearVINService:
    """Service for decoding VINs via ClearVIN API."""
    
    def __init__(self, api_key: Optional[str] = None):
        self.api_key = api_key or os.environ.get('CLEARVIN_API_KEY')
        if not self.api_key:
            logger.warning("ClearVIN API key not configured")
    
    def decode_vin(self, vin: str) -> dict:
        """
        Decode a VIN and return vehicle specifications.
        
        Args:
            vin: 17-character Vehicle Identification Number
            
        Returns:
            dict with vehicle data (year, make, model, etc.)
            
        Raises:
            ClearVINError: If API call fails
        """
        if not vin or len(vin) != 17:
            raise ClearVINError("VIN must be exactly 17 characters")
        
        vin = vin.upper().strip()
        
        # If no API key, return mock data for development
        if not self.api_key:
            logger.info("Using mock VIN data (no API key)", vin=vin)
            return self._mock_decode(vin)
        
        try:
            with httpx.Client(timeout=30.0) as client:
                response = client.get(
                    CLEARVIN_API_URL,
                    params={
                        'vin': vin,
                        'key': self.api_key,
                    }
                )
                response.raise_for_status()
                data = response.json()
                
                logger.info("VIN decoded successfully", vin=vin)
                return self._parse_response(data)
                
        except httpx.HTTPStatusError as e:
            logger.error("ClearVIN API error", vin=vin, status=e.response.status_code)
            raise ClearVINError(f"API error: {e.response.status_code}")
        except httpx.RequestError as e:
            logger.error("ClearVIN request failed", vin=vin, error=str(e))
            raise ClearVINError(f"Request failed: {str(e)}")
        except Exception as e:
            logger.error("VIN decode failed", vin=vin, error=str(e))
            raise ClearVINError(f"Decode failed: {str(e)}")
    
    def _parse_response(self, data: dict) -> dict:
        """Parse ClearVIN API response into our format."""
        # ClearVIN response structure varies - adapt as needed
        vehicle = data.get('vehicle', {})
        specs = data.get('specifications', {})
        
        return {
            'vin': data.get('vin', ''),
            'year': vehicle.get('year'),
            'make': vehicle.get('make'),
            'model': vehicle.get('model'),
            'trim': vehicle.get('trim'),
            'body_type': specs.get('body_type') or specs.get('body_style'),
            'engine': specs.get('engine') or specs.get('engine_description'),
            'transmission': specs.get('transmission'),
            'drivetrain': specs.get('drivetrain') or specs.get('drive_type'),
            'exterior_color': None,  # Not typically in VIN decode
            'interior_color': None,
            'fuel_type': specs.get('fuel_type'),
            'cylinders': specs.get('cylinders'),
            'displacement': specs.get('displacement'),
        }
    
    def _mock_decode(self, vin: str) -> dict:
        """Return mock data for development without API key."""
        # Use VIN characters to generate somewhat realistic mock data
        year_code = vin[9] if len(vin) > 9 else 'M'
        year_map = {
            'A': 2010, 'B': 2011, 'C': 2012, 'D': 2013, 'E': 2014,
            'F': 2015, 'G': 2016, 'H': 2017, 'J': 2018, 'K': 2019,
            'L': 2020, 'M': 2021, 'N': 2022, 'P': 2023, 'R': 2024,
            'S': 2025,
        }
        year = year_map.get(year_code, 2021)
        
        # Mock makes based on first character
        make_map = {
            '1': 'Chevrolet', '2': 'Ford', '3': 'Chrysler',
            '4': 'Buick', '5': 'Cadillac', 'J': 'Honda',
            'W': 'Volkswagen', 'S': 'Mercedes-Benz', 'Z': 'Ferrari',
        }
        make = make_map.get(vin[0], 'Toyota')
        
        model_map = {
            'Toyota': 'Camry',
            'Honda': 'Accord',
            'Ford': 'F-150',
            'Chevrolet': 'Silverado',
            'Chrysler': '300',
            'Buick': 'Enclave',
            'Cadillac': 'Escalade',
            'Volkswagen': 'Jetta',
            'Mercedes-Benz': 'C-Class',
            'Ferrari': '488',
        }
        model = model_map.get(make, 'Sedan')
        
        return {
            'vin': vin,
            'year': year,
            'make': make,
            'model': model,
            'trim': 'SE',
            'body_type': 'Sedan',
            'engine': '2.5L 4-Cylinder',
            'transmission': 'Automatic',
            'drivetrain': 'FWD',
            'exterior_color': None,
            'interior_color': None,
            'fuel_type': 'Gasoline',
            'cylinders': 4,
            'displacement': '2.5L',
        }


# Singleton instance
clearvin_service = ClearVINService()
