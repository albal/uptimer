import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { monitorsAPI } from '../api/client';
import { CreateMonitorRequest } from '../types';

export function useMonitors() {
  return useQuery({
    queryKey: ['monitors'],
    queryFn: monitorsAPI.list,
    refetchInterval: 30000,
  });
}

export function useMonitor(id: string) {
  return useQuery({
    queryKey: ['monitor', id],
    queryFn: () => monitorsAPI.get(id),
    enabled: !!id,
    refetchInterval: 15000,
  });
}

export function useMonitorResults(id: string, limit = 100) {
  return useQuery({
    queryKey: ['monitor-results', id, limit],
    queryFn: () => monitorsAPI.getResults(id, limit),
    enabled: !!id,
    refetchInterval: 30000,
  });
}

export function useCreateMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateMonitorRequest) => monitorsAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useDeleteMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => monitorsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function usePauseMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => monitorsAPI.pause(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}

export function useResumeMonitor() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => monitorsAPI.resume(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['monitors'] });
    },
  });
}
