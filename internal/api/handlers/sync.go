package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/yourusername/gatekeep/internal/sync"
)

// SyncHandler handles sync requests
type SyncHandler struct {
	orchestrator *sync.Orchestrator
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(orchestrator *sync.Orchestrator) *SyncHandler {
	return &SyncHandler{
		orchestrator: orchestrator,
	}
}

// TriggerSync handles POST /api/sync
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigPath string `json:"config_path"`
		DryRun     bool   `json:"dry_run"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest, err)
		return
	}

	if req.ConfigPath == "" {
		errorResponse(w, "config_path is required", http.StatusBadRequest, nil)
		return
	}

	// Check if config file exists
	if _, err := os.Stat(req.ConfigPath); os.IsNotExist(err) {
		errorResponse(w, "config file not found", http.StatusNotFound, err)
		return
	}

	// Execute sync (this will be implemented when orchestrator is complete)
	// For now, return a placeholder response
	syncID := uuid.New()

	response := map[string]interface{}{
		"sync_id":             syncID,
		"status":              "pending",
		"operations_executed": 0,
		"operations_failed":   0,
		"duration_ms":         0,
		"details_url":         fmt.Sprintf("/api/sync/history/%s", syncID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}

// DryRunSync handles POST /api/sync/dry-run
func (h *SyncHandler) DryRunSync(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfigPath string `json:"config_path"`
		DryRun     bool   `json:"dry_run"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, "invalid request body", http.StatusBadRequest, err)
		return
	}

	if req.ConfigPath == "" {
		errorResponse(w, "config_path is required", http.StatusBadRequest, nil)
		return
	}

	// Check if config file exists
	if _, err := os.Stat(req.ConfigPath); os.IsNotExist(err) {
		errorResponse(w, "config file not found", http.StatusNotFound, err)
		return
	}

	// Execute dry-run (this will be implemented when orchestrator is complete)
	// For now, return a placeholder response
	response := map[string]interface{}{
		"sync_id":             uuid.New(),
		"status":              "success",
		"operations_executed": 0,
		"operations_failed":   0,
		"duration_ms":         0,
		"operations": []map[string]string{
			{
				"type":   "PLACEHOLDER",
				"target": "N/A",
				"sql":    "-- Dry-run will show SQL statements here",
				"status": "pending",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}
