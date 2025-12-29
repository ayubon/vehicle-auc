import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Car, Gavel, Shield, Truck } from 'lucide-react';

export default function HomePage() {
  return (
    <div>
      {/* Hero Section */}
      <section className="bg-primary text-primary-foreground py-20">
        <div className="container mx-auto px-4 text-center">
          <h1 className="text-4xl md:text-6xl font-bold mb-6">
            Marketplace, Made Simple.
          </h1>
          <p className="text-xl md:text-2xl mb-8 opacity-90 max-w-2xl mx-auto">
            Find your next vehicle at auction prices. Browse thousands of cars, trucks, and SUVs from trusted sellers.
          </p>
          <div className="flex gap-4 justify-center">
            <Button size="lg" variant="secondary" asChild>
              <Link to="/vehicles">
                <Car className="mr-2 h-5 w-5" />
                Browse Inventory
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link to="/auctions">
                <Gavel className="mr-2 h-5 w-5" />
                Live Auctions
              </Link>
            </Button>
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-16 bg-muted/50">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-12">How It Works</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <Shield className="h-8 w-8 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-2">1. Get Verified</h3>
              <p className="text-muted-foreground">
                Create an account, verify your ID, and add a payment method.
              </p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <Gavel className="h-8 w-8 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-2">2. Bid & Win</h3>
              <p className="text-muted-foreground">
                Browse vehicles, place bids, and win auctions at great prices.
              </p>
            </div>
            <div className="text-center">
              <div className="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center mx-auto mb-4">
                <Truck className="h-8 w-8 text-primary" />
              </div>
              <h3 className="text-xl font-semibold mb-2">3. We Deliver</h3>
              <p className="text-muted-foreground">
                We handle title transfer and deliver the vehicle to your door.
              </p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
