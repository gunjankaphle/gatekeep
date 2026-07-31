package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/gatekeep/internal/api/handlers"
	"github.com/yourusername/gatekeep/internal/api/middleware"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/repository"
	"github.com/yourusername/gatekeep/internal/snowflake"
)

// RouterConfig contains dependencies for the API router (read-only mode)
type RouterConfig struct {
	AuditRepo       *repository.AuditRepository // Optional - for history endpoints
	CacheRepo       *repository.CacheRepository // Required - for cache endpoints
	ConfigParser    *config.Parser              // Required - for roles endpoint
	ConfigPath      string                      // Required - path to YAML config
	SnowflakeClient snowflake.Client            // Required - Snowflake client for cache refresh
	MaxWorkers      int                         // Max concurrent workers for Snowflake queries
}

// NewRouter creates a new HTTP router with all routes and middleware
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS)

	// Initialize handlers (read-only mode)
	healthHandler := handlers.NewHealthHandler(cfg.AuditRepo)
	rolesHandler := handlers.NewRolesHandler(cfg.ConfigParser, cfg.ConfigPath)
	historyHandler := handlers.NewHistoryHandler(cfg.AuditRepo)
	syncHandler := handlers.NewSyncHandler() // No orchestrator - read-only mode

	// Initialize cache handler if cache repo is provided
	var rolesCacheHandler *handlers.RolesCacheHandler
	if cfg.CacheRepo != nil && cfg.SnowflakeClient != nil {
		rolesCacheHandler = handlers.NewRolesCacheHandler(cfg.CacheRepo, cfg.SnowflakeClient, cfg.MaxWorkers)
	}

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.Handle)

		// Roles (legacy config-based endpoint)
		r.Get("/roles", rolesHandler.ListRoles)

		// Roles cache endpoints (new)
		if rolesCacheHandler != nil {
			r.Get("/roles/hierarchy", rolesCacheHandler.GetRoleHierarchy)
			r.Post("/roles/refresh", rolesCacheHandler.TriggerRefresh)
			r.Get("/roles/refresh/status", rolesCacheHandler.GetRefreshStatus)
			r.Get("/roles/{roleName}/grants", rolesCacheHandler.GetRoleGrants)
			r.Get("/roles/compare", rolesCacheHandler.CompareRoles)
		}

		// Sync operations
		r.Post("/sync", syncHandler.TriggerSync)
		r.Post("/sync/dry-run", syncHandler.DryRunSync)

		// History
		r.Get("/sync/history", historyHandler.ListHistory)
		r.Get("/sync/history/{id}", historyHandler.GetSyncRunDetail)
	})

	return r
}
