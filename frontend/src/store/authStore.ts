import { create } from 'zustand';

interface AuthState {
  isAuthenticated: boolean;
  userId: string | null;
  teamId: string | null;
  loading: boolean;
  setAuth: (userId: string, teamId: string) => void;
  clearAuth: () => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  userId: null,
  teamId: null,
  loading: true,
  setAuth: (userId, teamId) =>
    set({ isAuthenticated: true, userId, teamId, loading: false }),
  clearAuth: () =>
    set({ isAuthenticated: false, userId: null, teamId: null, loading: false }),
  setLoading: (loading) => set({ loading }),
}));
