// internal/handlers/users.go
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"userms/internal/auth"
	"userms/internal/models"
	"userms/internal/repository"
	"userms/internal/utils"

	"github.com/gorilla/mux"
	client "github.com/ory/kratos-client-go"
)

type UserHandler struct {
	userRepo            *repository.UserRepository
	kratosAdmin         *client.APIClient
	sessionManager      *auth.SessionManager
	verificationService *auth.VerificationService
}

func NewUserHandler(db *sql.DB, kratosAdmin *client.APIClient, sessionManager *auth.SessionManager, verificationService *auth.VerificationService) *UserHandler {
	return &UserHandler{
		userRepo:            repository.NewUserRepository(db),
		kratosAdmin:         kratosAdmin,
		sessionManager:      sessionManager,
		verificationService: verificationService,
	}
}

func (uh *UserHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	utils.LogAuth("Processing whoami request")

	session, err := uh.sessionManager.GetSessionFromRequest(r)
	if err != nil {
		utils.LogAuth("Unauthorized whoami request: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	utils.LogAuth("Whoami request authenticated for user: %s", session.Identity.Id)
	user := uh.mapIdentityToUser(session.Identity)

	// Get user from database for additional info
	dbUser, err := uh.userRepo.GetUserFromDB(user.ID)
	if err != nil {
		utils.LogWarning("Error getting user from database: %v", err)
	} else if dbUser != nil {
		// Merge database info with Kratos identity
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	orgs, err := uh.userRepo.GetUserOrganizations(user.ID)
	if err != nil {
		utils.LogWarning("Error getting user organizations: %v", err)
	} else {
		user.Organizations = orgs
		utils.LogInfo("Found %d organizations for user %s", len(orgs), user.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	utils.LogSuccess("Whoami response sent for user: %s", user.Email)
}

func (uh *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Listing all users")

	// Get all identities from Kratos
	identities, resp, err := uh.kratosAdmin.IdentityApi.ListIdentities(context.Background()).Execute()
	if err != nil || resp.StatusCode != 200 {
		utils.LogError("Failed to list identities: %v", err)
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	var users []models.User
	for _, identity := range identities {
		user := uh.mapIdentityToUser(identity)
		
		// Get additional info from database
		dbUser, err := uh.userRepo.GetUserFromDB(user.ID)
		if err == nil && dbUser != nil {
			user.FirstName = dbUser.FirstName
			user.LastName = dbUser.LastName
			user.TimeZone = dbUser.TimeZone
			user.UIMode = dbUser.UIMode
			user.CreatedAt = dbUser.CreatedAt
			user.UpdatedAt = dbUser.UpdatedAt
			user.LastLogin = dbUser.LastLogin
		}

		orgs, err := uh.userRepo.GetUserOrganizations(user.ID)
		if err == nil {
			user.Organizations = orgs
		}

		utils.LogInfo("Built user object for %s: verified=%t, org_count=%d", user.Email, user.Verified, len(user.Organizations))
		users = append(users, user)
	}

	utils.LogInfo("Found %d users in Kratos", len(users))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

	utils.LogSuccess("Users list sent successfully")
}

func (uh *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	utils.LogInfo("Getting user details for: %s", userID)

	identity, resp, err := uh.kratosAdmin.IdentityApi.GetIdentity(context.Background(), userID).Execute()
	if err != nil || resp.StatusCode != 200 {
		utils.LogWarning("User not found: %s", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user := uh.mapIdentityToUser(*identity)

	// Get additional info from database
	dbUser, err := uh.userRepo.GetUserFromDB(user.ID)
	if err == nil && dbUser != nil {
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	utils.LogSuccess("User details retrieved for: %s", user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (uh *UserHandler) mapIdentityToUser(identity client.Identity) models.User {
	verified := uh.verificationService.IsEmailVerified(identity)
	utils.LogInfo("Mapping user %s, verified status: %t", identity.Id, verified)

	user := models.User{
		ID:            identity.Id,
		Traits:        identity.Traits,
		Verified:      verified,
		Organizations: []models.OrgMember{}, // Initialize as empty slice
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

	return user
}
