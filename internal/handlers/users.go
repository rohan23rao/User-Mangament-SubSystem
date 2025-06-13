package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
	"strings"

	"github.com/gorilla/mux"
	client "github.com/ory/kratos-client-go"
	"userms/internal/auth"
	"userms/internal/logger"
	"userms/internal/models"
)

type UserHandler struct {
	authService  *auth.Service
	kratosAdmin  *client.APIClient
	db           *sql.DB
}

func NewUserHandler(authService *auth.Service, kratosAdmin *client.APIClient, db *sql.DB) *UserHandler {
	return &UserHandler{
		authService: authService,
		kratosAdmin: kratosAdmin,
		db:          db,
	}
}

func (h *UserHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	logger.Auth("Processing whoami request")

	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized whoami request: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Auth("Whoami request authenticated for user: %s", session.Identity.Id)
	user := h.mapIdentityToUser(&session.Identity)

	// Get user from database for additional info
	dbUser, err := h.getUserFromDB(user.ID)
	if err != nil {
		logger.Warning("Error getting user from database: %v", err)
	} else if dbUser != nil {
		// Merge database info with Kratos identity
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CanCreateOrganizations = dbUser.CanCreateOrganizations // ADDED: Include permission field
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	orgs, err := h.getUserOrganizations(user.ID)
	if err != nil {
		logger.Warning("Error getting user organizations: %v", err)
	} else {
		user.Organizations = orgs
		logger.Info("Found %d organizations for user %s", len(orgs), user.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	logger.Success("Whoami response sent for user: %s", user.Email)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing list users request")

	identities, resp, err := h.kratosAdmin.IdentityApi.ListIdentities(context.Background()).Execute()
	if err != nil {
		logger.Error("Failed to fetch identities from Kratos: %v", err)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	logger.Info("Found %d identities from Kratos", len(identities))

	var users []models.User
	for _, identity := range identities {
		user := h.mapIdentityToUser(&identity)
		
		// Get additional info from database
		dbUser, err := h.getUserFromDB(user.ID)
		if err != nil {
			logger.Warning("Error getting user %s from database: %v", user.ID, err)
		} else if dbUser != nil {
			user.FirstName = dbUser.FirstName
			user.LastName = dbUser.LastName
			user.TimeZone = dbUser.TimeZone
			user.UIMode = dbUser.UIMode
			user.CanCreateOrganizations = dbUser.CanCreateOrganizations // ADDED: Include permission field
			user.CreatedAt = dbUser.CreatedAt
			user.UpdatedAt = dbUser.UpdatedAt
			user.LastLogin = dbUser.LastLogin
		}

		users = append(users, user)
	}

	logger.Success("Returning %d users", len(users))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	logger.Info("Getting user details for ID: %s", userID)

	identity, resp, err := h.kratosAdmin.IdentityApi.GetIdentity(context.Background(), userID).Execute()
	if err != nil {
		logger.Error("Failed to fetch identity from Kratos: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	user := h.mapIdentityToUser(identity)

	// Get additional info from database
	dbUser, err := h.getUserFromDB(user.ID)
	if err != nil {
		logger.Warning("Error getting user from database: %v", err)
	} else if dbUser != nil {
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CanCreateOrganizations = dbUser.CanCreateOrganizations // ADDED: Include permission field
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	logger.Success("User details retrieved for: %s", user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) mapIdentityToUser(identity *client.Identity) models.User {
	user := models.User{
		ID:     identity.Id,
		Traits: identity.Traits,
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

	// Add verification status from Kratos identity
	user.EmailVerified = h.isEmailVerified(identity)

	return user
}

func (h *UserHandler) isEmailVerified(identity *client.Identity) bool {
	// Check if the email is verified in Kratos
	if identity.VerifiableAddresses != nil {
		for _, addr := range identity.VerifiableAddresses {
			// Check if this address is verified
			if addr.Verified {
				return true
			}
		}
	}
	return false
}

// UPDATED: getUserFromDB method to include can_create_organizations
func (h *UserHandler) getUserFromDB(userID string) (*models.User, error) {
	var user models.User
	err := h.db.QueryRow(`
		SELECT id, email, first_name, last_name, time_zone, ui_mode, can_create_organizations, created_at, updated_at, last_login
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.TimeZone, &user.UIMode, 
		&user.CanCreateOrganizations, // ADDED: Include permission field
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (h *UserHandler) getUserOrganizations(userID string) ([]models.OrgMember, error) {
	rows, err := h.db.Query(`
		SELECT o.id, o.name, o.org_type, uol.role, uol.joined_at
		FROM organizations o
		JOIN user_organization_links uol ON o.id = uol.organization_id
		WHERE uol.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.OrgMember
	for rows.Next() {
		var org models.OrgMember
		err := rows.Scan(&org.OrgID, &org.OrgName, &org.OrgType, &org.Role, &org.JoinedAt)
		if err != nil {
			logger.Warning("Error scanning organization row: %v", err)
			continue
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

func (h *UserHandler) DebugAuth(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing debug auth request")

	// Try to get session without failing
	session, err := h.authService.GetSessionFromRequest(r)
	
	debugInfo := map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"headers": map[string]interface{}{
			"authorization": r.Header.Get("Authorization"),
			"cookie_count":  len(r.Cookies()),
		},
	}

	if err != nil {
		debugInfo["authenticated"] = false
		debugInfo["error"] = err.Error()
		
		// Check cookies
		cookies := make(map[string]string)
		for _, cookie := range r.Cookies() {
			if strings.Contains(cookie.Name, "kratos") {
				cookies[cookie.Name] = "present (hidden)"
			}
		}
		debugInfo["cookies"] = cookies
	} else {
		debugInfo["authenticated"] = true
		debugInfo["user_id"] = session.Identity.Id
		if traits, ok := session.Identity.Traits.(map[string]interface{}); ok {
			if email, exists := traits["email"].(string); exists {
				debugInfo["email"] = email
			}
		}
		debugInfo["session_active"] = session.Active
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debugInfo)
	
	if err != nil {
		logger.Warning("Debug auth - not authenticated: %v", err)
	} else {
		logger.Success("Debug auth - authenticated user: %s", session.Identity.Id)
	}
}