package auth

import (
	"context"
	"net/http"
	"strings"

	client "github.com/ory/kratos-client-go"
	"userms/internal/logger"
)

type Service struct {
	kratosPublic *client.APIClient
}

func NewService(kratosPublic *client.APIClient) *Service {
	return &Service{
		kratosPublic: kratosPublic,
	}
}

func (s *Service) GetSessionFromRequest(r *http.Request) (*client.Session, error) {
	// Try cookie first
	cookie, err := r.Cookie("ory_kratos_session")
	if err == nil && cookie.Value != "" {
		logger.Auth("Attempting authentication with session cookie")
		session, resp, err := s.kratosPublic.FrontendApi.ToSession(context.Background()).
			Cookie(cookie.String()).
			Execute()
		if err != nil {
			logger.Auth("Cookie authentication failed: %v", err)
		} else {
			logger.Auth("Cookie authentication successful for user: %s", session.Identity.Id)
			return session, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			sessionToken := strings.TrimPrefix(authHeader, "Bearer ")
			logger.Auth("Attempting authentication with bearer token")
			
			session, resp, err := s.kratosPublic.FrontendApi.ToSession(context.Background()).
				XSessionToken(sessionToken).
				Execute()
			if err != nil {
				logger.Auth("Bearer token authentication failed: %v", err)
			} else {
				logger.Auth("Bearer token authentication successful for user: %s", session.Identity.Id)
				return session, nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	return nil, ErrUnauthorized
}

var ErrUnauthorized = &AuthError{Message: "unauthorized"}

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}