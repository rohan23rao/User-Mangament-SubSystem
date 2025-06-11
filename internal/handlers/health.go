package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"userms/internal/logger"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	logger.Info("Health check requested")

	// Check database connectivity
	if err := h.db.Ping(); err != nil {
		logger.Error("Database health check failed: %v", err)
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

	logger.Success("Health check: OK")
}