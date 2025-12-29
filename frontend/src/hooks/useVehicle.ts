/**
 * useVehicle - fetch a single vehicle by ID.
 */
import { useQuery } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';
import type { VehicleDetail } from '@/types';

export function useVehicle(id: number | undefined) {
  const query = useQuery({
    queryKey: ['vehicle', id],
    queryFn: () => vehiclesApi.get(id!),
    enabled: !!id,
  });

  return {
    vehicle: query.data?.data as VehicleDetail | undefined,
    isLoading: query.isLoading,
    error: query.error,
  };
}
