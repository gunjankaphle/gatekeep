import { useState } from 'react';
import { ChevronRight, ChevronDown, Circle } from 'lucide-react';
import type { Role } from '@/types';
import { cn } from '@/lib/utils';

interface RoleTreeViewProps {
  roles: Role[];
  onRoleClick: (role: Role) => void;
}

interface TreeNodeProps {
  role: Role;
  roles: Role[];
  level: number;
  onRoleClick: (role: Role) => void;
}

function TreeNode({ role, roles, level, onRoleClick }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(level < 2);

  // Find child roles (roles that have this role as a parent)
  const children = roles.filter((r) => r.parent_roles.includes(role.name));
  const hasChildren = children.length > 0;

  return (
    <div>
      <div
        className={cn(
          'flex items-center gap-2 rounded-md px-3 py-2 hover:bg-slate-100 cursor-pointer transition-colors',
          'border-l-2 border-slate-200'
        )}
        style={{ marginLeft: `${level * 24}px` }}
      >
        {hasChildren ? (
          <button
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
            className="flex h-5 w-5 items-center justify-center"
          >
            {expanded ? (
              <ChevronDown className="h-4 w-4 text-slate-600" />
            ) : (
              <ChevronRight className="h-4 w-4 text-slate-600" />
            )}
          </button>
        ) : (
          <Circle className="h-2 w-2 text-slate-400" />
        )}
        <button
          onClick={() => onRoleClick(role)}
          className="flex-1 text-left"
        >
          <span className="font-mono text-sm font-medium text-slate-900">
            {role.name}
          </span>
          {role.comment && (
            <span className="ml-2 text-xs text-slate-500">
              {role.comment}
            </span>
          )}
        </button>
      </div>
      {hasChildren && expanded && (
        <div className="mt-1">
          {children.map((child) => (
            <TreeNode
              key={child.name}
              role={child}
              roles={roles}
              level={level + 1}
              onRoleClick={onRoleClick}
            />
          ))}
        </div>
      )}
    </div>
  );
}

export function RoleTreeView({ roles, onRoleClick }: RoleTreeViewProps) {
  // Find root roles (roles with no parents)
  const rootRoles = roles.filter((role) => role.parent_roles.length === 0);

  if (roles.length === 0) {
    return (
      <div className="flex h-96 items-center justify-center rounded-lg border border-slate-200 bg-white">
        <p className="text-sm text-slate-500">No roles to display</p>
      </div>
    );
  }

  return (
    <div className="space-y-2 rounded-lg border border-slate-200 bg-white p-4">
      <h3 className="mb-4 font-semibold text-slate-900">Role Hierarchy</h3>
      <div className="space-y-1">
        {rootRoles.map((role) => (
          <TreeNode
            key={role.name}
            role={role}
            roles={roles}
            level={0}
            onRoleClick={onRoleClick}
          />
        ))}
      </div>
    </div>
  );
}
