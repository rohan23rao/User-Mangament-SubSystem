// internal/auth/verification.go
package auth

import (
	"database/sql"
	"strings"
	"userms/internal/utils"

	client "github.com/ory/kratos-client-go"
)

type VerificationService struct {
	db *sql.DB
}

func NewVerificationService(db *sql.DB) *VerificationService {
	return &VerificationService{db: db}
}

func (vs *VerificationService) IsEmailVerified(identity client.Identity) bool {
	utils.LogInfo("Checking verification for user %s", identity.Id)

	// Check if this is the first user - first user is automatically verified
	var userCount int
	err := vs.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err == nil && userCount <= 1 {
		utils.LogInfo("User %s is the first user - automatically verified", identity.Id)
		return true
	}

	// Check if user authenticated via Google OAuth
	// Google OAuth users are automatically verified
	if vs.IsGoogleOAuthUser(identity) {
		utils.LogInfo("User %s is verified via Google OAuth", identity.Id)
		return true
	}

	if identity.VerifiableAddresses == nil {
		utils.LogInfo("User %s has no verifiable addresses", identity.Id)
		return false
	}

	utils.LogInfo("User %s has %d verifiable addresses", identity.Id, len(identity.VerifiableAddresses))
	for i, addr := range identity.VerifiableAddresses {
		utils.LogInfo("  Address %d: Via=%s, Verified=%t, Value=%s", i, addr.Via, addr.Verified, addr.Value)
		if addr.Via == "email" && addr.Verified {
			utils.LogInfo("User %s is verified via email", identity.Id)
			return true
		}
	}
	utils.LogInfo("User %s is not verified", identity.Id)
	return false
}

// Check if user authenticated via Google OAuth
func (vs *VerificationService) IsGoogleOAuthUser(identity client.Identity) bool {
	// Check if the user has OAuth credentials from Google
	if identity.Credentials != nil {
		credentials := *identity.Credentials
		if oidcCreds, ok := credentials["oidc"]; ok {
			if oidcCreds.Type != nil && *oidcCreds.Type == "oidc" && oidcCreds.Identifiers != nil {
				for _, identifier := range oidcCreds.Identifiers {
					if strings.HasPrefix(identifier, "google:") {
						utils.LogInfo("User %s authenticated via Google OAuth: %s", identity.Id, identifier)
						return true
					}
				}
			}
		}
	}

	// Having a Gmail address or a verified email doesn't mean the user authenticated via Google OAuth
	return false
}
