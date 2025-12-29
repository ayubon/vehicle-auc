import { useUser, RedirectToSignIn } from '@clerk/clerk-react';
import { Shield, CreditCard, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function DashboardPage() {
  const { isLoaded, isSignedIn, user } = useUser();

  if (!isLoaded) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p>Loading...</p>
      </div>
    );
  }

  if (!isSignedIn) {
    return <RedirectToSignIn />;
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8">
        Welcome, {user.firstName || user.emailAddresses[0]?.emailAddress}
      </h1>
      
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Verification Status */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="h-6 w-6 text-primary" />
            <h2 className="text-lg font-semibold">ID Verification</h2>
          </div>
          <div className="flex items-center gap-2 mb-4">
            <XCircle className="h-5 w-5 text-yellow-500" />
            <span className="text-yellow-600">Not Verified</span>
          </div>
          <Button variant="outline" className="w-full">Verify ID</Button>
        </div>

        {/* Payment Method */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-center gap-3 mb-4">
            <CreditCard className="h-6 w-6 text-primary" />
            <h2 className="text-lg font-semibold">Payment Method</h2>
          </div>
          <div className="flex items-center gap-2 mb-4">
            <XCircle className="h-5 w-5 text-yellow-500" />
            <span className="text-yellow-600">No card</span>
          </div>
          <Button variant="outline" className="w-full">Add Payment</Button>
        </div>

        {/* Bidding Status */}
        <div className="border rounded-lg p-6 bg-card">
          <h2 className="text-lg font-semibold mb-4">Bidding Status</h2>
          <div className="text-muted-foreground text-sm">
            <p>Complete ID verification and add a payment method to start bidding.</p>
          </div>
        </div>

        {/* Account Info from Clerk */}
        <div className="border rounded-lg p-6 bg-card md:col-span-2 lg:col-span-3">
          <h2 className="text-lg font-semibold mb-4">Account Info</h2>
          <div className="grid gap-2 text-sm">
            <p><strong>Email:</strong> {user.emailAddresses[0]?.emailAddress}</p>
            <p><strong>Name:</strong> {user.fullName || 'Not set'}</p>
            <p><strong>Clerk ID:</strong> {user.id}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
