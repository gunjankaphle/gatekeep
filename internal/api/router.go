package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/gatekeep/internal/api/handlers"
	"github.com/yourusername/gatekeep/internal/api/middleware"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/repository"
	"github.com/yourusername/gatekeep/internal/snowflake"
	"github.com/yourusername/gatekeep/internal/sync"
)

// RouterConfig contains dependencies for the API router
type RouterConfig struct {
	AuditRepo       *repository.AuditRepository
	SnowflakeClient *snowflake.Client
	ConfigParser    *config.Parser
	ConfigPath      string
	Orchestrator    *sync.Orchestrator
}

// NewRouter creates a new HTTP router with all routes and middleware
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.CORS)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(cfg.AuditRepo)
	rolesHandler := handlers.NewRolesHandler(cfg.ConfigParser, cfg.ConfigPath)
	historyHandler := handlers.NewHistoryHandler(cfg.AuditRepo)
	syncHandler := handlers.NewSyncHandler(cfg.Orchestrator)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/health", healthHandler.Handle)

		// Roles
		r.Get("/roles", rolesHandler.ListRoles)

		// Sync operations
		r.Post("/sync", syncHandler.TriggerSync)
		r.Post("/sync/dry-run", syncHandler.DryRunSync)

		// History
		r.Get("/sync/history", historyHandler.ListHistory)
		r.Get("/sync/history/{id}", historyHandler.GetSyncRunDetail)
	})

	return r
}
