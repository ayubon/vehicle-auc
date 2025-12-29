# Frontend Onboarding Guide

Welcome to the Vehicle Auction Platform frontend! This guide will get you up to speed quickly.

---

## Quick Start

```bash
cd frontend

# Install dependencies
npm install

# Start dev server (port 3000)
npm run dev
```

**Verify it's working:** Open http://localhost:3000

> **Note:** The backend must be running on port 5001 for API calls to work.

---

## Tech Stack

| Technology | Purpose |
|------------|---------|
| **React 18** | UI framework |
| **TypeScript** | Type safety |
| **Vite** | Build tool & dev server |
| **Tailwind CSS v4** | Utility-first styling |
| **shadcn/ui** | Pre-built accessible components |
| **React Router** | Client-side routing |
| **TanStack Query** | Server state management & caching |
| **React Hook Form + Zod** | Form handling & validation |
| **Axios** | HTTP client |
| **Clerk** | Authentication (SSO) |
| **Lucide React** | Icons |

---

## Directory Structure

```
frontend/src/
├── main.tsx                 # Entry point
├── App.tsx                  # Routes + providers
├── index.css                # Tailwind CSS
│
├── components/
│   ├── ui/                  # shadcn/ui components
│   │   ├── button.tsx
│   │   ├── card.tsx
│   │   ├── input.tsx
│   │   └── ...
│   ├── Layout.tsx           # App layout with nav
│   └── ImageUpload.tsx      # Drag-and-drop S3 upload
│
├── pages/
│   ├── HomePage.tsx
│   ├── VehiclesPage.tsx     # Inventory grid + filters
│   ├── VehicleDetailPage.tsx
│   ├── VehicleCreatePage.tsx
│   ├── AuctionsPage.tsx
│   ├── AuctionDetailPage.tsx
│   └── DashboardPage.tsx
│
├── hooks/
│   ├── index.ts             # Exports all hooks
│   ├── useVehicles.ts       # List + filter vehicles
│   ├── useVehicle.ts        # Single vehicle fetch
│   └── useAuth.ts           # Clerk → Flask JWT sync
│
├── services/
│   └── api.ts               # Axios client + API functions
│
├── types/
│   ├── index.ts             # Exports all types
│   ├── vehicle.ts           # Vehicle interfaces
│   └── form.ts              # Zod form schemas
│
└── lib/
    └── utils.ts             # Utility functions (cn, etc.)
```

---

## Key Concepts

### 1. Component Library (shadcn/ui)

We use **shadcn/ui** — copy-paste components built on Radix UI + Tailwind.

```tsx
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';

function Example() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Hello</CardTitle>
      </CardHeader>
      <CardContent>
        <Input placeholder="Type here..." />
        <Button>Submit</Button>
      </CardContent>
    </Card>
  );
}
```

**Adding new components:**
```bash
npx shadcn@latest add dialog
npx shadcn@latest add dropdown-menu
```

### 2. Routing (React Router)

Routes are defined in `App.tsx`:

```tsx
<Routes>
  <Route path="/" element={<HomePage />} />
  <Route path="/vehicles" element={<VehiclesPage />} />
  <Route path="/vehicles/:id" element={<VehicleDetailPage />} />
  <Route path="/vehicles/new" element={<VehicleCreatePage />} />
</Routes>
```

**Navigation:**
```tsx
import { Link, useNavigate } from 'react-router-dom';

// Declarative
<Link to="/vehicles/123">View Vehicle</Link>

// Programmatic
const navigate = useNavigate();
navigate('/vehicles');
```

### 3. Data Fetching (TanStack Query)

We use TanStack Query for server state:

```tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { vehiclesApi } from '@/services/api';

// Fetching data
function VehicleDetail({ id }: { id: number }) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['vehicle', id],
    queryFn: () => vehiclesApi.getById(id),
  });

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading vehicle</div>;
  
  return <div>{data.make} {data.model}</div>;
}

// Mutating data
function CreateVehicle() {
  const queryClient = useQueryClient();
  
  const mutation = useMutation({
    mutationFn: vehiclesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['vehicles'] });
    },
  });

  const handleSubmit = (data) => {
    mutation.mutate(data);
  };
}
```

### 4. Forms (React Hook Form + Zod)

Forms use React Hook Form with Zod validation:

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';

// Define schema
const schema = z.object({
  make: z.string().min(1, 'Make is required'),
  model: z.string().min(1, 'Model is required'),
  year: z.coerce.number().min(1900).max(2030),
  price: z.coerce.number().positive(),
});

type FormData = z.infer<typeof schema>;

// Use in component
function VehicleForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = (data: FormData) => {
    console.log(data);
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Input {...register('make')} placeholder="Make" />
      {errors.make && <span>{errors.make.message}</span>}
      
      <Input {...register('year')} type="number" placeholder="Year" />
      {errors.year && <span>{errors.year.message}</span>}
      
      <Button type="submit">Submit</Button>
    </form>
  );
}
```

### 5. Authentication (Clerk)

We use Clerk for SSO, synced to Flask JWT:

```tsx
import { useAuth } from '@/hooks';
import { SignInButton, UserButton } from '@clerk/clerk-react';

function Header() {
  const { isSignedIn, isLoaded } = useAuth();

  if (!isLoaded) return null;

  return (
    <nav>
      {isSignedIn ? (
        <UserButton />
      ) : (
        <SignInButton mode="modal">
          <Button>Sign In</Button>
        </SignInButton>
      )}
    </nav>
  );
}
```

**Auth flow:**
1. User clicks Sign In → Clerk modal opens
2. User authenticates via Google/GitHub
3. `useAuth` hook syncs Clerk user to backend (`/api/auth/clerk-sync`)
4. Backend returns Flask JWT token
5. Token is stored and sent with all API requests

### 6. API Client

All API calls go through `services/api.ts`:

```tsx
import { vehiclesApi } from '@/services/api';

// List vehicles with filters
const vehicles = await vehiclesApi.getAll({ make: 'Toyota', year_min: 2020 });

// Get single vehicle
const vehicle = await vehiclesApi.getById(123);

// Create vehicle (requires auth)
const newVehicle = await vehiclesApi.create({ vin: '...', make: 'Toyota', ... });

// Upload image
const { upload_url, s3_key } = await vehiclesApi.getUploadUrl(vehicleId, 'photo.jpg', 'image/jpeg');
```

---

## Common Tasks

### Adding a New Page

1. **Create the page** in `pages/`:
```tsx
// pages/MyNewPage.tsx
export function MyNewPage() {
  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold">My New Page</h1>
    </div>
  );
}
```

2. **Add the route** in `App.tsx`:
```tsx
import { MyNewPage } from '@/pages/MyNewPage';

<Route path="/my-new-page" element={<MyNewPage />} />
```

3. **Add nav link** in `Layout.tsx` (if needed).

### Adding a New Hook

1. **Create the hook** in `hooks/`:
```tsx
// hooks/useMyData.ts
import { useQuery } from '@tanstack/react-query';
import { api } from '@/services/api';

export function useMyData(id: number) {
  return useQuery({
    queryKey: ['myData', id],
    queryFn: async () => {
      const { data } = await api.get(`/my-endpoint/${id}`);
      return data;
    },
  });
}
```

2. **Export it** in `hooks/index.ts`:
```tsx
export { useMyData } from './useMyData';
```

### Adding a New API Function

In `services/api.ts`:

```tsx
export const myApi = {
  getAll: async () => {
    const { data } = await api.get('/my-endpoint');
    return data;
  },
  
  create: async (payload: MyType) => {
    const { data } = await api.post('/my-endpoint', payload);
    return data;
  },
};
```

### Adding a shadcn/ui Component

```bash
# See available components
npx shadcn@latest add --help

# Add specific component
npx shadcn@latest add dialog
npx shadcn@latest add toast
npx shadcn@latest add tabs
```

Components are added to `components/ui/`.

---

## Styling with Tailwind

### Basic Usage

```tsx
<div className="flex items-center gap-4 p-4 bg-white rounded-lg shadow">
  <span className="text-lg font-semibold text-gray-900">Title</span>
  <span className="text-sm text-gray-500">Subtitle</span>
</div>
```

### Common Patterns

```tsx
// Responsive
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">

// Hover/Focus states
<button className="bg-blue-500 hover:bg-blue-600 focus:ring-2">

// Conditional classes (use cn utility)
import { cn } from '@/lib/utils';

<div className={cn(
  "p-4 rounded",
  isActive && "bg-blue-100",
  isError && "border-red-500"
)}>
```

### Theme Colors

We use CSS variables for theming (defined in `index.css`):

```tsx
// Primary color
<Button className="bg-primary text-primary-foreground">

// Muted text
<span className="text-muted-foreground">

// Destructive (red)
<Button variant="destructive">Delete</Button>
```

---

## TypeScript Types

### Defining Types

```tsx
// types/vehicle.ts
export interface Vehicle {
  id: number;
  vin: string;
  year: number;
  make: string;
  model: string;
  starting_price: number;
  images: VehicleImage[];
}

export interface VehicleImage {
  id: number;
  url: string;
  is_primary: boolean;
}

export interface VehicleFilters {
  make?: string;
  year_min?: number;
  year_max?: number;
  max_price?: number;
}
```

### Using Types

```tsx
import { Vehicle, VehicleFilters } from '@/types';

function VehicleCard({ vehicle }: { vehicle: Vehicle }) {
  return <div>{vehicle.year} {vehicle.make} {vehicle.model}</div>;
}
```

---

## File Upload Pattern

The `ImageUpload` component handles S3 uploads:

```tsx
import { ImageUpload } from '@/components/ImageUpload';

function VehicleForm({ vehicleId }: { vehicleId: number }) {
  const [images, setImages] = useState([]);

  return (
    <ImageUpload
      vehicleId={vehicleId}
      images={images}
      onImagesChange={setImages}
      maxImages={20}
    />
  );
}
```

**How it works:**
1. User drops/selects files
2. Component calls `POST /api/vehicles/{id}/upload-url` to get presigned S3 URL
3. Component uploads file directly to S3
4. Component calls `POST /api/vehicles/{id}/images` to register the image
5. Image appears in the grid

---

## Development Tips

### Hot Module Replacement

Vite provides instant updates. Just save a file and see changes immediately.

### React DevTools

Install the React DevTools browser extension to inspect component state.

### Network Tab

Use browser DevTools Network tab to debug API calls.

### Console Errors

Check browser console for React errors and API failures.

---

## Common Issues

| Issue | Solution |
|-------|----------|
| `CORS error` | Backend not running or wrong port |
| `401 Unauthorized` | Not signed in, or JWT expired |
| `Module not found` | Check import path, use `@/` alias |
| Styles not applying | Check Tailwind class names |
| Form not submitting | Check Zod validation errors |

---

## Build for Production

```bash
npm run build
```

Creates optimized static files in `dist/`:
- `index.html`
- `assets/index-xxx.js` (minified JS)
- `assets/index-xxx.css` (minified CSS)

Deploy `dist/` to any static host (Netlify, Vercel, S3, etc.).

---

## Questions?

- Check existing pages for patterns (especially `VehicleCreatePage.tsx`)
- Look at `services/api.ts` for API call examples
- Run `npm run dev` and experiment!
