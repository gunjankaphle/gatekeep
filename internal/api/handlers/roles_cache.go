package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/gatekeep/internal/repository"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// RolesCacheHandler handles role cache requests with Snowflake refresh
type RolesCacheHandler struct {
	cacheRepo       *repository.CacheRepository
	snowflakeClient snowflake.Client
	maxWorkers      int
	refreshMu       sync.Mutex
	isRefreshing    bool
}

// NewRolesCacheHandler creates a new roles cache handler
func NewRolesCacheHandler(cacheRepo *repository.CacheRepository, sfClient snowflake.Client, maxWorkers int) *RolesCacheHandler {
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default to 10 concurrent workers
	}

	return &RolesCacheHandler{
		cacheRepo:       cacheRepo,
		snowflakeClient: sfClient,
		maxWorkers:      maxWorkers,
	}
}

// GetRoleHierarchy handles GET /api/roles/hierarchy
func (h *RolesCacheHandler) GetRoleHierarchy(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 50

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 200 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize

	// Get cache metadata
	metadata, err := h.cacheRepo.GetCacheMetadata(r.Context())
	if err != nil {
		errorResponse(w, "failed to get cache metadata", http.StatusInternalServerError, err)
		return
	}

	// Get paginated roles
	filter := repository.RoleHierarchyFilter{
		Limit:  pageSize,
		Offset: offset,
	}

	result, err := h.cacheRepo.GetRoleHierarchy(r.Context(), filter)
	if err != nil {
		errorResponse(w, "failed to get role hierarchy", http.StatusInternalServerError, err)
		return
	}

	// Convert to API response
	type roleInfo struct {
		Name        string    `json:"name"`
		ParentRoles []string  `json:"parent_roles"`
		Comment     string    `json:"comment,omitempty"`
		Owner       string    `json:"owner,omitempty"`
		LastUpdated time.Time `json:"last_updated"`
	}

	roles := make([]roleInfo, len(result.Roles))
	for i, role := range result.Roles {
		roles[i] = roleInfo{
			Name:        role.Name,
			ParentRoles: role.ParentRoles,
			Comment:     role.Comment,
			Owner:       role.Owner,
			LastUpdated: role.LastUpdated,
		}
	}

	totalPages := (result.TotalCount + pageSize - 1) / pageSize

	response := map[string]interface{}{
		"roles": roles,
		"pagination": map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_count": result.TotalCount,
			"total_pages": totalPages,
		},
		"cache_metadata": map[string]interface{}{
			"last_refresh": metadata.LastRefresh,
			"status":       string(metadata.RefreshStatus),
			"total_roles":  metadata.TotalRoles,
			"total_grants": metadata.TotalGrants,
			"duration_ms":  metadata.DurationMs,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = err // Response already written
	}
}

// TriggerRefresh handles POST /api/roles/refresh
func (h *RolesCacheHandler) TriggerRefresh(w http.ResponseWriter, r *http.Request) {
	// Guard against concurrent refreshes within this process; the cache_metadata
	// row alone isn't enough since two requests could both read it before either
	// writes 'refreshing'.
	h.refreshMu.Lock()
	if h.isRefreshing {
		h.refreshMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		response := map[string]interface{}{
			"error":   "Refresh already in progress",
			"message": "A cache refresh is already running",
			"code":    http.StatusConflict,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			_ = err
		}
		return
	}
	h.isRefreshing = true
	h.refreshMu.Unlock()

	// Start refresh in background
	go h.refreshCache(context.Background())

	// Return immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	response := map[string]interface{}{
		"status":  "refreshing",
		"message": "Cache refresh initiated",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = err
	}
}

// GetRefreshStatus handles GET /api/roles/refresh/status
func (h *RolesCacheHandler) GetRefreshStatus(w http.ResponseWriter, r *http.Request) {
	metadata, err := h.cacheRepo.GetCacheMetadata(r.Context())
	if err != nil {
		errorResponse(w, "failed to get cache metadata", http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"status":        string(metadata.RefreshStatus),
		"last_refresh":  metadata.LastRefresh,
		"duration_ms":   metadata.DurationMs,
		"total_roles":   metadata.TotalRoles,
		"total_grants":  metadata.TotalGrants,
		"error_message": metadata.ErrorMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = err
	}
}

// GetRoleGrants handles GET /api/roles/{roleName}/grants
func (h *RolesCacheHandler) GetRoleGrants(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "roleName")
	if roleName == "" {
		errorResponse(w, "role name is required", http.StatusBadRequest, nil)
		return
	}

	grants, err := h.cacheRepo.GetGrantsForRole(r.Context(), roleName)
	if err != nil {
		errorResponse(w, "failed to get grants", http.StatusInternalServerError, err)
		return
	}

	// Convert to API response
	type grantInfo struct {
		GrantedOn   string `json:"granted_on"`
		ObjectName  string `json:"object_name"`
		Privilege   string `json:"privilege"`
		GranteeType string `json:"grantee_type"`
		GranteeName string `json:"grantee_name"`
	}

	grantsList := make([]grantInfo, len(grants))
	for i, grant := range grants {
		grantsList[i] = grantInfo{
			GrantedOn:   grant.GrantedOn,
			ObjectName:  grant.ObjectName,
			Privilege:   grant.Privilege,
			GranteeType: grant.GranteeType,
			GranteeName: grant.GranteeName,
		}
	}

	response := map[string]interface{}{
		"role_name": roleName,
		"grants":    grantsList,
		"count":     len(grantsList),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = err
	}
}

// CompareRoles handles GET /api/roles/compare?roleA={name}&roleB={name}
func (h *RolesCacheHandler) CompareRoles(w http.ResponseWriter, r *http.Request) {
	roleA := r.URL.Query().Get("roleA")
	roleB := r.URL.Query().Get("roleB")

	if roleA == "" || roleB == "" {
		errorResponse(w, "both roleA and roleB are required", http.StatusBadRequest, nil)
		return
	}

	// Get grants for both roles
	grantsByRole, err := h.cacheRepo.GetGrantsForRoles(r.Context(), []string{roleA, roleB})
	if err != nil {
		errorResponse(w, "failed to get grants", http.StatusInternalServerError, err)
		return
	}

	grantsA := grantsByRole[roleA]
	grantsB := grantsByRole[roleB]

	// Compute diff
	diff := computeGrantsDiff(grantsA, grantsB)

	response := map[string]interface{}{
		"roleA": roleA,
		"roleB": roleB,
		"diff":  diff,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		_ = err
	}
}

// refreshCache performs the actual cache refresh from Snowflake
func (h *RolesCacheHandler) refreshCache(ctx context.Context) {
	defer func() {
		h.refreshMu.Lock()
		h.isRefreshing = false
		h.refreshMu.Unlock()
	}()

	start := time.Now()

	// Update status to 'refreshing'
	_ = h.cacheRepo.UpdateCacheMetadata(ctx, repository.CacheStatusRefreshing, 0, 0, 0, nil) //nolint:errcheck // best-effort status write; refresh proceeds regardless

	// Fetch from Snowflake using parallel queries
	stateReader := snowflake.NewStateReader(h.snowflakeClient)
	roles, grants, err := stateReader.GetAllRolesWithGrants(h.maxWorkers)
	if err != nil {
		errMsg := err.Error()
		_ = h.cacheRepo.UpdateCacheMetadata(ctx, repository.CacheStatusFailed, 0, 0, 0, &errMsg) //nolint:errcheck // best-effort status write; refresh has already failed
		return
	}

	// Convert Snowflake types to cache types
	cachedRoles := make([]repository.CachedRole, len(roles))
	for i, role := range roles {
		cachedRoles[i] = repository.CachedRole{
			Name:        role.Name,
			ParentRoles: []string{}, // Will be populated from grants
			Comment:     role.Comment,
			Owner:       role.Owner,
		}
	}

	cachedGrants := make([]repository.CachedGrant, len(grants))
	for i, grant := range grants {
		cachedGrants[i] = repository.CachedGrant{
			RoleName:    grant.GranteeName,
			GrantedOn:   grant.GrantedOn,
			ObjectName:  grant.Name,
			Privilege:   grant.Privilege,
			GranteeType: grant.GranteeType,
			GranteeName: grant.GranteeName,
		}
	}

	// Update cache in transaction
	params := repository.CacheRefreshParams{
		Roles:  cachedRoles,
		Grants: cachedGrants,
	}

	err = h.cacheRepo.RefreshCache(ctx, params)
	if err != nil {
		errMsg := err.Error()
		_ = h.cacheRepo.UpdateCacheMetadata(ctx, repository.CacheStatusFailed, 0, 0, 0, &errMsg) //nolint:errcheck // best-effort status write; refresh has already failed
		return
	}

	// Update metadata with success
	durationMs := time.Since(start).Milliseconds()
	_ = h.cacheRepo.UpdateCacheMetadata(ctx, repository.CacheStatusSuccess, len(cachedRoles), len(cachedGrants), durationMs, nil) //nolint:errcheck // best-effort status write; refresh itself already succeeded
}

// computeGrantsDiff computes the difference between two sets of grants
func computeGrantsDiff(grantsA, grantsB []repository.CachedGrant) map[string]interface{} {
	// Create maps for fast lookup
	mapA := make(map[string]repository.CachedGrant)
	mapB := make(map[string]repository.CachedGrant)

	for _, grant := range grantsA {
		key := fmt.Sprintf("%s:%s:%s:%s", grant.GrantedOn, grant.ObjectName, grant.Privilege, grant.GranteeName)
		mapA[key] = grant
	}

	for _, grant := range grantsB {
		key := fmt.Sprintf("%s:%s:%s:%s", grant.GrantedOn, grant.ObjectName, grant.Privilege, grant.GranteeName)
		mapB[key] = grant
	}

	// Compute diff
	var added, removed, unchanged []map[string]string

	// Find added (in B but not in A)
	for key, grant := range mapB {
		if _, exists := mapA[key]; !exists {
			added = append(added, map[string]string{
				"granted_on":   grant.GrantedOn,
				"object_name":  grant.ObjectName,
				"privilege":    grant.Privilege,
				"grantee_name": grant.GranteeName,
			})
		} else {
			unchanged = append(unchanged, map[string]string{
				"granted_on":   grant.GrantedOn,
				"object_name":  grant.ObjectName,
				"privilege":    grant.Privilege,
				"grantee_name": grant.GranteeName,
			})
		}
	}

	// Find removed (in A but not in B)
	for key, grant := range mapA {
		if _, exists := mapB[key]; !exists {
			removed = append(removed, map[string]string{
				"granted_on":   grant.GrantedOn,
				"object_name":  grant.ObjectName,
				"privilege":    grant.Privilege,
				"grantee_name": grant.GranteeName,
			})
		}
	}

	return map[string]interface{}{
		"added":     added,
		"removed":   removed,
		"unchanged": unchanged,
	}
}
