package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	client "github.com/ory/kratos-client-go"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

type Server struct {
	kratosPublic *client.APIClient
	kratosAdmin  *client.APIClient
	db           *sql.DB
}

type User struct {
	ID            string      `json:"id"`
	Email         string      `json:"email"`
	FirstName     string      `json:"first_name"`
	LastName      string      `json:"last_name"`
	TimeZone      string      `json:"time_zone"`
	UIMode        string      `json:"ui_mode"`
	Traits        interface{} `json:"traits"`
	Organizations []OrgMember `json:"organizations,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	LastLogin     *time.Time  `json:"last_login"`
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

// Colored logging functions
func logInfo(message string, args ...interface{}) {
	log.Printf(ColorBlue+"[INFO]"+ColorReset+" "+message, args...)
}

func logSuccess(message string, args ...interface{}) {
	log.Printf(ColorGreen+"[SUCCESS]"+ColorReset+" "+message, args...)
}

func logWarning(message string, args ...interface{}) {
	log.Printf(ColorYellow+"[WARNING]"+ColorReset+" "+message, args...)
}

func logError(message string, args ...interface{}) {
	log.Printf(ColorRed+"[ERROR]"+ColorReset+" "+message, args...)
}

func logRequest(method, path, userID string) {
	log.Printf(ColorCyan+"[REQUEST]"+ColorReset+" %s %s | User: %s", method, path, userID)
}

func logAuth(message string, args ...interface{}) {
	log.Printf(ColorPurple+"[AUTH]"+ColorReset+" "+message, args...)
}

func logDB(message string, args ...interface{}) {
	log.Printf(ColorWhite+"[DB]"+ColorReset+" "+message, args...)
}

func NewServer() *Server {
	kratosPublicURL := getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433")
	kratosAdminURL := getEnv("KRATOS_ADMIN_URL", "http://localhost:4434")

	logInfo("Initializing server with Kratos URLs:")
	logInfo("  Public: %s", kratosPublicURL)
	logInfo("  Admin: %s", kratosAdminURL)

	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{{URL: kratosPublicURL}}

	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{{URL: kratosAdminURL}}

	logInfo("Initializing database...")
	db, err := initDB()
	if err != nil {
		logError("Failed to initialize database: %v", err)
		log.Fatal("Database initialization failed")
	}
	logSuccess("Database initialized successfully")

	return &Server{
		kratosPublic: client.NewAPIClient(publicConfig),
		kratosAdmin:  client.NewAPIClient(adminConfig),
		db:           db,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func initDB() (*sql.DB, error) {
	databaseURL := getEnv("DATABASE_URL", "postgres://userms:userms_password@localhost:5432/userms?sslmode=disable")

	logDB("Connecting to PostgreSQL database...")
	logDB("Database URL: %s", strings.ReplaceAll(databaseURL, "userms_password", "***"))

	// Retry connection with backoff
	var db *sql.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			logError("Failed to open database connection: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			logWarning("Database not ready, retrying in 2 seconds... (attempt %d/30)", i+1)
			time.Sleep(2 * time.Second)
			continue
		}

		logSuccess("Connected to PostgreSQL database")
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 30 attempts: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test if our tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count)
	if err != nil {
		logError("Failed to check if tables exist: %v", err)
		return nil, err
	}

	if count == 0 {
		logWarning("Tables don't exist yet - they should be created by init.sql")
	} else {
		logDB("Database tables verified")
	}

	return db, nil
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		session, _ := s.getSessionFromRequest(r)
		userID := "anonymous"
		if session != nil {
			userID = session.Identity.Id[:8] + "..."
		}

		logRequest(r.Method, r.URL.Path, userID)

		wrapper := &responseWrapper{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		statusColor := ColorGreen
		if wrapper.statusCode >= 400 {
			statusColor = ColorRed
		} else if wrapper.statusCode >= 300 {
			statusColor = ColorYellow
		}

		log.Printf(ColorCyan+"[RESPONSE]"+ColorReset+" %s%d"+ColorReset+" | %s | %v",
			statusColor, wrapper.statusCode, r.URL.Path, duration)
	})
}

func (s *Server) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(s.loggingMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	// User endpoints
	api.HandleFunc("/whoami", s.whoAmI).Methods("GET")
	api.HandleFunc("/users", s.listUsers).Methods("GET")
	api.HandleFunc("/users/{id}", s.getUser).Methods("GET")

	// Organization endpoints
	api.HandleFunc("/organizations", s.createOrganization).Methods("POST")
	api.HandleFunc("/organizations", s.listOrganizations).Methods("GET")
	api.HandleFunc("/organizations/{id}", s.getOrganization).Methods("GET")
	api.HandleFunc("/organizations/{id}", s.updateOrganization).Methods("PUT")
	api.HandleFunc("/organizations/{id}", s.deleteOrganization).Methods("DELETE")

	// Organization member endpoints
	api.HandleFunc("/organizations/{id}/members", s.addMember).Methods("POST")
	api.HandleFunc("/organizations/{id}/members", s.getMembers).Methods("GET")
	api.HandleFunc("/organizations/{id}/members/{userId}", s.removeMember).Methods("DELETE")
	api.HandleFunc("/organizations/{id}/members/{userId}/role", s.updateMemberRole).Methods("PUT")

	// Debug endpoint
	api.HandleFunc("/debug/auth", s.debugAuth).Methods("GET")

	// Webhook endpoints
	hooks := r.PathPrefix("/hooks").Subrouter()
	hooks.HandleFunc("/after-registration", s.handleAfterRegistration).Methods("POST")
	hooks.HandleFunc("/after-login", s.handleAfterLogin).Methods("POST")

	// System endpoints
	r.HandleFunc("/health", s.healthCheck).Methods("GET")
	r.HandleFunc("/auth/session", s.getSession).Methods("GET")
	r.HandleFunc("/auth/logout", s.logout).Methods("POST")

	logInfo("Routes configured successfully")
	return r
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) getSessionFromRequest(r *http.Request) (*client.Session, error) {
	logAuth("=== SESSION VALIDATION START ===")

	// Log all cookies for debugging
	logAuth("All cookies in request:")
	for _, cookie := range r.Cookies() {
		if cookie.Name == "ory_kratos_session" {
			logAuth("  %s: %s (length: %d)", cookie.Name, cookie.Value[:min(len(cookie.Value), 30)]+"...", len(cookie.Value))
		} else {
			logAuth("  %s: %s", cookie.Name, cookie.Value)
		}
	}

	// Log headers for debugging
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		logAuth("Authorization header found: %s", authHeader[:min(len(authHeader), 50)]+"...")
	}

	var sessionToken string

	// Method 1: Try Authorization header (Bearer token)
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
		logAuth("Extracted Bearer token: %s...", sessionToken[:min(len(sessionToken), 20)])

		session, resp, err := s.kratosPublic.FrontendApi.ToSession(context.Background()).
			XSessionToken(sessionToken).
			Execute()

		if err != nil {
			logAuth("Bearer token validation failed: %v", err)
			if resp != nil {
				logAuth("Response status: %d", resp.StatusCode)
			}
		} else if resp.StatusCode == 200 {
			logAuth("✅ Bearer token validated successfully for user: %s", session.Identity.Id)
			return session, nil
		}
	}

	// Method 2: Try session cookie
	sessionCookie, err := r.Cookie("ory_kratos_session")
	if err != nil {
		logAuth("❌ No ory_kratos_session cookie found: %v", err)
		return nil, fmt.Errorf("no session found")
	}

	sessionToken = sessionCookie.Value
	logAuth("Found session cookie value: %s... (length: %d)", sessionToken[:min(len(sessionToken), 20)], len(sessionToken))

	// Try validation method 1: X-Session-Token
	logAuth("Trying validation with X-Session-Token header...")
	session, resp, err := s.kratosPublic.FrontendApi.ToSession(context.Background()).
		XSessionToken(sessionToken).
		Execute()

	if err == nil && resp != nil && resp.StatusCode == 200 {
		logAuth("✅ Session validated via X-Session-Token for user: %s", session.Identity.Id)
		return session, nil
	}

	if err != nil {
		logAuth("X-Session-Token validation failed: %v", err)
	}
	if resp != nil {
		logAuth("X-Session-Token response status: %d", resp.StatusCode)
	}

	// Try validation method 2: Cookie header
	logAuth("Trying validation with Cookie header...")
	cookieHeader := fmt.Sprintf("ory_kratos_session=%s", sessionToken)
	logAuth("Cookie header: %s...", cookieHeader[:min(len(cookieHeader), 50)])

	session, resp, err = s.kratosPublic.FrontendApi.ToSession(context.Background()).
		Cookie(cookieHeader).
		Execute()

	if err != nil {
		logAuth("❌ Cookie validation failed: %v", err)
		if resp != nil {
			logAuth("Response status: %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("invalid session from cookie: %v", err)
	}

	if resp.StatusCode != 200 {
		logAuth("❌ Cookie validation bad status: %d", resp.StatusCode)
		return nil, fmt.Errorf("invalid session status: %d", resp.StatusCode)
	}

	logAuth("✅ Session validated via Cookie for user: %s", session.Identity.Id)
	logAuth("=== SESSION VALIDATION END ===")
	return session, nil
}

func (s *Server) debugAuth(w http.ResponseWriter, r *http.Request) {
	logAuth("=== DEBUG AUTH ENDPOINT ===")

	session, err := s.getSessionFromRequest(r)

	response := map[string]interface{}{
		"timestamp": time.Now(),
		"method":    r.Method,
		"url":       r.URL.String(),
		"headers":   map[string][]string{},
		"cookies":   map[string]string{},
	}

	// Add headers (sanitized)
	for name, values := range r.Header {
		if name == "Authorization" && len(values) > 0 {
			response["headers"].(map[string][]string)[name] = []string{values[0][:min(len(values[0]), 30)] + "..."}
		} else {
			response["headers"].(map[string][]string)[name] = values
		}
	}

	// Add cookies (sanitized)
	for _, cookie := range r.Cookies() {
		if cookie.Name == "ory_kratos_session" {
			response["cookies"].(map[string]string)[cookie.Name] = cookie.Value[:min(len(cookie.Value), 30)] + "..."
		} else {
			response["cookies"].(map[string]string)[cookie.Name] = cookie.Value
		}
	}

	if err != nil {
		response["auth_status"] = "failed"
		response["auth_error"] = err.Error()
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		response["auth_status"] = "success"
		response["user_id"] = session.Identity.Id
		response["user_email"] = s.getEmailFromIdentity(session.Identity)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	logAuth("=== DEBUG AUTH ENDPOINT END ===")
}

func (s *Server) mapIdentityToUser(identity client.Identity) User {
	user := User{
		ID:     identity.Id,
		Traits: identity.Traits,
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

	return user
}

// User Management Endpoints

func (s *Server) whoAmI(w http.ResponseWriter, r *http.Request) {
	logAuth("Processing whoami request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized whoami request: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logAuth("Whoami request authenticated for user: %s", session.Identity.Id)
	user := s.mapIdentityToUser(session.Identity)

	// Get user from database for additional info
	dbUser, err := s.getUserFromDB(user.ID)
	if err != nil {
		logWarning("Error getting user from database: %v", err)
	} else if dbUser != nil {
		// Merge database info with Kratos identity
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	orgs, err := s.getUserOrganizations(user.ID)
	if err != nil {
		logWarning("Error getting user organizations: %v", err)
	} else {
		user.Organizations = orgs
		logInfo("Found %d organizations for user %s", len(orgs), user.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
	logSuccess("Whoami response sent for user: %s", user.Email)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing list users request")

	identities, resp, err := s.kratosAdmin.IdentityApi.ListIdentities(context.Background()).Execute()
	if err != nil || resp.StatusCode != 200 {
		logError("Failed to fetch users from Kratos: %v (status: %d)", err, resp.StatusCode)
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	var users []User
	for _, identity := range identities {
		user := s.mapIdentityToUser(identity)

		// Get additional info from database
		dbUser, err := s.getUserFromDB(user.ID)
		if err == nil && dbUser != nil {
			user.FirstName = dbUser.FirstName
			user.LastName = dbUser.LastName
			user.TimeZone = dbUser.TimeZone
			user.UIMode = dbUser.UIMode
			user.CreatedAt = dbUser.CreatedAt
			user.UpdatedAt = dbUser.UpdatedAt
			user.LastLogin = dbUser.LastLogin
		}

		users = append(users, user)
	}

	logInfo("Found %d users in Kratos", len(users))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

	logSuccess("Users list sent successfully")
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	logInfo("Getting user details for: %s", userID)

	identity, resp, err := s.kratosAdmin.IdentityApi.GetIdentity(context.Background(), userID).Execute()
	if err != nil || resp.StatusCode != 200 {
		logWarning("User not found: %s", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user := s.mapIdentityToUser(*identity)

	// Get additional info from database
	dbUser, err := s.getUserFromDB(user.ID)
	if err == nil && dbUser != nil {
		user.FirstName = dbUser.FirstName
		user.LastName = dbUser.LastName
		user.TimeZone = dbUser.TimeZone
		user.UIMode = dbUser.UIMode
		user.CreatedAt = dbUser.CreatedAt
		user.UpdatedAt = dbUser.UpdatedAt
		user.LastLogin = dbUser.LastLogin
	}

	logSuccess("User details retrieved for: %s", user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Organization Management Endpoints

func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing organization creation request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized organization creation: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logAuth("Organization creation authorized for user: %s", session.Identity.Id)

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError("Invalid request body for organization creation: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		logWarning("Organization creation failed: name is required")
		http.Error(w, "Organization name is required", http.StatusBadRequest)
		return
	}

	if req.OrgType == "" {
		req.OrgType = "organization"
	}

	// Validate org_type
	validTypes := map[string]bool{"domain": true, "organization": true, "tenant": true}
	if !validTypes[req.OrgType] {
		logWarning("Invalid org_type: %s", req.OrgType)
		http.Error(w, "Invalid org_type. Must be 'domain', 'organization', or 'tenant'", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

	logInfo("Creating organization '%s' for user %s", req.Name, session.Identity.Id)

	orgID := uuid.New().String()
	dataJSON, _ := json.Marshal(req.Data)

	_, err = s.db.Exec(`
		INSERT INTO organizations (id, domain_id, org_id, org_type, name, description, owner_id, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		orgID, req.DomainID, req.OrgID, req.OrgType, req.Name, req.Description, session.Identity.Id, dataJSON,
	)
	if err != nil {
		logError("Failed to create organization in database: %v", err)
		http.Error(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	logDB("Organization created with ID: %s", orgID)

	// Add owner as admin member
	_, err = s.db.Exec(`
		INSERT INTO user_organization_links (user_id, organization_id, role)
		VALUES ($1, $2, $3)`,
		session.Identity.Id, orgID, "admin",
	)
	if err != nil {
		logError("Failed to add owner to organization: %v", err)
		http.Error(w, "Failed to add owner to organization", http.StatusInternalServerError)
		return
	}

	logDB("Owner added as admin to organization %s", orgID)
	s.saveUserProfile(session.Identity)

	org := Organization{
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

	logSuccess("Organization '%s' created successfully with ID: %s", req.Name, orgID)
}

func (s *Server) listOrganizations(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing list organizations request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized list organizations: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logAuth("List organizations authorized for user: %s", session.Identity.Id)

	rows, err := s.db.Query(`
		SELECT o.id, o.domain_id, o.org_id, o.org_type, o.name, o.description, o.owner_id, 
		       o.data, o.created_at, o.updated_at, uol.role
		FROM organizations o
		JOIN user_organization_links uol ON o.id = uol.organization_id
		WHERE uol.user_id = $1
	`, session.Identity.Id)
	if err != nil {
		logError("Failed to fetch organizations from database: %v", err)
		http.Error(w, "Failed to fetch organizations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var org Organization
		var role string
		var dataJSON []byte
		var domainID, orgID, ownerID sql.NullString

		err := rows.Scan(&org.ID, &domainID, &orgID, &org.OrgType, &org.Name, &org.Description,
			&ownerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt, &role)
		if err != nil {
			logWarning("Error scanning organization row: %v", err)
			continue
		}

		if domainID.Valid {
			org.DomainID = &domainID.String
		}
		if orgID.Valid {
			org.OrgID = &orgID.String
		}
		if ownerID.Valid {
			org.OwnerID = &ownerID.String
		}

		if len(dataJSON) > 0 {
			json.Unmarshal(dataJSON, &org.Data)
		} else {
			org.Data = make(map[string]interface{})
		}

		organizations = append(organizations, org)
	}

	logInfo("Found %d organizations for user", len(organizations))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(organizations)

	logSuccess("Organizations list sent successfully")
}

func (s *Server) getOrganization(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized get organization: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	logInfo("Getting organization %s for user %s", orgID, session.Identity.Id)

	if !s.isOrgMember(session.Identity.Id, orgID) {
		logAuth("User %s not authorized for organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var org Organization
	var dataJSON []byte
	var domainID, parentOrgID, ownerID sql.NullString

	err = s.db.QueryRow(`
		SELECT id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at
		FROM organizations WHERE id = $1`,
		orgID,
	).Scan(&org.ID, &domainID, &parentOrgID, &org.OrgType, &org.Name, &org.Description,
		&ownerID, &dataJSON, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			logWarning("Organization %s not found", orgID)
			http.Error(w, "Organization not found", http.StatusNotFound)
		} else {
			logError("Failed to fetch organization %s: %v", orgID, err)
			http.Error(w, "Failed to fetch organization", http.StatusInternalServerError)
		}
		return
	}

	if domainID.Valid {
		org.DomainID = &domainID.String
	}
	if parentOrgID.Valid {
		org.OrgID = &parentOrgID.String
	}
	if ownerID.Valid {
		org.OwnerID = &ownerID.String
	}

	if len(dataJSON) > 0 {
		json.Unmarshal(dataJSON, &org.Data)
	} else {
		org.Data = make(map[string]interface{})
	}

	members, err := s.getOrgMembers(orgID)
	if err != nil {
		logWarning("Error getting organization members: %v", err)
	} else {
		org.Members = members
		logInfo("Found %d members for organization %s", len(members), orgID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)

	logSuccess("Organization %s details sent successfully", orgID)
}

func (s *Server) updateOrganization(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing organization update request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized organization update: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	// Check if user is admin of the organization
	if !s.isOrgAdmin(session.Identity.Id, orgID) {
		logAuth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError("Invalid request body for organization update: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		logWarning("Organization update failed: name is required")
		http.Error(w, "Organization name is required", http.StatusBadRequest)
		return
	}

	if req.Data == nil {
		req.Data = make(map[string]interface{})
	}

	logInfo("Updating organization %s: '%s'", orgID, req.Name)

	dataJSON, _ := json.Marshal(req.Data)

	// Update organization in database
	result, err := s.db.Exec(`
		UPDATE organizations 
		SET name = $1, description = $2, org_type = $3, domain_id = $4, org_id = $5, data = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7`,
		req.Name, req.Description, req.OrgType, req.DomainID, req.OrgID, dataJSON, orgID,
	)
	if err != nil {
		logError("Failed to update organization in database: %v", err)
		http.Error(w, "Failed to update organization", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logWarning("Organization %s not found for update", orgID)
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	logDB("Organization %s updated successfully", orgID)

	// Get the updated organization
	var org Organization
	var dataJSONResult []byte
	var domainID, parentOrgID, ownerID sql.NullString

	err = s.db.QueryRow(`
		SELECT id, domain_id, org_id, org_type, name, description, owner_id, data, created_at, updated_at
		FROM organizations WHERE id = $1`,
		orgID,
	).Scan(&org.ID, &domainID, &parentOrgID, &org.OrgType, &org.Name, &org.Description,
		&ownerID, &dataJSONResult, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		logError("Failed to fetch updated organization: %v", err)
		http.Error(w, "Failed to fetch updated organization", http.StatusInternalServerError)
		return
	}

	if domainID.Valid {
		org.DomainID = &domainID.String
	}
	if parentOrgID.Valid {
		org.OrgID = &parentOrgID.String
	}
	if ownerID.Valid {
		org.OwnerID = &ownerID.String
	}

	if len(dataJSONResult) > 0 {
		json.Unmarshal(dataJSONResult, &org.Data)
	} else {
		org.Data = make(map[string]interface{})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(org)

	logSuccess("Organization %s updated successfully to '%s'", orgID, req.Name)
}

func (s *Server) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing organization deletion request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized organization deletion: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	// Check if user is the owner of the organization
	var ownerID sql.NullString
	err = s.db.QueryRow("SELECT owner_id FROM organizations WHERE id = $1", orgID).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			logWarning("Organization %s not found for deletion", orgID)
			http.Error(w, "Organization not found", http.StatusNotFound)
		} else {
			logError("Failed to check organization ownership: %v", err)
			http.Error(w, "Failed to check organization", http.StatusInternalServerError)
		}
		return
	}

	if !ownerID.Valid || ownerID.String != session.Identity.Id {
		logAuth("User %s not owner of organization %s (owner: %s)", session.Identity.Id, orgID, ownerID.String)
		http.Error(w, "Forbidden - Only organization owner can delete", http.StatusForbidden)
		return
	}

	logInfo("Deleting organization %s and all members", orgID)

	// Start transaction for atomic deletion
	tx, err := s.db.Begin()
	if err != nil {
		logError("Failed to start transaction: %v", err)
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Delete all organization members first
	_, err = tx.Exec("DELETE FROM user_organization_links WHERE organization_id = $1", orgID)
	if err != nil {
		logError("Failed to delete organization members: %v", err)
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	// Delete the organization
	result, err := tx.Exec("DELETE FROM organizations WHERE id = $1", orgID)
	if err != nil {
		logError("Failed to delete organization: %v", err)
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logWarning("Organization %s not found for deletion", orgID)
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		logError("Failed to commit deletion transaction: %v", err)
		http.Error(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	logDB("Organization %s and all members deleted successfully", orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Organization deleted successfully"})

	logSuccess("Organization %s deleted successfully", orgID)
}

// Organization Member Management Endpoints

func (s *Server) addMember(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized add member: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !s.isOrgAdmin(session.Identity.Id, orgID) {
		logAuth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError("Invalid request body for add member: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	logInfo("Adding member %s with role %s to organization %s", req.Email, req.Role, orgID)

	identities, _, err := s.kratosAdmin.IdentityApi.ListIdentities(context.Background()).Execute()
	if err != nil {
		logError("Failed to search users in Kratos: %v", err)
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
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
		logWarning("User not found: %s", req.Email)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	logInfo("Found user %s for email %s", targetUserID, req.Email)

	_, err = s.db.Exec(`
		INSERT INTO user_organization_links (user_id, organization_id, role) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id, organization_id) 
		DO UPDATE SET role = $3, joined_at = CURRENT_TIMESTAMP`,
		targetUserID, orgID, req.Role,
	)
	if err != nil {
		logError("Failed to add member to database: %v", err)
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	logDB("Member %s added to organization %s with role %s", req.Email, orgID, req.Role)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Member added successfully"})

	logSuccess("Member %s added successfully to organization %s", req.Email, orgID)
}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request) {
	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized get members: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]

	if !s.isOrgMember(session.Identity.Id, orgID) {
		logAuth("User %s not member of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	logInfo("Getting members for organization %s", orgID)

	members, err := s.getOrgMembers(orgID)
	if err != nil {
		logError("Failed to fetch members: %v", err)
		http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
		return
	}

	logInfo("Found %d members for organization %s", len(members), orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)

	logSuccess("Members list sent for organization %s", orgID)
}

func (s *Server) removeMember(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing remove member request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized remove member: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]
	userID := vars["userId"]

	if userID == "" {
		logWarning("User ID is required for member removal")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if requesting user is admin of the organization
	if !s.isOrgAdmin(session.Identity.Id, orgID) {
		logAuth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	// Check if target user is the organization owner
	var ownerID sql.NullString
	err = s.db.QueryRow("SELECT owner_id FROM organizations WHERE id = $1", orgID).Scan(&ownerID)
	if err != nil {
		logError("Failed to check organization ownership: %v", err)
		http.Error(w, "Failed to check organization", http.StatusInternalServerError)
		return
	}

	if ownerID.Valid && userID == ownerID.String {
		logWarning("Cannot remove organization owner %s from organization %s", userID, orgID)
		http.Error(w, "Cannot remove organization owner", http.StatusBadRequest)
		return
	}

	logInfo("Removing user %s from organization %s", userID, orgID)

	// Remove the member
	result, err := s.db.Exec(`
		DELETE FROM user_organization_links 
		WHERE organization_id = $1 AND user_id = $2`,
		orgID, userID,
	)
	if err != nil {
		logError("Failed to remove member from database: %v", err)
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logWarning("Member %s not found in organization %s", userID, orgID)
		http.Error(w, "Member not found in organization", http.StatusNotFound)
		return
	}

	logDB("Member %s removed from organization %s", userID, orgID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})

	logSuccess("Member %s removed successfully from organization %s", userID, orgID)
}

func (s *Server) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing update member role request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("Unauthorized update member role: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	orgID := vars["id"]
	userID := vars["userId"]

	if userID == "" {
		logWarning("User ID is required for role update")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if requesting user is admin of the organization
	if !s.isOrgAdmin(session.Identity.Id, orgID) {
		logAuth("User %s not admin of organization %s", session.Identity.Id, orgID)
		http.Error(w, "Forbidden - Admin access required", http.StatusForbidden)
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logError("Invalid request body for role update: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		logWarning("Role is required for member role update")
		http.Error(w, "Role is required", http.StatusBadRequest)
		return
	}

	// Validate role
	validRoles := map[string]bool{"member": true, "admin": true}
	if !validRoles[req.Role] {
		logWarning("Invalid role: %s", req.Role)
		http.Error(w, "Invalid role. Must be 'member' or 'admin'", http.StatusBadRequest)
		return
	}

	// Check if target user is the organization owner
	var ownerID sql.NullString
	err = s.db.QueryRow("SELECT owner_id FROM organizations WHERE id = $1", orgID).Scan(&ownerID)
	if err != nil {
		logError("Failed to check organization ownership: %v", err)
		http.Error(w, "Failed to check organization", http.StatusInternalServerError)
		return
	}

	if ownerID.Valid && userID == ownerID.String {
		logWarning("Cannot change role of organization owner %s", userID)
		http.Error(w, "Cannot change organization owner's role", http.StatusBadRequest)
		return
	}

	logInfo("Updating role of user %s in organization %s to %s", userID, orgID, req.Role)

	// Update the member's role
	result, err := s.db.Exec(`
		UPDATE user_organization_links 
		SET role = $1 
		WHERE organization_id = $2 AND user_id = $3`,
		req.Role, orgID, userID,
	)
	if err != nil {
		logError("Failed to update member role in database: %v", err)
		http.Error(w, "Failed to update member role", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logWarning("Member %s not found in organization %s", userID, orgID)
		http.Error(w, "Member not found in organization", http.StatusNotFound)
		return
	}

	logDB("Member %s role updated to %s in organization %s", userID, req.Role, orgID)

	// Get updated member information
	var member Member
	err = s.db.QueryRow(`
		SELECT uol.user_id, uol.role, uol.joined_at, u.email, u.first_name, u.last_name
		FROM user_organization_links uol
		LEFT JOIN users u ON uol.user_id = u.id
		WHERE uol.organization_id = $1 AND uol.user_id = $2
	`, orgID, userID).Scan(&member.UserID, &member.Role, &member.JoinedAt, &member.Email, &member.FirstName, &member.LastName)

	if err != nil {
		logError("Failed to fetch updated member info: %v", err)
		http.Error(w, "Failed to fetch updated member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)

	logSuccess("Member %s role updated successfully to %s in organization %s", userID, req.Role, orgID)
}

// Helper Functions

func (s *Server) getOrgMembers(orgID string) ([]Member, error) {
	rows, err := s.db.Query(`
		SELECT uol.user_id, uol.role, uol.joined_at, u.email, u.first_name, u.last_name
		FROM user_organization_links uol
		LEFT JOIN users u ON uol.user_id = u.id
		WHERE uol.organization_id = $1
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var member Member
		var email, firstName, lastName sql.NullString
		err := rows.Scan(&member.UserID, &member.Role, &member.JoinedAt, &email, &firstName, &lastName)
		if err != nil {
			logWarning("Error scanning member row: %v", err)
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

func (s *Server) getUserOrganizations(userID string) ([]OrgMember, error) {
	rows, err := s.db.Query(`
		SELECT o.id, o.name, o.org_type, uol.role, uol.joined_at
		FROM organizations o
		JOIN user_organization_links uol ON o.id = uol.organization_id
		WHERE uol.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []OrgMember
	for rows.Next() {
		var org OrgMember
		err := rows.Scan(&org.OrgID, &org.OrgName, &org.OrgType, &org.Role, &org.JoinedAt)
		if err != nil {
			logWarning("Error scanning organization row: %v", err)
			continue
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

func (s *Server) getUserFromDB(userID string) (*User, error) {
	var user User
	var lastLogin sql.NullTime

	err := s.db.QueryRow(`
		SELECT id, email, first_name, last_name, time_zone, ui_mode, created_at, updated_at, last_login
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.TimeZone,
		&user.UIMode, &user.CreatedAt, &user.UpdatedAt, &lastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found in database
		}
		return nil, err
	}

	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	return &user, nil
}

func (s *Server) isOrgMember(userID string, orgID string) bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM user_organization_links 
		WHERE user_id = $1 AND organization_id = $2`,
		userID, orgID,
	).Scan(&count)
	return err == nil && count > 0
}

func (s *Server) isOrgAdmin(userID string, orgID string) bool {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM user_organization_links 
		WHERE user_id = $1 AND organization_id = $2 AND role IN ('admin')`,
		userID, orgID,
	).Scan(&count)

	if err == nil && count > 0 {
		return true
	}

	// Also check if user is the owner
	var ownerID sql.NullString
	err = s.db.QueryRow("SELECT owner_id FROM organizations WHERE id = $1", orgID).Scan(&ownerID)
	return err == nil && ownerID.Valid && ownerID.String == userID
}

func (s *Server) saveUserProfile(identity client.Identity) {
	user := s.mapIdentityToUser(identity)

	logDB("Saving user profile for: %s", user.Email)

	_, err := s.db.Exec(`
		INSERT INTO users (id, email, first_name, last_name, last_login)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (id) 
		DO UPDATE SET 
			email = $2,
			first_name = $3,
			last_name = $4,
			last_login = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`, user.ID, user.Email, user.FirstName, user.LastName)

	if err != nil {
		logError("Error saving user profile: %v", err)
	} else {
		logDB("User profile saved successfully for: %s", user.Email)
	}
}

// Webhook Handlers

func (s *Server) handleAfterRegistration(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing registration webhook")

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logError("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logSuccess("New user registered: %s (%s)", payload.Identity.Id, s.getEmailFromIdentity(payload.Identity))

	s.saveUserProfile(payload.Identity)

	w.WriteHeader(http.StatusOK)
	logInfo("Registration webhook processed successfully")
}

func (s *Server) handleAfterLogin(w http.ResponseWriter, r *http.Request) {
	logInfo("Processing login webhook")

	var payload WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		logError("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	logSuccess("User logged in: %s (%s)", payload.Identity.Id, s.getEmailFromIdentity(payload.Identity))

	s.saveUserProfile(payload.Identity)

	w.WriteHeader(http.StatusOK)
	logInfo("Login webhook processed successfully")
}

func (s *Server) getEmailFromIdentity(identity client.Identity) string {
	if traits, ok := identity.Traits.(map[string]interface{}); ok {
		if email, exists := traits["email"].(string); exists {
			return email
		}
	}
	return "unknown"
}

// System Endpoints

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	logAuth("Processing get session request")

	session, err := s.getSessionFromRequest(r)
	if err != nil {
		logAuth("No valid session found: %v", err)
		http.Error(w, "No valid session", http.StatusUnauthorized)
		return
	}

	logAuth("Session retrieved for user: %s", session.Identity.Id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	logAuth("Processing logout request")

	var sessionToken string

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		sessionToken = strings.TrimPrefix(authHeader, "Bearer ")
		logAuth("Found session token in Authorization header")
	} else {
		sessionCookie, err := r.Cookie("ory_kratos_session")
		if err != nil {
			logWarning("No session found for logout")
			http.Error(w, "No session found", http.StatusBadRequest)
			return
		}
		sessionToken = sessionCookie.Value
		logAuth("Found session token in cookie")
	}

	// First, get the session details to extract the session ID
	logAuth("Getting session details to extract session ID")
	session, resp, err := s.kratosPublic.FrontendApi.ToSession(context.Background()).
		XSessionToken(sessionToken).
		Execute()

	if err != nil || resp.StatusCode != 200 {
		logWarning("Could not get session details for logout: %v (status: %d)", err, resp.StatusCode)
		// Session might already be invalid, continue with clearing cookie
	} else {
		logAuth("Found session ID: %s", session.Id)

		// Use the session ID (not token) to disable the session
		_, err = s.kratosAdmin.IdentityApi.DisableSession(context.Background(), session.Id).Execute()
		if err != nil {
			logWarning("Error revoking session with ID %s: %v", session.Id, err)
		} else {
			logSuccess("Session %s revoked successfully", session.Id)
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

	logSuccess("Logout completed successfully")
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	logInfo("Health check requested")

	// Check database connectivity
	if err := s.db.Ping(); err != nil {
		logError("Database health check failed: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"database": "connected",
	})

	logSuccess("Health check: OK")
}

func main() {
	fmt.Printf("%s%s", ColorBold, ColorGreen)
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    🚀 User Management System 🚀     ║")
	fmt.Println("║           Starting Server           ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Printf("%s", ColorReset)

	server := NewServer()
	router := server.setupRoutes()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:3000",
			"http://localhost:3001", // Frontend development server
			"http://localhost:8080",
			"http://172.16.1.65:3000",
			"http://172.16.1.65:3001",
			"http://172.16.1.65:8080",
			"http://172.16.1.66:3000",
			"http://172.16.1.66:3001",
			"http://172.16.1.66:8080",
			"file://", // For local HTML files
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Cookie"}),
		handlers.AllowCredentials(),
	)(router)

	port := getEnv("PORT", "3000")

	logInfo("Server configuration:")
	logInfo("  Port: %s", port)
	logInfo("  Kratos Public URL: %s", getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433"))
	logInfo("  Kratos Admin URL: %s", getEnv("KRATOS_ADMIN_URL", "http://localhost:4434"))
	logInfo("  Database URL: %s", strings.ReplaceAll(getEnv("DATABASE_URL", "postgres://userms:userms_password@localhost:5432/userms?sslmode=disable"), "userms_password", "***"))

	fmt.Printf("\n%s%s🌟 Server ready! Listening on:%s http://localhost:%s %s\n\n",
		ColorBold, ColorGreen, ColorReset, port, ColorGreen)
	fmt.Printf("%sEndpoints available:%s\n", ColorCyan, ColorReset)
	fmt.Printf("  📊 Health: http://localhost:%s/health\n", port)
	fmt.Printf("  👤 Users:  http://localhost:%s/api/users\n", port)
	fmt.Printf("  🏢 Orgs:   http://localhost:%s/api/organizations\n", port)
	fmt.Printf("  🔐 Auth:   Bearer token or Cookie authentication\n")
	fmt.Printf("  🔍 Debug:  http://localhost:%s/api/debug/auth\n", port)
	fmt.Printf("%s\n", ColorReset)

	logSuccess("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}
