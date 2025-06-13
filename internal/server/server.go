package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	client "github.com/ory/kratos-client-go"
	hydra "github.com/ory/hydra-client-go/v2"

	"userms/internal/auth"
	"userms/internal/config"
	"userms/internal/database"
	handlersPackage "userms/internal/handlers"
	"userms/internal/logger"
	"userms/internal/middleware"
	"userms/internal/oauth2"
)

type Server struct {
	config                *config.Config
	authService           *auth.Service
	oauth2Service         *oauth2.Service
	userHandler           *handlersPackage.UserHandler
	orgHandler            *handlersPackage.OrganizationHandler
	oauth2Handler         *handlersPackage.OAuth2Handler
	healthHandler         *handlersPackage.HealthHandler
	webhookHandler        *handlersPackage.WebhookHandler
	verificationHandler   *handlersPackage.VerificationHandler
}

func New(cfg *config.Config) *Server {
	logger.Info("Initializing server with Kratos and Hydra URLs:")
	logger.Info("  Kratos Public: %s", cfg.KratosPublicURL)
	logger.Info("  Kratos Admin: %s", cfg.KratosAdminURL)
	logger.Info("  Hydra Public: %s", cfg.HydraPublicURL)
	logger.Info("  Hydra Admin: %s", cfg.HydraAdminURL)

	// Initialize Kratos clients
	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosPublicURL}}

	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosAdminURL}}

	kratosPublic := client.NewAPIClient(publicConfig)
	kratosAdmin := client.NewAPIClient(adminConfig)

	// Initialize Hydra clients
	hydraAdminConfig := hydra.NewConfiguration()
	hydraAdminConfig.Servers = []hydra.ServerConfiguration{{URL: cfg.HydraAdminURL}}
	hydraAdmin := hydra.NewAPIClient(hydraAdminConfig)

	// Initialize database
	logger.Info("Initializing database...")
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		log.Fatal("Database initialization failed")
	}
	logger.Success("Database initialized successfully")

	// Initialize services
	authService := auth.NewService(kratosPublic)
	oauth2Service := oauth2.NewService(hydraAdmin, db)

	// Initialize handlers
	userHandler := handlersPackage.NewUserHandler(authService, kratosAdmin, db)
	orgHandler := handlersPackage.NewOrganizationHandler(authService, kratosAdmin, db)
	oauth2Handler := handlersPackage.NewOAuth2Handler(authService, oauth2Service)
	healthHandler := handlersPackage.NewHealthHandler(db)
	webhookHandler := handlersPackage.NewWebhookHandler(userHandler)
	verificationHandler := handlersPackage.NewVerificationHandler(authService, kratosAdmin)

	return &Server{
		config:                cfg,
		authService:           authService,
		oauth2Service:         oauth2Service,
		userHandler:           userHandler,
		orgHandler:            orgHandler,
		oauth2Handler:         oauth2Handler,
		healthHandler:         healthHandler,
		webhookHandler:        webhookHandler,
		verificationHandler:   verificationHandler,
	}
}

func (s *Server) setupRoutes() http.Handler {
	r := mux.NewRouter()

	// Add logging middleware
	r.Use(middleware.LoggingMiddleware(s.authService))

	// Health check endpoint
	r.HandleFunc("/health", s.healthHandler.HealthCheck).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// User endpoints - CORRECTED: Using the actual method names from your UserHandler
	api.HandleFunc("/whoami", s.userHandler.WhoAmI).Methods("GET")
	api.HandleFunc("/users", s.userHandler.ListUsers).Methods("GET")  // FIXED: Was GetUsers, now ListUsers
	api.HandleFunc("/users/{id}", s.userHandler.GetUser).Methods("GET")
	// REMOVED: DebugAuth method doesn't exist in your UserHandler

	// Organization endpoints - CORRECTED: Using the actual method names from your OrganizationHandler
	api.HandleFunc("/organizations", s.orgHandler.ListOrganizations).Methods("GET")  // FIXED: Was GetOrganizations, now ListOrganizations
	api.HandleFunc("/organizations", s.orgHandler.CreateOrganization).Methods("POST")
	api.HandleFunc("/organizations/{id}", s.orgHandler.GetOrganization).Methods("GET")
	api.HandleFunc("/organizations/{id}", s.orgHandler.UpdateOrganization).Methods("PUT")
	api.HandleFunc("/organizations/{id}", s.orgHandler.DeleteOrganization).Methods("DELETE")

	// Organization member endpoints - CORRECTED: Using the actual method names
	api.HandleFunc("/organizations/{id}/members", s.orgHandler.GetMembers).Methods("GET")  // FIXED: Was GetOrganizationMembers, now GetMembers
	api.HandleFunc("/organizations/{id}/members", s.orgHandler.AddMember).Methods("POST")  // FIXED: Was InviteUser, now AddMember
	api.HandleFunc("/organizations/{id}/members/{user_id}", s.orgHandler.UpdateMemberRole).Methods("PUT")
	api.HandleFunc("/organizations/{id}/members/{user_id}", s.orgHandler.RemoveMember).Methods("DELETE")
	// REMOVED: RemoveMember method doesn't exist in your OrganizationHandler

	// NEW: OAuth2 M2M endpoints
	oauth2Routes := api.PathPrefix("/oauth2").Subrouter()
	oauth2Routes.HandleFunc("/clients", s.oauth2Handler.CreateM2MClient).Methods("POST")
	oauth2Routes.HandleFunc("/clients", s.oauth2Handler.ListM2MClients).Methods("GET")
	oauth2Routes.HandleFunc("/clients/{clientId}", s.oauth2Handler.GetM2MClientInfo).Methods("GET")
	oauth2Routes.HandleFunc("/clients/{clientId}", s.oauth2Handler.RevokeM2MClient).Methods("DELETE")
	oauth2Routes.HandleFunc("/clients/{clientId}/regenerate", s.oauth2Handler.RegenerateM2MClientSecret).Methods("POST")
	
	// Token endpoints (public endpoints for M2M authentication)
	api.HandleFunc("/oauth2/token", s.oauth2Handler.GenerateM2MToken).Methods("POST")
	api.HandleFunc("/oauth2/validate", s.oauth2Handler.ValidateM2MToken).Methods("POST")

	// Verification endpoints
	api.HandleFunc("/users/{id}/verification/status", s.verificationHandler.GetVerificationStatus).Methods("GET")
	api.HandleFunc("/verification/flow", s.verificationHandler.CreateVerificationFlow).Methods("GET")

	// Webhook endpoints
	hooks := r.PathPrefix("/hooks").Subrouter()
	hooks.HandleFunc("/after-registration", s.webhookHandler.HandleAfterRegistration).Methods("POST")
	hooks.HandleFunc("/after-login", s.webhookHandler.HandleAfterLogin).Methods("POST")
	hooks.HandleFunc("/after-verification", s.webhookHandler.HandleAfterVerification).Methods("POST")

	return r
}

func (s *Server) Start() error {
	fmt.Printf("%s%s", logger.ColorBold, logger.ColorGreen)
	fmt.Println("╔═══════════════════════════════════════════════════╗")
	fmt.Println("║    🚀 User Management System with OAuth2 🚀      ║")
	fmt.Println("║              Starting Server                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════╝")
	fmt.Printf("%s", logger.ColorReset)

	router := s.setupRoutes()

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://172.16.1.65:3000",
			"http://172.16.1.65:3001",
			"http://172.16.1.65:8080",
			"http://172.16.1.66:3000",
			"http://172.16.1.66:3001",
			"http://172.16.1.66:8080",
			"file://",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Cookie"}),
		handlers.AllowCredentials(),
	)(router)

	logger.Info("Server configuration:")
	logger.Info("  Port: %s", s.config.Port)
	logger.Info("  Kratos Public URL: %s", s.config.KratosPublicURL)
	logger.Info("  Kratos Admin URL: %s", s.config.KratosAdminURL)
	logger.Info("  Hydra Public URL: %s", s.config.HydraPublicURL)
	logger.Info("  Hydra Admin URL: %s", s.config.HydraAdminURL)
	logger.Info("  Database URL: %s", strings.ReplaceAll(s.config.DatabaseURL, "userms_password", "***"))

	fmt.Printf("\n%s%s🌟 Server ready! Listening on:%s http://localhost:%s %s\n\n",
		logger.ColorBold, logger.ColorGreen, logger.ColorReset, s.config.Port, logger.ColorGreen)
	fmt.Printf("%sEndpoints available:%s\n", logger.ColorCyan, logger.ColorReset)
	fmt.Printf("  📊 Health: http://localhost:%s/health\n", s.config.Port)
	fmt.Printf("  👤 Users:  http://localhost:%s/api/users\n", s.config.Port)
	fmt.Printf("  🏢 Orgs:   http://localhost:%s/api/organizations\n", s.config.Port)
	fmt.Printf("  🔐 Auth:   Bearer token or Cookie authentication\n")
	fmt.Printf("  🔑 OAuth2: http://localhost:%s/api/oauth2/clients\n", s.config.Port)
	fmt.Printf("  🎫 Token:  http://localhost:%s/api/oauth2/token\n", s.config.Port)
	fmt.Printf("  ✅ Validate: http://localhost:%s/api/oauth2/validate\n", s.config.Port)
	fmt.Printf("  🎣 Hooks:  http://localhost:%s/hooks/*\n", s.config.Port)
	fmt.Printf("%s\n", logger.ColorReset)

	logger.Success("Server starting on port %s", s.config.Port)
	return http.ListenAndServe(":"+s.config.Port, corsHandler)
}