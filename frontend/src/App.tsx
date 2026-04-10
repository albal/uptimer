import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Layout } from './components/layout/Layout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Monitors } from './pages/Monitors';
import { MonitorDetail } from './pages/MonitorDetail';
import { Incidents } from './pages/Incidents';
import { StatusPages } from './pages/StatusPages';
import { Integrations } from './pages/Integrations';
import { Settings } from './pages/Settings';
import { Maintenance } from './pages/Maintenance';
import { PublicStatusPage } from './pages/PublicStatusPage';
import { useAuth } from './hooks/useAuth';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 10000,
      retry: 1,
    },
  },
});

function ProtectedRoutes() {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100vh' }}>
        <div className="spinner" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <Layout />;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/status/:slug" element={<PublicStatusPage />} />
          <Route path="/*" element={<ProtectedRoutes />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="monitors" element={<Monitors />} />
            <Route path="monitors/:id" element={<MonitorDetail />} />
            <Route path="incidents" element={<Incidents />} />
            <Route path="status-pages" element={<StatusPages />} />
            <Route path="integrations" element={<Integrations />} />
            <Route path="maintenance" element={<Maintenance />} />
            <Route path="settings" element={<Settings />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
