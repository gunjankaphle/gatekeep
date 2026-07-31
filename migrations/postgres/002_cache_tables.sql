-- GateKeep Cache Tables Schema
-- Version: 2.0
-- Description: Cache tables for Snowflake data (roles, grants)

-- Cache roles table: stores role hierarchy snapshot from Snowflake
CREATE TABLE cached_roles (
    name VARCHAR(255) PRIMARY KEY,
    parent_roles TEXT[], -- Array of parent role names
    comment TEXT,
    owner VARCHAR(255),
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Cache grants table: stores all grants by role
CREATE TABLE cached_grants (
    id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(255) NOT NULL,
    granted_on VARCHAR(50) NOT NULL,      -- ROLE, TABLE, WAREHOUSE, DATABASE, etc.
    object_name VARCHAR(500) NOT NULL,    -- Name of the granted object
    privilege VARCHAR(50) NOT NULL,       -- SELECT, INSERT, USAGE, etc.
    grantee_type VARCHAR(50) NOT NULL,    -- ROLE or USER
    grantee_name VARCHAR(255) NOT NULL,   -- Name of the grantee
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (role_name, granted_on, object_name, privilege, grantee_name)
);

-- Cache metadata table: tracks cache refresh status
CREATE TABLE cache_metadata (
    cache_key VARCHAR(50) PRIMARY KEY,
    last_refresh TIMESTAMPTZ,
    refresh_status VARCHAR(20) NOT NULL DEFAULT 'idle', -- 'idle', 'refreshing', 'success', 'failed'
    error_message TEXT,
    total_roles INT DEFAULT 0,
    total_grants INT DEFAULT 0,
    duration_ms BIGINT DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_cached_roles_updated ON cached_roles(last_updated DESC);
CREATE INDEX idx_cached_grants_role ON cached_grants(role_name);
CREATE INDEX idx_cached_grants_object ON cached_grants(object_name);
CREATE INDEX idx_cached_grants_updated ON cached_grants(last_updated DESC);

-- Initialize cache metadata
INSERT INTO cache_metadata (cache_key, refresh_status)
VALUES ('roles_and_grants', 'idle')
ON CONFLICT (cache_key) DO NOTHING;

-- Trigger to update cache_metadata.updated_at
CREATE TRIGGER update_cache_metadata_updated_at
    BEFORE UPDATE ON cache_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE cached_roles IS 'Cached snapshot of Snowflake roles hierarchy';
COMMENT ON TABLE cached_grants IS 'Cached snapshot of Snowflake grants by role';
COMMENT ON TABLE cache_metadata IS 'Metadata tracking cache refresh status and statistics';
COMMENT ON COLUMN cached_roles.parent_roles IS 'Array of parent role names (role inheritance)';
COMMENT ON COLUMN cache_metadata.refresh_status IS 'Current refresh status: idle, refreshing, success, failed';
