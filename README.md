# GateKeep

Open-source Snowflake permissions management with declarative YAML, a parallel sync engine, PostgreSQL audit logging, and a React UI for inspecting roles and comparing permissions.

> **Development preview:** local inspection and cache refresh work with a real Snowflake account. The CLI supports additive sync and dry-run; API sync endpoints are disabled, and authentication and full reconciliation remain unfinished. Evaluate locally before adopting for production permission changes.

## Current capabilities

| Feature | Status |
| --- | --- |
| YAML parsing and CLI `validate` | Available |
| Snowflake role, user, grant, database and warehouse reads | Verified against a real account |
| PostgreSQL roles/grants cache and on-demand refresh | Verified against real Snowflake and local PostgreSQL |
| Web UI: dashboard, logs, role graph/tree and permission comparison | Connected to the API and browser-checked |
| Object-name/type filters and searchable operation logs | Available |
| Parallel sync orchestrator with optional audit logging | Available internally; isolated role creation and an unchanged second run verified |
| Generic grants for dynamic tables, semantic views and warehouses | YAML, SQL planning and existing direct-grant detection tested; dynamic-table/semantic-view execution not yet verified live |
| CLI `sync`, including `--dry-run` | Wired to orchestrator; additive execution, JSON/text output and failure exit codes |
| API `/api/sync` and `/api/sync/dry-run` | Return HTTP 501 |
| Strict reconciliation and authentication | Incomplete |

See [completion status](docs/completion-status.md) for validation evidence and remaining release work. Parallel execution is implemented; published performance targets are goals, not measured speedup guarantees.

## Is PostgreSQL optional?

PostgreSQL is optional for the internal sync engine. Without `POSTGRES_DSN`, it uses a no-op audit logger; audit logging failures are non-fatal.

The current web UI needs PostgreSQL for its data:

| Component | PostgreSQL dependency |
| --- | --- |
| YAML validation | None |
| Internal Snowflake sync engine | Optional audit logging |
| Dashboard and Logs | Sync runs and operation audit records |
| Roles and Permission Diff | Cached Snowflake roles and grants |
| API health and config-based role listing | Can run without PostgreSQL |

The API starts without PostgreSQL, but history returns HTTP 503 and cache routes are unavailable. Cached roles can be read without Snowflake connected; refreshing the cache requires Snowflake credentials.

## Run locally

### Requirements

- Go 1.25 or newer, as required by `go.mod`.
- Node.js 20.19+ within Node 20, or 22.12+, and npm for the frontend.
- Docker Compose for the included PostgreSQL service, or an existing PostgreSQL instance.
- A Snowflake account and a role with visibility into the roles/grants you want to inspect. The current client uses password authentication.

### Build and configure

From a checkout of this repository:

```sh
go mod download
make build
cp .env.example .env
```

Skip the copy if `.env` already exists. Edit the ignored `.env` file:

```dotenv
SNOWFLAKE_ACCOUNT=your-account
SNOWFLAKE_USER=your-user
SNOWFLAKE_PASSWORD=your-password
SNOWFLAKE_ROLE=your-role

# Optional defaults. Leave empty unless these resources exist and are accessible.
SNOWFLAKE_DATABASE=
SNOWFLAKE_WAREHOUSE=

# Needed for the current web UI's cache and audit history.
POSTGRES_DSN=postgres://gatekeep:gatekeep@localhost:5432/gatekeep?sslmode=disable

SERVER_HOST=127.0.0.1
SERVER_PORT=8080
GATEKEEP_CONFIG_PATH=configs/example.yaml
```

Database and warehouse defaults are unnecessary for account-level role/grant inspection. An invalid or inaccessible default can prevent the connection from opening. The example file defaults to `ACCOUNTADMIN`; choose credentials and a role appropriate to your evaluation rather than treating that as a requirement.

The Go entry points do **not** automatically load `.env`. Export it in the terminal where you start the server:

```sh
set -a
source .env
set +a
```

When using `source`, quote values that contain spaces or shell-special characters. Keep credentials out of Git and chat.

### Start PostgreSQL and the API

```sh
docker compose up -d postgres
./bin/gatekeep-server
```

The PostgreSQL service applies both [audit](migrations/postgres/001_init.sql) and [cache](migrations/postgres/002_cache_tables.sql) migrations when initializing a **new** data volume. For an existing database, apply missing migrations in order with your migration tooling or `psql`; restarting the container does not apply new migrations to an existing volume. Preserve existing data.

The API listens at `http://localhost:8080`. The supplied Dockerfile builds the Go binaries; the frontend is started separately below. LocalStack is an optional testing service and is not needed to connect to real Snowflake.

### Start the web UI

In another terminal:

```sh
cd web
npm ci
npm run dev
```

Open the URL Vite prints, normally `http://localhost:5173`. The development proxy forwards `/api` requests to port 8080.

1. Open **Roles** and click **Refresh from Snowflake**.
2. Wait for the refresh to report success, then inspect the graph or tree.
3. Open **Diff**, select two roles, and filter by object name or type.
4. Open **Logs** to inspect recorded sync operations.

Cache refresh reads Snowflake and writes the PostgreSQL cache; it does not change Snowflake permissions or create sync audit entries. Logs remain empty until an audited orchestrator execution occurs. The UI shows operations from the latest 100 recorded sync runs.

## Declarative object grants

Use the generic `objects` section to grant privileges on existing named objects. New Snowflake types using ordinary named-object GRANT syntax can use this model without a separate YAML section.

```yaml
version: 1.0
roles:
  - name: ANALYST_ROLE
objects:
  - type: DATABASE
    name: [ANALYTICS_DB]
    grants:
      - to_role: ANALYST_ROLE
        privileges: [USAGE]
  - type: SCHEMA
    name: [ANALYTICS_DB, PUBLIC]
    grants:
      - to_role: ANALYST_ROLE
        privileges: [USAGE]
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

Validate the example without any database credentials:

```sh
./bin/gatekeep validate configs/generic-objects.yaml
```

Names are lists of exact identifier parts; the planner quotes each part separately. Types and privileges use uppercase SQL keywords. Local validation checks syntax and role references; Snowflake checks whether the object/privilege combination is supported at execution. Include prerequisite access explicitly.

Run a real preview after exporting your Snowflake environment:

```sh
./bin/gatekeep sync --config configs/generic-objects.yaml --mode additive --dry-run --format json
```

Apply a reviewed configuration in a sandbox:

```sh
./bin/gatekeep sync --config your-sandbox.yaml --mode additive
```

`--workers` and `--timeout` configure concurrency and per-operation execution timeout. Defaults come from `SYNC_WORKERS`, `SYNC_TIMEOUT` and `SYNC_MODE`, falling back to 10 workers, 30 seconds and additive mode. Strict mode is available for dry-runs only; execution is rejected. Dry-runs do not execute planned SQL or write audit history. Exit code 0 means success, 1 means validation/connection/execution/output failure, and 2 means invalid command/options. JSON failures produce a structured failure result after valid option parsing; flag-parser errors go to stderr.

The existing nested table and warehouse syntax remains supported. Generic objects grant access to existing resources; they do not create dynamic tables, semantic views or warehouses. ALL/FUTURE scopes, wildcard discovery, ownership transfer, grant option, function/procedure signatures and database-role recipients are not implemented in this model.

See [object grant documentation](docs/object-grants.md) and [the complete example](configs/generic-objects.yaml).

## How teams can adopt GateKeep

### 1. Start with an inspection pilot

Run the API and UI locally with Snowflake credentials and a dedicated PostgreSQL database. Refresh the cache and compare a small set of representative roles. Confirm that the Snowflake role sees the objects and grants you expect; the cache reflects what that role can read.

This is the supported starting point today. Authentication is not implemented, so keep the service on a trusted local interface during evaluation.

### 2. Describe one team's intended access

Create a team-owned YAML file in version control. Begin with a small role and a few existing objects. Include database/schema USAGE and warehouse privileges where needed. Review the file with the people responsible for that data, and run `gatekeep validate` in CI.

Treat configuration as an explicit access contract. Retain the current permission-management process during this pilot. GateKeep does not yet import existing permissions into a complete desired-state configuration automatically.

### 3. Gate write adoption on the remaining implementation work

Before using GateKeep as an operational replacement, complete and verify:

- Verify the wired CLI and audit records against your sandbox, including meaningful failure exit codes and machine-readable output.
- Authentication and authorization for any remotely accessible API.
- A defined managed scope for reconciliation, with tests covering additions, revocations and preservation of unrelated access.
- Live grant tests for the Snowflake object types and privilege combinations your organization uses.

The current strict-mode differ can select non-system roles omitted from YAML for deletion and user role assignments outside the configuration for revocation; object and hierarchy revocation are still incomplete. Do not treat strict mode as ready for a shared account. The CLI honors `SYNC_MODE`, but rejects strict execution; use `--mode additive` to override an older environment setting.

### 4. Prove writes in an isolated sandbox

Start with additive mode and review a real dry-run. Test only dedicated roles and resources. Execute the reviewed plan, verify Snowflake state and audit records, then run it again and confirm zero changes. Test a deliberate failure and recovery before increasing scope.

An internal orchestrator smoke test has created one isolated role and verified a zero-change second run. That does not establish complete reconciliation or live coverage for every object type.

### 5. Roll out gradually with reviewed changes

After those gates pass, expand team by team. Keep configurations in Git, require review for permission changes, store credentials in your secret manager, and retain audit history in PostgreSQL. Establish a recovery procedure before introducing revocations.

[Preview](.github/workflows/gatekeep-preview.yml) and [sync](.github/workflows/gatekeep-sync.yml) workflow scaffolds exist, but pin an older Go version than this project requires. They need updating and validation before use; a successful scaffold run is not evidence of a Snowflake change.

## API endpoints

| Method | Endpoint | Behavior |
| --- | --- | --- |
| GET | `/api/health` | API/database health; does not probe Snowflake |
| GET | `/api/roles` | Roles from the configured YAML file |
| GET | `/api/roles/hierarchy` | Paginated cached hierarchy |
| POST | `/api/roles/refresh` | Background Snowflake-to-PostgreSQL cache refresh |
| GET | `/api/roles/refresh/status` | Refresh status and counts |
| GET | `/api/roles/{roleName}/grants` | Cached direct grants |
| GET | `/api/roles/compare?roleA=A&roleB=B` | Compare cached direct grants |
| GET | `/api/sync/history` | Recorded sync runs |
| GET | `/api/sync/history/{id}` | Run details and operations |
| POST | `/api/sync` | Disabled: HTTP 501 |
| POST | `/api/sync/dry-run` | Disabled: HTTP 501 |

## Development

Go linting uses golangci-lint v2 with the version-2 configuration. `make install-tools` installs it from source using the current Go toolchain; `make lint` upgrades an older v1 installation.

```sh
make build
make test
make lint

cd web
npm ci
npm run build
npm run lint
npx playwright install chromium
npm test
```

Browser tests use API fixtures. Short Go tests skip database-dependent integration tests. PostgreSQL repository tests can use a local `POSTGRES_DSN`; LocalStack Snowflake tests require the optional emulator setup described in [test/README.md](test/README.md). Tests using live infrastructure should run against isolated resources.

## Project structure

```text
cmd/cli/                 CLI sync, dry-run and YAML validation
cmd/server/              API server
internal/config/         YAML models and validation
internal/snowflake/      Connection and state readers
internal/diff/           Diff engine and SQL planner
internal/sync/           Orchestrator and parallel executor
internal/audit/          Optional audit logging
internal/repository/     PostgreSQL audit and cache storage
migrations/postgres/     Database migrations
web/                     React UI and browser tests
configs/                 Configuration examples
```

## Contributing and roadmap

Contributions are welcome. Priorities are sandbox CLI validation, scoped reconciliation, ALL/FUTURE grants, live coverage for more object types, authentication, and production frontend hosting. Run the relevant Go/frontend checks with your changes and describe what you verified in your pull request.

Older [getting-started notes](docs/getting-started.md) describe the intended end-to-end workflow; use this README and [completion status](docs/completion-status.md) for current availability.

Inspired by [Permifrost](https://gitlab.com/gitlab-data/permifrost). Built with Go, Chi, pgx and React.

## License

[MIT](LICENSE).
