package models

import (
	"time"

	client "github.com/ory/kratos-client-go"
)

type User struct {
	ID                     string      `json:"id"`
	Email                  string      `json:"email"`
	EmailVerified          bool        `json:"email_verified"`
	FirstName              string      `json:"first_name"`
	LastName               string      `json:"last_name"`
	TimeZone               string      `json:"time_zone"`
	UIMode                 string      `json:"ui_mode"`
	CanCreateOrganizations bool        `json:"can_create_organizations"` // ADDED: New permission field
	Traits                 interface{} `json:"traits"`
	Organizations          []OrgMember `json:"organizations,omitempty"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
	LastLogin              *time.Time  `json:"last_login"`
}

type Organization struct {
	ID          string                 `json:"id"`
	DomainID    *string                `json:"domain_id"`
	OrgID       *string                `json:"org_id"`
	OrgType     string                 `json:"org_type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OwnerID     *string                `json:"owner_id"`
	Data        map[string]interface{} `json:"data"`
	Members     []Member               `json:"members,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type Member struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

type OrgMember struct {
	OrgID    string    `json:"org_id"`
	OrgName  string    `json:"org_name"`
	OrgType  string    `json:"org_type"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type WebhookPayload struct {
	Identity client.Identity `json:"identity"`
	Flow     interface{}     `json:"flow"`
}

type CreateOrgRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OrgType     string                 `json:"org_type"`
	DomainID    *string                `json:"domain_id"`
	OrgID       *string                `json:"org_id"`
	Data        map[string]interface{} `json:"data"`
}

type InviteUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

// Add these to your existing internal/models/models.go file

// OAuth2Client represents a machine-to-machine OAuth2 client
type OAuth2Client struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret,omitempty"` // Omit in API responses for security
	UserID       string    `json:"user_id"`
	OrgID        string    `json:"org_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Scopes       string    `json:"scopes"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

// CreateM2MClientRequest represents a request to create a machine-to-machine client
type CreateM2MClientRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	OrgID       string `json:"org_id" validate:"required"`
	Scopes      string `json:"scopes,omitempty"` // Optional, defaults to data_pipeline
}

// TokenRequest represents a request to generate an OAuth2 token
type TokenRequest struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	GrantType    string `json:"grant_type,omitempty"` // Defaults to client_credentials
	Scope        string `json:"scope,omitempty"`
}

// TokenResponse represents an OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// TokenInfo represents information about a validated token
type TokenInfo struct {
	Active    bool      `json:"active"`
	ClientID  string    `json:"client_id"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
	Subject   string    `json:"subject,omitempty"`
}

// APIKey represents an API key for authentication (alternative to OAuth2)
type APIKey struct {
	ID          string     `json:"id"`
	KeyHash     string     `json:"-"` // Never expose the hash
	KeyPrefix   string     `json:"key_prefix"` // First 8 chars for identification
	UserID      string     `json:"user_id"`
	OrgID       string     `json:"org_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Scopes      string     `json:"scopes"`
	IsActive    bool       `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// OAuth2TokenLog represents a log entry for token usage
type OAuth2TokenLog struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	GrantedScopes string    `json:"granted_scopes"`
	IPAddress     string    `json:"ip_address,omitempty"`
	UserAgent     string    `json:"user_agent,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// ClientIPWhitelist represents IP whitelist entries for OAuth2 clients
type ClientIPWhitelist struct {
	ID          string    `json:"id"`
	ClientID    string    `json:"client_id"`
	IPAddress   string    `json:"ip_address"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}