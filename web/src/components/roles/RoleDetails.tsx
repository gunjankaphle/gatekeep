import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { Role } from '@/types';
import { Users, Lock, ArrowUp } from 'lucide-react';

interface RoleDetailsProps {
  role: Role | null;
}

export function RoleDetails({ role }: RoleDetailsProps) {
  if (!role) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Role Details</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-slate-500">
            Select a role to view details
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock className="h-5 w-5" />
          {role.name}
        </CardTitle>
        {role.comment && (
          <p className="text-sm text-slate-600">{role.comment}</p>
        )}
      </CardHeader>
      <CardContent className="space-y-6">
        <div>
          <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700">
            <ArrowUp className="h-4 w-4" />
            Parent Roles
          </div>
          {role.parent_roles.length > 0 ? (
            <div className="flex flex-wrap gap-2">
              {role.parent_roles.map((parent) => (
                <Badge key={parent} variant="outline">
                  {parent}
                </Badge>
              ))}
            </div>
          ) : (
            <p className="text-sm text-slate-500">No parent roles (root role)</p>
          )}
        </div>

        <div>
          <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-700">
            <Users className="h-4 w-4" />
            Role Type
          </div>
          <Badge>
            {role.parent_roles.length === 0 ? 'Root Role' : 'Inherited Role'}
          </Badge>
        </div>

        <div className="rounded-md border border-slate-200 bg-slate-50 p-4">
          <h4 className="mb-2 text-sm font-semibold text-slate-700">Role Information</h4>
          <dl className="space-y-2">
            <div className="flex justify-between text-sm">
              <dt className="text-slate-600">Role Name:</dt>
              <dd className="font-mono text-slate-900">{role.name}</dd>
            </div>
            <div className="flex justify-between text-sm">
              <dt className="text-slate-600">Parent Count:</dt>
              <dd className="font-mono text-slate-900">{role.parent_roles.length}</dd>
            </div>
          </dl>
        </div>

        <div className="text-xs text-slate-400">
          <p>
            💡 Tip: Click on a role in the graph or tree to view its details
          </p>
        </div>
      </CardContent>
    </Card>
  );
}
