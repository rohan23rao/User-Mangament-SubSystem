package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"userms/internal/auth"
	"userms/internal/logger"
	"userms/internal/models"
	"userms/internal/oauth2"
)

type OAuth2Handler struct {
	authService   *auth.Service
	oauth2Service *oauth2.Service
}

func NewOAuth2Handler(authService *auth.Service, oauth2Service *oauth2.Service) *OAuth2Handler {
	return &OAuth2Handler{
		authService:   authService,
		oauth2Service: oauth2Service,
	}
}

// CreateM2MClient creates a new machine-to-machine OAuth2 client
func (h *OAuth2Handler) CreateM2MClient(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M client creation request")

	// Authenticate user
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized M2M client creation: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req models.CreateM2MClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for M2M client creation: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Client name is required", http.StatusBadRequest)
		return
	}

	if req.OrgID == "" {
		http.Error(w, "Organization ID is required", http.StatusBadRequest)
		return
	}

	// Create M2M client - FIXED: Use session variable that was properly declared above
	client, err := h.oauth2Service.CreateM2MClient(r.Context(), session.Identity.Id, req.OrgID, req.Name, req.Description)
	if err != nil {
		logger.Error("Failed to create M2M client: %v", err)
		http.Error(w, "Failed to create M2M client", http.StatusInternalServerError)
		return
	}

	// FIXED: Use session variable that was properly declared above
	logger.Success("M2M client created for user %s: %s", session.Identity.Id, client.ClientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client_id":     client.ClientID,
		"client_secret": client.ClientSecret,
		"name":          client.Name,
		"description":   client.Description,
		"scopes":        client.Scopes,
		"created_at":    client.CreatedAt,
		"message":       "Store the client_secret securely - it will not be shown again",
	})
}

// ListM2MClients lists all M2M clients for the authenticated user
func (h *OAuth2Handler) ListM2MClients(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M client list request")

	// Authenticate user
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized M2M client list: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's M2M clients
	clients, err := h.oauth2Service.ListUserM2MClients(r.Context(), session.Identity.Id)
	if err != nil {
		logger.Error("Failed to list M2M clients: %v", err)
		http.Error(w, "Failed to list M2M clients", http.StatusInternalServerError)
		return
	}

	logger.Success("Listed %d M2M clients for user: %s", len(clients), session.Identity.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// RevokeM2MClient revokes a machine-to-machine OAuth2 client
func (h *OAuth2Handler) RevokeM2MClient(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M client revocation request")

	// Authenticate user
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized M2M client revocation: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get client ID from URL
	// Path parameters extracted with r.PathValue
	clientID := r.PathValue("clientId")

	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Verify that the client belongs to the authenticated user
	// This requires checking the database first

	// Revoke the client
	err = h.oauth2Service.RevokeM2MClient(r.Context(), clientID)
	if err != nil {
		logger.Error("Failed to revoke M2M client: %v", err)
		http.Error(w, "Failed to revoke M2M client", http.StatusInternalServerError)
		return
	}

	logger.Success("M2M client revoked by user %s: %s", session.Identity.Id, clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "M2M client revoked successfully",
		"client_id": clientID,
	})
}

// GenerateM2MToken generates an access token for machine-to-machine authentication
func (h *OAuth2Handler) GenerateM2MToken(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M token generation request")

	// Parse request
	var req models.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for token generation: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials
	if req.ClientID == "" || req.ClientSecret == "" {
		http.Error(w, "Client ID and client secret are required", http.StatusBadRequest)
		return
	}

	// Generate token
	tokenResponse, err := h.oauth2Service.GenerateM2MToken(r.Context(), req.ClientID, req.ClientSecret)
	if err != nil {
		logger.Error("Failed to generate M2M token: %v", err)
		http.Error(w, "Failed to generate token", http.StatusUnauthorized)
		return
	}

	logger.Success("M2M token generated for client: %s", req.ClientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokenResponse)
}

// ValidateM2MToken validates a machine-to-machine token (for internal use by data pipeline)
func (h *OAuth2Handler) ValidateM2MToken(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M token validation request")

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusBadRequest)
		return
	}

	// Check for Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Invalid authorization header format", http.StatusBadRequest)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	tokenInfo, err := h.oauth2Service.ValidateM2MToken(r.Context(), token)
	if err != nil {
		logger.Warning("Invalid M2M token validation attempt: %v", err)
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	logger.Success("M2M token validated for client: %s", tokenInfo.ClientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":      tokenInfo.Active,
		"client_id":  tokenInfo.ClientID,
		"scope":      tokenInfo.Scope,
		"expires_at": tokenInfo.ExpiresAt,
	})
}

// GetM2MClientInfo gets information about a specific M2M client
func (h *OAuth2Handler) GetM2MClientInfo(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M client info request")

	// Authenticate user
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized M2M client info: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get client ID from URL
	// Path parameters extracted with r.PathValue
	clientID := r.PathValue("clientId")

	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	// Get client info from database
	clients, err := h.oauth2Service.ListUserM2MClients(r.Context(), session.Identity.Id)
	if err != nil {
		logger.Error("Failed to get M2M client info: %v", err)
		http.Error(w, "Failed to get client info", http.StatusInternalServerError)
		return
	}

	// Find the specific client
	var targetClient *models.OAuth2Client
	for _, client := range clients {
		if client.ClientID == clientID {
			targetClient = &client
			break
		}
	}

	if targetClient == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	logger.Success("M2M client info retrieved for user %s: %s", session.Identity.Id, clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targetClient)
}

// RegenerateM2MClientSecret regenerates the client secret for a M2M client
func (h *OAuth2Handler) RegenerateM2MClientSecret(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing M2M client secret regeneration request")

	// Authenticate user
	_, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized M2M client secret regeneration: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get client ID from URL
	// Path parameters extracted with r.PathValue
	clientID := r.PathValue("clientId")

	if clientID == "" {
		http.Error(w, "Client ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement secret regeneration in oauth2 service
	// This would involve:
	// 1. Verify client belongs to user
	// 2. Generate new secret
	// 3. Update in Hydra
	// 4. Update in database
	// 5. Return new secret

	logger.Warning("M2M client secret regeneration not yet implemented")
	http.Error(w, "Feature not yet implemented", http.StatusNotImplemented)
}