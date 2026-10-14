# Scoped permission reconciliation

Additive mode adds missing permissions and leaves existing access in place. Both additive and strict modes compare direct role hierarchy and object grants with Snowflake state, so an unchanged configuration produces no repeated grants.

Strict mode requires an explicit `managed` section, including at least one custom role. It reconciles direct permissions within that scope. It never drops roles, users, databases, warehouses, or other resources.

```yaml
version: "1.0"
managed:
  roles: [TEAM_READER, TEAM_PARENT]
  users: [service_user]
  objects:
    - type: WAREHOUSE
      name: [TEAM_WH]
      privileges: [USAGE, MONITOR]
roles:
  - name: TEAM_READER
    parent_roles: [TEAM_PARENT]
  - name: TEAM_PARENT
objects:
  - type: WAREHOUSE
    name: [TEAM_WH]
    grants:
      - to_role: TEAM_READER
        privileges: [USAGE]
```

This configuration preserves `TEAM_READER` → `TEAM_PARENT` and USAGE on `TEAM_WH` for `TEAM_READER`. It may revoke MONITOR on that warehouse from either managed role and USAGE from `TEAM_PARENT`, because those scoped grants are absent from the desired configuration. OPERATE, ownership, other warehouses, and other recipient roles remain untouched.

Revocation boundaries are explicit:

- **Hierarchy:** both the child and parent must appear in `managed.roles`.
- **User assignments:** the user must appear in `managed.users`, and the assigned role must appear in `managed.roles`.
- **Object permissions:** the recipient role, exact object type/name, and individual privilege must all appear in the managed scope.

Scope entries do not grant permissions or create resources. They may remain after removing a desired grant or user entry, allowing GateKeep to revoke that access. To stop managing access without changing it, remove its scope entry. Additions still come from the entire desired configuration, including entries outside the revocation scope.

System roles cannot appear in `managed.roles`. Ownership transfers, ALL privileges, account/global permissions, database-role recipients, function/procedure signatures, and future grants are outside this model. Managed object name parts containing dots or double quotes are rejected because the current SHOW state model cannot match those boundaries unambiguously; ordinary additive named-object grants still support these identifiers. Matching is case-sensitive for identifiers; SHOW object types such as `DYNAMIC_TABLE` normalize to their SQL form.

Preview and review before executing:

```sh
./bin/gatekeep validate configs/scoped-reconciliation.yaml
./bin/gatekeep sync --config your-sandbox.yaml --mode strict --dry-run --format json
./bin/gatekeep sync --config your-sandbox.yaml --mode strict
```

Execution follows the planner's dependency phases. Work within each phase can run in parallel; a failed or skipped phase stops subsequent phases, including revocations. Skipped operations are reported in CLI output and audit history. Execution is not transactional: successful operations remain applied after a failure. A dry-run is a preview, and execute mode reads current state again rather than applying a saved plan.

Use a Snowflake execution role that can read and manage the scoped grants. GateKeep reconciles visible direct grants; it does not expand inherited permissions or discover future grants. Test additions, removal, unchanged reruns, and failure recovery against isolated resources before adopting strict mode for a team.
