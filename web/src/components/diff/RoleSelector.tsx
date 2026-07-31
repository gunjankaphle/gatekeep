import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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
    <Card>
      <CardHeader>
        <CardTitle>Compare Roles</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">
              Role A
            </label>
            <select
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
              onClick={onSwap}
              disabled={!roleA || !roleB}
            >
              <ArrowLeftRight className="h-4 w-4" />
            </Button>
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">
              Role B
            </label>
            <select
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
      </CardContent>
    </Card>
  );
}
