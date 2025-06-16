package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"userms/internal/auth"
	"userms/internal/logger"
)

// LoggingMiddleware provides logging for HTTP requests (updated for net/http)
func LoggingMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			session, _ := authService.GetSessionFromRequest(r)
			userID := "anonymous"
			if session != nil {
				userID = session.Identity.Id[:8] + "..."
			}

			logger.Request(r.Method, r.URL.Path, userID)

			// Create response wrapper to capture status code
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			statusColor := logger.ColorGreen
			if wrapper.statusCode >= 400 {
				statusColor = logger.ColorRed
			} else if wrapper.statusCode >= 300 {
				statusColor = logger.ColorYellow
			}

			// Log with structured logging
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("user", userID).
				Int("status", wrapper.statusCode).
				Dur("duration", duration).
				Msg("Request completed")

			// Keep the old logger format for compatibility
			logger.Info("[RESPONSE] %s%d%s | %s | %v",
				statusColor, wrapper.statusCode, logger.ColorReset, r.URL.Path, duration)
		})
	}
}

// responseWrapper wraps ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWrapper) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWrapper) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(data)
}