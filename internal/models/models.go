// internal/models/models.go
package models

import (
	"time"
	client "github.com/ory/kratos-client-go"
)

type User struct {
	ID                  string              `json:"id"`
	Email               string              `json:"email"`
	FirstName           string              `json:"first_name"`
	LastName            string              `json:"last_name"`
	TimeZone            string              `json:"time_zone"`
	UIMode              string              `json:"ui_mode"`
	Traits              interface{}         `json:"traits"`
	Organizations       []OrgMember         `json:"organizations"`
	VerifiableAddresses []VerifiableAddress `json:"verifiable_addresses,omitempty"`
	RecoveryAddresses   []RecoveryAddress   `json:"recovery_addresses,omitempty"`
	Verified            bool                `json:"verified"`
	CreatedAt           time.Time           `json:"created_at"`
	UpdatedAt           time.Time           `json:"updated_at"`
	LastLogin           *time.Time          `json:"last_login"`
}

type VerifiableAddress struct {
	ID       string `json:"id"`
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
	Via      string `json:"via"`
	Status   string `json:"status"`
}

type RecoveryAddress struct {
	ID    string `json:"id"`
	Value string `json:"value"`
	Via   string `json:"via"`
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
