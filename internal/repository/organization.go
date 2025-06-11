// internal/repository/organization.go
package repository

import (
	"database/sql"
	"encoding/json"
	"userms/internal/models"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (or *OrganizationRepository) CreateOrganization(org *models.Organization) error {
	dataJSON, err := json.Marshal(org.Data)
	if err != nil {
		return err
	}

	_, err = or.db.Exec(`
		INSERT INTO organizations (id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, org.ID, org.DomainID, org.OrgID, org.OrgType, org.Name, org.Description, org.OwnerID, dataJSON)
	return err
}

func (or *OrganizationRepository) AddMember(orgID, userID, role string) error {
	_, err := or.db.Exec(`
		INSERT INTO user_organization_links (user_id, organization_id, role, joined_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, organization_id) DO UPDATE SET
			role = EXCLUDED.role
	`, userID, orgID, role)
	return err
}

func (or *OrganizationRepository) GetOrganization(orgID string) (*models.Organization, error) {
	org := &models.Organization{}
	var dataJSON []byte
	
	err := or.db.QueryRow(`
		SELECT id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at
		FROM organizations WHERE id = $1
	`, orgID).Scan(
		&org.ID, &org.DomainID, &org.OrgID, &org.OrgType, &org.Name,
		&org.Description, &org.OwnerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse JSON data
	if err := json.Unmarshal(dataJSON, &org.Data); err != nil {
		org.Data = make(map[string]interface{})
	}

	return org, nil
}

func (or *OrganizationRepository) ListOrganizations() ([]models.Organization, error) {
	rows, err := or.db.Query(`
		SELECT id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at
		FROM organizations ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []models.Organization
	for rows.Next() {
		var org models.Organization
		var dataJSON []byte
		
		err := rows.Scan(
			&org.ID, &org.DomainID, &org.OrgID, &org.OrgType, &org.Name,
			&org.Description, &org.OwnerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON data
		if err := json.Unmarshal(dataJSON, &org.Data); err != nil {
			org.Data = make(map[string]interface{})
		}

		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (or *OrganizationRepository) GetOrganizationMembers(orgID string) ([]models.Member, error) {
	rows, err := or.db.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, uol.role, uol.joined_at
		FROM user_organization_links uol
		JOIN users u ON uol.user_id = u.id
		WHERE uol.organization_id = $1
		ORDER BY uol.joined_at ASC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var member models.Member
		err := rows.Scan(&member.UserID, &member.Email, &member.FirstName, &member.LastName, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

func (or *OrganizationRepository) UpdateOrganization(orgID string, org *models.Organization) error {
	dataJSON, err := json.Marshal(org.Data)
	if err != nil {
		return err
	}

	_, err = or.db.Exec(`
		UPDATE organizations 
		SET name = $2, description = $3, data = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, orgID, org.Name, org.Description, dataJSON)
	return err
}

func (or *OrganizationRepository) DeleteOrganization(orgID string) error {
	_, err := or.db.Exec("DELETE FROM organizations WHERE id = $1", orgID)
	return err
}

func (or *OrganizationRepository) RemoveMember(orgID, userID string) error {
	_, err := or.db.Exec(`
		DELETE FROM user_organization_links 
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID)
	return err
}

func (or *OrganizationRepository) UpdateMemberRole(orgID, userID, role string) error {
	_, err := or.db.Exec(`
		UPDATE user_organization_links 
		SET role = $3 
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID, role)
	return err
}
