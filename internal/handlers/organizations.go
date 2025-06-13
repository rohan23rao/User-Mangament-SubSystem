package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	client "github.com/ory/kratos-client-go"
	"userms/internal/auth"
	"userms/internal/logger"
	"userms/internal/models"
)

type OrganizationHandler struct {
	authService *auth.Service
	kratosAdmin *client.APIClient
	db          *sql.DB
}

func NewOrganizationHandler(authService *auth.Service, kratosAdmin *client.APIClient, db *sql.DB) *OrganizationHandler {
	return &OrganizationHandler{
		authService: authService,
		kratosAdmin: kratosAdmin,
		db:          db,
	}
}

func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing organization creation request")

	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized organization creation: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Auth("Organization creation authorized for user: %s", session.Identity.Id)

	// Check if user has permission to create organizations
	var canCreateOrgs bool
	err = h.db.QueryRow("SELECT can_create_organizations FROM users WHERE id = $1", session.Identity.Id).Scan(&canCreateOrgs)
	if err != nil {
		logger.Error("Failed to check user permissions: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !canCreateOrgs {
		logger.Auth("User %s does not have permission to create organizations", session.Identity.Id)
		http.Error(w, "Forbidden: You do not have permission to create organizations", http.StatusForbidden)
		return
	}

	logger.Auth("User %s has permission to create organizations", session.Identity.Id)

	var req models.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for organization creation: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		logger.Warning("Organization creation failed: name is required")
		http.Error(w, "Organization name is required", http.StatusBadRequest)
		return
	}

	if req.OrgType == "" {
		req.OrgType = "organization"
	}

	// Validate org_type
	validTypes := map[string]bool{"domain": true, "organization": true, "tenant": true}
	if !validTypes[req.OrgType] {
		logger.Warning("Invalid org_type: %s", req.OrgType)
		http.Error(w, "Invalid org_type. Must be 'domain', 'organization', or 'tenant'", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

	logger.Info("Creating organization '%s' for user %s", req.Name, session.Identity.Id)

	orgID := uuid.New().String()
	dataJSON, _ := json.Marshal(req.Data)

	_, err = h.db.Exec(`
		INSERT INTO organizations (id, domain_id, org_id, org_type, name, description, owner_id, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		orgID, req.DomainID, req.OrgID, req.OrgType, req.Name, req.Description, session.Identity.Id, dataJSON,
	)
	if err != nil {
		logger.Error("Failed to create organization in database: %v", err)
		http.Error(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	logger.DB("Organization created with ID: %s", orgID)

	// Add owner as 'owner' role
	_, err = h.db.Exec(`
		INSERT INTO user_organization_links (user_id, organization_id, role)
		VALUES ($1, $2, $3)`,
		session.Identity.Id, orgID, "owner",
	)
	if err != nil {
		logger.Error("Failed to add owner to organization: %v", err)
		http.Error(w, "Failed to add owner to organization", http.StatusInternalServerError)
		return
	}

	logger.DB("Owner added as owner to organization %s", orgID)
	h.saveUserProfile(&session.Identity)

	org := models.Organization{
		ID:          orgID,
		DomainID:    req.DomainID,
		OrgID:       req.OrgID,
		OrgType:     req.OrgType,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     &session.Identity.Id,
		Data:        req.Data,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)

	logger.Success("Organization '%s' created successfully with ID: %s", req.Name, orgID)
}

func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	logger.Info("Processing list organizations request")

	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized list organizations: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Auth("List organizations authorized for user: %s", session.Identity.Id)

	rows, err := h.db.Query(`
		SELECT o.id, o.domain_id, o.org_id, o.org_type, o.name, o.description, o.owner_id, o.data, o.created_at, o.updated_at
		FROM organizations o
		JOIN user_organization_links uol ON o.id = uol.organization_id
		WHERE uol.user_id = $1
	`, session.Identity.Id)
	if err != nil {
		logger.Error("Failed to fetch organizations: %v", err)
		http.Error(w, "Failed to fetch organizations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var organizations []models.Organization
	for rows.Next() {
		var org models.Organization
		var dataJSON []byte
		err := rows.Scan(
			&org.ID, &org.DomainID, &org.OrgID, &org.OrgType, &org.Name, 
			&org.Description, &org.OwnerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			logger.Warning("Error scanning organization row: %v", err)
			continue
		}

		if len(dataJSON) > 0 {
			json.Unmarshal(dataJSON, &org.Data)
		}

		organizations = append(organizations, org)
	}

	logger.Success("Found %d organizations for user %s", len(organizations), session.Identity.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)
}

func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized get organization: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !h.isOrgMember(session.Identity.Id, orgID) {
		logger.Auth("User %s not member of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	logger.Info("Getting organization details for ID: %s", orgID)

	var org models.Organization
	var dataJSON []byte
	err = h.db.QueryRow(`
		SELECT id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at
		FROM organizations WHERE id = $1
	`, orgID).Scan(
		&org.ID, &org.DomainID, &org.OrgID, &org.OrgType, &org.Name,
		&org.Description, &org.OwnerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		logger.Warning("Organization not found: %s", orgID)
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}
	if err != nil {
		logger.Error("Failed to fetch organization: %v", err)
		http.Error(w, "Failed to fetch organization", http.StatusInternalServerError)
		return
	}

	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &org.Data)
	}

	members, err := h.getOrgMembers(orgID)
	if err != nil {
		logger.Warning("Failed to fetch organization members: %v", err)
	} else {
		org.Members = members
	}

	logger.Success("Organization details retrieved for ID: %s", orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized update organization: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !h.isOrgAdmin(session.Identity.Id, orgID) {
		logger.Auth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for organization update: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logger.Info("Updating organization %s", orgID)

	dataJSON, _ := json.Marshal(req.Data)
	_, err = h.db.Exec(`
		UPDATE organizations 
		SET name = $1, description = $2, org_type = $3, data = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5`,
		req.Name, req.Description, req.OrgType, dataJSON, orgID,
	)
	if err != nil {
		logger.Error("Failed to update organization: %v", err)
		http.Error(w, "Failed to update organization", http.StatusInternalServerError)
		return
	}

	logger.Success("Organization %s updated successfully", orgID)

	// Return updated organization
	h.GetOrganization(w, r)
}

func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized delete organization: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !h.isOrgOwner(session.Identity.Id, orgID) {
		logger.Auth("User %s not owner of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden: Only organization owners can delete organizations", http.StatusForbidden)
		return
	}

	logger.Info("Deleting organization %s", orgID)

	_, err = h.db.Exec("DELETE FROM organizations WHERE id = $1", orgID)
	if err != nil {
		logger.Error("Failed to delete organization: %v", err)
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	logger.Success("Organization %s deleted successfully", orgID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized add member: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !h.isOrgAdmin(session.Identity.Id, orgID) {
		logger.Auth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for add member: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validRoles := map[string]bool{"admin": true, "member": true}
	if !validRoles[req.Role] {
		logger.Warning("Invalid role: %s", req.Role)
		http.Error(w, "Invalid role. Must be 'admin' or 'member'", http.StatusBadRequest)
		return
	}

	logger.Info("Adding member %s to organization %s with role %s", req.Email, orgID, req.Role)

	// Find user by email from Kratos
	identities, resp, err := h.kratosAdmin.IdentityApi.ListIdentities(context.Background()).Execute()
	if err != nil {
		logger.Error("Failed to fetch identities from Kratos: %v", err)
		http.Error(w, "Failed to lookup user", http.StatusInternalServerError)
		return
	}
	if resp != nil {
		defer resp.Body.Close()
	}

	var targetUserID string
	for _, identity := range identities {
		if traits, ok := identity.Traits.(map[string]interface{}); ok {
			if email, exists := traits["email"].(string); exists && email == req.Email {
				targetUserID = identity.Id
				break
			}
		}
	}

	if targetUserID == "" {
		logger.Warning("User not found: %s", req.Email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logger.Info("Found user %s for email %s", targetUserID, req.Email)

	_, err = h.db.Exec(`
		INSERT INTO user_organization_links (user_id, organization_id, role) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, organization_id) 
		DO UPDATE SET role = $3, joined_at = CURRENT_TIMESTAMP`,
		targetUserID, orgID, req.Role,
	)
	if err != nil {
		logger.Error("Failed to add member to database: %v", err)
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	logger.DB("Member %s added to organization %s with role %s", req.Email, orgID, req.Role)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added successfully"})

	logger.Success("Member %s added successfully to organization %s", req.Email, orgID)
}

func (h *OrganizationHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized get members: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !h.isOrgMember(session.Identity.Id, orgID) {
		logger.Auth("User %s not member of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	logger.Info("Getting members for organization %s", orgID)

	members, err := h.getOrgMembers(orgID)
	if err != nil {
		logger.Error("Failed to fetch members: %v", err)
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}

	logger.Info("Found %d members for organization %s", len(members), orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized update member role: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]
	userID := vars["user_id"]

	if !h.isOrgAdmin(session.Identity.Id, orgID) {
		logger.Auth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Invalid request body for update member role: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validRoles := map[string]bool{"admin": true, "member": true, "owner": true}
	if !validRoles[req.Role] {
		logger.Warning("Invalid role: %s", req.Role)
		http.Error(w, "Invalid role. Must be 'admin', 'member', or 'owner'", http.StatusBadRequest)
		return
	}

	// Check if target user is currently an owner - prevent owner demotion
	var currentRole string
	err = h.db.QueryRow(`
		SELECT role FROM user_organization_links 
		WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID,
	).Scan(&currentRole)
	
	if err == nil && currentRole == "owner" && req.Role != "owner" {
		logger.Auth("Attempt to demote owner %s blocked", userID)
		http.Error(w, "Forbidden: Cannot demote organization owner", http.StatusForbidden)
		return
	}

	// Only owners can promote to owner
	if req.Role == "owner" && !h.isOrgOwner(session.Identity.Id, orgID) {
		logger.Auth("Non-owner %s attempted to promote user to owner", session.Identity.Id)
		http.Error(w, "Forbidden: Only owners can promote users to owner", http.StatusForbidden)
		return
	}

	logger.Info("Updating member %s role to %s in organization %s", userID, req.Role, orgID)

	// If promoting to owner, handle ownership transfer
	if req.Role == "owner" {
		tx, err := h.db.Begin()
		if err != nil {
			logger.Error("Failed to begin transaction: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Update user_organization_links
		_, err = tx.Exec(`
			UPDATE user_organization_links 
			SET role = $1 
			WHERE user_id = $2 AND organization_id = $3`,
			req.Role, userID, orgID,
		)
		if err != nil {
			logger.Error("Failed to update member role: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}

		// Update organizations.owner_id
		_, err = tx.Exec(`
			UPDATE organizations 
			SET owner_id = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2`,
			userID, orgID,
		)
		if err != nil {
			logger.Error("Failed to update organization owner: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}

		// Demote previous owner to admin
		_, err = tx.Exec(`
			UPDATE user_organization_links 
			SET role = 'admin' 
			WHERE organization_id = $1 AND role = 'owner' AND user_id != $2`,
			orgID, userID,
		)
		if err != nil {
			logger.Error("Failed to demote previous owner: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}

		if err = tx.Commit(); err != nil {
			logger.Error("Failed to commit ownership transfer: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}
	} else {
		// Regular role update
		_, err = h.db.Exec(`
			UPDATE user_organization_links 
			SET role = $1 
			WHERE user_id = $2 AND organization_id = $3`,
			req.Role, userID, orgID,
		)
		if err != nil {
			logger.Error("Failed to update member role: %v", err)
			http.Error(w, "Failed to update member role", http.StatusInternalServerError)
			return
		}
	}

	// Get updated member info
	var member models.Member
	err = h.db.QueryRow(`
		SELECT uol.user_id, uol.role, uol.joined_at, u.email, u.first_name, u.last_name
		FROM user_organization_links uol
		LEFT JOIN users u ON uol.user_id = u.id
		WHERE uol.user_id = $1 AND uol.organization_id = $2
	`, userID, orgID).Scan(&member.UserID, &member.Role, &member.JoinedAt, 
		&member.Email, &member.FirstName, &member.LastName)

	if err != nil {
		logger.Error("Failed to fetch updated member info: %v", err)
		http.Error(w, "Failed to fetch updated member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)

	logger.Success("Member %s role updated successfully to %s in organization %s", userID, req.Role, orgID)
}
func (h *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	session, err := h.authService.GetSessionFromRequest(r)
	if err != nil {
		logger.Auth("Unauthorized remove member: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]
	targetUserID := vars["user_id"]

	if !h.isOrgAdmin(session.Identity.Id, orgID) {
		logger.Auth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden: Only admins can remove members", http.StatusForbidden)
		return
	}

	// Check if target user is an owner - cannot remove owners
	var targetRole string
	err = h.db.QueryRow(`
		SELECT role FROM user_organization_links 
		WHERE user_id = $1 AND organization_id = $2`,
		targetUserID, orgID,
	).Scan(&targetRole)
	
	if err == nil && targetRole == "owner" {
		logger.Auth("Attempt to remove owner %s blocked", targetUserID)
		http.Error(w, "Forbidden: Cannot remove organization owner", http.StatusForbidden)
		return
	}

	result, err := h.db.Exec(`
		DELETE FROM user_organization_links 
		WHERE user_id = $1 AND organization_id = $2`,
		targetUserID, orgID,
	)
	if err != nil {
		logger.Error("Failed to remove member: %v", err)
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}

	logger.Success("Member %s removed from organization %s", targetUserID, orgID)
	w.WriteHeader(http.StatusNoContent)
}
// Helper functions
func (h *OrganizationHandler) getOrgMembers(orgID string) ([]models.Member, error) {
	rows, err := h.db.Query(`
		SELECT uol.user_id, uol.role, uol.joined_at, u.email, u.first_name, u.last_name
		FROM user_organization_links uol
		LEFT JOIN users u ON uol.user_id = u.id
		WHERE uol.organization_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var member models.Member
		var email, firstName, lastName sql.NullString
		err := rows.Scan(&member.UserID, &member.Role, &member.JoinedAt, &email, &firstName, &lastName)
		if err != nil {
			logger.Warning("Error scanning member row: %v", err)
			continue
		}

		if email.Valid {
			member.Email = email.String
		}
		if firstName.Valid {
			member.FirstName = firstName.String
		}
		if lastName.Valid {
			member.LastName = lastName.String
		}

		members = append(members, member)
	}

	return members, nil
}

func (h *OrganizationHandler) isOrgMember(userID, orgID string) bool {
	var count int
	err := h.db.QueryRow(
		"SELECT COUNT(*) FROM user_organization_links WHERE user_id = $1 AND organization_id = $2",
		userID, orgID,
	).Scan(&count)
	return err == nil && count > 0
}

// UPDATED: Enhanced isOrgAdmin to include owners
func (h *OrganizationHandler) isOrgAdmin(userID, orgID string) bool {
	var count int
	err := h.db.QueryRow(
		"SELECT COUNT(*) FROM user_organization_links WHERE user_id = $1 AND organization_id = $2 AND role IN ('admin', 'owner')",
		userID, orgID,
	).Scan(&count)
	return err == nil && count > 0
}

// ADDED: Helper function to check if user is organization owner
func (h *OrganizationHandler) isOrgOwner(userID, orgID string) bool {
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM user_organization_links uol
		JOIN organizations o ON uol.organization_id = o.id
		WHERE uol.user_id = $1 AND uol.organization_id = $2 
		AND (uol.role = 'owner' OR o.owner_id = $1)`,
		userID, orgID,
	).Scan(&count)
	return err == nil && count > 0
}

func (h *OrganizationHandler) saveUserProfile(identity *client.Identity) {
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

	_, err := h.db.Exec(`
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