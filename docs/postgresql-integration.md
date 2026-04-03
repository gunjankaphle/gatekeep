# PostgreSQL Integration (Optional)

GateKeep supports **optional** PostgreSQL integration for audit logging. If PostgreSQL is not configured, GateKeep will run with a no-op audit logger (no audit data is stored).

## Why PostgreSQL is Optional

- **Simplicity**: Run GateKeep without setting up a database for simple use cases
- **Flexibility**: Choose when you need audit logging
- **Graceful Degradation**: GateKeep continues to work even if PostgreSQL connection fails

## When to Use PostgreSQL

Use PostgreSQL integration when you need:
- **Audit Trail**: Track all sync operations, timestamps, and outcomes
- **Compliance**: Maintain a record of permission changes for compliance/security
- **History**: View past sync runs and their results via API
- **Troubleshooting**: Investigate failed operations with detailed logs

## Configuration

### Enabling PostgreSQL (Recommended for Production)

Set the `POSTGRES_DSN` environment variable:

```bash
export POSTGRES_DSN="postgres://user:password@localhost:5432/gatekeep?sslmode=disable"
```

**Alternative**: Set individual PostgreSQL variables:

```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=gatekeep
export POSTGRES_PASSWORD=your_password
export POSTGRES_DB=gatekeep
```

### Running Without PostgreSQL

Simply **omit** the PostgreSQL environment variables. GateKeep will automatically use a no-op audit logger:

```bash
# PostgreSQL not configured - audit logging disabled
./gatekeep sync --config prod.yaml
```

You'll see this log message:
```
PostgreSQL not configured (POSTGRES_DSN not set) - audit logging disabled
```

## Database Setup

If using PostgreSQL, run the migration to create the required tables:

```bash
psql $POSTGRES_DSN -f migrations/postgres/001_init.sql
```

This creates:
- `sync_runs` - High-level sync execution tracking
- `sync_operations` - Individual SQL operations
- Indexes for performance
- Cleanup functions for retention management

## Database Schema

### sync_runs table

Tracks each sync execution:

| Column | Type | Description |
|--------|------|-------------|
| id | BIGSERIAL | Primary key |
| sync_id | UUID | Unique identifier for this sync |
| started_at | TIMESTAMPTZ | When sync started |
| completed_at | TIMESTAMPTZ | When sync completed (nullable) |
| status | VARCHAR(20) | `pending`, `running`, `success`, `failed`, `partial` |
| config_hash | VARCHAR(64) | SHA256 hash of config file |
| config_path | VARCHAR(500) | Path to config file |
| total_operations | INT | Total operations planned |
| successful_operations | INT | Operations that succeeded |
| failed_operations | INT | Operations that failed |
| duration_ms | BIGINT | Total execution time in milliseconds |
| error_message | TEXT | Error message if sync failed |

### sync_operations table

Tracks individual SQL operations:

| Column | Type | Description |
|--------|------|-------------|
| id | BIGSERIAL | Primary key |
| sync_run_id | BIGINT | Foreign key to sync_runs |
| operation_type | VARCHAR(50) | `CREATE_ROLE`, `GRANT`, `REVOKE`, etc. |
| target_object | VARCHAR(500) | Role/user/database being modified |
| sql_statement | TEXT | Exact SQL executed |
| status | VARCHAR(20) | `pending`, `success`, `failed`, `skipped` |
| error_message | TEXT | Error message if operation failed |
| execution_time_ms | INT | Execution time in milliseconds |
| executed_at | TIMESTAMPTZ | When operation was executed |

## API Endpoints (PostgreSQL Required)

These endpoints require PostgreSQL to be configured:

### GET /api/sync/history
List recent sync runs with pagination.

**Query Parameters:**
- `page` - Page number (default: 1)
- `page_size` - Results per page (default: 20, max: 100)

**Response:**
```json
{
  "sync_runs": [
    {
      "id": 123,
      "sync_id": "uuid",
      "started_at": "2026-03-30T10:00:00Z",
      "completed_at": "2026-03-30T10:00:05Z",
      "status": "success",
      "config_path": "/path/to/config.yaml",
      "total_operations": 45,
      "successful_operations": 45,
      "failed_operations": 0,
      "duration_ms": 5000
    }
  ],
  "total_count": 10,
  "page": 1,
  "page_size": 20
}
```

### GET /api/sync/history/:id
Get detailed information about a specific sync run.

**Response:**
```json
{
  "id": 123,
  "sync_id": "uuid",
  "started_at": "2026-03-30T10:00:00Z",
  "completed_at": "2026-03-30T10:00:05Z",
  "status": "success",
  "total_operations": 45,
  "successful_operations": 45,
  "failed_operations": 0,
  "duration_ms": 5000,
  "operations": [
    {
      "id": 1,
      "operation_type": "CREATE_ROLE",
      "target_object": "ANALYST",
      "sql_statement": "CREATE ROLE ANALYST;",
      "status": "success",
      "execution_time_ms": 120,
      "executed_at": "2026-03-30T10:00:01Z"
    }
  ]
}
```

### GET /api/health
Health check includes database status.

**Response (with PostgreSQL):**
```json
{
  "status": "healthy",
  "timestamp": "2026-03-30T10:00:00Z",
  "services": {
    "database": "ok"
  }
}
```

**Response (without PostgreSQL):**
```json
{
  "status": "healthy",
  "timestamp": "2026-03-30T10:00:00Z",
  "services": {
    "database": "not_configured"
  }
}
```

## Data Retention

By default, audit logs older than **30 days** are eligible for cleanup.

### Manual Cleanup

```sql
SELECT * FROM cleanup_old_sync_runs();
```

### Automated Cleanup (pg_cron)

Install pg_cron extension:

```sql
CREATE EXTENSION pg_cron;

-- Schedule daily cleanup at 2 AM
SELECT cron.schedule('cleanup-sync-runs', '0 2 * * *', 'SELECT cleanup_old_sync_runs()');
```

### Custom Retention Period

Modify the `cleanup_old_sync_runs()` function in the migration file to adjust retention:

```sql
-- Change 30 days to 90 days
WHERE started_at < NOW() - INTERVAL '90 days'
```

## Performance Optimization

For high-volume deployments (>100k operations/day), consider:

### 1. Partitioning by Date

```sql
-- Create monthly partitions
CREATE TABLE sync_runs_2026_03 PARTITION OF sync_runs
FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE sync_runs_2026_04 PARTITION OF sync_runs
FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
```

### 2. TimescaleDB (Optional)

For even better time-series performance:

```sql
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Convert to hypertable
SELECT create_hypertable('sync_runs', 'started_at', if_not_exists => TRUE);

-- Automatic retention policy
SELECT add_retention_policy('sync_runs', INTERVAL '30 days');
```

## Troubleshooting

### Connection Errors

If PostgreSQL connection fails, GateKeep will:
1. Log a warning: `⚠ PostgreSQL connection failed: <error> - using no-op audit logger`
2. Continue running without audit logging
3. Return `database: "not_configured"` in health checks

### Migration Errors

If migration fails, check:
- PostgreSQL version (9.6+)
- User has `CREATE` permissions
- `uuid-ossp` extension can be installed

### Query Performance

If history queries are slow:
- Check if indexes are created: `\d sync_runs` in psql
- Ensure PostgreSQL has adequate resources
- Consider partitioning for large datasets
- Use `EXPLAIN ANALYZE` to diagnose slow queries

## Example: Programmatic Usage

```go
import (
    "context"
    "github.com/yourusername/gatekeep/internal/database"
    "github.com/yourusername/gatekeep/internal/sync"
)

func main() {
    ctx := context.Background()

    // Create audit logger (gracefully handles missing PostgreSQL)
    auditLogger, pool, err := database.NewAuditLogger(ctx)
    if err != nil {
        // Non-fatal - continues with no-op logger
        log.Printf("Audit logging unavailable: %v", err)
    }
    defer func() {
        if pool != nil {
            pool.Close()
        }
    }()

    // Create orchestrator with audit logger
    orchestrator := sync.NewOrchestrator(
        configParser,
        stateReader,
        executor,
        syncMode,
    ).WithAuditLogger(auditLogger)

    // Run sync - audit logging happens automatically if configured
    result, err := orchestrator.Sync(ctx, "config.yaml", sync.DefaultConfig())
}
```

## Summary

- **PostgreSQL is OPTIONAL** - GateKeep works without it
- **Set `POSTGRES_DSN`** to enable audit logging
- **Omit PostgreSQL config** to run without audit logging
- **Graceful degradation** - connection failures don't break syncs
- **Use for production** where audit trails are important
- **Skip for dev/testing** where simplicity is preferred
