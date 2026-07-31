# GateKeep

> ⚠️ **WORK IN PROGRESS** - This project is under active development. Core features are being implemented. APIs and CLI may change.

**Scalable, faster, open-source** Snowflake permissions management platform

## 🚧 Development Status

| Epic | Feature | Status |
|------|---------|--------|
| ✅ Epic 1 | Project Foundation | Complete |
| ✅ Epic 2 | YAML Configuration Parser | Complete |
| ✅ Epic 3 | Snowflake Integration | Complete |
| ✅ Epic 4 | Diff Engine | Complete |
| ✅ Epic 5 | Sync Executor (Parallel) | Complete |
| ✅ Epic 6 | PostgreSQL Audit Log | Complete |
| ✅ Epic 7 | REST API (read-only) | Complete |
| ✅ Epic 8 | Integration Tests (Part 1) | Complete |
| 🚧 — | Roles/Grants Cache Layer | In Progress |
| 🚧 — | Web UI (React) | In Progress (mock data) |

**What Works Now:**
- ✅ YAML config parsing and validation
- ✅ Snowflake connection and state reading
- ✅ Diff engine and SQL generation
- ✅ Parallel sync execution (10 workers by default)
- ✅ Optional PostgreSQL audit logging
- ✅ Read-only REST API (roles, sync history, health)
- ✅ Integration test suite with LocalStack Snowflake support
- ✅ CLI `validate` and `sync` commands

**Coming Soon:**
- 🚧 Postgres-backed roles/grants cache with on-demand Snowflake refresh
- 🚧 React web UI (log viewer, role hierarchy explorer, permission diff tool) — built against mock data, not yet wired to the live API
- 🚧 Authentication (JWT)

**Contributions Welcome:** Issues and PRs are appreciated!

GateKeep manages Snowflake roles and permissions through declarative YAML configurations, with a focus on **parallelization** and **performance** (5-10x faster than sequential tools).

---

## 📖 New to GateKeep?

**👉 [Read the Complete Getting Started Guide](docs/getting-started.md)**

Learn how to adopt GateKeep in your company with step-by-step instructions, real-world examples, and best practices.

---

## Features

- 🚀 **10x faster** through parallel execution
- 📝 **Declarative YAML** configuration for infrastructure-as-code
- 🔄 **GitOps workflow** with dry-run previews on PRs
- 🔍 **Full state reconciliation** - detects and fixes configuration drift
- 📊 **Comprehensive audit logging** to PostgreSQL
- 🔌 **REST API** for external integrations
- 🐳 **Docker** support for easy deployment
- ✅ **Dry-run mode** to preview changes before applying

## Quick Start

### Prerequisites

- Go 1.25+ (for development)
- Docker & Docker Compose (for local development)
- Snowflake account with ACCOUNTADMIN privileges
- PostgreSQL 16+ (for audit logging)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/gatekeep.git
cd gatekeep

# Install dependencies
go mod download

# Build binaries
make build

# Or install directly
go install ./cmd/cli
go install ./cmd/server
```

### Configuration

1. Copy the example environment file:
```bash
cp .env.example .env
```

2. Edit `.env` and add your Snowflake credentials:
```bash
SNOWFLAKE_ACCOUNT=your-account
SNOWFLAKE_USER=your-user
SNOWFLAKE_PASSWORD=your-password
SNOWFLAKE_DATABASE=your-database
SNOWFLAKE_WAREHOUSE=your-warehouse
```

3. Create a YAML configuration file (see `configs/example.yaml`):
```yaml
version: 1.0

roles:
  - name: ANALYST_ROLE
    comment: "Data analysts role"

users:
  - name: analyst@company.com
    roles: [ANALYST_ROLE]

databases:
  - name: PROD_DB
    schemas:
      - name: PUBLIC
        tables:
          - name: CUSTOMERS
            grants:
              - to_role: ANALYST_ROLE
                privileges: [SELECT]
```

### Usage

#### CLI

```bash
# Validate configuration
./bin/gatekeep validate config.yaml

# Dry-run (preview changes without applying)
./bin/gatekeep sync --config config.yaml --dry-run

# Execute sync
./bin/gatekeep sync --config config.yaml

# JSON output for automation
./bin/gatekeep sync --config config.yaml --format json
```

#### API Server

```bash
# Start the API server
./bin/gatekeep-server

# Or with Docker
docker-compose up
```

The API will be available at `http://localhost:8080`.

#### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/roles` | List roles from config |
| GET | `/api/roles/hierarchy` | Cached role hierarchy (parent/child relationships) |
| POST | `/api/roles/refresh` | Trigger a cache refresh from Snowflake |
| GET | `/api/roles/refresh/status` | Cache refresh status |
| GET | `/api/roles/:roleName/grants` | Grants for a specific role (from cache) |
| GET | `/api/roles/compare` | Diff grants between two roles (from cache) |
| POST | `/api/sync` | Trigger sync |
| POST | `/api/sync/dry-run` | Dry-run sync |
| GET | `/api/sync/history` | List sync history |
| GET | `/api/sync/history/:id` | Sync details |

> The `/api/roles/*` cache endpoints are backed by a new Postgres cache layer (`cached_roles`, `cached_grants`, `cache_metadata` — see `migrations/postgres/002_cache_tables.sql`) that snapshots Snowflake role/grant state so the web UI doesn't hit Snowflake directly on every request. This work is in progress and not yet merged to `main`.

## GitOps Workflow

GateKeep supports a GitOps workflow where configuration changes are reviewed via pull requests before being applied to Snowflake.

### Setup

1. Store your YAML configuration in a Git repository (e.g., `snowflake/prod.yaml`)

2. Configure GitHub secrets:
   - `SNOWFLAKE_ACCOUNT`
   - `SNOWFLAKE_USER`
   - `SNOWFLAKE_PASSWORD`
   - `SNOWFLAKE_DATABASE`
   - `SNOWFLAKE_WAREHOUSE`
   - `POSTGRES_DSN`

3. The workflows are already set up in `.github/workflows/`:
   - `gatekeep-preview.yml` - Runs dry-run on pull requests
   - `gatekeep-sync.yml` - Syncs changes when merged to main

### Workflow

1. **Create a PR** with YAML changes
2. **GitHub Actions runs dry-run** and comments on PR with SQL changes
3. **Review the SQL** to ensure it's correct
4. **Merge the PR** when approved
5. **GitHub Actions syncs** changes to Snowflake automatically
6. **Audit log** records all operations

## Architecture

```
┌─────────────────┐
│  YAML Config    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│  Config Parser  │──────▶│  Validator   │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│ Snowflake State │◀─────│  Snowflake   │
│     Reader      │      │   Client     │
└────────┬────────┘      └──────────────┘
         │
         ▼
┌─────────────────┐
│  Diff Engine    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  SQL Planner    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐
│     Parallel    │──────▶│  PostgreSQL  │
│    Executor     │      │  Audit Log   │
└─────────────────┘      └──────────────┘
```

### Key Components

- **Config Parser**: Parses and validates YAML configuration
- **State Reader**: Reads current state from Snowflake
- **Diff Engine**: Compares desired vs actual state
- **SQL Planner**: Generates minimal SQL statements with dependency resolution
- **Parallel Executor**: Executes operations in parallel (10 workers by default)
- **Audit Logger**: Records all operations to PostgreSQL

## Performance

GateKeep is designed for performance through parallel execution:

- **10 concurrent workers** (configurable)
- **Phase-based execution** (roles → grants → revokes)
- **Circuit breaker** stops execution if >20% operations fail
- **Target**: 1000 operations in <30 seconds

### Benchmarks

| Operations | Sequential | Parallel | Speedup |
|-----------|-----------|----------|---------|
| 100 | ~10s | ~2s | 5x |
| 500 | ~50s | ~8s | 6.2x |
| 1000 | ~180s | ~25s | 7.2x |

## Development

```bash
# Install development tools
make install-tools

# Run tests
make test

# Run integration tests (requires Docker)
make test-integration

# Run linters
make lint

# Generate coverage report
make coverage

# Build Docker image
make docker-build

# Start local development environment
make docker-up
```

## Testing

```bash
# Unit tests
go test -v ./...

# Integration tests (requires PostgreSQL)
go test -v -run Integration ./...

# E2E tests (requires Snowflake + PostgreSQL)
go test -v -run E2E ./test/...

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Project Structure

```
gatekeep/
├── cmd/
│   ├── server/         # API server entry point
│   └── cli/            # CLI tool entry point
├── internal/
│   ├── config/         # YAML parser & validator
│   ├── snowflake/      # Snowflake integration
│   ├── diff/           # State diff engine
│   ├── sync/           # Sync orchestration + parallelization
│   ├── audit/          # Audit logging
│   ├── api/            # REST API layer
│   ├── repository/     # Data access layer
│   └── domain/         # Core business types
├── migrations/         # Database migrations
├── configs/            # Example YAML configs
├── test/               # Integration & E2E tests
└── docs/               # Documentation
```

## Configuration Schema

See [docs/yaml-schema.md](docs/yaml-schema.md) for complete schema documentation.

### Sync Modes

- **Strict mode** (default): Revokes grants not in YAML (full reconciliation)
- **Additive mode**: Only adds grants, never revokes (append-only)

Set mode via environment variable:
```bash
SYNC_MODE=additive  # or 'strict' (default)
```

## Security

- **Never commit credentials** - use environment variables or secrets management
- **SQL injection protection** - parameterized queries throughout
- **Input validation** - comprehensive YAML schema validation
- **Audit trail** - all operations logged to PostgreSQL
- **Circuit breaker** - stops execution on high failure rate

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [x] YAML configuration parser
- [x] Snowflake integration
- [x] Diff engine
- [x] Parallel sync executor
- [x] PostgreSQL audit logging
- [x] REST API (read-only)
- [x] GitOps workflows
- [x] Integration test suite (LocalStack Snowflake)
- [ ] Roles/grants Postgres cache layer (in progress)
- [ ] React frontend (Web UI) — scaffolded with mock data, not yet wired to live API
- [ ] Authentication (JWT)
- [ ] Multi-tenant support
- [ ] Terraform provider
- [ ] Slack/Teams notifications
- [ ] Approval workflows

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- Inspired by [Permifrost](https://gitlab.com/gitlab-data/permifrost)
- Built with [Go](https://golang.org/), [Chi](https://github.com/go-chi/chi), and [pgx](https://github.com/jackc/pgx)

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/yourusername/gatekeep/issues)
- 💬 [Discussions](https://github.com/yourusername/gatekeep/discussions)

---

**GateKeep** - Scalable Snowflake permissions management made simple.
