/**
 * useAuth - syncs Clerk session token with API client.
 */
import { useEffect } from 'react';
import { useAuth as useClerkAuth } from '@clerk/clerk-react';
import { setAuthToken } from '@/services/api';

export function useAuth() {
  const { getToken, isSignedIn, isLoaded } = useClerkAuth();

  useEffect(() => {
    const syncToken = async () => {
      if (isLoaded && isSignedIn) {
        const token = await getToken();
        setAuthToken(token);
      } else {
        setAuthToken(null);
      }
    };
    syncToken();
  }, [getToken, isSignedIn, isLoaded]);

  return { isSignedIn, isLoaded };
}
