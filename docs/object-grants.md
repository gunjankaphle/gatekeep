# Extensible object grants

GateKeep accepts grants on existing named Snowflake objects through a generic `objects` section. New object types that use the same named-object GRANT syntax do not require a new Go struct, YAML section or fixed privilege allowlist.

```yaml
version: 1.0
roles:
  - name: ANALYST_ROLE
objects:
  - type: DYNAMIC TABLE
    name: [ANALYTICS_DB, PUBLIC, DAILY_REVENUE]
    grants:
      - to_role: ANALYST_ROLE
        privileges: [SELECT, MONITOR]
  - type: SEMANTIC VIEW
    name: [ANALYTICS_DB, PUBLIC, REVENUE_MODEL]
    grants:
      - to_role: ANALYST_ROLE
        privileges: [SELECT, REFERENCES]
  - type: WAREHOUSE
    name: [ANALYTICS_WH]
    grants:
      - to_role: ANALYST_ROLE
        privileges: [USAGE, MONITOR]
```

Use uppercase SQL keywords for `type` and privileges. Names are lists of one to three exact identifier parts: warehouse/database `[NAME]`, schema `[DB, SCHEMA]`, and a schema object `[DB, SCHEMA, OBJECT]`. Each part is quoted separately and embedded double quotes are escaped. A dot inside an individual part stays part of that identifier. Names are case-sensitive because SQL uses quoted identifiers.

The SQL planner generates statements such as:

```sql
GRANT SELECT ON DYNAMIC TABLE "ANALYTICS_DB"."PUBLIC"."DAILY_REVENUE" TO ROLE "ANALYST_ROLE";
GRANT REFERENCES ON SEMANTIC VIEW "ANALYTICS_DB"."PUBLIC"."REVENUE_MODEL" TO ROLE "ANALYST_ROLE";
GRANT USAGE ON WAREHOUSE "ANALYTICS_WH" TO ROLE "ANALYST_ROLE";
```

`objects` grants access to existing objects; it does not create them. Declare database/schema USAGE and warehouse access explicitly where required. The full example in `configs/generic-objects.yaml` includes these prerequisite grants. The executor needs sufficient authority to grant the requested access.

Snowflake documents SELECT, MONITOR and OPERATE for dynamic tables; SELECT, REFERENCES and MONITOR for semantic views. See [GRANT privileges to role](https://docs.snowflake.com/en/sql-reference/sql/grant-privilege) and [dynamic table access control](https://docs.snowflake.com/en/user-guide/dynamic-tables/privileges).

Validation checks keyword syntax, nonempty identifier parts, role references and individual privileges. It deliberately does not maintain a fixed catalog of Snowflake object/privilege combinations: Snowflake validates those at execution. Unsupported combinations therefore fail at execution, not during local YAML validation. Ownership transfers and ALL PRIVILEGES are rejected; list specific permissions.

Existing nested `databases.schemas.tables` and `warehouses` configuration remains supported. The diff suppresses existing direct grants and duplicate desired grants, normalizing multiword SHOW grant types such as `DYNAMIC_TABLE`. It does not expand inherited permissions.

## Current limits

This implementation handles explicit named objects with ordinary privilege grants. Account/global privileges, function/procedure argument signatures, database-role recipients, grant option, ownership transfers, wildcard discovery, ALL/FUTURE object scopes and strict object-grant revocation require additional models. These forms must not be encoded as freeform `type` or `name` values.

The orchestrator can execute these planned grants. The CLI `sync` entry point is still a placeholder and must be wired before this is usable through the documented CLI. Tests cover configuration validation, identifier quoting, dynamic table/semantic view/warehouse SQL and an unchanged second diff; these new object types have not been executed against live Snowflake objects.
