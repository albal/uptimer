package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/albal/uptimer/internal/config"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"github.com/albal/uptimer/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	authService *service.AuthService
	teamRepo    *repository.TeamRepo
	cfg         *config.Config
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *service.AuthService, teamRepo *repository.TeamRepo, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, teamRepo: teamRepo, cfg: cfg}
}

// GetProviders returns the configured OAuth providers.
func (h *AuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers": h.authService.GetAvailableProviders(),
	})
}

// GoogleLogin redirects to Google OAuth.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, _ := h.authService.GenerateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	url := h.authService.GetGoogleAuthURL(state)
	if url == "" {
		writeError(w, http.StatusBadRequest, "Google OAuth not configured")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the Google OAuth callback.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	user, err := h.authService.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OAuth callback failed: "+err.Error())
		return
	}

	h.completeLogin(w, r, user)
}

// MicrosoftLogin redirects to Microsoft OAuth.
func (h *AuthHandler) MicrosoftLogin(w http.ResponseWriter, r *http.Request) {
	state, _ := h.authService.GenerateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	url := h.authService.GetMicrosoftAuthURL(state)
	if url == "" {
		writeError(w, http.StatusBadRequest, "Microsoft OAuth not configured")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// MicrosoftCallback handles the Microsoft OAuth callback.
func (h *AuthHandler) MicrosoftCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	user, err := h.authService.HandleMicrosoftCallback(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OAuth callback failed: "+err.Error())
		return
	}

	h.completeLogin(w, r, user)
}

// AppleLogin redirects to Apple OAuth.
func (h *AuthHandler) AppleLogin(w http.ResponseWriter, r *http.Request) {
	state, _ := h.authService.GenerateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	url := h.authService.GetAppleAuthURL(state)
	if url == "" {
		writeError(w, http.StatusBadRequest, "Apple OAuth not configured")
		return
	}
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// AppleCallback handles the Apple OAuth callback (POST).
func (h *AuthHandler) AppleCallback(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	idToken := r.FormValue("id_token")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing authorization code")
		return
	}

	user, err := h.authService.HandleAppleCallback(r.Context(), code, idToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OAuth callback failed: "+err.Error())
		return
	}

	h.completeLogin(w, r, user)
}

// completeLogin generates a JWT and redirects to the frontend.
func (h *AuthHandler) completeLogin(w http.ResponseWriter, r *http.Request, user *models.User) {

	// Get user's teams
	teams, err := h.teamRepo.FindByUserID(r.Context(), user.ID)
	if err != nil || len(teams) == 0 {
		writeError(w, http.StatusInternalServerError, "user has no team")
		return
	}

	teamID := teams[0].ID
	token, err := h.authService.GenerateJWT(user, teamID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	service.SetAuthCookie(w, token, h.cfg.JWTExpiryHours*3600)

	// Redirect to frontend dashboard
	http.Redirect(w, r, h.cfg.FrontendURL+"/dashboard", http.StatusTemporaryRedirect)
}

// GetMe returns the current authenticated user.
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := GetUserID(r.Context())
	teamID := GetTeamID(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"team_id": teamID,
	})
}

// Logout clears the auth cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	service.SetAuthCookie(w, "", -1)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// WithAuth wraps a handler with authentication middleware.
func (h *AuthHandler) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		cookie, err := r.Cookie("uptimer_token")
		if err == nil {
			tokenStr = cookie.Value
		}
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
		claims, err := h.authService.ValidateJWT(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
		ctx = context.WithValue(ctx, ContextTeamID, claims.TeamID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AuthMiddleware is the Chi middleware version.
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return AuthMiddlewareFunc(h.authService)(next)
}
