package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	mux                   *http.ServeMux
	server                *http.Server
}

func New(cfg *config.Config) *Server {
	// Setup structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

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
		log.Fatal().Err(err).Msg("Database initialization failed")
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

	// Create ServeMux
	mux := http.NewServeMux()

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
		mux:                   mux,
	}
}

func (s *Server) setupRoutes() http.Handler {
	// Health check endpoint
	s.mux.HandleFunc("GET /health", s.healthHandler.HealthCheck)

	// API routes with authentication middleware
	s.mux.HandleFunc("GET /api/whoami", s.withAuth(s.userHandler.WhoAmI))
	s.mux.HandleFunc("GET /api/users", s.withAuth(s.userHandler.ListUsers))
	s.mux.HandleFunc("GET /api/users/{id}", s.withAuth(s.userHandler.GetUser))
	s.mux.HandleFunc("GET /api/debug/auth", s.userHandler.DebugAuth) // No auth for debug

	// Organization endpoints
	s.mux.HandleFunc("GET /api/organizations", s.withAuth(s.orgHandler.ListOrganizations))
	s.mux.HandleFunc("POST /api/organizations", s.withAuth(s.orgHandler.CreateOrganization))
	s.mux.HandleFunc("GET /api/organizations/{id}", s.withAuth(s.orgHandler.GetOrganization))
	s.mux.HandleFunc("PUT /api/organizations/{id}", s.withAuth(s.orgHandler.UpdateOrganization))
	s.mux.HandleFunc("DELETE /api/organizations/{id}", s.withAuth(s.orgHandler.DeleteOrganization))

	// Organization member endpoints
	s.mux.HandleFunc("GET /api/organizations/{id}/members", s.withAuth(s.orgHandler.GetMembers))
	s.mux.HandleFunc("POST /api/organizations/{id}/members", s.withAuth(s.orgHandler.AddMember))
	s.mux.HandleFunc("PUT /api/organizations/{id}/members/{user_id}", s.withAuth(s.orgHandler.UpdateMemberRole))
	s.mux.HandleFunc("DELETE /api/organizations/{id}/members/{user_id}", s.withAuth(s.orgHandler.RemoveMember))
	s.mux.HandleFunc("GET /api/organizations/{id}/tenants", s.withAuth(s.orgHandler.GetOrganizationWithTenants))

	// OAuth2 M2M endpoints
	s.mux.HandleFunc("POST /api/oauth2/clients", s.withAuth(s.oauth2Handler.CreateM2MClient))
	s.mux.HandleFunc("GET /api/oauth2/clients", s.withAuth(s.oauth2Handler.ListM2MClients))
	s.mux.HandleFunc("GET /api/oauth2/clients/{clientId}", s.withAuth(s.oauth2Handler.GetM2MClientInfo))
	s.mux.HandleFunc("DELETE /api/oauth2/clients/{clientId}", s.withAuth(s.oauth2Handler.RevokeM2MClient))
	s.mux.HandleFunc("POST /api/oauth2/clients/{clientId}/regenerate", s.withAuth(s.oauth2Handler.RegenerateM2MClientSecret))

	// Token endpoints (public endpoints for M2M authentication)
	s.mux.HandleFunc("POST /api/oauth2/token", s.oauth2Handler.GenerateM2MToken)
	s.mux.HandleFunc("POST /api/oauth2/validate", s.oauth2Handler.ValidateM2MToken)

	// Verification endpoints
	s.mux.HandleFunc("GET /api/users/{id}/verification/status", s.withAuth(s.verificationHandler.GetVerificationStatus))
	s.mux.HandleFunc("GET /api/verification/flow", s.verificationHandler.CreateVerificationFlow)

	// Webhook endpoints
	s.mux.HandleFunc("POST /hooks/after-registration", s.webhookHandler.HandleAfterRegistration)
	s.mux.HandleFunc("POST /hooks/after-login", s.webhookHandler.HandleAfterLogin)
	s.mux.HandleFunc("POST /hooks/after-verification", s.webhookHandler.HandleAfterVerification)

	// Setup CORS
	corsOptions := cors.Options{
		AllowedOrigins: []string{
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
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "Cookie"},
		AllowCredentials: true,
	}

	corsHandler := cors.New(corsOptions)

	// Wrap with middlewares
	handler := middleware.LoggingMiddleware(s.authService)(corsHandler.Handler(s.mux))

	return handler
}

// withAuth wraps handlers with authentication middleware
func (s *Server) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := s.authService.GetSessionFromRequest(r)
		if err != nil {
			logger.Auth("Unauthorized request: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add session to request context if needed
		handler.ServeHTTP(w, r)
	}
}

func (s *Server) Start() error {
	fmt.Printf("%s%s", logger.ColorBold, logger.ColorGreen)
	fmt.Println("╔═══════════════════════════════════════════════════╗")
	fmt.Println("║    🚀 User Management System with OAuth2 🚀      ║")
	fmt.Println("║              Starting Server                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════╝")
	fmt.Printf("%s", logger.ColorReset)

	router := s.setupRoutes()

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

	s.server = &http.Server{
		Addr:              ":" + s.config.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return s.server.ListenAndServe()
}