package oauth2

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	hydra "github.com/ory/hydra-client-go/v2"
	"userms/internal/logger"
	"userms/internal/models"
)

type Service struct {
	hydraAdmin *hydra.APIClient
	db         *sql.DB
}

func NewService(hydraAdmin *hydra.APIClient, db *sql.DB) *Service {
	return &Service{
		hydraAdmin: hydraAdmin,
		db:         db,
	}
}

// CreateM2MClient creates a machine-to-machine OAuth2 client for a user/organization
func (s *Service) CreateM2MClient(ctx context.Context, userID, orgID, name, description string) (*models.OAuth2Client, error) {
	logger.Info("Creating M2M OAuth2 client for user: %s, org: %s", userID, orgID)

	// Generate client credentials
	clientID := fmt.Sprintf("m2m_%s_%s", userID[:8], uuid.New().String()[:8])
	clientSecret := uuid.New().String() + uuid.New().String() // 72 chars

	// Create client in Hydra
	client := *hydra.NewOAuth2Client()
	client.SetClientId(clientID)
	client.SetClientSecret(clientSecret)
	client.SetClientName(name)
	
	// Set grant types for M2M
	client.SetGrantTypes([]string{"client_credentials"})
	client.SetResponseTypes([]string{"token"})
	client.SetScope("data_pipeline data_export telemetry_ingest")
	client.SetTokenEndpointAuthMethod("client_secret_basic")
	
	// M2M specific settings
	client.SetSkipConsent(true) // Skip consent for M2M flows
	
	// Metadata for tracking
	metadata := map[string]interface{}{
		"user_id":       userID,
		"org_id":        orgID,
		"client_type":   "machine_to_machine",
		"pipeline_type": "data_collection",
		"created_by":    "userms",
		"description":   description,
	}
	client.SetMetadata(metadata)

	// Create in Hydra using the correct API
	_, resp, err := s.hydraAdmin.OAuth2API.CreateOAuth2Client(ctx).OAuth2Client(client).Execute()
	if err != nil {
		logger.Error("Failed to create OAuth2 client in Hydra: %v", err)
		return nil, fmt.Errorf("failed to create OAuth2 client: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// Store in our database for management
	oauth2Client := &models.OAuth2Client{
		ID:           uuid.New().String(),
		ClientID:     clientID,
		ClientSecret: clientSecret, // Store encrypted in production
		UserID:       userID,
		OrgID:        orgID,
		Name:         name,
		Description:  description,
		Scopes:       "data_pipeline data_export telemetry_ingest",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert into database
	_, err = s.db.Exec(`
		INSERT INTO oauth2_clients (id, client_id, client_secret, user_id, org_id, name, description, scopes, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		oauth2Client.ID, oauth2Client.ClientID, oauth2Client.ClientSecret,
		oauth2Client.UserID, oauth2Client.OrgID, oauth2Client.Name,
		oauth2Client.Description, oauth2Client.Scopes, oauth2Client.IsActive,
		oauth2Client.CreatedAt, oauth2Client.UpdatedAt)

	if err != nil {
		logger.Error("Failed to store OAuth2 client in database: %v", err)
		// Try to cleanup Hydra client
		s.hydraAdmin.OAuth2API.DeleteOAuth2Client(ctx, clientID)
		return nil, fmt.Errorf("failed to store OAuth2 client: %v", err)
	}

	logger.Success("M2M OAuth2 client created: %s", clientID)
	return oauth2Client, nil
}

// RevokeM2MClient revokes and deletes a machine-to-machine client
func (s *Service) RevokeM2MClient(ctx context.Context, clientID string) error {
	logger.Info("Revoking M2M OAuth2 client: %s", clientID)

	// Delete from Hydra
	resp, err := s.hydraAdmin.OAuth2API.DeleteOAuth2Client(ctx, clientID).Execute()
	if err != nil {
		logger.Error("Failed to delete OAuth2 client from Hydra: %v", err)
		return fmt.Errorf("failed to delete OAuth2 client: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// Mark as inactive in our database
	_, err = s.db.Exec(`UPDATE oauth2_clients SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE client_id = $1`, clientID)
	if err != nil {
		logger.Warning("Failed to update OAuth2 client status in database: %v", err)
	}

	logger.Success("M2M OAuth2 client revoked: %s", clientID)
	return nil
}

// ListUserM2MClients lists all M2M clients for a user
func (s *Service) ListUserM2MClients(ctx context.Context, userID string) ([]models.OAuth2Client, error) {
	logger.Info("Listing M2M OAuth2 clients for user: %s", userID)

	rows, err := s.db.Query(`
		SELECT id, client_id, name, description, scopes, is_active, created_at, updated_at
		FROM oauth2_clients 
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC`, userID)

	if err != nil {
		logger.Error("Failed to query OAuth2 clients: %v", err)
		return nil, fmt.Errorf("failed to query OAuth2 clients: %v", err)
	}
	defer rows.Close()

	var clients []models.OAuth2Client
	for rows.Next() {
		var client models.OAuth2Client
		err := rows.Scan(&client.ID, &client.ClientID, &client.Name,
			&client.Description, &client.Scopes, &client.IsActive,
			&client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			logger.Error("Failed to scan OAuth2 client: %v", err)
			continue
		}
		client.UserID = userID
		// Don't return client secret in list operations
		client.ClientSecret = ""
		clients = append(clients, client)
	}

	logger.Success("Found %d M2M OAuth2 clients for user: %s", len(clients), userID)
	return clients, nil
}

// GenerateM2MToken generates an access token for machine-to-machine authentication
func (s *Service) GenerateM2MToken(ctx context.Context, clientID, clientSecret string) (*models.TokenResponse, error) {
	logger.Info("Generating M2M token for client: %s", clientID)

	// Create a custom HTTP client with Basic Auth
	client := &http.Client{
		Transport: &BasicAuthTransport{
			Username: clientID,
			Password: clientSecret,
		},
	}

	// Create a new configuration with the authenticated client
	publicConfig := hydra.NewConfiguration()
	// Use the public API URL (typically port 4444)
	publicConfig.HTTPClient = client
	publicConfig.Servers = []hydra.ServerConfiguration{
		{URL: "http://hydra:4444"}, // Use your Hydra public URL
	}

	// Create public API client
	publicClient := hydra.NewAPIClient(publicConfig)

	// Use the token endpoint - client auth is via HTTP Basic Auth header
	// Scope is determined by the client configuration, not the token request
	tokenResponse, resp, err := publicClient.OAuth2API.Oauth2TokenExchange(ctx).
		GrantType("client_credentials").
		Execute()
	if err != nil {
		logger.Error("Failed to generate M2M token: %v", err)
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// Log token generation (without exposing token)
	expiresAt := time.Now().Add(time.Duration(tokenResponse.GetExpiresIn()) * time.Second)
	_, err = s.db.Exec(`
		INSERT INTO oauth2_token_logs (client_id, granted_scopes, expires_at, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`,
		clientID, tokenResponse.GetScope(), expiresAt)

	if err != nil {
		logger.Warning("Failed to log token generation: %v", err)
	}

	response := &models.TokenResponse{
		AccessToken: tokenResponse.GetAccessToken(),
		TokenType:   tokenResponse.GetTokenType(),
		ExpiresIn:   int(tokenResponse.GetExpiresIn()),
		Scope:       tokenResponse.GetScope(),
	}

	// Add refresh token if present
	if tokenResponse.RefreshToken != nil {
		response.RefreshToken = *tokenResponse.RefreshToken
	}

	logger.Success("M2M token generated for client: %s", clientID)
	return response, nil
}

// ValidateM2MToken validates a machine-to-machine token
func (s *Service) ValidateM2MToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	// Use Hydra's introspection endpoint with the correct API
	tokenInfo, resp, err := s.hydraAdmin.OAuth2API.IntrospectOAuth2Token(ctx).Token(token).Execute()
	if err != nil {
		logger.Error("Failed to introspect token: %v", err)
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	if !tokenInfo.GetActive() {
		return nil, fmt.Errorf("token is inactive or expired")
	}

	info := &models.TokenInfo{
		Active:    tokenInfo.GetActive(),
		ClientID:  tokenInfo.GetClientId(),
		Scope:     tokenInfo.GetScope(),
		ExpiresAt: time.Unix(tokenInfo.GetExp(), 0),
	}

	// Add subject if present
	if tokenInfo.Sub != nil {
		info.Subject = *tokenInfo.Sub
	}

	// Add issued at if present
	if tokenInfo.Iat != nil {
		info.IssuedAt = time.Unix(*tokenInfo.Iat, 0)
	}

	return info, nil
}

// BasicAuthTransport implements HTTP Basic Authentication
type BasicAuthTransport struct {
	Username string
	Password string
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add Basic Auth header
	auth := base64.StdEncoding.EncodeToString([]byte(t.Username + ":" + t.Password))
	req.Header.Set("Authorization", "Basic "+auth)
	
	// Use default transport
	return http.DefaultTransport.RoundTrip(req)
}