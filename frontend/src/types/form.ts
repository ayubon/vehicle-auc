/**
 * Form-related types for vehicle creation/editing.
 */
import { z } from 'zod';

export const vehicleFormSchema = z.object({
  vin: z.string().length(17, 'VIN must be exactly 17 characters'),
  year: z.number().min(1900).max(new Date().getFullYear() + 1),
  make: z.string().min(1, 'Make is required'),
  model: z.string().min(1, 'Model is required'),
  trim: z.string().optional().or(z.literal('')),
  body_type: z.string().optional().or(z.literal('')),
  engine: z.string().optional().or(z.literal('')),
  transmission: z.string().optional().or(z.literal('')),
  drivetrain: z.string().optional().or(z.literal('')),
  exterior_color: z.string().optional().or(z.literal('')),
  interior_color: z.string().optional().or(z.literal('')),
  mileage: z.number().min(0).optional(),
  condition: z.enum(['runs_drives', 'starts', 'non_running', 'parts_only']),
  title_type: z.enum(['clean', 'salvage', 'rebuilt', 'flood', 'lemon']),
  title_state: z.string().max(2).optional().or(z.literal('')),
  has_keys: z.boolean(),
  description: z.string().optional().or(z.literal('')),
  starting_price: z.number().min(1, 'Starting price is required'),
  reserve_price: z.number().optional(),
  buy_now_price: z.number().optional(),
  location_address: z.string().optional().or(z.literal('')),
  location_city: z.string().optional().or(z.literal('')),
  location_state: z.string().max(2).optional().or(z.literal('')),
  location_zip: z.string().optional().or(z.literal('')),
});

export type VehicleFormData = z.infer<typeof vehicleFormSchema>;
