// API Response Types for GateKeep

export interface Role {
  name: string;
  parent_roles: string[];
  comment?: string;
}

export interface RolesResponse {
  roles: Role[];
}

export type SyncStatus = 'pending' | 'running' | 'success' | 'failed' | 'partial';

export interface SyncRun {
  id: number;
  sync_id: string;
  status: SyncStatus;
  started_at: string;
  completed_at?: string;
  total_operations: number;
  successful_operations: number;
  failed_operations: number;
  duration_ms: number;
}

export interface SyncHistoryResponse {
  sync_runs: SyncRun[];
  pagination: {
    page: number;
    page_size: number;
    total_count: number;
    total_pages: number;
  };
}

export type OperationStatus = 'success' | 'failed' | 'skipped';

export interface Operation {
  id: number;
  operation_type: string;
  target_object: string;
  sql_statement: string;
  status: OperationStatus;
  error_message?: string;
  execution_time_ms: number;
  executed_at: string;
}

export interface SyncDetailsResponse {
  sync_run: SyncRun;
  operations: Operation[];
}

export interface HealthResponse {
  status: 'healthy' | 'unhealthy' | 'degraded';
  snowflake_connected: boolean;
  database_connected: boolean;
  timestamp: string;
}

// Filter types
export interface LogFilters {
  status?: SyncStatus;
  operation_type?: string;
  date_from?: string;
  date_to?: string;
  search?: string;
}

// Diff types
export interface Grant {
  object_type: string;
  object_name: string;
  privilege: string;
  granted_on: string;
}

export interface PermissionDiff {
  added: Grant[];
  removed: Grant[];
  unchanged: Grant[];
}
