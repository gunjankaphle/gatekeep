package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yourusername/gatekeep/internal/api/handlers"
	"github.com/yourusername/gatekeep/internal/api/middleware"
	"github.com/yourusername/gatekeep/internal/config"
	"github.com/yourusername/gatekeep/internal/repository"
)

// RouterConfig contains dependencies for the API router (read-only mode)
type RouterConfig struct {
	AuditRepo    *repository.AuditRepository // Optional - for history endpoints
	ConfigParser *config.Parser              // Required - for roles endpoint
	ConfigPath   string                      // Required - path to YAML config
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
