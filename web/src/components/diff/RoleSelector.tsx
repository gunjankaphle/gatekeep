import { ArrowLeftRight } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface RoleSelectorProps {
  roleA: string;
  roleB: string;
  roles: string[];
  onRoleAChange: (role: string) => void;
  onRoleBChange: (role: string) => void;
  onSwap: () => void;
}

export function RoleSelector({
  roleA,
  roleB,
  roles,
  onRoleAChange,
  onRoleBChange,
  onSwap,
}: RoleSelectorProps) {
  return (
    <div className="flex flex-wrap items-end gap-4 rounded-lg border border-slate-200 bg-white p-4">
          <div className="min-w-[240px] flex-1">
            <label className="mb-2 block text-sm font-medium text-slate-700">
              Role A
            </label>
            <select
              aria-label="Role A"
              value={roleA}
              onChange={(e) => onRoleAChange(e.target.value)}
              className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-400"
            >
              <option value="">Select a role...</option>
              {roles.map((role) => (
                <option key={role} value={role}>
                  {role}
                </option>
              ))}
            </select>
          </div>

          <div className="flex justify-center">
            <Button
              variant="outline"
              size="icon"
              aria-label="Swap roles"
              onClick={onSwap}
              disabled={!roleA || !roleB}
            >
              <ArrowLeftRight className="h-4 w-4" />
            </Button>
          </div>

          <div className="min-w-[240px] flex-1">
            <label className="mb-2 block text-sm font-medium text-slate-700">
              Role B
            </label>
            <select
              aria-label="Role B"
              value={roleB}
              onChange={(e) => onRoleBChange(e.target.value)}
              className="w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-400"
            >
              <option value="">Select a role...</option>
              {roles.map((role) => (
                <option key={role} value={role}>
                  {role}
                </option>
              ))}
            </select>
          </div>
    </div>
  );
}
