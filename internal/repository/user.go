// internal/repository/user.go
package repository

import (
	"database/sql"
	"userms/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUserCount() (int, error) {
	var count int
	err := ur.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func (ur *UserRepository) UpsertUser(user *models.User) error {
	_, err := ur.db.Exec(`
		INSERT INTO users (id, email, first_name, last_name, time_zone, ui_mode, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			email = EXCLUDED.email,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			updated_at = CURRENT_TIMESTAMP
	`, user.ID, user.Email, user.FirstName, user.LastName, user.TimeZone, user.UIMode)
	return err
}

func (ur *UserRepository) UpdateLastLogin(userID string) error {
	_, err := ur.db.Exec(`
		UPDATE users 
		SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, userID)
	return err
}

func (ur *UserRepository) GetUserFromDB(userID string) (*models.User, error) {
	user := &models.User{}
	err := ur.db.QueryRow(`
		SELECT id, email, first_name, last_name, time_zone, ui_mode, created_at, updated_at, last_login
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.TimeZone, &user.UIMode, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (ur *UserRepository) GetUserOrganizations(userID string) ([]models.OrgMember, error) {
	rows, err := ur.db.Query(`
		SELECT o.id, o.name, o.org_type, uol.role, uol.joined_at
		FROM user_organization_links uol
		JOIN organizations o ON uol.organization_id = o.id
		WHERE uol.user_id = $1
		ORDER BY uol.joined_at ASC
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
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (ur *UserRepository) IsUserAdmin(userID string) bool {
	var count int
	err := ur.db.QueryRow(`
		SELECT COUNT(*) 
		FROM user_organization_links 
		WHERE user_id = $1 AND role IN ('admin', 'owner')
	`, userID).Scan(&count)
	return err == nil && count > 0
}
