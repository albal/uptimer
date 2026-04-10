import { Monitor, MonitorResult, Incident, AlertContact, StatusPage, MaintenanceWindow, Team, TeamMember, APIKey, CreateMonitorRequest } from '../types';

const API_BASE = '/api';

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (res.status === 401) {
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

// Auth
export const authAPI = {
  getProviders: () => request<{ providers: string[] }>('/auth/providers'),
  getMe: () => request<{ user_id: string; team_id: string }>('/auth/me'),
  logout: () => request<{ message: string }>('/auth/logout', { method: 'POST' }),
};

// Monitors
export const monitorsAPI = {
  list: () => request<Monitor[]>('/monitors'),
  get: (id: string) => request<Monitor>(`/monitors/${id}`),
  create: (data: CreateMonitorRequest) =>
    request<Monitor>('/monitors', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: CreateMonitorRequest) =>
    request<Monitor>(`/monitors/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) =>
    request<{ message: string }>(`/monitors/${id}`, { method: 'DELETE' }),
  pause: (id: string) =>
    request<{ message: string }>(`/monitors/${id}/pause`, { method: 'POST' }),
  resume: (id: string) =>
    request<{ message: string }>(`/monitors/${id}/resume`, { method: 'POST' }),
  getResults: (id: string, limit = 100, offset = 0) =>
    request<{ results: MonitorResult[]; total_count: number; has_more: boolean }>(
      `/monitors/${id}/results?limit=${limit}&offset=${offset}`
    ),
};

// Incidents
export const incidentsAPI = {
  list: (limit = 50, offset = 0) =>
    request<{ incidents: Incident[]; total_count: number; has_more: boolean }>(
      `/incidents?limit=${limit}&offset=${offset}`
    ),
};

// Alert Contacts
export const alertContactsAPI = {
  list: () => request<AlertContact[]>('/alert-contacts'),
  create: (data: Partial<AlertContact>) =>
    request<AlertContact>('/alert-contacts', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<AlertContact>) =>
    request<AlertContact>(`/alert-contacts/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) =>
    request<{ message: string }>(`/alert-contacts/${id}`, { method: 'DELETE' }),
};

// Status Pages
export const statusPagesAPI = {
  list: () => request<StatusPage[]>('/status-pages'),
  get: (id: string) => request<StatusPage>(`/status-pages/${id}`),
  create: (data: Partial<StatusPage>) =>
    request<StatusPage>('/status-pages', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<StatusPage>) =>
    request<StatusPage>(`/status-pages/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) =>
    request<{ message: string }>(`/status-pages/${id}`, { method: 'DELETE' }),
  setMonitors: (id: string, monitorIds: string[]) =>
    request<{ message: string }>(`/status-pages/${id}/monitors`, {
      method: 'PUT',
      body: JSON.stringify({ monitor_ids: monitorIds }),
    }),
};

// Maintenance Windows
export const maintenanceAPI = {
  list: () => request<MaintenanceWindow[]>('/maintenance-windows'),
  create: (data: Partial<MaintenanceWindow>) =>
    request<MaintenanceWindow>('/maintenance-windows', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  delete: (id: string) =>
    request<{ message: string }>(`/maintenance-windows/${id}`, { method: 'DELETE' }),
};

// Team
export const teamAPI = {
  get: () => request<Team>('/team'),
  listMembers: () => request<TeamMember[]>('/team/members'),
  inviteMember: (email: string, role: string) =>
    request<{ message: string }>('/team/members', {
      method: 'POST',
      body: JSON.stringify({ email, role }),
    }),
  removeMember: (userId: string) =>
    request<{ message: string }>(`/team/members/${userId}`, { method: 'DELETE' }),
};

// API Keys
export const apiKeysAPI = {
  list: () => request<APIKey[]>('/api-keys'),
  create: (name: string, scopes: string[]) =>
    request<APIKey & { key: string }>('/api-keys', {
      method: 'POST',
      body: JSON.stringify({ name, scopes }),
    }),
  delete: (id: string) =>
    request<{ message: string }>(`/api-keys/${id}`, { method: 'DELETE' }),
};

// Public Status Page
export const publicAPI = {
  getStatusPage: (slug: string) =>
    fetch(`/status/${slug}`).then(r => r.json()),
};
