import { useAuthStore } from '@/store/auth';
import { Shield, CreditCard, CheckCircle, XCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

export default function DashboardPage() {
  const { user } = useAuthStore();

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p>Please log in to view your dashboard.</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-8">Welcome, {user.first_name || user.email}</h1>
      
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
        {/* Verification Status */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-center gap-3 mb-4">
            <Shield className="h-6 w-6 text-primary" />
            <h2 className="text-lg font-semibold">ID Verification</h2>
          </div>
          <div className="flex items-center gap-2 mb-4">
            {user.is_id_verified ? (
              <>
                <CheckCircle className="h-5 w-5 text-green-500" />
                <span className="text-green-600">Verified</span>
              </>
            ) : (
              <>
                <XCircle className="h-5 w-5 text-yellow-500" />
                <span className="text-yellow-600">Not Verified</span>
              </>
            )}
          </div>
          {!user.is_id_verified && (
            <Button variant="outline" className="w-full">Verify ID</Button>
          )}
        </div>

        {/* Payment Method */}
        <div className="border rounded-lg p-6 bg-card">
          <div className="flex items-center gap-3 mb-4">
            <CreditCard className="h-6 w-6 text-primary" />
            <h2 className="text-lg font-semibold">Payment Method</h2>
          </div>
          <div className="flex items-center gap-2 mb-4">
            {user.has_payment_method ? (
              <>
                <CheckCircle className="h-5 w-5 text-green-500" />
                <span className="text-green-600">Card on file</span>
              </>
            ) : (
              <>
                <XCircle className="h-5 w-5 text-yellow-500" />
                <span className="text-yellow-600">No card</span>
              </>
            )}
          </div>
          {!user.has_payment_method && (
            <Button variant="outline" className="w-full">Add Payment</Button>
          )}
        </div>

        {/* Bidding Status */}
        <div className="border rounded-lg p-6 bg-card">
          <h2 className="text-lg font-semibold mb-4">Bidding Status</h2>
          {user.can_bid ? (
            <div className="flex items-center gap-2 text-green-600">
              <CheckCircle className="h-5 w-5" />
              <span>Ready to bid!</span>
            </div>
          ) : (
            <div className="text-muted-foreground text-sm">
              <p>Complete ID verification and add a payment method to start bidding.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
