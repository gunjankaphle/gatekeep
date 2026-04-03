# GateKeep API Server (Read-Only Mode)

The GateKeep API server provides **read-only** HTTP endpoints for querying roles, permissions, and sync history. YAML files remain the single source of truth for configuration.

## Philosophy: Read-Only by Design

**Why Read-Only?**
- **GitOps Best Practice**: YAML files in Git provide audit trail, versioning, and code review
- **Security**: No accidental permission changes through API
- **Simplicity**: API is for viewing/querying only, CLI handles mutations
- **Compliance**: All changes go through Git workflow (PR → Review → Merge)

**For Write Operations**: Use the CLI
```bash
# Sync permissions
gatekeep sync --config prod.yaml

# Dry-run (preview changes)
gatekeep sync --config prod.yaml --dry-run
```

## Quick Start

### 1. Start the Server

```bash
# Using the binary
./bin/gatekeep-server

# Or with go run
go run ./cmd/server

# Or with make
make run-server
```

### 2. Configure (Optional)

```bash
# Set config file path (default: configs/example.yaml)
export GATEKEEP_CONFIG_PATH=/path/to/your/config.yaml

# Set PostgreSQL for audit history (optional)
export POSTGRES_DSN="postgres://user:pass@localhost:5432/gatekeep"

# Set server port (default: 8080)
export SERVER_PORT=8080
```

### 3. Access the API

```bash
# Health check
curl http://localhost:8080/api/health

# List roles from config
curl http://localhost:8080/api/roles

# View sync history (requires PostgreSQL)
curl http://localhost:8080/api/sync/history
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `GATEKEEP_CONFIG_PATH` | `configs/example.yaml` | Path to YAML config file |
| `SERVER_READ_TIMEOUT` | `30s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | `30s` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `POSTGRES_DSN` | *(optional)* | PostgreSQL connection string |

## API Endpoints

### GET /api/health

Health check endpoint - returns server and database status.

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2026-03-30T10:00:00Z",
  "services": {
    "database": "ok"
  }
}
```

**Without PostgreSQL**:
```json
{
  "status": "healthy",
  "timestamp": "2026-03-30T10:00:00Z",
  "services": {
    "database": "not_configured"
  }
}
```

---

### GET /api/roles

List all roles from the YAML configuration file.

**Query Parameters**: None

**Response** (200 OK):
```json
{
  "roles": [
    {
      "name": "ANALYST_ROLE",
      "parent_roles": ["READ_ONLY_ROLE"],
      "comment": "Data analysts role"
    },
    {
      "name": "ENGINEER_ROLE",
      "parent_roles": [],
      "comment": "Engineering team role"
    }
  ],
  "count": 2
}
```

**Example**:
```bash
curl http://localhost:8080/api/roles | jq .
```

---

### GET /api/sync/history

List recent sync runs from audit log.

**Requires**: PostgreSQL configured

**Query Parameters**:
- `page` - Page number (default: 1)
- `page_size` - Results per page (default: 20, max: 100)

**Response** (200 OK):
```json
{
  "sync_runs": [
    {
      "id": 123,
      "sync_id": "550e8400-e29b-41d4-a716-446655440000",
      "started_at": "2026-03-30T10:00:00Z",
      "completed_at": "2026-03-30T10:00:05Z",
      "status": "success",
      "config_path": "/path/to/prod.yaml",
      "total_operations": 45,
      "successful_operations": 45,
      "failed_operations": 0,
      "duration_ms": 5000
    }
  ],
  "total_count": 1,
  "page": 1,
  "page_size": 20
}
```

**Example**:
```bash
# Get last 10 sync runs
curl "http://localhost:8080/api/sync/history?page_size=10" | jq .

# Get page 2
curl "http://localhost:8080/api/sync/history?page=2&page_size=20" | jq .
```

---

### GET /api/sync/history/:id

Get detailed information about a specific sync run, including all operations.

**Requires**: PostgreSQL configured

**Path Parameters**:
- `id` - Sync run ID (integer)

**Response** (200 OK):
```json
{
  "id": 123,
  "sync_id": "550e8400-e29b-41d4-a716-446655440000",
  "started_at": "2026-03-30T10:00:00Z",
  "completed_at": "2026-03-30T10:00:05Z",
  "status": "success",
  "total_operations": 2,
  "successful_operations": 2,
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
    },
    {
      "id": 2,
      "operation_type": "GRANT",
      "target_object": "ANALYST",
      "sql_statement": "GRANT SELECT ON TABLE customers TO ROLE ANALYST;",
      "status": "success",
      "execution_time_ms": 95,
      "executed_at": "2026-03-30T10:00:02Z"
    }
  ]
}
```

**Example**:
```bash
curl http://localhost:8080/api/sync/history/123 | jq .
```

---

### POST /api/sync

**⚠️ Not Implemented (Read-Only Mode)**

This endpoint returns `501 Not Implemented` with instructions to use the CLI.

**Response** (501 Not Implemented):
```json
{
  "error": "not_implemented",
  "message": "Sync operations are not available through the API in read-only mode",
  "details": "Use the CLI for sync operations: gatekeep sync --config <file>",
  "cli_usage": {
    "sync": "gatekeep sync --config prod.yaml",
    "dry-run": "gatekeep sync --config prod.yaml --dry-run"
  }
}
```

---

### POST /api/sync/dry-run

**⚠️ Not Implemented (Read-Only Mode)**

This endpoint returns `501 Not Implemented` with instructions to use the CLI.

**Response** (501 Not Implemented):
```json
{
  "error": "not_implemented",
  "message": "Dry-run operations are not available through the API in read-only mode",
  "details": "Use the CLI for dry-run operations: gatekeep sync --config <file> --dry-run",
  "cli_usage": {
    "dry-run": "gatekeep sync --config prod.yaml --dry-run",
    "format": "gatekeep sync --config prod.yaml --dry-run --format json"
  }
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_type",
  "message": "Human-readable error message",
  "code": 400
}
```

### Common Error Codes

| Code | Error Type | Description |
|------|------------|-------------|
| 400 | Bad Request | Invalid request parameters |
| 404 | Not Found | Resource not found |
| 500 | Internal Server Error | Server error |
| 501 | Not Implemented | Feature not available (write operations) |
| 503 | Service Unavailable | Service unhealthy |

## Middleware

The API server includes the following middleware:

1. **Recovery**: Catches panics and returns 500 errors
2. **Logger**: Logs all requests (method, path, duration, status)
3. **Request ID**: Adds unique `X-Request-ID` header to each request
4. **CORS**: Enables cross-origin requests (configured for web UIs)

## Usage Examples

### Check Server Health

```bash
curl http://localhost:8080/api/health
```

### Query Roles

```bash
# Get all roles
curl http://localhost:8080/api/roles | jq '.roles'

# Count roles
curl http://localhost:8080/api/roles | jq '.count'

# Find specific role
curl http://localhost:8080/api/roles | jq '.roles[] | select(.name=="ANALYST_ROLE")'
```

### View Sync History

```bash
# Last 5 sync runs
curl "http://localhost:8080/api/sync/history?page_size=5" | jq '.sync_runs'

# Get details of sync run #123
curl http://localhost:8080/api/sync/history/123 | jq .

# Show failed operations from sync #123
curl http://localhost:8080/api/sync/history/123 | \
  jq '.operations[] | select(.status=="failed")'
```

## Docker Deployment

```dockerfile
# Dockerfile already configured
docker build -t gatekeep:latest .

# Run server
docker run -p 8080:8080 \
  -e GATEKEEP_CONFIG_PATH=/app/config.yaml \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  gatekeep:latest gatekeep-server
```

## Docker Compose

```yaml
version: '3.8'

services:
  gatekeep-api:
    image: gatekeep:latest
    command: gatekeep-server
    ports:
      - "8080:8080"
    environment:
      - SERVER_PORT=8080
      - GATEKEEP_CONFIG_PATH=/app/config.yaml
      - POSTGRES_DSN=postgres://gatekeep:password@postgres:5432/gatekeep
    volumes:
      - ./prod.yaml:/app/config.yaml:ro
    depends_on:
      - postgres

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_USER=gatekeep
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=gatekeep
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations/postgres:/docker-entrypoint-initdb.d:ro

volumes:
  postgres_data:
```

## Security Considerations

1. **Read-Only by Design**: No state mutations through API
2. **YAML as Source of Truth**: All changes require Git workflow
3. **Audit Logging**: All sync operations logged (if PostgreSQL configured)
4. **No Credentials in API**: Snowflake credentials only in CLI environment
5. **CORS Configured**: Review and restrict in production
6. **Request Logging**: All requests logged for audit trail

## Monitoring

### Prometheus Metrics (Future)

*Not yet implemented - planned for future release*

### Health Check for Load Balancers

```bash
# Use health endpoint
curl http://localhost:8080/api/health
```

Returns 200 if healthy, 503 if unhealthy.

## Future Enhancements

- [ ] Prometheus metrics endpoint (`/metrics`)
- [ ] OpenAPI/Swagger documentation endpoint
- [ ] Authentication/authorization (JWT, OAuth2)
- [ ] Rate limiting
- [ ] GraphQL endpoint for complex queries
- [ ] WebSocket for real-time sync status updates
- [ ] Export endpoints (CSV, JSON, Parquet)

## Troubleshooting

### Server won't start

```bash
# Check if port is already in use
lsof -i :8080

# Check config file path
ls -la $GATEKEEP_CONFIG_PATH

# Check logs
./bin/gatekeep-server 2>&1 | tee server.log
```

### History endpoints return empty

- Ensure PostgreSQL is configured and running
- Check `POSTGRES_DSN` environment variable
- Verify migrations were run: `psql $POSTGRES_DSN -f migrations/postgres/001_init.sql`
- Check server logs for PostgreSQL connection errors

### Roles endpoint returns empty

- Verify config file exists at `$GATEKEEP_CONFIG_PATH`
- Check config file is valid YAML
- Validate config: `gatekeep validate $GATEKEEP_CONFIG_PATH`

## Support

- **Documentation**: See `docs/` directory
- **Issues**: https://github.com/yourusername/gatekeep/issues
- **CLI Help**: `gatekeep --help`

---

**Read-Only API Server** - YAML files remain the source of truth
