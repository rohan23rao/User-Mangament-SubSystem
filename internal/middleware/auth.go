// internal/middleware/auth.go
package middleware

import (
	"encoding/json"
	"net/http"
	"userms/internal/auth"
	"userms/internal/utils"
)

func RequireVerifiedUser(sessionManager *auth.SessionManager, verificationService *auth.VerificationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionManager.GetSessionFromRequest(r)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user is verified
			if !verificationService.IsEmailVerified(session.Identity) {
				utils.LogAuth("Unverified user %s attempting to access protected resource", session.Identity.Id)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "Email verification required",
					"code":    "EMAIL_NOT_VERIFIED",
					"message": "Please verify your email address before accessing this resource",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
