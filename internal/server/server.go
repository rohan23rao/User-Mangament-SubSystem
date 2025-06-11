// internal/server/server.go
package server

import (
	"database/sql"
	"net/http"
	"userms/internal/auth"
	"userms/internal/config"
	"userms/internal/handlers"
	"userms/internal/middleware"
	"userms/internal/utils"

	"github.com/gorilla/mux"
	gorillaHandlers "github.com/gorilla/handlers"
	client "github.com/ory/kratos-client-go"
)

type Server struct {
	config               *config.Config
	db                   *sql.DB
	kratosPublic         *client.APIClient
	kratosAdmin          *client.APIClient
	sessionManager       *auth.SessionManager
	verificationService  *auth.VerificationService
	webhookHandler       *handlers.WebhookHandler
	userHandler          *handlers.UserHandler
	organizationHandler  *handlers.OrganizationHandler
	authHandler          *handlers.AuthHandler
	healthHandler        *handlers.HealthHandler
}

func New(cfg *config.Config, db *sql.DB) *Server {
	utils.LogInfo("Initializing server with Kratos URLs:")
	utils.LogInfo("  Public: %s", cfg.KratosPublicURL)
	utils.LogInfo("  Admin: %s", cfg.KratosAdminURL)

	// Initialize Kratos clients
	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosPublicURL}}

	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosAdminURL}}

	kratosPublic := client.NewAPIClient(publicConfig)
	kratosAdmin := client.NewAPIClient(adminConfig)

	// Initialize services
	sessionManager := auth.NewSessionManager(kratosPublic)
	verificationService := auth.NewVerificationService(db)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(db)
	userHandler := handlers.NewUserHandler(db, kratosAdmin, sessionManager, verificationService)
	organizationHandler := handlers.NewOrganizationHandler(db, sessionManager, verificationService)
	authHandler := handlers.NewAuthHandler(kratosPublic, kratosAdmin, sessionManager)
	healthHandler := handlers.NewHealthHandler(db)

	return &Server{
		config:               cfg,
		db:                   db,
		kratosPublic:         kratosPublic,
		kratosAdmin:          kratosAdmin,
		sessionManager:       sessionManager,
		verificationService:  verificationService,
		webhookHandler:       webhookHandler,
		userHandler:          userHandler,
		organizationHandler:  organizationHandler,
		authHandler:          authHandler,
		healthHandler:        healthHandler,
	}
}

func (s *Server) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	
	// Global middleware
	r.Use(middleware.Logging(s.sessionManager))

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// User endpoints
	api.HandleFunc("/whoami", s.userHandler.WhoAmI).Methods("GET")
	api.HandleFunc("/users", s.userHandler.ListUsers).Methods("GET")
	api.HandleFunc("/users/{id}", s.userHandler.GetUser).Methods("GET")

	// Organization endpoints (protected by verification)
	orgRouter := api.PathPrefix("/organizations").Subrouter()
	orgRouter.Use(middleware.RequireVerifiedUser(s.sessionManager, s.verificationService))
	orgRouter.HandleFunc("", s.organizationHandler.CreateOrganization).Methods("POST")
	orgRouter.HandleFunc("", s.organizationHandler.ListOrganizations).Methods("GET")
	orgRouter.HandleFunc("/{id}", s.organizationHandler.GetOrganization).Methods("GET")
	orgRouter.HandleFunc("/{id}", s.organizationHandler.UpdateOrganization).Methods("PUT")
	orgRouter.HandleFunc("/{id}", s.organizationHandler.DeleteOrganization).Methods("DELETE")

	// Organization member endpoints (protected by verification)
	orgRouter.HandleFunc("/{id}/members", s.organizationHandler.AddMember).Methods("POST")
	orgRouter.HandleFunc("/{id}/members", s.organizationHandler.GetMembers).Methods("GET")
	orgRouter.HandleFunc("/{id}/members/{userId}", s.organizationHandler.RemoveMember).Methods("DELETE")
	orgRouter.HandleFunc("/{id}/members/{userId}/role", s.organizationHandler.UpdateMemberRole).Methods("PUT")

	// Debug endpoint
	api.HandleFunc("/debug/auth", s.authHandler.DebugAuth).Methods("GET")

	// Webhook endpoints
	hooks := r.PathPrefix("/hooks").Subrouter()
	hooks.HandleFunc("/after-registration", s.webhookHandler.HandleAfterRegistration).Methods("POST")
	hooks.HandleFunc("/after-login", s.webhookHandler.HandleAfterLogin).Methods("POST")

	// System endpoints
	r.HandleFunc("/health", s.healthHandler.HealthCheck).Methods("GET")
	r.HandleFunc("/auth/session", s.authHandler.GetSession).Methods("GET")
	r.HandleFunc("/auth/logout", s.authHandler.Logout).Methods("POST")

	utils.LogInfo("Routes configured successfully")
	return r
}

func (s *Server) Start() error {
	router := s.setupRoutes()

	corsHandler := gorillaHandlers.CORS(
		gorillaHandlers.AllowedOrigins(s.config.AllowedOrigins),
		gorillaHandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Cookie"}),
		gorillaHandlers.AllowCredentials(),
	)(router)

	return http.ListenAndServe(":"+s.config.Port, corsHandler)
}
