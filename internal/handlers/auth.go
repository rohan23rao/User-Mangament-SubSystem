// internal/handlers/auth.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"userms/internal/auth"
	"userms/internal/utils"

	client "github.com/ory/kratos-client-go"
)

type AuthHandler struct {
	kratosPublic   *client.APIClient
	kratosAdmin    *client.APIClient
	sessionManager *auth.SessionManager
}

func NewAuthHandler(kratosPublic, kratosAdmin *client.APIClient, sessionManager *auth.SessionManager) *AuthHandler {
	return &AuthHandler{
		kratosPublic:   kratosPublic,
		kratosAdmin:    kratosAdmin,
		sessionManager: sessionManager,
	}
}

func (ah *AuthHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	utils.LogAuth("Getting session information")

	session, err := ah.sessionManager.GetSessionFromRequest(r)
	if err != nil {
		utils.LogAuth("No valid session found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
	utils.LogSuccess("Session information sent for user: %s", session.Identity.Id)
}

func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	utils.LogAuth("Processing logout request")

	// Get session token for logout
	sessionToken := ""
	if cookie, err := r.Cookie("ory_kratos_session"); err == nil {
		sessionToken = cookie.Value
	} else if authHeader := r.Header.Get("Authorization"); authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if sessionToken == "" {
		utils.LogWarning("No session token found for logout")
		http.Error(w, "No session to logout", http.StatusBadRequest)
		return
	}

	utils.LogAuth("Attempting to logout session: %s...", sessionToken[:utils.Min(len(sessionToken), 20)])

	// Get session details first
	session, resp, err := ah.kratosPublic.FrontendApi.ToSession(context.Background()).
		XSessionToken(sessionToken).
		Execute()

	if err != nil || resp.StatusCode != 200 {
		utils.LogWarning("Could not get session details for logout: %v (status: %d)", err, resp.StatusCode)
		// Session might already be invalid, continue with clearing cookie
	} else {
		utils.LogAuth("Found session ID: %s", session.Id)

		// Use the session ID (not token) to disable the session
		_, err = ah.kratosAdmin.IdentityApi.DisableSession(context.Background(), session.Id).Execute()
		if err != nil {
			utils.LogWarning("Error revoking session with ID %s: %v", session.Id, err)
		} else {
			utils.LogSuccess("Session %s revoked successfully", session.Id)
		}
	}

	// Clear cookie regardless of session revocation status
	http.SetCookie(w, &http.Cookie{
		Name:     "ory_kratos_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})

	utils.LogSuccess("Logout completed successfully")
}

func (ah *AuthHandler) DebugAuth(w http.ResponseWriter, r *http.Request) {
	utils.LogAuth("=== DEBUG AUTH ENDPOINT START ===")

	response := map[string]interface{}{
		"headers": make(map[string][]string),
		"cookies": make(map[string]string),
	}

	// Log all headers (sanitized)
	for name, values := range r.Header {
		if strings.ToLower(name) == "authorization" && len(values) > 0 {
			response["headers"].(map[string][]string)[name] = []string{values[0][:utils.Min(len(values[0]), 30)] + "..."}
		} else {
			response["headers"].(map[string][]string)[name] = values
		}
	}

	// Add cookies (sanitized)
	for _, cookie := range r.Cookies() {
		if cookie.Name == "ory_kratos_session" {
			response["cookies"].(map[string]string)[cookie.Name] = cookie.Value[:utils.Min(len(cookie.Value), 30)] + "..."
		} else {
			response["cookies"].(map[string]string)[cookie.Name] = cookie.Value
		}
	}

	session, err := ah.sessionManager.GetSessionFromRequest(r)
	if err != nil {
		response["auth_status"] = "failed"
		response["auth_error"] = err.Error()
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		response["auth_status"] = "success"
		response["user_id"] = session.Identity.Id
		response["user_email"] = getEmailFromIdentity(session.Identity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	utils.LogAuth("=== DEBUG AUTH ENDPOINT END ===")
}
