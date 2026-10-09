import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, X } from 'lucide-react';
import type { LogFilters as LogFiltersType } from '@/types';

interface LogFiltersProps {
  filters: LogFiltersType;
  onFiltersChange: (filters: LogFiltersType) => void;
}

export function LogFilters({ filters, onFiltersChange }: LogFiltersProps) {
  const statuses = ['success', 'failed', 'partial'] as const;
  const operationTypes = ['CREATE_ROLE', 'GRANT', 'REVOKE'] as const;

  const handleStatusToggle = (status: typeof statuses[number]) => {
    onFiltersChange({
      ...filters,
      status: filters.status === status ? undefined : status,
    });
  };

  const handleOperationTypeToggle = (type: string) => {
    onFiltersChange({
      ...filters,
      operation_type: filters.operation_type === type ? undefined : type,
    });
  };

  const handleClearFilters = () => {
    onFiltersChange({});
  };

  const hasActiveFilters = filters.status || filters.operation_type || filters.search;

  return (
    <div className="flex flex-wrap items-end gap-4 rounded-lg border border-slate-200 bg-white p-4">
      <div className="min-w-[240px] flex-1">
        <label htmlFor="log-search" className="mb-2 block text-sm font-medium text-slate-700">Search</label>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <Input
            id="log-search"
            placeholder="Search by role or object..."
            value={filters.search || ''}
            onChange={(e) =>
              onFiltersChange({ ...filters, search: e.target.value || undefined })
            }
            className="pl-10"
          />
        </div>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-slate-700">Status</label>
        <div className="flex flex-wrap gap-2">
          {statuses.map((status) => (
            <Button
              key={status}
              variant={filters.status === status ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleStatusToggle(status)}
            >
              {status.charAt(0).toUpperCase() + status.slice(1)}
            </Button>
          ))}
        </div>
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-slate-700">Operation Type</label>
        <div className="flex flex-wrap gap-2">
          {operationTypes.map((type) => (
            <Button
              key={type}
              variant={filters.operation_type === type ? 'default' : 'outline'}
              size="sm"
              onClick={() => handleOperationTypeToggle(type)}
            >
              {type}
            </Button>
          ))}
        </div>
      </div>
      {hasActiveFilters && (
        <Button variant="ghost" size="sm" onClick={handleClearFilters}>
          <X className="mr-1 h-4 w-4" />Clear
        </Button>
      )}
    </div>
  );
}
