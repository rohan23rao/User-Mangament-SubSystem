// internal/handlers/organizations.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"userms/internal/auth"
	"userms/internal/models"
	"userms/internal/repository"
	"userms/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OrganizationHandler struct {
	orgRepo             *repository.OrganizationRepository
	userRepo            *repository.UserRepository
	sessionManager      *auth.SessionManager
	verificationService *auth.VerificationService
}

func NewOrganizationHandler(db *sql.DB, sessionManager *auth.SessionManager, verificationService *auth.VerificationService) *OrganizationHandler {
	return &OrganizationHandler{
		orgRepo:             repository.NewOrganizationRepository(db),
		userRepo:            repository.NewUserRepository(db),
		sessionManager:      sessionManager,
		verificationService: verificationService,
	}
}

func (oh *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Processing organization creation request")

	session, err := oh.sessionManager.GetSessionFromRequest(r)
	if err != nil {
		utils.LogAuth("Unauthorized organization creation: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError("Invalid organization creation request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	utils.LogInfo("Creating organization '%s' for user %s", req.Name, session.Identity.Id)

	orgID := uuid.New().String()
	org := &models.Organization{
		ID:          orgID,
		DomainID:    req.DomainID,
		OrgID:       req.OrgID,
		OrgType:     req.OrgType,
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     &session.Identity.Id,
		Data:        req.Data,
	}

	if org.Data == nil {
		org.Data = make(map[string]interface{})
	}

	err = oh.orgRepo.CreateOrganization(org)
	if err != nil {
		utils.LogError("Failed to create organization: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Add creator as admin
	err = oh.orgRepo.AddMember(orgID, session.Identity.Id, "admin")
	if err != nil {
		utils.LogError("Failed to add creator as admin: %v", err)
		http.Error(w, "Failed to set admin permissions", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Organization '%s' created successfully with ID: %s", req.Name, orgID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (oh *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Listing organizations")

	orgs, err := oh.orgRepo.ListOrganizations()
	if err != nil {
		utils.LogError("Failed to list organizations: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogInfo("Found %d organizations", len(orgs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

func (oh *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]

	utils.LogInfo("Getting organization: %s", orgID)

	org, err := oh.orgRepo.GetOrganization(orgID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Organization not found", http.StatusNotFound)
		} else {
			utils.LogError("Failed to get organization: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Get members
	members, err := oh.orgRepo.GetOrganizationMembers(orgID)
	if err != nil {
		utils.LogWarning("Failed to get organization members: %v", err)
	} else {
		org.Members = members
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)
}

func (oh *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]

	utils.LogInfo("Updating organization: %s", orgID)

	var req models.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	org := &models.Organization{
		Name:        req.Name,
		Description: req.Description,
		Data:        req.Data,
	}

	if org.Data == nil {
		org.Data = make(map[string]interface{})
	}

	err := oh.orgRepo.UpdateOrganization(orgID, org)
	if err != nil {
		utils.LogError("Failed to update organization: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Organization %s updated successfully", orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Organization updated successfully"})
}

func (oh *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]

	utils.LogInfo("Deleting organization: %s", orgID)

	err := oh.orgRepo.DeleteOrganization(orgID)
	if err != nil {
		utils.LogError("Failed to delete organization: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Organization %s deleted successfully", orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Organization deleted successfully"})
}

func (oh *OrganizationHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]

	utils.LogInfo("Adding member to organization: %s", orgID)

	var req models.InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO: In a real implementation, you'd send an invitation email
	// For now, we'll assume the user exists and add them directly

	err := oh.orgRepo.AddMember(orgID, req.Email, req.Role) // Note: Using email as userID for now
	if err != nil {
		utils.LogError("Failed to add member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Member %s added to organization %s with role %s", req.Email, orgID, req.Role)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added successfully"})
}

func (oh *OrganizationHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]

	utils.LogInfo("Getting members for organization: %s", orgID)

	members, err := oh.orgRepo.GetOrganizationMembers(orgID)
	if err != nil {
		utils.LogError("Failed to get organization members: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (oh *OrganizationHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]
	userID := vars["userId"]

	utils.LogInfo("Removing member %s from organization: %s", userID, orgID)

	err := oh.orgRepo.RemoveMember(orgID, userID)
	if err != nil {
		utils.LogError("Failed to remove member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Member %s removed from organization %s", userID, orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})
}

func (oh *OrganizationHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["id"]
	userID := vars["userId"]

	utils.LogInfo("Updating member role for %s in organization: %s", userID, orgID)

	var req models.UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := oh.orgRepo.UpdateMemberRole(orgID, userID, req.Role)
	if err != nil {
		utils.LogError("Failed to update member role: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	utils.LogSuccess("Member %s role updated to %s in organization %s", userID, req.Role, orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member role updated successfully"})
}
