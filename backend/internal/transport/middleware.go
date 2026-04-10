package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/service"
)

// ContextKey type for context values.
type ContextKey string

const (
	ContextUserID ContextKey = "user_id"
	ContextTeamID ContextKey = "team_id"
	ContextEmail  ContextKey = "email"
)

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ContextUserID).(uuid.UUID)
	return id
}

// GetTeamID extracts the team ID from the request context.
func GetTeamID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ContextTeamID).(uuid.UUID)
	return id
}

// AuthMiddlewareFunc creates an auth middleware from the auth service.
func AuthMiddlewareFunc(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try cookie first
			var tokenStr string
			cookie, err := r.Cookie("uptimer_token")
			if err == nil {
				tokenStr = cookie.Value
			}

			// Try Authorization header
			if tokenStr == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					tokenStr = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			if tokenStr == "" {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			claims, err := authService.ValidateJWT(tokenStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextTeamID, claims.TeamID)
			ctx = context.WithValue(ctx, ContextEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// readJSON reads and decodes JSON from the request body.
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
