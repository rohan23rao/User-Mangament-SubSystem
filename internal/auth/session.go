// internal/auth/session.go
package auth

import (
	"context"
	"net/http"
	"strings"
	"userms/internal/utils"

	client "github.com/ory/kratos-client-go"
)

type SessionManager struct {
	kratosPublic *client.APIClient
}

func NewSessionManager(kratosPublic *client.APIClient) *SessionManager {
	return &SessionManager{
		kratosPublic: kratosPublic,
	}
}

func (sm *SessionManager) GetSessionFromRequest(r *http.Request) (*client.Session, error) {
	utils.LogAuth("=== SESSION VALIDATION START ===")

	// Log all cookies for debugging
	utils.LogAuth("All cookies in request:")
	for _, cookie := range r.Cookies() {
		if cookie.Name == "ory_kratos_session" {
			utils.LogAuth("  %s: %s (length: %d)", cookie.Name, cookie.Value[:utils.Min(len(cookie.Value), 30)]+"...", len(cookie.Value))
		} else {
			utils.LogAuth("  %s: %s", cookie.Name, cookie.Value)
		}
	}

	// Log headers for debugging
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		utils.LogAuth("Authorization header found: %s", authHeader[:utils.Min(len(authHeader), 50)]+"...")
	}

	var sessionToken string

	// Method 1: Try Authorization header (Bearer token)
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
		utils.LogAuth("Extracted Bearer token: %s...", sessionToken[:utils.Min(len(sessionToken), 20)])

		session, resp, err := sm.kratosPublic.FrontendApi.ToSession(context.Background()).
			XSessionToken(sessionToken).
			Execute()

		if err != nil || resp.StatusCode != 200 {
			utils.LogWarning("Bearer token validation failed: %v (status: %d)", err, resp.StatusCode)
		} else {
			utils.LogSuccess("Bearer token validated successfully for user: %s", session.Identity.Id)
			utils.LogAuth("=== SESSION VALIDATION END (SUCCESS) ===")
			return session, nil
		}
	}

	// Method 2: Try session cookie
	sessionCookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		utils.LogAuth("No session cookie found: %v", err)
		utils.LogAuth("=== SESSION VALIDATION END (FAILED) ===")
		return nil, err
	}

	sessionToken = sessionCookie.Value
	utils.LogAuth("Found session cookie: %s...", sessionToken[:utils.Min(len(sessionToken), 20)])

	session, resp, err := sm.kratosPublic.FrontendApi.ToSession(context.Background()).
		XSessionToken(sessionToken).
		Execute()

	if err != nil || resp.StatusCode != 200 {
		utils.LogWarning("Cookie session validation failed: %v (status: %d)", err, resp.StatusCode)
		utils.LogAuth("=== SESSION VALIDATION END (FAILED) ===")
		return nil, err
	}

	utils.LogSuccess("Cookie session validated successfully for user: %s", session.Identity.Id)
	utils.LogAuth("=== SESSION VALIDATION END (SUCCESS) ===")
	return session, nil
}
