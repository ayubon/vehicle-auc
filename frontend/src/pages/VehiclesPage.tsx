import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { vehiclesApi } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Car, MapPin, Gauge, FileText } from 'lucide-react';

interface Vehicle {
  id: number;
  vin: string;
  year: number;
  make: string;
  model: string;
  trim?: string;
  mileage?: number;
  condition?: string;
  title_type?: string;
  starting_price?: number;
  buy_now_price?: number;
  location_city?: string;
  location_state?: string;
  primary_image_url?: string;
}

export default function VehiclesPage() {
  const [filters, setFilters] = useState({
    make: '',
    year_min: '',
    year_max: '',
    price_max: '',
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ['vehicles', filters],
    queryFn: () => vehiclesApi.list(filters),
  });

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12 text-destructive">
          Failed to load vehicles. Please try again.
        </div>
      </div>
    );
  }

  const vehicles: Vehicle[] = data?.data?.vehicles || [];
  const total = data?.data?.total || 0;

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold">Vehicle Inventory</h1>
          <p className="text-muted-foreground">{total} vehicles available</p>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-card border rounded-lg p-4 mb-8">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium mb-1">Make</label>
            <input
              type="text"
              className="w-full border rounded-md px-3 py-2 bg-background"
              placeholder="Any make"
              value={filters.make}
              onChange={(e) => setFilters({ ...filters, make: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Year Min</label>
            <input
              type="number"
              className="w-full border rounded-md px-3 py-2 bg-background"
              placeholder="2010"
              value={filters.year_min}
              onChange={(e) => setFilters({ ...filters, year_min: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Year Max</label>
            <input
              type="number"
              className="w-full border rounded-md px-3 py-2 bg-background"
              placeholder="2024"
              value={filters.year_max}
              onChange={(e) => setFilters({ ...filters, year_max: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Max Price</label>
            <input
              type="number"
              className="w-full border rounded-md px-3 py-2 bg-background"
              placeholder="50000"
              value={filters.price_max}
              onChange={(e) => setFilters({ ...filters, price_max: e.target.value })}
            />
          </div>
        </div>
      </div>

      {vehicles.length === 0 ? (
        <div className="text-center py-12">
          <Car className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No vehicles match your criteria.</p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {vehicles.map((vehicle) => (
            <Link
              key={vehicle.id}
              to={`/vehicles/${vehicle.id}`}
              className="border rounded-lg overflow-hidden bg-card hover:shadow-lg transition-shadow"
            >
              <div className="aspect-video bg-muted relative">
                {vehicle.primary_image_url ? (
                  <img
                    src={vehicle.primary_image_url}
                    alt={`${vehicle.year} ${vehicle.make} ${vehicle.model}`}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center">
                    <Car className="h-12 w-12 text-muted-foreground" />
                  </div>
                )}
                {vehicle.title_type && vehicle.title_type !== 'clean' && (
                  <span className="absolute top-2 right-2 bg-yellow-500 text-white text-xs px-2 py-1 rounded">
                    {vehicle.title_type.replace('_', ' ')}
                  </span>
                )}
              </div>
              <div className="p-4">
                <h3 className="font-semibold text-lg">
                  {vehicle.year} {vehicle.make} {vehicle.model}
                </h3>
                {vehicle.trim && (
                  <p className="text-muted-foreground text-sm">{vehicle.trim}</p>
                )}
                
                <div className="mt-3 flex flex-wrap gap-3 text-sm text-muted-foreground">
                  {vehicle.mileage && (
                    <span className="flex items-center gap-1">
                      <Gauge className="h-4 w-4" />
                      {vehicle.mileage.toLocaleString()} mi
                    </span>
                  )}
                  {vehicle.location_city && (
                    <span className="flex items-center gap-1">
                      <MapPin className="h-4 w-4" />
                      {vehicle.location_city}, {vehicle.location_state}
                    </span>
                  )}
                  {vehicle.condition && (
                    <span className="flex items-center gap-1">
                      <FileText className="h-4 w-4" />
                      {vehicle.condition.replace('_', ' ')}
                    </span>
                  )}
                </div>

                <div className="mt-4 flex justify-between items-center">
                  <div>
                    <span className="font-bold text-lg">
                      ${vehicle.starting_price?.toLocaleString()}
                    </span>
                    {vehicle.buy_now_price && (
                      <span className="text-sm text-muted-foreground ml-2">
                        Buy Now: ${vehicle.buy_now_price.toLocaleString()}
                      </span>
                    )}
                  </div>
                  <Button size="sm">View</Button>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
