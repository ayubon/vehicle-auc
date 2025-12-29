import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';

export default function LoginPage() {
  return (
    <div className="container mx-auto px-4 py-16 max-w-md">
      <div className="border rounded-lg p-6 bg-card">
        <h1 className="text-2xl font-bold text-center mb-6">Log In</h1>
        <form className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Email</label>
            <input type="email" className="w-full border rounded-md px-3 py-2" placeholder="you@example.com" />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input type="password" className="w-full border rounded-md px-3 py-2" placeholder="••••••••" />
          </div>
          <Button className="w-full">Log In</Button>
        </form>
        <p className="text-center mt-4 text-sm text-muted-foreground">
          Don't have an account? <Link to="/register" className="text-primary hover:underline">Sign up</Link>
        </p>
      </div>
    </div>
  );
}
