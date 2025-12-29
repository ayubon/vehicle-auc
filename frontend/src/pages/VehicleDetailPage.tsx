import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Car, MapPin, Gauge, FileText, Key, ArrowLeft, Clock } from 'lucide-react';

export default function VehicleDetailPage() {
  const { id } = useParams();

  const { data, isLoading, error } = useQuery({
    queryKey: ['vehicle', id],
    queryFn: () => vehiclesApi.get(Number(id)),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Skeleton className="h-6 w-40 mb-6" />
        <div className="grid lg:grid-cols-2 gap-8">
          <div>
            <Skeleton className="aspect-video w-full rounded-lg mb-4" />
            <div className="grid grid-cols-4 gap-2">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="aspect-video rounded" />
              ))}
            </div>
          </div>
          <div>
            <Skeleton className="h-10 w-3/4 mb-2" />
            <Skeleton className="h-6 w-1/2 mb-6" />
            <Skeleton className="h-32 w-full mb-6" />
            <Skeleton className="h-24 w-full" />
          </div>
        </div>
      </div>
    );
  }

  if (error || !data?.data) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center py-12">
          <p className="text-destructive mb-4">Vehicle not found</p>
          <Button asChild>
            <Link to="/vehicles">Back to Inventory</Link>
          </Button>
        </div>
      </div>
    );
  }

  const vehicle = data.data;
  const images = vehicle.images || [];
  const auction = vehicle.auction;

  return (
    <div className="container mx-auto px-4 py-8">
      <Link to="/vehicles" className="inline-flex items-center text-muted-foreground hover:text-foreground mb-6">
        <ArrowLeft className="h-4 w-4 mr-2" />
        Back to Inventory
      </Link>

      <div className="grid lg:grid-cols-2 gap-8">
        {/* Images */}
        <div>
          <div className="aspect-video bg-muted rounded-lg overflow-hidden mb-4">
            {images.length > 0 ? (
              <img
                src={images.find((i: any) => i.is_primary)?.url || images[0]?.url}
                alt={`${vehicle.year} ${vehicle.make} ${vehicle.model}`}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <Car className="h-16 w-16 text-muted-foreground" />
              </div>
            )}
          </div>
          {images.length > 1 && (
            <div className="grid grid-cols-4 gap-2">
              {images.slice(0, 4).map((img: any, idx: number) => (
                <div key={idx} className="aspect-video bg-muted rounded overflow-hidden">
                  <img src={img.url} alt="" className="w-full h-full object-cover" />
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Details */}
        <div>
          <h1 className="text-3xl font-bold mb-2">
            {vehicle.year} {vehicle.make} {vehicle.model}
          </h1>
          {vehicle.trim && (
            <p className="text-xl text-muted-foreground mb-4">{vehicle.trim}</p>
          )}

          {/* Auction Info */}
          {auction && (
            <Card className="mb-6 border-primary/20 bg-primary/5">
              <CardContent className="p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="font-semibold">Current Bid</span>
                  <span className="text-2xl font-bold">${auction.current_bid?.toLocaleString() || 0}</span>
                </div>
                <div className="flex items-center justify-between text-sm text-muted-foreground">
                  <Badge variant="secondary">{auction.bid_count || 0} bids</Badge>
                  <span className="flex items-center gap-1">
                    <Clock className="h-4 w-4" />
                    {auction.time_remaining > 0 
                      ? `${Math.floor(auction.time_remaining / 3600)}h ${Math.floor((auction.time_remaining % 3600) / 60)}m left`
                      : 'Ended'}
                  </span>
                </div>
                <Button className="w-full mt-4">Place Bid</Button>
              </CardContent>
            </Card>
          )}

          {/* Price */}
          <Card className="mb-6">
            <CardContent className="p-4">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Starting Price</span>
                <span className="text-2xl font-bold">${vehicle.starting_price?.toLocaleString()}</span>
              </div>
              {vehicle.buy_now_price && (
                <div className="flex justify-between items-center mt-2 pt-2 border-t">
                  <span className="text-muted-foreground">Buy Now Price</span>
                  <Badge variant="outline" className="text-green-600 text-lg px-3 py-1">
                    ${vehicle.buy_now_price.toLocaleString()}
                  </Badge>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Quick Info */}
          <div className="grid grid-cols-2 gap-4 mb-6">
            <div className="flex items-center gap-2 text-sm">
              <Gauge className="h-5 w-5 text-muted-foreground" />
              <span>{vehicle.mileage?.toLocaleString() || 'N/A'} miles</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <span className="capitalize">{vehicle.title_type?.replace('_', ' ') || 'Clean'} Title</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <MapPin className="h-5 w-5 text-muted-foreground" />
              <span>{vehicle.location?.city}, {vehicle.location?.state}</span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <Key className="h-5 w-5 text-muted-foreground" />
              <span>{vehicle.has_keys ? 'Has Keys' : 'No Keys'}</span>
            </div>
          </div>

          {/* Specs */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-lg">Specifications</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 gap-y-2 text-sm">
                <span className="text-muted-foreground">VIN</span>
                <span className="font-mono">{vehicle.vin}</span>
                <span className="text-muted-foreground">Condition</span>
                <Badge variant="outline" className="w-fit capitalize">{vehicle.condition?.replace('_', ' ')}</Badge>
                <span className="text-muted-foreground">Engine</span>
                <span>{vehicle.engine || 'N/A'}</span>
                <span className="text-muted-foreground">Transmission</span>
                <span>{vehicle.transmission || 'N/A'}</span>
                <span className="text-muted-foreground">Drivetrain</span>
                <span>{vehicle.drivetrain || 'N/A'}</span>
                <span className="text-muted-foreground">Exterior</span>
                <span>{vehicle.exterior_color || 'N/A'}</span>
                <span className="text-muted-foreground">Interior</span>
                <span>{vehicle.interior_color || 'N/A'}</span>
              </div>
            </CardContent>
          </Card>

          {/* Description */}
          {vehicle.description && (
            <div className="mt-6">
              <h3 className="font-semibold mb-2">Description</h3>
              <p className="text-muted-foreground whitespace-pre-wrap">{vehicle.description}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
