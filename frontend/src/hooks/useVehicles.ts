/**
 * useVehicles - encapsulates vehicle list fetching and filtering logic.
 * Separates "how to get data" from "how to display it".
 */
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';
import type { Vehicle, VehicleFilters } from '@/types';

export function useVehicles(filters: VehicleFilters) {
  const query = useQuery({
    queryKey: ['vehicles'],
    queryFn: () => vehiclesApi.list({}),
  });

  // Client-side filtering for instant response
  const vehicles = useMemo(() => {
    const allVehicles: Vehicle[] = query.data?.data?.vehicles || [];
    return allVehicles.filter((v) => {
      if (filters.make && !v.make.toLowerCase().includes(filters.make.toLowerCase())) {
        return false;
      }
      if (filters.year_min && v.year < parseInt(filters.year_min)) {
        return false;
      }
      if (filters.year_max && v.year > parseInt(filters.year_max)) {
        return false;
      }
      if (filters.price_max && v.starting_price && v.starting_price > parseFloat(filters.price_max)) {
        return false;
      }
      return true;
    });
  }, [query.data, filters]);

  return {
    vehicles,
    total: vehicles.length,
    isLoading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
  };
}
