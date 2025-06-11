// internal/handlers/webhooks.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"userms/internal/auth"
	"userms/internal/models"
	"userms/internal/repository"
	"userms/internal/utils"

	"github.com/google/uuid"
	client "github.com/ory/kratos-client-go"
)

type WebhookHandler struct {
	userRepo            *repository.UserRepository
	orgRepo             *repository.OrganizationRepository
	verificationService *auth.VerificationService
}

func NewWebhookHandler(db *sql.DB) *WebhookHandler {
	return &WebhookHandler{
		userRepo:            repository.NewUserRepository(db),
		orgRepo:             repository.NewOrganizationRepository(db),
		verificationService: auth.NewVerificationService(db),
	}
}

func (wh *WebhookHandler) HandleAfterRegistration(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("=== REGISTRATION WEBHOOK START ===")

	var payload models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.LogError("Failed to decode registration webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	identity := payload.Identity
	email := getEmailFromIdentity(identity)
	firstName, lastName := getNameFromIdentity(identity)

	utils.LogInfo("Processing registration for user %s (%s)", identity.Id, email)

	// Check if this is a Google OAuth user
	isGoogleUser := wh.verificationService.IsGoogleOAuthUser(identity)
	utils.LogInfo("Google OAuth user: %t", isGoogleUser)

	// Check if this is the first user in the system
	userCount, err := wh.userRepo.GetUserCount()
	if err != nil {
		utils.LogError("Failed to get user count: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	isFirstUser := userCount == 0
	utils.LogInfo("First user in system: %t (current user count: %d)", isFirstUser, userCount)

	// Insert or update user
	user := &models.User{
		ID:        identity.Id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		TimeZone:  "UTC",
		UIMode:    "system",
	}

	err = wh.userRepo.UpsertUser(user)
	if err != nil {
		utils.LogError("Failed to upsert user: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("User %s created/updated successfully", identity.Id)

	// If this is the first user, automatically create a default organization and make them admin
	if isFirstUser {
		utils.LogInfo("Creating default organization for first user %s", identity.Id)
		
		orgID := uuid.New().String()
		orgName := "Default Organization"
		if firstName != "" && lastName != "" {
			orgName = fmt.Sprintf("%s %s's Organization", firstName, lastName)
		} else if firstName != "" {
			orgName = fmt.Sprintf("%s's Organization", firstName)
		}

		// Create the organization
		org := &models.Organization{
			ID:          orgID,
			OrgType:     "organization",
			Name:        orgName,
			Description: "Default organization for the first user",
			OwnerID:     &identity.Id,
			Data:        make(map[string]interface{}),
		}

		err = wh.orgRepo.CreateOrganization(org)
		if err != nil {
			utils.LogError("Failed to create default organization: %v", err)
		} else {
			// Add user as admin of the organization
			err = wh.orgRepo.AddMember(orgID, identity.Id, "admin")
			if err != nil {
				utils.LogError("Failed to add user as admin of default organization: %v", err)
			} else {
				utils.LogSuccess("First user %s granted admin access to default organization %s", identity.Id, orgID)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})

	utils.LogInfo("=== REGISTRATION WEBHOOK END ===")
}

func (wh *WebhookHandler) HandleAfterLogin(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Processing login webhook")

	var payload models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.LogError("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	utils.LogSuccess("User logged in: %s (%s)", payload.Identity.Id, getEmailFromIdentity(payload.Identity))

	// Update last login time
	err := wh.userRepo.UpdateLastLogin(payload.Identity.Id)
	if err != nil {
		utils.LogError("Error updating last login: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	utils.LogInfo("Login webhook processed successfully")
}

func getEmailFromIdentity(identity client.Identity) string {
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"].(string); exists {
			return email
		}
	}
	return ""
}

func getNameFromIdentity(identity client.Identity) (string, string) {
	var firstName, lastName string
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if nameObj, exists := traits["name"].(map[string]interface{}); exists {
			if first, ok := nameObj["first"].(string); ok {
				firstName = first
			}
			if last, ok := nameObj["last"].(string); ok {
				lastName = last
			}
		}
	}
	return firstName, lastName
}
