package repository

import "time"

// CachedRole represents a role stored in the cache
type CachedRole struct {
	Name        string
	ParentRoles []string
	Comment     string
	Owner       string
	LastUpdated time.Time
	CreatedAt   time.Time
}

// CachedGrant represents a grant stored in the cache
type CachedGrant struct {
	ID          int64
	RoleName    string
	GrantedOn   string
	ObjectName  string
	Privilege   string
	GranteeType string
	GranteeName string
	LastUpdated time.Time
	CreatedAt   time.Time
}

// CacheMetadata represents cache refresh metadata
type CacheMetadata struct {
	CacheKey      string
	LastRefresh   *time.Time
	RefreshStatus CacheStatus
	ErrorMessage  *string
	TotalRoles    int
	TotalGrants   int
	DurationMs    int64
	UpdatedAt     time.Time
}

// CacheStatus represents the cache refresh status
type CacheStatus string

// Cache refresh status values.
const (
	CacheStatusIdle       CacheStatus = "idle"
	CacheStatusRefreshing CacheStatus = "refreshing"
	CacheStatusSuccess    CacheStatus = "success"
	CacheStatusFailed     CacheStatus = "failed"
)

// CacheRefreshParams contains data for refreshing the cache
type CacheRefreshParams struct {
	Roles  []CachedRole
	Grants []CachedGrant
}

// RoleHierarchyFilter represents filters for querying role hierarchy
type RoleHierarchyFilter struct {
	Limit  int
	Offset int
}

// RoleHierarchyResponse represents paginated role hierarchy
type RoleHierarchyResponse struct {
	Roles      []CachedRole
	TotalCount int
}
