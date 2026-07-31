import { useState } from 'react';
import { RoleGraph } from './RoleGraph';
import { RoleTreeView } from './RoleTreeView';
import { RoleDetails } from './RoleDetails';
import { Button } from '@/components/ui/button';
import type { Role } from '@/types';
import { Network, List } from 'lucide-react';

// Mock roles for demonstration
const mockRoles: Role[] = [
  { name: 'READ_ONLY', parent_roles: [], comment: 'Read-only access to production' },
  { name: 'ANALYST_ROLE', parent_roles: ['READ_ONLY'], comment: 'Data analysts with read access' },
  { name: 'ENGINEER_ROLE', parent_roles: ['ANALYST_ROLE'], comment: 'Engineers with write access' },
  { name: 'DATA_SCIENTIST_ROLE', parent_roles: ['ANALYST_ROLE'], comment: 'Data scientists with ML access' },
  { name: 'DATA_ENGINEER_ROLE', parent_roles: ['ENGINEER_ROLE'], comment: 'Data engineers for ETL pipelines' },
  { name: 'ADMIN_ROLE', parent_roles: ['ENGINEER_ROLE', 'DATA_SCIENTIST_ROLE'], comment: 'System administrators' },
  { name: 'DBA_ROLE', parent_roles: ['ADMIN_ROLE'], comment: 'Database administrators' },
  { name: 'SECURITY_ADMIN', parent_roles: ['DBA_ROLE'], comment: 'Security administrators' },
];

export function RoleHierarchy() {
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

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2">
          {view === 'graph' ? (
            <RoleGraph roles={mockRoles} onRoleClick={setSelectedRole} />
          ) : (
            <RoleTreeView roles={mockRoles} onRoleClick={setSelectedRole} />
          )}
        </div>
        <div className="col-span-1">
          <RoleDetails role={selectedRole} />
        </div>
      </div>
    </div>
  );
}
