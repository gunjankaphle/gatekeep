package handlers

import (
	"encoding/json"
	"net/http"
)

// SyncHandler handles sync requests
// In read-only mode, write operations return "not implemented"
type SyncHandler struct{}

// NewSyncHandler creates a new sync handler
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// TriggerSync handles POST /api/sync
// READ-ONLY MODE: This endpoint is not available
func (h *SyncHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"error":   "not_implemented",
		"message": "Sync operations are not available through the API in read-only mode",
		"details": "Use the CLI for sync operations: gatekeep sync --config <file>",
		"cli_usage": map[string]string{
			"sync":    "gatekeep sync --config prod.yaml",
			"dry-run": "gatekeep sync --config prod.yaml --dry-run",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}

// DryRunSync handles POST /api/sync/dry-run
// READ-ONLY MODE: This endpoint is not available
func (h *SyncHandler) DryRunSync(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"error":   "not_implemented",
		"message": "Dry-run operations are not available through the API in read-only mode",
		"details": "Use the CLI for dry-run operations: gatekeep sync --config <file> --dry-run",
		"cli_usage": map[string]string{
			"dry-run": "gatekeep sync --config prod.yaml --dry-run",
			"format":  "gatekeep sync --config prod.yaml --dry-run --format json",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}
