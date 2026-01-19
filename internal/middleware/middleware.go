package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	// SessionCookieName is the name of the session cookie
	SessionCookieName = "mdspace_session"
	// SessionMaxAge is the maximum age of the session cookie (24 hours)
	SessionMaxAge = 24 * 60 * 60
)

// SessionMiddleware ensures each request has a session ID
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			sessionID := uuid.New().String()
			http.SetCookie(w, &http.Cookie{
				Name:     SessionCookieName,
				Value:    sessionID,
				Path:     "/",
				MaxAge:   SessionMaxAge,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				Secure:   false, // Set to true in production with HTTPS
			})
			r.Header.Set("X-Session-ID", sessionID)
		} else {
			r.Header.Set("X-Session-ID", cookie.Value)
		}
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequestLoggerMiddleware logs incoming requests
func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		if r.URL.Path != "/health" {
			println(r.Method, r.URL.Path, duration.String())
		}
	})
}
