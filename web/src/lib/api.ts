import type {
  RolesResponse,
  SyncHistoryResponse,
  SyncDetailsResponse,
  HealthResponse,
} from '@/types';

const API_BASE = '/api';

class ApiClient {
  private async fetch<T>(endpoint: string): Promise<T> {
    const response = await fetch(`${API_BASE}${endpoint}`);
    if (!response.ok) {
      throw new Error(`API error: ${response.statusText}`);
    }
    return response.json();
  }

  getRoles() {
    return this.fetch<RolesResponse>('/roles');
  }

  getSyncHistory(page = 1, pageSize = 20) {
    return this.fetch<SyncHistoryResponse>(
      `/sync/history?page=${page}&page_size=${pageSize}`
    );
  }

  getSyncDetails(id: number) {
    return this.fetch<SyncDetailsResponse>(`/sync/history/${id}`);
  }

  getHealth() {
    return this.fetch<HealthResponse>('/health');
  }
}

export const api = new ApiClient();
