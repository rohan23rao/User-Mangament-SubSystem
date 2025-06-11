// internal/middleware/logging.go
package middleware

import (
	"log"
	"net/http"
	"time"
	"userms/internal/auth"
	"userms/internal/utils"
)

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(sessionManager *auth.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			session, _ := sessionManager.GetSessionFromRequest(r)
			userID := "anonymous"
			if session != nil {
				userID = session.Identity.Id[:8] + "..."
			}

			utils.LogRequest(r.Method, r.URL.Path, userID)

			wrapper := &responseWrapper{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			statusColor := utils.ColorGreen
			if wrapper.statusCode >= 400 {
				statusColor = utils.ColorRed
			} else if wrapper.statusCode >= 300 {
				statusColor = utils.ColorYellow
			}

			log.Printf(utils.ColorCyan+"[RESPONSE]"+utils.ColorReset+" %s%d"+utils.ColorReset+" | %s | %v",
				statusColor, wrapper.statusCode, r.URL.Path, duration)
		})
	}
}
