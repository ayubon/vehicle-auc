/**
 * useAuth - syncs Clerk user with backend and gets Flask JWT token.
 */
import { useEffect, useState } from 'react';
import { useAuth as useClerkAuth, useUser } from '@clerk/clerk-react';
import { setAuthToken } from '@/services/api';

export function useAuth() {
  const { isSignedIn, isLoaded } = useClerkAuth();
  const { user } = useUser();
  const [synced, setSynced] = useState(false);

  useEffect(() => {
    const syncWithBackend = async () => {
      if (isLoaded && isSignedIn && user && !synced) {
        try {
          // Sync Clerk user with backend and get Flask JWT
          const response = await fetch('/api/auth/clerk-sync', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              clerk_user_id: user.id,
              email: user.primaryEmailAddress?.emailAddress,
              first_name: user.firstName || '',
              last_name: user.lastName || '',
            }),
          });
          
          if (response.ok) {
            const data = await response.json();
            setAuthToken(data.access_token);
            setSynced(true);
          }
        } catch (error) {
          console.error('Failed to sync with backend:', error);
        }
      } else if (!isSignedIn) {
        setAuthToken(null);
        setSynced(false);
      }
    };
    syncWithBackend();
  }, [isSignedIn, isLoaded, user, synced]);

  return { isSignedIn, isLoaded, synced };
}
