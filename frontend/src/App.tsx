import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ClerkProvider, SignIn, SignUp } from '@clerk/clerk-react';
import Layout from '@/components/Layout';
import HomePage from '@/pages/HomePage';
import VehiclesPage from '@/pages/VehiclesPage';
import VehicleDetailPage from '@/pages/VehicleDetailPage';
import VehicleCreatePage from '@/pages/VehicleCreatePage';
import AuctionsPage from '@/pages/AuctionsPage';
import AuctionDetailPage from '@/pages/AuctionDetailPage';
import DashboardPage from '@/pages/DashboardPage';
import MyBidsPage from '@/pages/MyBidsPage';
import WatchlistPage from '@/pages/WatchlistPage';
import NotFoundPage from '@/pages/NotFoundPage';
import { Toaster } from '@/components/ui/toaster';

const CLERK_PUBLISHABLE_KEY = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
});

function App() {
  // If Clerk key not configured, show warning
  if (!CLERK_PUBLISHABLE_KEY) {
    console.warn('Clerk publishable key not found. SSO will not work.');
  }

  const app = (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<HomePage />} />
            <Route path="vehicles" element={<VehiclesPage />} />
            <Route path="vehicles/new" element={<VehicleCreatePage />} />
            <Route path="vehicles/:id" element={<VehicleDetailPage />} />
            <Route path="auctions" element={<AuctionsPage />} />
            <Route path="auctions/:id" element={<AuctionDetailPage />} />
            <Route path="sign-in/*" element={<SignIn routing="path" path="/sign-in" />} />
            <Route path="sign-up/*" element={<SignUp routing="path" path="/sign-up" />} />
            <Route path="dashboard" element={<DashboardPage />} />
            <Route path="my-bids" element={<MyBidsPage />} />
            <Route path="watchlist" element={<WatchlistPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Route>
        </Routes>
      </BrowserRouter>
      <Toaster />
    </QueryClientProvider>
  );

  // Wrap with ClerkProvider if key is available
  if (CLERK_PUBLISHABLE_KEY) {
    return (
      <ClerkProvider publishableKey={CLERK_PUBLISHABLE_KEY}>
        {app}
      </ClerkProvider>
    );
  }

  return app;
}

export default App;
