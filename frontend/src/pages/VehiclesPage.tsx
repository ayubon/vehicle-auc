import { useQuery } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Car } from 'lucide-react';

export default function VehiclesPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['vehicles'],
    queryFn: () => vehiclesApi.list(),
  });

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">Loading vehicles...</div>
      </div>
    );
  }

  const vehicles = data?.data?.vehicles || [];

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8">Vehicle Inventory</h1>
      
      {vehicles.length === 0 ? (
        <div className="text-center py-12">
          <Car className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No vehicles available at the moment.</p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {vehicles.map((vehicle: any) => (
            <div key={vehicle.id} className="border rounded-lg overflow-hidden bg-card">
              <div className="aspect-video bg-muted" />
              <div className="p-4">
                <h3 className="font-semibold">{vehicle.year} {vehicle.make} {vehicle.model}</h3>
                <p className="text-muted-foreground text-sm">{vehicle.mileage?.toLocaleString()} miles</p>
                <div className="mt-4 flex justify-between items-center">
                  <span className="font-bold">${vehicle.starting_price?.toLocaleString()}</span>
                  <Button size="sm">View Details</Button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
