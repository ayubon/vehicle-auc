/**
 * Form-related types for vehicle creation/editing.
 */
import { z } from 'zod';

export const vehicleFormSchema = z.object({
  vin: z.string().length(17, 'VIN must be exactly 17 characters'),
  year: z.coerce.number().min(1900).max(new Date().getFullYear() + 1),
  make: z.string().min(1, 'Make is required'),
  model: z.string().min(1, 'Model is required'),
  trim: z.string().optional(),
  body_type: z.string().optional(),
  engine: z.string().optional(),
  transmission: z.string().optional(),
  drivetrain: z.string().optional(),
  exterior_color: z.string().optional(),
  interior_color: z.string().optional(),
  mileage: z.coerce.number().min(0).optional().or(z.literal('')),
  condition: z.enum(['runs_drives', 'starts', 'non_running', 'parts_only']),
  title_type: z.enum(['clean', 'salvage', 'rebuilt', 'flood', 'lemon']),
  title_state: z.string().max(2).optional(),
  has_keys: z.boolean(),
  description: z.string().optional(),
  starting_price: z.coerce.number().min(1, 'Starting price is required'),
  reserve_price: z.coerce.number().optional().or(z.literal('')),
  buy_now_price: z.coerce.number().optional().or(z.literal('')),
  location_address: z.string().optional(),
  location_city: z.string().optional(),
  location_state: z.string().max(2).optional(),
  location_zip: z.string().optional(),
});

export type VehicleFormData = z.infer<typeof vehicleFormSchema>;
