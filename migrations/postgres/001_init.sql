-- GateKeep Database Schema
-- Version: 1.0
-- Description: Initial schema for sync runs and operations audit log

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- High-level sync runs table
CREATE TABLE sync_runs (
    id BIGSERIAL PRIMARY KEY,
    sync_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failed', 'partial')),
    config_hash VARCHAR(64) NOT NULL,
    config_path VARCHAR(500),
    total_operations INT DEFAULT 0,
    successful_operations INT DEFAULT 0,
    failed_operations INT DEFAULT 0,
    duration_ms BIGINT,
    error_message TEXT,
    triggered_by VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Individual operations within a sync run
CREATE TABLE sync_operations (
    id BIGSERIAL PRIMARY KEY,
    sync_run_id BIGINT NOT NULL REFERENCES sync_runs(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,
    target_object VARCHAR(500) NOT NULL,
    sql_statement TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'success', 'failed', 'skipped')),
    error_message TEXT,
    execution_time_ms INT,
    executed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_sync_runs_started_at ON sync_runs(started_at DESC);
CREATE INDEX idx_sync_runs_sync_id ON sync_runs(sync_id);
CREATE INDEX idx_sync_runs_status ON sync_runs(status);
CREATE INDEX idx_sync_runs_config_hash ON sync_runs(config_hash);
CREATE INDEX idx_sync_operations_sync_run_id ON sync_operations(sync_run_id);
CREATE INDEX idx_sync_operations_status ON sync_operations(status);
CREATE INDEX idx_sync_operations_operation_type ON sync_operations(operation_type);

-- Index for recent queries (30 days)
CREATE INDEX idx_sync_runs_recent
ON sync_runs(started_at DESC)
WHERE started_at > NOW() - INTERVAL '30 days';

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at
CREATE TRIGGER update_sync_runs_updated_at
    BEFORE UPDATE ON sync_runs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to cleanup old sync runs (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_sync_runs()
RETURNS TABLE(deleted_count BIGINT) AS $$
DECLARE
    count BIGINT;
BEGIN
    DELETE FROM sync_runs
    WHERE started_at < NOW() - INTERVAL '30 days'
    AND status IN ('success', 'failed', 'partial');

    GET DIAGNOSTICS count = ROW_COUNT;
    RETURN QUERY SELECT count;
END;
$$ LANGUAGE plpgsql;

-- Optional: Create a view for recent sync history with summary
CREATE VIEW recent_sync_history AS
SELECT
    sr.id,
    sr.sync_id,
    sr.started_at,
    sr.completed_at,
    sr.status,
    sr.config_path,
    sr.total_operations,
    sr.successful_operations,
    sr.failed_operations,
    sr.duration_ms,
    EXTRACT(EPOCH FROM (sr.completed_at - sr.started_at)) * 1000 AS calculated_duration_ms,
    COUNT(so.id) AS operation_count
FROM sync_runs sr
LEFT JOIN sync_operations so ON sr.id = so.sync_run_id
WHERE sr.started_at > NOW() - INTERVAL '30 days'
GROUP BY sr.id
ORDER BY sr.started_at DESC;

-- Comments for documentation
COMMENT ON TABLE sync_runs IS 'High-level sync run tracking';
COMMENT ON TABLE sync_operations IS 'Individual operations executed during sync';
COMMENT ON COLUMN sync_runs.config_hash IS 'SHA256 hash of config file to detect changes';
COMMENT ON COLUMN sync_runs.status IS 'pending: not started, running: in progress, success: all ops succeeded, failed: all ops failed, partial: some ops failed';
COMMENT ON COLUMN sync_operations.operation_type IS 'CREATE_ROLE, DROP_ROLE, GRANT_ROLE, GRANT_PRIVILEGE, REVOKE_ROLE, REVOKE_PRIVILEGE';
