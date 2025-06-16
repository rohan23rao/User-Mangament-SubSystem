package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	client "github.com/ory/kratos-client-go"
	"userms/internal/auth"
	"userms/internal/logger"
)

type VerificationHandler struct {
	authService *auth.Service
	kratosAdmin *client.APIClient
}

func NewVerificationHandler(authService *auth.Service, kratosAdmin *client.APIClient) *VerificationHandler {
	return &VerificationHandler{
		authService: authService,
		kratosAdmin: kratosAdmin,
	}
}

// Get verification status for a user
func (h *VerificationHandler) GetVerificationStatus(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing verification status request")

	// Path parameters extracted with r.PathValue
	userID := r.PathValue("id")

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get the identity from Kratos
	identity, resp, err := h.kratosAdmin.IdentityAPI.GetIdentity(context.Background(), userID).Execute()
	if err != nil {
		logger.Error("Failed to get identity: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	// Get verification status
	verificationStatus := map[string]interface{}{
		"user_id":   userID,
		"email":     "",
		"verified":  false,
		"addresses": []map[string]interface{}{},
	}

	if identity.VerifiableAddresses != nil {
		var addresses []map[string]interface{}
		for _, addr := range identity.VerifiableAddresses {
			address := map[string]interface{}{
				"id":       addr.Id,
				"value":    addr.Value,
				"verified": addr.Verified,
				"via":      addr.Via,
			}
			
			if verificationStatus["email"] == "" {
				verificationStatus["email"] = addr.Value
			}
			if addr.Verified {
				verificationStatus["verified"] = true
			}
			
			addresses = append(addresses, address)
		}
		verificationStatus["addresses"] = addresses
	}

	logger.Info("Verification status for user %s: verified=%v", userID, verificationStatus["verified"])

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(verificationStatus)
}

// Simple endpoint to trigger verification flow creation for testing
func (h *VerificationHandler) CreateVerificationFlow(w http.ResponseWriter, r *http.Request) {
	logger.Info("Creating verification flow for testing")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Visit http://localhost:4433/self-service/verification/browser to create a verification flow",
		"kratos_url": "http://localhost:4433/self-service/verification/browser",
	})
}