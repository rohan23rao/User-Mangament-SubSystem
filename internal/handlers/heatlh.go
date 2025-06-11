// internal/handlers/health.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"userms/internal/utils"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (hh *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Health check requested")

	// Check database connectivity
	if err := hh.db.Ping(); err != nil {
		utils.LogError("Database health check failed: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"database": "connected",
	})

	utils.LogSuccess("Health check: OK")
}
