/**
 * useAuth - syncs Clerk user with backend and sets up auth token.
 * Uses Clerk's JWT token directly for API authentication.
 */
import { useEffect, useState, useCallback } from 'react';
import { useAuth as useClerkAuth, useUser } from '@clerk/clerk-react';
import { setAuthToken } from '@/services/api';

export function useAuth() {
  const { isSignedIn, isLoaded, getToken } = useClerkAuth();
  const { user } = useUser();
  const [synced, setSynced] = useState(false);

  // Update auth token whenever needed
  const updateToken = useCallback(async () => {
    if (isSignedIn) {
      const token = await getToken();
      setAuthToken(token);
    } else {
      setAuthToken(null);
    }
  }, [isSignedIn, getToken]);

  useEffect(() => {
    const syncWithBackend = async () => {
      if (isLoaded && isSignedIn && user && !synced) {
        try {
          // Get Clerk token and set it for API calls
          const token = await getToken();
          console.log('[useAuth] Got Clerk token:', token ? `${token.substring(0, 50)}...` : 'null');
          setAuthToken(token);

          // Sync Clerk user with backend database
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
  }, [isSignedIn, isLoaded, user, synced, getToken]);

  // Refresh token periodically (Clerk tokens expire)
  useEffect(() => {
    if (!isSignedIn) return;
    
    const interval = setInterval(updateToken, 50000); // Refresh every 50s
    return () => clearInterval(interval);
  }, [isSignedIn, updateToken]);

  return { isSignedIn, isLoaded, synced, updateToken };
}
