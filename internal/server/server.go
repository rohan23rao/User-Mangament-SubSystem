package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	client "github.com/ory/kratos-client-go"

	"userms/internal/auth"
	"userms/internal/config"
	"userms/internal/database"
	handlersPackage "userms/internal/handlers"
	"userms/internal/logger"
	"userms/internal/middleware"
)

type Server struct {
	config         *config.Config
	authService    *auth.Service
	userHandler    *handlersPackage.UserHandler
	orgHandler     *handlersPackage.OrganizationHandler
	healthHandler  *handlersPackage.HealthHandler
	webhookHandler *handlersPackage.WebhookHandler
}

func New(cfg *config.Config) *Server {
	logger.Info("Initializing server with Kratos URLs:")
	logger.Info("  Public: %s", cfg.KratosPublicURL)
	logger.Info("  Admin: %s", cfg.KratosAdminURL)

	// Initialize Kratos clients
	publicConfig := client.NewConfiguration()
	publicConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosPublicURL}}

	adminConfig := client.NewConfiguration()
	adminConfig.Servers = []client.ServerConfiguration{{URL: cfg.KratosAdminURL}}

	kratosPublic := client.NewAPIClient(publicConfig)
	kratosAdmin := client.NewAPIClient(adminConfig)

	// Initialize database
	logger.Info("Initializing database...")
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to initialize database: %v", err)
		log.Fatal("Database initialization failed")
	}
	logger.Success("Database initialized successfully")

	// Initialize services and handlers
	authService := auth.NewService(kratosPublic)
	userHandler := handlersPackage.NewUserHandler(authService, kratosAdmin, db)
	orgHandler := handlersPackage.NewOrganizationHandler(authService, kratosAdmin, db)
	healthHandler := handlersPackage.NewHealthHandler(db)
	webhookHandler := handlersPackage.NewWebhookHandler(userHandler)

	return &Server{
		config:         cfg,
		authService:    authService,
		userHandler:    userHandler,
		orgHandler:     orgHandler,
		healthHandler:  healthHandler,
		webhookHandler: webhookHandler,
	}
}

func (s *Server) setupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware(s.authService))

	// Health endpoint
	r.HandleFunc("/health", s.healthHandler.HealthCheck).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()

	// User endpoints
	api.HandleFunc("/whoami", s.userHandler.WhoAmI).Methods("GET")
	api.HandleFunc("/users", s.userHandler.ListUsers).Methods("GET")
	api.HandleFunc("/users/{id}", s.userHandler.GetUser).Methods("GET")

	// Organization endpoints
	api.HandleFunc("/organizations", s.orgHandler.CreateOrganization).Methods("POST")
	api.HandleFunc("/organizations", s.orgHandler.ListOrganizations).Methods("GET")
	api.HandleFunc("/organizations/{id}", s.orgHandler.GetOrganization).Methods("GET")
	api.HandleFunc("/organizations/{id}", s.orgHandler.UpdateOrganization).Methods("PUT")
	api.HandleFunc("/organizations/{id}", s.orgHandler.DeleteOrganization).Methods("DELETE")

	// Organization member endpoints
	api.HandleFunc("/organizations/{id}/members", s.orgHandler.AddMember).Methods("POST")
	api.HandleFunc("/organizations/{id}/members", s.orgHandler.GetMembers).Methods("GET")
	api.HandleFunc("/organizations/{id}/members/{user_id}", s.orgHandler.UpdateMemberRole).Methods("PUT")

	// Webhook endpoints
	hooks := r.PathPrefix("/hooks").Subrouter()
	hooks.HandleFunc("/after-registration", s.webhookHandler.HandleAfterRegistration).Methods("POST")
	hooks.HandleFunc("/after-login", s.webhookHandler.HandleAfterLogin).Methods("POST")

	return r
}

func (s *Server) Start() error {
	fmt.Printf("%s%s", logger.ColorBold, logger.ColorGreen)
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    🚀 User Management System 🚀     ║")
	fmt.Println("║           Starting Server           ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Printf("%s", logger.ColorReset)

	router := s.setupRoutes()

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

	logger.Info("Server configuration:")
	logger.Info("  Port: %s", s.config.Port)
	logger.Info("  Kratos Public URL: %s", s.config.KratosPublicURL)
	logger.Info("  Kratos Admin URL: %s", s.config.KratosAdminURL)
	logger.Info("  Database URL: %s", strings.ReplaceAll(s.config.DatabaseURL, "userms_password", "***"))

	fmt.Printf("\n%s%s🌟 Server ready! Listening on:%s http://localhost:%s %s\n\n",
		logger.ColorBold, logger.ColorGreen, logger.ColorReset, s.config.Port, logger.ColorGreen)
	fmt.Printf("%sEndpoints available:%s\n", logger.ColorCyan, logger.ColorReset)
	fmt.Printf("  📊 Health: http://localhost:%s/health\n", s.config.Port)
	fmt.Printf("  👤 Users:  http://localhost:%s/api/users\n", s.config.Port)
	fmt.Printf("  🏢 Orgs:   http://localhost:%s/api/organizations\n", s.config.Port)
	fmt.Printf("  🔐 Auth:   Bearer token or Cookie authentication\n")
	fmt.Printf("  🔍 Debug:  http://localhost:%s/api/debug/auth\n", s.config.Port)
	fmt.Printf("%s\n", logger.ColorReset)

	logger.Success("Server starting on port %s", s.config.Port)
	return http.ListenAndServe(":"+s.config.Port, corsHandler)
}