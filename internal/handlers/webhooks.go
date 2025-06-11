package handlers

import (
	"encoding/json"
	"net/http"

	client "github.com/ory/kratos-client-go"
	"userms/internal/logger"
	"userms/internal/models"
)

type WebhookHandler struct {
	userHandler *UserHandler
}

func NewWebhookHandler(userHandler *UserHandler) *WebhookHandler {
	return &WebhookHandler{
		userHandler: userHandler,
	}
}

func (h *WebhookHandler) HandleAfterRegistration(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing registration webhook")

	var payload models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Error("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logger.Success("New user registered: %s (%s)", payload.Identity.Id, h.getEmailFromIdentity(&payload.Identity))

	h.saveUserProfile(&payload.Identity)

	w.WriteHeader(http.StatusOK)
	logger.Info("Registration webhook processed successfully")
}

func (h *WebhookHandler) HandleAfterLogin(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing login webhook")

	var payload models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Error("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logger.Success("User logged in: %s (%s)", payload.Identity.Id, h.getEmailFromIdentity(&payload.Identity))

	h.saveUserProfile(&payload.Identity)

	w.WriteHeader(http.StatusOK)
	logger.Info("Login webhook processed successfully")
}

func (h *WebhookHandler) HandleAfterVerification(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing verification webhook")

	var payload models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logger.Error("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logger.Success("User email verified: %s (%s)", payload.Identity.Id, h.getEmailFromIdentity(&payload.Identity))

	// Update user profile and potentially trigger additional verification logic
	h.saveUserProfile(&payload.Identity)

	// You can add custom verification logic here
	// For example: send welcome email, update user permissions, etc.

	w.WriteHeader(http.StatusOK)
	logger.Info("Verification webhook processed successfully")
}

func (h *WebhookHandler) getEmailFromIdentity(identity *client.Identity) string {
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"].(string); exists {
			return email
		}
	}
	return "unknown"
}

func (h *WebhookHandler) saveUserProfile(identity *client.Identity) {
	user := models.User{
		ID: identity.Id,
	}

	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"].(string); exists {
			user.Email = email
		}
		if nameObj, exists := traits["name"].(map[string]interface{}); exists {
			if first, ok := nameObj["first"].(string); ok {
				user.FirstName = first
			}
			if last, ok := nameObj["last"].(string); ok {
				user.LastName = last
			}
		}
	}

	_, err := h.userHandler.db.Exec(`
		INSERT INTO users (id, email, first_name, last_name, last_login) 
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (id) 
		DO UPDATE SET 
			email = $2, 
			first_name = $3, 
			last_name = $4, 
			last_login = CURRENT_TIMESTAMP, 
			updated_at = CURRENT_TIMESTAMP`,
		user.ID, user.Email, user.FirstName, user.LastName,
	)
	if err != nil {
		logger.Warning("Failed to save user profile: %v", err)
	} else {
		logger.DB("User profile saved/updated for: %s", user.Email)
	}
}