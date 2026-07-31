package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

// CacheRepository handles cache operations in Postgres
type CacheRepository struct {
	pool *pgxpool.Pool
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(pool *pgxpool.Pool) *CacheRepository {
	return &CacheRepository{pool: pool}
}

// RefreshCache atomically replaces all cached data
func (r *CacheRepository) RefreshCache(ctx context.Context, params CacheRefreshParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx) //nolint:errcheck // safe to call after commit; rollback error is not actionable
	}()

	// 1. Update metadata to 'refreshing'
	_, err = tx.Exec(ctx, `
		UPDATE cache_metadata
		SET refresh_status = $1, updated_at = NOW()
		WHERE cache_key = $2
	`, CacheStatusRefreshing, "roles_and_grants")
	if err != nil {
		return fmt.Errorf("failed to update metadata status: %w", err)
	}

	// 2. Truncate existing cache data
	_, err = tx.Exec(ctx, "TRUNCATE cached_roles, cached_grants")
	if err != nil {
		return fmt.Errorf("failed to truncate cache tables: %w", err)
	}

	// 3. Bulk insert roles
	if len(params.Roles) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"cached_roles"},
			[]string{"name", "parent_roles", "comment", "owner", "last_updated"},
			pgx.CopyFromSlice(len(params.Roles), func(i int) ([]interface{}, error) {
				role := params.Roles[i]
				return []interface{}{
					role.Name,
					pq.Array(role.ParentRoles),
					role.Comment,
					role.Owner,
					time.Now(),
				}, nil
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to insert roles: %w", err)
		}
	}

	// 4. Bulk insert grants
	if len(params.Grants) > 0 {
		_, err = tx.CopyFrom(
			ctx,
			pgx.Identifier{"cached_grants"},
			[]string{"role_name", "granted_on", "object_name", "privilege", "grantee_type", "grantee_name", "last_updated"},
			pgx.CopyFromSlice(len(params.Grants), func(i int) ([]interface{}, error) {
				grant := params.Grants[i]
				return []interface{}{
					grant.RoleName,
					grant.GrantedOn,
					grant.ObjectName,
					grant.Privilege,
					grant.GranteeType,
					grant.GranteeName,
					time.Now(),
				}, nil
			}),
		)
		if err != nil {
			return fmt.Errorf("failed to insert grants: %w", err)
		}
	}

	// 5. Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateCacheMetadata updates cache metadata after refresh
func (r *CacheRepository) UpdateCacheMetadata(ctx context.Context, status CacheStatus, totalRoles, totalGrants int, durationMs int64, errorMsg *string) error {
	query := `
		UPDATE cache_metadata
		SET refresh_status = $1,
		    last_refresh = NOW(),
		    total_roles = $2,
		    total_grants = $3,
		    duration_ms = $4,
		    error_message = $5,
		    updated_at = NOW()
		WHERE cache_key = $6
	`

	_, err := r.pool.Exec(ctx, query, status, totalRoles, totalGrants, durationMs, errorMsg, "roles_and_grants")
	if err != nil {
		return fmt.Errorf("failed to update cache metadata: %w", err)
	}

	return nil
}

// GetCacheMetadata retrieves current cache metadata
func (r *CacheRepository) GetCacheMetadata(ctx context.Context) (*CacheMetadata, error) {
	query := `
		SELECT cache_key, last_refresh, refresh_status, error_message,
		       total_roles, total_grants, duration_ms, updated_at
		FROM cache_metadata
		WHERE cache_key = $1
	`

	var metadata CacheMetadata
	err := r.pool.QueryRow(ctx, query, "roles_and_grants").Scan(
		&metadata.CacheKey,
		&metadata.LastRefresh,
		&metadata.RefreshStatus,
		&metadata.ErrorMessage,
		&metadata.TotalRoles,
		&metadata.TotalGrants,
		&metadata.DurationMs,
		&metadata.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get cache metadata: %w", err)
	}

	return &metadata, nil
}

// GetRoleHierarchy retrieves cached roles with pagination
func (r *CacheRepository) GetRoleHierarchy(ctx context.Context, filter RoleHierarchyFilter) (*RoleHierarchyResponse, error) {
	// Get total count
	var totalCount int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM cached_roles").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated roles
	query := `
		SELECT name, parent_roles, comment, owner, last_updated, created_at
		FROM cached_roles
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []CachedRole
	for rows.Next() {
		var role CachedRole
		var parentRoles []string

		err = rows.Scan(
			&role.Name,
			pq.Array(&parentRoles),
			&role.Comment,
			&role.Owner,
			&role.LastUpdated,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		role.ParentRoles = parentRoles
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return &RoleHierarchyResponse{
		Roles:      roles,
		TotalCount: totalCount,
	}, nil
}

// GetGrantsForRole retrieves all grants for a specific role
func (r *CacheRepository) GetGrantsForRole(ctx context.Context, roleName string) ([]CachedGrant, error) {
	query := `
		SELECT id, role_name, granted_on, object_name, privilege,
		       grantee_type, grantee_name, last_updated, created_at
		FROM cached_grants
		WHERE role_name = $1
		ORDER BY granted_on, object_name, privilege
	`

	rows, err := r.pool.Query(ctx, query, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to query grants: %w", err)
	}
	defer rows.Close()

	var grants []CachedGrant
	for rows.Next() {
		var grant CachedGrant
		err = rows.Scan(
			&grant.ID,
			&grant.RoleName,
			&grant.GrantedOn,
			&grant.ObjectName,
			&grant.Privilege,
			&grant.GranteeType,
			&grant.GranteeName,
			&grant.LastUpdated,
			&grant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grant: %w", err)
		}

		grants = append(grants, grant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating grants: %w", err)
	}

	return grants, nil
}

// GetGrantsForRoles retrieves grants for multiple roles (for comparison)
func (r *CacheRepository) GetGrantsForRoles(ctx context.Context, roleNames []string) (map[string][]CachedGrant, error) {
	query := `
		SELECT id, role_name, granted_on, object_name, privilege,
		       grantee_type, grantee_name, last_updated, created_at
		FROM cached_grants
		WHERE role_name = ANY($1)
		ORDER BY role_name, granted_on, object_name, privilege
	`

	rows, err := r.pool.Query(ctx, query, roleNames)
	if err != nil {
		return nil, fmt.Errorf("failed to query grants: %w", err)
	}
	defer rows.Close()

	grantsByRole := make(map[string][]CachedGrant)
	for rows.Next() {
		var grant CachedGrant
		err = rows.Scan(
			&grant.ID,
			&grant.RoleName,
			&grant.GrantedOn,
			&grant.ObjectName,
			&grant.Privilege,
			&grant.GranteeType,
			&grant.GranteeName,
			&grant.LastUpdated,
			&grant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grant: %w", err)
		}

		grantsByRole[grant.RoleName] = append(grantsByRole[grant.RoleName], grant)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating grants: %w", err)
	}

	return grantsByRole, nil
}
