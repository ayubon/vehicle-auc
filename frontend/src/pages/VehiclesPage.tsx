import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useVehicles } from '@/hooks';
import type { VehicleFilters } from '@/types';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Car, MapPin, Gauge, FileText } from 'lucide-react';

export default function VehiclesPage() {
  const [filters, setFilters] = useState<VehicleFilters>({
    make: '',
    year_min: '',
    year_max: '',
    price_max: '',
  });

  const { vehicles, total, isLoading, error } = useVehicles(filters);

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Vehicle Inventory</h1>
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Card key={i}>
              <Skeleton className="aspect-video w-full" />
              <CardContent className="p-4">
                <Skeleton className="h-6 w-3/4 mb-2" />
                <Skeleton className="h-4 w-1/2 mb-4" />
                <Skeleton className="h-8 w-full" />
              </CardContent>
            </Card>
          ))}
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


  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold">Vehicle Inventory</h1>
          <p className="text-muted-foreground">{total} vehicles available</p>
        </div>
        <Link to="/vehicles/new">
          <Button>List a Vehicle</Button>
        </Link>
      </div>

      {/* Filters */}
      <Card className="mb-8">
        <CardContent className="p-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="space-y-2">
              <Label htmlFor="make">Make</Label>
              <Input
                id="make"
                placeholder="Any make"
                value={filters.make}
                onChange={(e) => setFilters({ ...filters, make: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="year_min">Year Min</Label>
              <Input
                id="year_min"
                type="number"
                placeholder="2010"
                value={filters.year_min}
                onChange={(e) => setFilters({ ...filters, year_min: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="year_max">Year Max</Label>
              <Input
                id="year_max"
                type="number"
                placeholder="2024"
                value={filters.year_max}
                onChange={(e) => setFilters({ ...filters, year_max: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="price_max">Max Price</Label>
              <Input
                id="price_max"
                type="number"
                placeholder="50000"
                value={filters.price_max}
                onChange={(e) => setFilters({ ...filters, price_max: e.target.value })}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {vehicles.length === 0 ? (
        <div className="text-center py-12">
          <Car className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No vehicles match your criteria.</p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {vehicles.map((vehicle) => (
            <Card key={vehicle.id} className="overflow-hidden hover:shadow-lg transition-shadow">
              <Link to={`/vehicles/${vehicle.id}`}>
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
                    <Badge variant="destructive" className="absolute top-2 right-2">
                      {vehicle.title_type.replace('_', ' ')}
                    </Badge>
                  )}
                </div>
                <CardContent className="p-4">
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
                </CardContent>
              </Link>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
