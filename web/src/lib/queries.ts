import { useQuery } from '@tanstack/react-query';
import { api } from './api';

export function useRoles() {
  return useQuery({ queryKey: ['roles'], queryFn: () => api.getRoles() });
}
export function useAuditData() {
  return useQuery({ queryKey: ['audit'], queryFn: () => api.getAuditData() });
}
