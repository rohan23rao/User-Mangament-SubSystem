package middleware

import (
	"log"
	"net/http"
	"time"

	"userms/internal/auth"
	"userms/internal/logger"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

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

			wrapper := &responseWrapper{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			statusColor := logger.ColorGreen
			if wrapper.statusCode >= 400 {
				statusColor = logger.ColorRed
			} else if wrapper.statusCode >= 300 {
				statusColor = logger.ColorYellow
			}

			log.Printf(logger.ColorCyan+"[RESPONSE]"+logger.ColorReset+" %s%d"+logger.ColorReset+" | %s | %v",
				statusColor, wrapper.statusCode, r.URL.Path, duration)
		})
	}
}