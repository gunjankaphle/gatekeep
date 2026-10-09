# GateKeep completion status

Updated 2026-10-08. Work resumed on `epic9-roles-cache-ui` from commit `04d4a90`.

## Where development stopped

The CLI scaffold, YAML parser, sync orchestrator, parallel executor, audit repository and read-only API existed. The latest commits added a PostgreSQL cache and a React UI, but the UI still displayed mock data. Several core reconciliation methods remain placeholders despite earlier completion claims.

## Completed in this pass

- Dashboard and log viewer read the latest 100 audit runs and their operation details from the real API. Filtering and paging operate within that window.
- Role hierarchy loads every page of cached roles and offers a Snowflake refresh button with status/error reporting.
- Permission comparison uses cached grants through the API.
- Header health comes from the API; API health currently checks PostgreSQL, not Snowflake.
- Loading, empty and error states replace mock data in application screens.
- Grant comparison matches object type, name and privilege independently of the recipient role, with deterministic ordering and empty JSON arrays.
- Cache hierarchy relationships are populated from role grants. Native pgx arrays are used for cache reads and bulk insertion.
- Cache reads remain available without a Snowflake connection. Missing history/refresh dependencies return HTTP 503.
- Snowflake SHOW output is read by column name to tolerate additional columns. Grant reads propagate failures instead of silently returning incomplete state. CLI state reading now retrieves role grants.
- Corrected ignore rules that excluded the log UI source directory from fresh checkouts.
- Frontend pagination no longer changes state during render; frontend lint errors resolved.

## Verified locally

- `go test -race -short ./...`
- `npm run build` and `npm run lint` in `web`
- `npm test` in `web`: browser tests covering dashboard, logs, hierarchy, comparison and missing-database handling using intercepted API fixtures

Live validation completed after credentials were supplied:

- Snowflake connection and SHOW readers succeeded: 8 roles, 369 grants, 2 users, 4 databases and 3 warehouses.
- PostgreSQL started through Docker Compose with both migrations present.
- A live API refresh saved all 8 roles and 369 grants to PostgreSQL in approximately 3.3 seconds.
- Health, hierarchy, role-grants, comparison and audit-history endpoints returned HTTP 200 with the live cache.
- The supplied default database caused Snowflake error 390201 (database does not exist or is unauthorized). Probes omitted optional database/warehouse defaults; account-level inspection does not require them.
- Live browser validation passed for the role hierarchy, permission comparison and empty dashboard/log history, with zero browser runtime errors.
- PostgreSQL-backed repository tests passed with the race detector.
- Initial connection/cache validation performed no Snowflake mutations.

Follow-up UI and smoke-sync validation:

- Graph role labels truncate within their cards, with full names/comments on hover. Layout graphs are recreated for each layout to avoid stale nodes/edges.
- Permission comparisons filter by object-name substring and object type; counts and rows reflect those filters, including all matching unchanged grants.
- User-requested smoke execution used the existing orchestrator in additive mode with `configs/smoke-test.yaml`. A reviewed dry-run allowed only creation of `GATEKEEP_SMOKE_TEST_20261008`, with no permission grants or revocations.
- First execution succeeded with one audited role creation; second execution succeeded with zero operations. The isolated test role remains in Snowflake.
- Live browser checks verified label bounds, both object filters and the actual operation in Logs/details. Build, frontend lint and browser tests passed.

## Remaining release work, in order

1. Add persistent regression coverage for cache rollback, refresh failure and concurrent refresh behavior. Live cache array encoding and hierarchy reads have passed.
2. Fix or omit the invalid default database in local configuration. Live SHOW output and cache refresh have passed; test a real CLI additive dry-run.
3. CLI sync/dry-run wiring is implemented locally with additive execution, JSON/text output, option validation and nonzero failure exits. Strict execution is disabled. Live additive CLI dry-run has been verified against the isolated smoke-test role. Next complete reconciliation in `internal/diff/differ.go`: suppress already-existing object/hierarchy grants and implement role/object grant revocation. Define managed scope before enabling strict mode on a shared account: current strict mode can drop non-system roles omitted from YAML and revoke user assignments outside the configuration.
4. Extend generic named-object grants with scoped ALL/FUTURE grants, function signatures and strict reconciliation. Named-object SQL now quotes identifier parts separately; dynamic table/semantic view/warehouse planning and direct-grant idempotency have regression tests. See `docs/object-grants.md`.
5. Validate sync against an isolated Snowflake sandbox, then verify a second run produces no changes and audit history records results.
6. Add authentication and access restrictions before exposing the API. Sync API endpoints deliberately return HTTP 501; decide whether writes belong in the first release or remain CLI-only.
7. Finish production frontend serving/deployment and update old documentation/claims (performance benchmarks are not validated here).

## Local connection setup

Fill the ignored root `.env` file with `SNOWFLAKE_ACCOUNT`, `SNOWFLAKE_USER`, `SNOWFLAKE_PASSWORD`, `SNOWFLAKE_ROLE`, and, if needed, `SNOWFLAKE_DATABASE` and `SNOWFLAKE_WAREHOUSE`. The current client supports password authentication. Do not paste credentials into chat.

For a local PostgreSQL instance:

```sh
docker compose up -d postgres
```

Set `POSTGRES_DSN=postgres://gatekeep:gatekeep@localhost:5432/gatekeep?sslmode=disable` in `.env`. Both SQL migrations must be applied. For an existing database, use a migration tool or `psql` to apply missing migrations; do not recreate its volume.

The Go entry points do not automatically load `.env`. Export it in the terminal before starting the API:

```sh
set -a
source .env
set +a
go run ./cmd/server
```

In a separate terminal:

```sh
cd web
npm run dev
```

Open the URL Vite prints and use Roles → Refresh from Snowflake. The refresh reads Snowflake state and writes only the PostgreSQL cache.


## CLI follow-up

The CLI now calls the existing orchestrator. It validates YAML before connecting, honors mode/workers/timeout settings and supports `--dry-run`, `--format json`, and additive execution. Dry-run results label SQL operations as `planned`; dry-runs skip audit initialization. Strict execution is rejected pending managed-scope work. Unknown commands/options and execution failures return nonzero exits.

Automated integration tests use simulated Snowflake SQL rows to exercise CLI-to-orchestrator planning/execution and ensure dry-run executes no write SQL. This is not a new live Snowflake execution test.

Git writes and network access are now available. The follow-up branch `feat/cli-sync-lint-v2` was rebased onto merged `origin/main` (`76a6f23`), preserving the local changes. No follow-up commit or push has been performed.

Validation for this follow-up: Go short tests with the race detector, CLI integration tests, golangci-lint and CLI build passed. The built CLI validated the generic-object example and rejected strict execution with structured JSON and exit code 2. The live additive CLI dry-run against `configs/smoke-test.yaml` succeeded with zero planned operations, confirming the existing smoke-test role needs no changes. No new execute-mode request was made. CI and local tooling now use golangci-lint v2 to support current Go export data; the v2 configuration and lint checks pass.
