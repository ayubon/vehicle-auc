import { Outlet, Link } from 'react-router-dom';
import { Car, LogIn } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SignedIn, SignedOut, UserButton } from '@clerk/clerk-react';

export default function Layout() {

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="border-b bg-background sticky top-0 z-50">
        <div className="container mx-auto px-4 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 font-bold text-xl">
            <Car className="h-6 w-6" />
            <span>Vehicle Auction</span>
          </Link>

          <nav className="hidden md:flex items-center gap-6">
            <Link to="/vehicles" className="text-muted-foreground hover:text-foreground transition-colors">
              Inventory
            </Link>
            <Link to="/auctions" className="text-muted-foreground hover:text-foreground transition-colors">
              Auctions
            </Link>
          </nav>

          <div className="flex items-center gap-2">
            <SignedIn>
              <Button variant="ghost" asChild>
                <Link to="/dashboard">Dashboard</Link>
              </Button>
              <UserButton afterSignOutUrl="/" />
            </SignedIn>
            <SignedOut>
              <Button variant="ghost" asChild>
                <Link to="/sign-in">
                  <LogIn className="h-4 w-4 mr-2" />
                  Log In
                </Link>
              </Button>
              <Button asChild>
                <Link to="/sign-up">Sign Up</Link>
              </Button>
            </SignedOut>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1">
        <Outlet />
      </main>

      {/* Footer */}
      <footer className="border-t py-8 bg-muted/50">
        <div className="container mx-auto px-4 text-center text-muted-foreground">
          <p>&copy; {new Date().getFullYear()} Vehicle Auction. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
