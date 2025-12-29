/**
 * VehicleCreatePage - Form to create a new vehicle listing with image upload.
 */
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';
import { vehicleFormSchema, type VehicleFormData } from '@/types';
import { useAuth } from '@/hooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ImageUpload } from '@/components/ImageUpload';
import { ArrowLeft, Loader2, Search } from 'lucide-react';
import { Link } from 'react-router-dom';

interface UploadedImage {
  url: string;
  s3_key: string;
  is_primary: boolean;
}

export default function VehicleCreatePage() {
  const navigate = useNavigate();
  const { isSignedIn, isLoaded } = useAuth();
  const [vehicleId, setVehicleId] = useState<number | null>(null);
  const [images, setImages] = useState<UploadedImage[]>([]);
  const [vinDecoding, setVinDecoding] = useState(false);

  // Require authentication
  if (isLoaded && !isSignedIn) {
    return (
      <div className="container mx-auto px-4 py-8 text-center">
        <h1 className="text-2xl font-bold mb-4">Sign in Required</h1>
        <p className="text-muted-foreground mb-4">You need to sign in to list a vehicle.</p>
        <Link to="/sign-in">
          <Button>Sign In</Button>
        </Link>
      </div>
    );
  }

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(vehicleFormSchema),
    defaultValues: {
      vin: '',
      year: undefined as unknown as number,
      make: '',
      model: '',
      condition: 'runs_drives' as const,
      title_type: 'clean' as const,
      has_keys: true,
      starting_price: undefined as unknown as number,
    },
  });

  const createMutation = useMutation({
    mutationFn: (data: VehicleFormData) => vehiclesApi.create(data),
    onSuccess: (response) => {
      const id = response.data.vehicle_id;
      setVehicleId(id);
    },
  });

  const submitMutation = useMutation({
    mutationFn: (id: number) => 
      fetch(`/api/vehicles/${id}/submit`, { method: 'POST' }).then(r => r.json()),
    onSuccess: () => {
      navigate('/vehicles');
    },
  });

  const decodeVin = async () => {
    const vin = watch('vin');
    if (!vin || vin.length !== 17) return;

    setVinDecoding(true);
    try {
      const response = await fetch('/api/decode-vin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ vin }),
      });
      const data = await response.json();
      
      if (data.success && data.data) {
        const v = data.data;
        if (v.year) setValue('year', v.year);
        if (v.make) setValue('make', v.make);
        if (v.model) setValue('model', v.model);
        if (v.trim) setValue('trim', v.trim);
        if (v.body_type) setValue('body_type', v.body_type);
        if (v.engine) setValue('engine', v.engine);
        if (v.transmission) setValue('transmission', v.transmission);
        if (v.drivetrain) setValue('drivetrain', v.drivetrain);
      }
    } catch (error) {
      console.error('VIN decode failed:', error);
    } finally {
      setVinDecoding(false);
    }
  };

  const onSubmit = async (data: unknown) => {
    if (!vehicleId) {
      // First create the vehicle
      createMutation.mutate(data as VehicleFormData);
    } else {
      // Submit for review
      submitMutation.mutate(vehicleId);
    }
  };

  const isCreated = vehicleId !== null;

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      <Link to="/vehicles" className="inline-flex items-center text-muted-foreground hover:text-foreground mb-6">
        <ArrowLeft className="h-4 w-4 mr-2" />
        Back to Inventory
      </Link>

      <h1 className="text-3xl font-bold mb-8">
        {isCreated ? 'Add Photos & Submit' : 'List a Vehicle'}
      </h1>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
        {/* VIN Section */}
        <Card>
          <CardHeader>
            <CardTitle>Vehicle Identification</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex gap-2">
              <div className="flex-1 space-y-2">
                <Label htmlFor="vin">VIN *</Label>
                <Input
                  id="vin"
                  {...register('vin')}
                  placeholder="Enter 17-character VIN"
                  maxLength={17}
                  className="font-mono uppercase"
                  disabled={isCreated}
                />
                {errors.vin && (
                  <p className="text-sm text-destructive">{errors.vin.message}</p>
                )}
              </div>
              <div className="flex items-end">
                <Button
                  type="button"
                  variant="outline"
                  onClick={decodeVin}
                  disabled={vinDecoding || isCreated}
                >
                  {vinDecoding ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Search className="h-4 w-4" />
                  )}
                  <span className="ml-2">Decode</span>
                </Button>
              </div>
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="space-y-2">
                <Label htmlFor="year">Year *</Label>
                <Input id="year" type="number" {...register('year')} disabled={isCreated} />
                {errors.year && <p className="text-sm text-destructive">{errors.year.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="make">Make *</Label>
                <Input id="make" {...register('make')} disabled={isCreated} />
                {errors.make && <p className="text-sm text-destructive">{errors.make.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="model">Model *</Label>
                <Input id="model" {...register('model')} disabled={isCreated} />
                {errors.model && <p className="text-sm text-destructive">{errors.model.message}</p>}
              </div>
              <div className="space-y-2">
                <Label htmlFor="trim">Trim</Label>
                <Input id="trim" {...register('trim')} disabled={isCreated} />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Specs Section */}
        <Card>
          <CardHeader>
            <CardTitle>Specifications</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="body_type">Body Type</Label>
                <Input id="body_type" {...register('body_type')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="engine">Engine</Label>
                <Input id="engine" {...register('engine')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="transmission">Transmission</Label>
                <Input id="transmission" {...register('transmission')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="drivetrain">Drivetrain</Label>
                <Input id="drivetrain" {...register('drivetrain')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="exterior_color">Exterior Color</Label>
                <Input id="exterior_color" {...register('exterior_color')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="interior_color">Interior Color</Label>
                <Input id="interior_color" {...register('interior_color')} disabled={isCreated} />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Condition Section */}
        <Card>
          <CardHeader>
            <CardTitle>Condition</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="space-y-2">
                <Label htmlFor="mileage">Mileage</Label>
                <Input id="mileage" type="number" {...register('mileage')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="condition">Condition</Label>
                <select
                  id="condition"
                  {...register('condition')}
                  className="w-full h-9 rounded-md border bg-transparent px-3"
                  disabled={isCreated}
                >
                  <option value="runs_drives">Runs & Drives</option>
                  <option value="starts">Starts Only</option>
                  <option value="non_running">Non-Running</option>
                  <option value="parts_only">Parts Only</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="title_type">Title Type</Label>
                <select
                  id="title_type"
                  {...register('title_type')}
                  className="w-full h-9 rounded-md border bg-transparent px-3"
                  disabled={isCreated}
                >
                  <option value="clean">Clean</option>
                  <option value="salvage">Salvage</option>
                  <option value="rebuilt">Rebuilt</option>
                  <option value="flood">Flood</option>
                  <option value="lemon">Lemon</option>
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="title_state">Title State</Label>
                <Input id="title_state" {...register('title_state')} maxLength={2} placeholder="MN" disabled={isCreated} />
              </div>
            </div>
            <div className="mt-4 flex items-center gap-2">
              <input
                type="checkbox"
                id="has_keys"
                {...register('has_keys')}
                className="h-4 w-4"
                disabled={isCreated}
              />
              <Label htmlFor="has_keys">Has Keys</Label>
            </div>
          </CardContent>
        </Card>

        {/* Pricing Section */}
        <Card>
          <CardHeader>
            <CardTitle>Pricing</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label htmlFor="starting_price">Starting Price *</Label>
                <Input id="starting_price" type="number" {...register('starting_price')} disabled={isCreated} />
                {errors.starting_price && (
                  <p className="text-sm text-destructive">{errors.starting_price.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="reserve_price">Reserve Price</Label>
                <Input id="reserve_price" type="number" {...register('reserve_price')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="buy_now_price">Buy Now Price</Label>
                <Input id="buy_now_price" type="number" {...register('buy_now_price')} disabled={isCreated} />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Location Section */}
        <Card>
          <CardHeader>
            <CardTitle>Location</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="col-span-2 space-y-2">
                <Label htmlFor="location_address">Address</Label>
                <Input id="location_address" {...register('location_address')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="location_city">City</Label>
                <Input id="location_city" {...register('location_city')} disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="location_state">State</Label>
                <Input id="location_state" {...register('location_state')} maxLength={2} placeholder="MN" disabled={isCreated} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="location_zip">ZIP</Label>
                <Input id="location_zip" {...register('location_zip')} disabled={isCreated} />
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Description */}
        <Card>
          <CardHeader>
            <CardTitle>Description</CardTitle>
          </CardHeader>
          <CardContent>
            <Textarea
              {...register('description')}
              placeholder="Describe the vehicle condition, features, history..."
              rows={4}
              disabled={isCreated}
            />
          </CardContent>
        </Card>

        {/* Images - only show after vehicle is created */}
        {isCreated && (
          <Card>
            <CardHeader>
              <CardTitle>Photos</CardTitle>
            </CardHeader>
            <CardContent>
              <ImageUpload
                vehicleId={vehicleId}
                images={images}
                onImagesChange={setImages}
              />
            </CardContent>
          </Card>
        )}

        {/* Submit */}
        <div className="flex gap-4">
          {!isCreated ? (
            <Button
              type="submit"
              size="lg"
              disabled={createMutation.isPending}
            >
              {createMutation.isPending && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Create Vehicle
            </Button>
          ) : (
            <Button
              type="submit"
              size="lg"
              disabled={submitMutation.isPending || images.length === 0}
            >
              {submitMutation.isPending && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Submit for Review
            </Button>
          )}
          <Button type="button" variant="outline" size="lg" onClick={() => navigate('/vehicles')}>
            Cancel
          </Button>
        </div>

        {createMutation.isError && (
          <p className="text-destructive">Failed to create vehicle. Please try again.</p>
        )}
      </form>
    </div>
  );
}
