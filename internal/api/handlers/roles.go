package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/gatekeep/internal/config"
)

// RolesHandler handles role-related requests
type RolesHandler struct {
	configParser *config.Parser
	configPath   string
}

// NewRolesHandler creates a new roles handler
func NewRolesHandler(configParser *config.Parser, configPath string) *RolesHandler {
	return &RolesHandler{
		configParser: configParser,
		configPath:   configPath,
	}
}

// ListRoles handles GET /api/roles
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	// Parse the config file
	cfg, err := h.configParser.ParseFile(h.configPath)
	if err != nil {
		errorResponse(w, "failed to parse config", http.StatusInternalServerError, err)
		return
	}

	// Convert to API response format
	type roleResponse struct {
		Name        string   `json:"name"`
		ParentRoles []string `json:"parent_roles,omitempty"`
		Comment     string   `json:"comment,omitempty"`
	}

	roles := make([]roleResponse, len(cfg.Roles))
	for i, role := range cfg.Roles {
		roles[i] = roleResponse{
			Name:        role.Name,
			ParentRoles: role.ParentRoles,
			Comment:     role.Comment,
		}
	}

	response := map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}

// errorResponse sends an error response
func errorResponse(w http.ResponseWriter, message string, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	fullMessage := message
	if err != nil {
		fullMessage = message + ": " + err.Error()
	}

	response := map[string]interface{}{
		"error":   http.StatusText(statusCode),
		"message": fullMessage,
		"code":    statusCode,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Response already written, can only log
		_ = err
	}
}
