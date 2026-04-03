package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yourusername/gatekeep/internal/repository"
)

// HistoryHandler handles sync history requests
type HistoryHandler struct {
	auditRepo *repository.AuditRepository
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler(auditRepo *repository.AuditRepository) *HistoryHandler {
	return &HistoryHandler{
		auditRepo: auditRepo,
	}
}

// ListHistory handles GET /api/sync/history
func (h *HistoryHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize

	// Query sync runs
	filter := repository.SyncRunFilter{
		Limit:  pageSize,
		Offset: offset,
	}

	syncRuns, err := h.auditRepo.ListSyncRuns(r.Context(), filter)
	if err != nil {
		errorResponse(w, "failed to list sync runs", http.StatusInternalServerError, err)
		return
	}

	// Convert to API response
	type syncRunInfo struct {
		ID                   int64      `json:"id"`
		SyncID               uuid.UUID  `json:"sync_id"`
		StartedAt            time.Time  `json:"started_at"`
		CompletedAt          *time.Time `json:"completed_at,omitempty"`
		Status               string     `json:"status"`
		ConfigPath           string     `json:"config_path,omitempty"`
		TotalOperations      int        `json:"total_operations"`
		SuccessfulOperations int        `json:"successful_operations"`
		FailedOperations     int        `json:"failed_operations"`
		DurationMs           *int64     `json:"duration_ms,omitempty"`
	}

	runs := make([]syncRunInfo, len(syncRuns))
	for i, sr := range syncRuns {
		runs[i] = syncRunInfo{
			ID:                   sr.ID,
			SyncID:               sr.SyncID,
			StartedAt:            sr.StartedAt,
			CompletedAt:          sr.CompletedAt,
			Status:               string(sr.Status),
			ConfigPath:           sr.ConfigPath,
			TotalOperations:      sr.TotalOperations,
			SuccessfulOperations: sr.SuccessfulOperations,
			FailedOperations:     sr.FailedOperations,
			DurationMs:           sr.DurationMs,
		}
	}

	response := map[string]interface{}{
		"sync_runs":   runs,
		"total_count": len(runs),
		"page":        page,
		"page_size":   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}

// GetSyncRunDetail handles GET /api/sync/history/:id
func (h *HistoryHandler) GetSyncRunDetail(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errorResponse(w, "invalid sync run ID", http.StatusBadRequest, err)
		return
	}

	// Get sync run
	syncRun, err := h.auditRepo.GetSyncRun(r.Context(), id)
	if err != nil {
		errorResponse(w, "sync run not found", http.StatusNotFound, err)
		return
	}

	// Get operations
	operations, err := h.auditRepo.GetOperationsBySyncRun(r.Context(), id)
	if err != nil {
		errorResponse(w, "failed to get operations", http.StatusInternalServerError, err)
		return
	}

	// Convert to API response
	type operationDetail struct {
		ID              int64      `json:"id"`
		OperationType   string     `json:"operation_type"`
		TargetObject    string     `json:"target_object"`
		SQLStatement    string     `json:"sql_statement"`
		Status          string     `json:"status"`
		ErrorMessage    *string    `json:"error_message,omitempty"`
		ExecutionTimeMs *int       `json:"execution_time_ms,omitempty"`
		ExecutedAt      *time.Time `json:"executed_at,omitempty"`
	}

	ops := make([]operationDetail, len(operations))
	for i, op := range operations {
		ops[i] = operationDetail{
			ID:              op.ID,
			OperationType:   op.OperationType,
			TargetObject:    op.TargetObject,
			SQLStatement:    op.SQLStatement,
			Status:          string(op.Status),
			ErrorMessage:    op.ErrorMessage,
			ExecutionTimeMs: op.ExecutionTimeMs,
			ExecutedAt:      op.ExecutedAt,
		}
	}

	response := map[string]interface{}{
		"id":                    syncRun.ID,
		"sync_id":               syncRun.SyncID,
		"started_at":            syncRun.StartedAt,
		"completed_at":          syncRun.CompletedAt,
		"status":                string(syncRun.Status),
		"config_path":           syncRun.ConfigPath,
		"total_operations":      syncRun.TotalOperations,
		"successful_operations": syncRun.SuccessfulOperations,
		"failed_operations":     syncRun.FailedOperations,
		"duration_ms":           syncRun.DurationMs,
		"operations":            ops,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}
