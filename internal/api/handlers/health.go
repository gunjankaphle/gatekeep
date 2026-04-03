package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourusername/gatekeep/internal/repository"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	auditRepo *repository.AuditRepository
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(auditRepo *repository.AuditRepository) *HealthHandler {
	return &HealthHandler{
		auditRepo: auditRepo,
	}
}

// Handle processes health check requests
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// Check database connection
	if h.auditRepo != nil {
		if err := h.auditRepo.Ping(ctx); err != nil {
			services["database"] = "unhealthy"
		} else {
			services["database"] = "ok"
		}
	} else {
		services["database"] = "not_configured"
	}

	// Determine overall status
	status := "healthy"
	for _, svcStatus := range services {
		if svcStatus == "unhealthy" {
			status = "unhealthy"
			break
		}
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"services":  services,
	}

	w.Header().Set("Content-Type", "application/json")
	if status == "unhealthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}
