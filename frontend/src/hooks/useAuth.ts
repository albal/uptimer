import { useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { authAPI } from '../api/client';

export function useAuth() {
  const { isAuthenticated, userId, teamId, loading, setAuth, clearAuth, setLoading } = useAuthStore();

  useEffect(() => {
    checkAuth();
  }, []);

  async function checkAuth() {
    setLoading(true);
    try {
      const data = await authAPI.getMe();
      setAuth(data.user_id, data.team_id);
    } catch {
      clearAuth();
    }
  }

  async function logout() {
    try {
      await authAPI.logout();
    } catch {
      // ignore
    }
    clearAuth();
    window.location.href = '/login';
  }

  return { isAuthenticated, userId, teamId, loading, logout, checkAuth };
}
