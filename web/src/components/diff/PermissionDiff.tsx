import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useRoles } from '@/lib/queries';
import { QueryState } from '@/components/ui/QueryState';
import { RoleSelector } from './RoleSelector';
import { DiffResults } from './DiffResults';


export function PermissionDiff() {
  const [roleA, setRoleA] = useState('');
  const [roleB, setRoleB] = useState('');

  const handleSwap = () => {
    const temp = roleA;
    setRoleA(roleB);
    setRoleB(temp);
  };

  const rolesQuery = useRoles();
  const availableRoles = (rolesQuery.data ?? []).map(role => role.name);
  const comparison = useQuery({ queryKey: ['comparison', roleA, roleB], queryFn: () => api.compareRoles(roleA, roleB), enabled: Boolean(roleA && roleB) });
  const diff = comparison.data ?? { added: [], removed: [], unchanged: [] };
  if (rolesQuery.isPending || rolesQuery.isError) return <QueryState pending={rolesQuery.isPending} error={rolesQuery.error} />;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold text-slate-900">Permission Diff</h2>
        <p className="mt-2 text-slate-600">
          Compare privileges between two roles
        </p>
      </div>

      <div className="space-y-4">
        <div className="w-full">
          <RoleSelector
            roleA={roleA}
            roleB={roleB}
            roles={availableRoles}
            onRoleAChange={setRoleA}
            onRoleBChange={setRoleB}
            onSwap={handleSwap}
          />
        </div>

        <div className="min-w-0">
          {roleA && roleB && (comparison.isPending || comparison.isError) ? <QueryState pending={comparison.isPending} error={comparison.error} /> : <DiffResults
            added={diff.added}
            removed={diff.removed}
            unchanged={diff.unchanged}
            roleA={roleA}
            roleB={roleB}
          />}
        </div>
      </div>
    </div>
  );
}
