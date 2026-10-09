import { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useRoles } from '@/lib/queries';
import { QueryState } from '@/components/ui/QueryState';
import { RoleGraph } from './RoleGraph';
import { RoleTreeView } from './RoleTreeView';
import { RoleDetails } from './RoleDetails';
import { Button } from '@/components/ui/button';
import type { Role } from '@/types';
import { Network, List } from 'lucide-react';

export function RoleHierarchy() {
  const query = useRoles();
  const client = useQueryClient();
  const status = useQuery({ queryKey: ['refresh-status'], queryFn: () => api.getRefreshStatus(), refetchInterval: 2000 });
  useEffect(() => {
    if (status.data?.status === 'success') void client.invalidateQueries({ queryKey: ['roles'] });
  }, [client, status.data?.status, status.data?.last_refresh]);
  const refresh = useMutation({ mutationFn: () => api.refreshRoles(), onSuccess: () => {
    void client.invalidateQueries({ queryKey: ['refresh-status'] });
  } });
  const roles = query.data ?? [];
  const [view, setView] = useState<'graph' | 'tree'>('graph');
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold text-slate-900">Role Hierarchy</h2>
          <p className="mt-2 text-slate-600">
            Explore role relationships and inheritance
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={view === 'graph' ? 'default' : 'outline'}
            onClick={() => setView('graph')}
          >
            <Network className="mr-2 h-4 w-4" />
            Graph View
          </Button>
          <Button
            variant={view === 'tree' ? 'default' : 'outline'}
            onClick={() => setView('tree')}
          >
            <List className="mr-2 h-4 w-4" />
            Tree View
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Button onClick={() => refresh.mutate()} disabled={refresh.isPending || status.data?.status === 'refreshing'}>Refresh from Snowflake</Button>
        <span>{status.data?.status}{status.data?.last_refresh ? ` · ${new Date(status.data.last_refresh).toLocaleString()}` : ''}</span>
      </div>
      {(refresh.error || status.error || status.data?.error_message) && <p role="alert">{refresh.error?.message ?? status.error?.message ?? status.data?.error_message}</p>}
      {(query.isPending || query.isError) && <QueryState pending={query.isPending} error={query.error} />}
      {!query.isPending && !query.isError && roles.length === 0 && <p>No cached roles yet. Refresh from Snowflake to load them.</p>}
      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2">
          {view === 'graph' ? (
            <RoleGraph roles={roles} onRoleClick={setSelectedRole} />
          ) : (
            <RoleTreeView roles={roles} onRoleClick={setSelectedRole} />
          )}
        </div>
        <div className="col-span-1">
          <RoleDetails role={selectedRole} />
        </div>
      </div>
    </div>
  );
}
