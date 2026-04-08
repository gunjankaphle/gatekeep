# GateKeep Integration Tests

This directory contains integration tests for GateKeep using LocalStack's Snowflake emulator.

## Quick Start

### 1. Start LocalStack with Snowflake

```bash
# Start LocalStack (requires LocalStack Pro with Snowflake support)
docker-compose up -d localstack

# Check if LocalStack is healthy
curl http://localhost:4566/_localstack/health
```

### 2. Run Integration Tests

```bash
# Run all integration tests
USE_LOCALSTACK=true go test -v ./test/integration/...

# Run specific test
USE_LOCALSTACK=true go test -v ./test/integration/ -run TestSnowflakeConnectionToLocalStack

# Run without LocalStack (skips integration tests)
go test -short ./test/integration/...
```

## LocalStack Setup

### Prerequisites

- Docker and Docker Compose
- LocalStack Pro account (for Snowflake emulator support)
- Set `LOCALSTACK_AUTH_TOKEN` environment variable:

```bash
export LOCALSTACK_AUTH_TOKEN=your-token-here
```

### Configuration

LocalStack connection settings (defaults):

```bash
LOCALSTACK_HOST=localhost
LOCALSTACK_PORT=4566
LOCALSTACK_SNOWFLAKE_ACCOUNT=test
LOCALSTACK_SNOWFLAKE_USER=test
LOCALSTACK_SNOWFLAKE_PASSWORD=test
LOCALSTACK_SNOWFLAKE_DATABASE=TEST_DB
LOCALSTACK_SNOWFLAKE_WAREHOUSE=ANALYTICS_WH
```

### Docker Compose Services

```bash
# Start only LocalStack
docker-compose up -d localstack

# Start LocalStack + PostgreSQL
docker-compose up -d localstack postgres

# View LocalStack logs
docker-compose logs -f localstack

# Stop services
docker-compose down
```

## Test Fixtures

Test YAML configurations are in `test/fixtures/`:

### Simple Config (`simple_config.yaml`)
- 3 roles (READ_ONLY_ROLE, ANALYST_ROLE, ENGINEER_ROLE)
- 2 users
- 1 database with basic permissions
- Good for quick tests

### Complex Config (`complex_config.yaml`)
- 7 roles with hierarchy (BASE → DEPARTMENT → ADMIN)
- 5 users
- 2 databases with multiple schemas
- Tests role inheritance and complex permissions

## Integration Test Structure

```
test/
├── fixtures/              # Test YAML configs
│   ├── simple_config.yaml
│   └── complex_config.yaml
├── integration/           # Integration tests
│   ├── helpers.go        # Test utilities
│   └── config_sync_test.go
└── localstack/           # LocalStack init scripts
    └── init-snowflake.sh
```

## Writing Integration Tests

### Basic Pattern

```go
func TestYourFeature(t *testing.T) {
    // Setup LocalStack connection
    db := SetupLocalStackSnowflake(t)
    defer db.Close()

    // Your test logic
    ctx := context.Background()
    _, err := db.ExecContext(ctx, "CREATE ROLE MY_TEST_ROLE")
    require.NoError(t, err)

    // Cleanup
    defer CleanupSnowflake(t, db, []string{"MY_TEST_ROLE"})
}
```

### Helper Functions

- `SetupLocalStackSnowflake(t)` - Connect to LocalStack Snowflake
- `CleanupSnowflake(t, db, roles)` - Drop test roles
- `CleanupUsers(t, db, users)` - Drop test users
- `GetCurrentRoles(t, db)` - Query all roles
- `GetCurrentUsers(t, db)` - Query all users
- `IsLocalStackAvailable()` - Check if LocalStack is running

### Skipping Tests

Tests automatically skip if:
- Running in short mode (`go test -short`)
- LocalStack is not available
- `USE_LOCALSTACK` env var is not set

## Test Coverage

Current integration tests:

- ✅ Config parser with YAML files
- ✅ Snowflake connection to LocalStack
- ✅ Role creation and verification
- ✅ StateReader for reading Snowflake state
- ✅ Config validation (valid/invalid YAML)
- 🚧 Full sync workflow (WIP)
- 🚧 Diff engine (TODO)
- 🚧 Parallel executor (TODO)

## Troubleshooting

### LocalStack won't start

```bash
# Check Docker logs
docker-compose logs localstack

# Verify auth token
echo $LOCALSTACK_AUTH_TOKEN

# Check if port 4566 is available
lsof -i :4566
```

### Tests fail with connection error

```bash
# Check LocalStack health
curl http://localhost:4566/_localstack/health

# Verify Snowflake service is running
curl http://localhost:4566/_localstack/health | jq '.services.snowflake'

# Check environment variables
env | grep LOCALSTACK
```

### "Skip: LocalStack is not available"

```bash
# Make sure LocalStack is running
docker-compose ps

# Set USE_LOCALSTACK env var
export USE_LOCALSTACK=true

# Verify connection
curl http://localhost:4566/_localstack/health
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: LocalStack/setup-localstack@v0.2.2
        with:
          image-tag: 'latest'

      - name: Run integration tests
        env:
          USE_LOCALSTACK: true
          LOCALSTACK_AUTH_TOKEN: ${{ secrets.LOCALSTACK_AUTH_TOKEN }}
        run: go test -v ./test/integration/...
```

## Performance

LocalStack Snowflake emulator performance:
- **Startup time**: ~10-15 seconds
- **Query latency**: <100ms for simple queries
- **Parallel ops**: Supports concurrent connections
- **Limitations**: Not all Snowflake features supported

Good for:
- ✅ Unit and integration testing
- ✅ CI/CD pipelines
- ✅ Local development
- ✅ Fast iteration

Not recommended for:
- ❌ Production use
- ❌ Performance benchmarking
- ❌ Testing Snowflake-specific optimizations

## Future Enhancements

- [ ] Add benchmark tests for parallel execution
- [ ] Test circuit breaker failure scenarios
- [ ] E2E tests with full sync workflow
- [ ] Test PostgreSQL audit logging integration
- [ ] Snapshot testing for SQL generation
- [ ] Chaos testing (random failures, network issues)

## Resources

- [LocalStack Snowflake Docs](https://docs.localstack.cloud/snowflake)
- [GateKeep Implementation Plan](../docs/implementation-plan.md)
- [API Documentation](../docs/api-server.md)

---

**Last Updated**: 2026-04-03
**Maintainer**: @gunjankaphle
