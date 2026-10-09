import type { Role, Grant, SyncRun, Operation, HealthResponse } from '@/types';

export interface HierarchyResponse {
  roles: Role[];
  pagination: { total_pages: number };
}
export interface RefreshStatus {
  status: 'idle' | 'refreshing' | 'success' | 'failed';
  last_refresh: string | null;
  error_message: string | null;
}

class ApiClient {
  private async fetch<T>(endpoint: string, method = 'GET'): Promise<T> {
    const response = await fetch(`/api${endpoint}`, { method });
    if (!response.ok) {
      const body = await response.json().catch(() => null);
      throw new Error(body?.message ?? `API request failed (${response.status})`);
    }
    return response.json();
  }

  async getRoles() {
    const roles: Role[] = [];
    let page = 1;
    let pages = 1;
    do {
      const result = await this.fetch<HierarchyResponse>(`/roles/hierarchy?page=${page}&page_size=200`);
      roles.push(...result.roles.map(role => ({ ...role, parent_roles: role.parent_roles ?? [] })));
      pages = result.pagination.total_pages;
      page++;
    } while (page <= pages);
    return roles;
  }

  async getAuditData() {
    const result = await this.fetch<{ sync_runs: SyncRun[] }>('/sync/history?page=1&page_size=100');
    const runs = result.sync_runs.map(run => ({ ...run, duration_ms: run.duration_ms ?? 0 }));
    const operations: Operation[] = [];
    // Bound concurrent requests while retrieving the recent runs' operation details.
    for (let i = 0; i < runs.length; i += 5) {
      const details = await Promise.all(runs.slice(i, i + 5).map(run => this.fetch<SyncRun & { operations: Operation[] }>(`/sync/history/${run.id}`)));
      operations.push(...details.flatMap(detail => detail.operations.map(op => ({ ...op, execution_time_ms: op.execution_time_ms ?? 0, executed_at: op.executed_at ?? detail.started_at }))));
    }
    operations.sort((a, b) => b.executed_at.localeCompare(a.executed_at));
    return { runs, operations };
  }

  async compareRoles(roleA: string, roleB: string) {
    type CachedGrant = Omit<Grant, 'object_type'>;
    const result = await this.fetch<{ diff: Record<'added' | 'removed' | 'unchanged', CachedGrant[]> }>(`/roles/compare?${new URLSearchParams({ roleA, roleB })}`);
    const convert = (grants: CachedGrant[]) => grants.map(grant => ({ ...grant, object_type: grant.granted_on }));
    return { added: convert(result.diff.added), removed: convert(result.diff.removed), unchanged: convert(result.diff.unchanged) };
  }

  refreshRoles() { return this.fetch('/roles/refresh', 'POST'); }
  getRefreshStatus() { return this.fetch<RefreshStatus>('/roles/refresh/status'); }
  getHealth() { return this.fetch<HealthResponse>('/health'); }
}

export const api = new ApiClient();
