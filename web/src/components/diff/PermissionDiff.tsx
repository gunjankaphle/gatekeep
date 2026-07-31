import { useState, useMemo } from 'react';
import { RoleSelector } from './RoleSelector';
import { DiffResults } from './DiffResults';
import type { Grant } from '@/types';

// Mock permission data
const mockPermissions: Record<string, Grant[]> = {
  READ_ONLY: [
    { object_type: 'DATABASE', object_name: 'PROD_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'SCHEMA', object_name: 'PROD_DB.PUBLIC', privilege: 'USAGE', granted_on: 'SCHEMA' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.CUSTOMERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.ORDERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'WAREHOUSE', object_name: 'ANALYTICS_WH', privilege: 'USAGE', granted_on: 'WAREHOUSE' },
  ],
  ANALYST_ROLE: [
    { object_type: 'DATABASE', object_name: 'PROD_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'SCHEMA', object_name: 'PROD_DB.PUBLIC', privilege: 'USAGE', granted_on: 'SCHEMA' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.CUSTOMERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.ORDERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'ANALYTICS_DB.PUBLIC.USER_METRICS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'WAREHOUSE', object_name: 'ANALYTICS_WH', privilege: 'USAGE', granted_on: 'WAREHOUSE' },
    { object_type: 'VIEW', object_name: 'ANALYTICS_DB.PUBLIC.REVENUE_VIEW', privilege: 'SELECT', granted_on: 'VIEW' },
  ],
  ENGINEER_ROLE: [
    { object_type: 'DATABASE', object_name: 'PROD_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'DATABASE', object_name: 'DEV_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'SCHEMA', object_name: 'PROD_DB.PUBLIC', privilege: 'USAGE', granted_on: 'SCHEMA' },
    { object_type: 'SCHEMA', object_name: 'DEV_DB.PUBLIC', privilege: 'CREATE TABLE', granted_on: 'SCHEMA' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.CUSTOMERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'PROD_DB.PUBLIC.ORDERS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'DEV_DB.PUBLIC.EVENTS', privilege: 'INSERT', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'DEV_DB.PUBLIC.EVENTS', privilege: 'UPDATE', granted_on: 'TABLE' },
    { object_type: 'TABLE', object_name: 'DEV_DB.PUBLIC.EVENTS', privilege: 'DELETE', granted_on: 'TABLE' },
    { object_type: 'WAREHOUSE', object_name: 'ANALYTICS_WH', privilege: 'USAGE', granted_on: 'WAREHOUSE' },
    { object_type: 'WAREHOUSE', object_name: 'COMPUTE_WH', privilege: 'USAGE', granted_on: 'WAREHOUSE' },
    { object_type: 'WAREHOUSE', object_name: 'COMPUTE_WH', privilege: 'OPERATE', granted_on: 'WAREHOUSE' },
  ],
  DATA_SCIENTIST_ROLE: [
    { object_type: 'DATABASE', object_name: 'PROD_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'DATABASE', object_name: 'SANDBOX_DB', privilege: 'USAGE', granted_on: 'DATABASE' },
    { object_type: 'SCHEMA', object_name: 'SANDBOX_DB.PUBLIC', privilege: 'ALL', granted_on: 'SCHEMA' },
    { object_type: 'TABLE', object_name: 'ANALYTICS_DB.PUBLIC.USER_METRICS', privilege: 'SELECT', granted_on: 'TABLE' },
    { object_type: 'WAREHOUSE', object_name: 'ETL_WH', privilege: 'USAGE', granted_on: 'WAREHOUSE' },
  ],
};

const availableRoles = Object.keys(mockPermissions);

export function PermissionDiff() {
  const [roleA, setRoleA] = useState('');
  const [roleB, setRoleB] = useState('');

  const handleSwap = () => {
    const temp = roleA;
    setRoleA(roleB);
    setRoleB(temp);
  };

  const diff = useMemo(() => {
    if (!roleA || !roleB) {
      return { added: [], removed: [], unchanged: [] };
    }

    const permissionsA = mockPermissions[roleA] || [];
    const permissionsB = mockPermissions[roleB] || [];

    const added = permissionsB.filter(
      (grantB) =>
        !permissionsA.some(
          (grantA) =>
            grantA.object_name === grantB.object_name &&
            grantA.privilege === grantB.privilege
        )
    );

    const removed = permissionsA.filter(
      (grantA) =>
        !permissionsB.some(
          (grantB) =>
            grantA.object_name === grantB.object_name &&
            grantA.privilege === grantB.privilege
        )
    );

    const unchanged = permissionsA.filter((grantA) =>
      permissionsB.some(
        (grantB) =>
          grantA.object_name === grantB.object_name &&
          grantA.privilege === grantB.privilege
      )
    );

    return { added, removed, unchanged };
  }, [roleA, roleB]);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold text-slate-900">Permission Diff</h2>
        <p className="mt-2 text-slate-600">
          Compare privileges between two roles
        </p>
      </div>

      <div className="grid grid-cols-4 gap-6">
        <div className="col-span-1">
          <RoleSelector
            roleA={roleA}
            roleB={roleB}
            roles={availableRoles}
            onRoleAChange={setRoleA}
            onRoleBChange={setRoleB}
            onSwap={handleSwap}
          />
        </div>

        <div className="col-span-3">
          <DiffResults
            added={diff.added}
            removed={diff.removed}
            unchanged={diff.unchanged}
            roleA={roleA}
            roleB={roleB}
          />
        </div>
      </div>
    </div>
  );
}
