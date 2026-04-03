# Claude AI Development Guidelines for GateKeep

This file contains important guidelines and context for AI assistants (Claude) working on this project.

## Go Version Management 🔴 CRITICAL

**IMPORTANT**: We've encountered Go version compatibility issues multiple times. Follow these rules strictly:

### The Problem
- Dependencies (like `pgx v5`) may require newer Go versions (e.g., Go 1.25)
- Pinning specific Go versions (e.g., `1.22`, `1.23`) causes build failures when dependencies upgrade
- Multiple build systems need to stay in sync (CI, Docker, local development)

### The Solution: Always Use "stable" or "latest"

#### 1. GitHub Actions Workflows (`.github/workflows/*.yml`)
```yaml
✅ CORRECT:
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: 'stable'  # Uses latest stable Go
    cache: true

❌ WRONG:
- name: Set up Go
  uses: actions/setup-go@v5
  with:
    go-version: '1.23'  # DO NOT pin specific versions
```

#### 2. Dockerfile
```dockerfile
✅ CORRECT:
FROM golang:alpine AS builder  # Uses latest Go

❌ WRONG:
FROM golang:1.23-alpine AS builder  # DO NOT pin specific versions
```

#### 3. golangci-lint Installation
```yaml
✅ CORRECT:
- name: Install golangci-lint
  run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

❌ WRONG:
- uses: golangci/golangci-lint-action@v3  # May download old pre-built binaries
  with:
    version: v1.64.8  # DO NOT pin specific versions
```

**Why**: Installing from source ensures golangci-lint is built with the same Go version as the project.

#### 4. go.mod
```go
// Let `go mod tidy` manage this automatically
// DO NOT manually edit the Go version in go.mod
go 1.25.0  // This is set by dependencies, not by us
```

**Why**: Dependencies determine the minimum Go version. Let Go tooling manage this.

### Quick Reference: Version Strategy

| File/System | Setting | Why |
|-------------|---------|-----|
| `.github/workflows/*.yml` | `go-version: 'stable'` | Always use latest stable Go |
| `Dockerfile` | `FROM golang:alpine` | Always use latest Go |
| `golangci-lint` | Install from source with `@latest` | Ensures compatibility with current Go |
| `go.mod` | Let `go mod tidy` manage | Dependencies determine minimum version |

### When You See Go Version Errors

If you encounter errors like:
```
Error: the Go language version (go1.24) used to build golangci-lint
is lower than the targeted Go version (1.25.0)
```

Or:
```
go: go.mod requires go >= 1.25.0 (running go 1.23.12)
```

**DO THIS**:
1. Update all pinned Go versions to use `'stable'` or `'latest'` or just `alpine`
2. Install golangci-lint from source (not pre-built binaries)
3. Run `go mod tidy` to let it set the correct version
4. **DO NOT** manually downgrade go.mod's Go version

**DON'T DO THIS**:
1. ❌ Manually edit go.mod to lower the Go version
2. ❌ Pin CI workflows to specific Go versions
3. ❌ Pin Dockerfile to specific Go versions
4. ❌ Use pre-built golangci-lint binaries

## PostgreSQL Integration

PostgreSQL is **OPTIONAL** in GateKeep. The system works with or without it.

### Environment Variables
```bash
# PostgreSQL is OPTIONAL - omit these to run without audit logging
POSTGRES_DSN=postgres://user:pass@localhost:5432/gatekeep?sslmode=disable

# Snowflake is REQUIRED
SNOWFLAKE_ACCOUNT=your-account
SNOWFLAKE_USER=your-user
SNOWFLAKE_PASSWORD=your-password
```

### Code Guidelines
- All audit logging must be **non-fatal**
- Use `nolint:errcheck // audit errors are non-fatal` for audit logger calls
- The orchestrator defaults to `NoOpLogger` if PostgreSQL is not configured
- Never crash or fail sync operations due to audit logging failures

## Testing Strategy

### Unit Tests
- Target: 80%+ coverage for core modules
- Run with: `go test ./... -short`
- Skip database-dependent tests in short mode

### Integration Tests
- PostgreSQL tests skip automatically if no database
- Mark with: `if testing.Short() { t.Skip(...) }`

### Linting
- Run locally: `golangci-lint run --timeout=5m`
- All linting rules in `.golangci.yml`
- Use `nolint:` comments sparingly and with clear justification

## Project Structure

```
gatekeep/
├── cmd/
│   ├── cli/         # CLI binary (gatekeep)
│   └── server/      # REST API server (gatekeep-server)
├── internal/        # Private application code
│   ├── api/         # REST API (Epic 7)
│   ├── audit/       # Audit logging (Epic 6)
│   ├── config/      # YAML parser (Epic 2)
│   ├── database/    # DB helpers (Epic 6)
│   ├── diff/        # State diff engine (Epic 4)
│   ├── domain/      # Core types
│   ├── repository/  # Data access (Epic 6)
│   ├── snowflake/   # Snowflake integration (Epic 3)
│   └── sync/        # Sync orchestration (Epic 5)
├── migrations/      # Database migrations
├── docs/           # Documentation
└── test/           # Integration tests
```

## Error Handling Philosophy

1. **Config Errors**: Fail fast with clear messages
2. **Snowflake Errors**: Retry with exponential backoff (3 attempts)
3. **Operation Errors**: Continue with others, log failures
4. **Audit Errors**: Never fail the sync (non-fatal)
5. **Critical Failures**: Circuit breaker stops at >20% failure rate

## Commit Message Format

Follow conventional commits:

```
feat: Add new feature
fix: Fix bug
docs: Update documentation
test: Add tests
chore: Update dependencies
refactor: Refactor code
```

Always include:
- Clear description of what changed
- Why the change was needed (if not obvious)
- Co-authored-by tag for AI assistance

Example:
```
feat: Add optional PostgreSQL integration for audit logging (Epic 6)

Implements comprehensive audit logging with optional PostgreSQL integration.
GateKeep now supports full audit trail while remaining functional without a database.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
```

## Dependencies

### Core Dependencies
- `github.com/snowflakedb/gosnowflake` - Snowflake driver
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/jackc/pgx/v5` - PostgreSQL driver (optional)
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/google/uuid` - UUID generation

### Adding New Dependencies
1. Use `go get package@latest` (not pinned versions)
2. Run `go mod tidy`
3. Verify tests still pass
4. Update documentation if it changes behavior

## CI/CD Pipeline

### GitHub Actions Workflows
1. **ci.yml**: Lint, test, build on every PR
2. **gatekeep-preview.yml**: Dry-run on PR (GitOps)
3. **gatekeep-sync.yml**: Sync on merge to main (GitOps)

### Pre-commit Hooks
Located in `.git/hooks/pre-commit`:
- `gofmt` - Format code
- `golangci-lint` - Lint code

## Performance Targets

- **Sync Operations**: 1000 operations in <30 seconds
- **Parallelization**: 5-10x speedup vs sequential
- **Workers**: Default 10 (configurable via env)
- **Database Queries**: <100ms for typical operations

## Security Guidelines

1. **No credentials in logs**: Use `nolint:gosec` for config file reads
2. **Parameterized queries**: Never use string concatenation for SQL
3. **Input validation**: Validate all YAML config thoroughly
4. **Secrets management**: Environment variables only, no hardcoding
5. **Least privilege**: Run as non-root user in Docker

## Common Pitfalls to Avoid

### 1. Import Cycles
❌ **Don't**: Import `api` package from `handlers`
✅ **Do**: Define types inline in handlers or in a separate `types` package

### 2. Error Handling
❌ **Don't**: Ignore errors silently with `_ = ...` without nolint
✅ **Do**: Add clear nolint comments explaining why errors are non-fatal

### 3. Testing
❌ **Don't**: Require PostgreSQL for all tests
✅ **Do**: Use `testing.Short()` to skip database tests

### 4. Versioning
❌ **Don't**: Pin Go versions in CI/Docker
✅ **Do**: Use `stable`/`latest`/`alpine` for automatic updates

## Documentation Requirements

When adding new features:
1. Update relevant `docs/*.md` files
2. Add inline code comments for complex logic
3. Update API documentation if endpoints change
4. Add examples in README if user-facing
5. Update CLAUDE.md if it affects development workflow

## Questions?

If you're unsure about:
- **Architecture decisions**: Check `docs/architecture.md`
- **API design**: Check `docs/api.md`
- **PostgreSQL setup**: Check `docs/postgresql-integration.md`
- **Deployment**: Check `docs/deployment.md`

## Version History

- **v0.1.0-alpha**: Epic 1-5 (Foundation, Config, Snowflake, Diff, Sync)
- **v0.2.0-alpha**: Epic 6 (PostgreSQL Integration - Optional)
- **Next**: Epic 7 (REST API), Epic 8 (Testing & Docs)

---

**Last Updated**: 2026-03-30
**Maintainer**: @gunjankaphle
**AI Assistant**: Claude Sonnet 4.5
