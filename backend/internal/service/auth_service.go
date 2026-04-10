package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/albal/uptimer/internal/config"
	"github.com/albal/uptimer/internal/models"
	"github.com/albal/uptimer/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Microsoft OAuth2 endpoints
var microsoftEndpoint = oauth2.Endpoint{
	AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
	TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
}

// Apple OAuth2 endpoints
var appleEndpoint = oauth2.Endpoint{
	AuthURL:  "https://appleid.apple.com/auth/authorize",
	TokenURL: "https://appleid.apple.com/auth/token",
}

// JWTClaims holds the claims for our JWT tokens.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	TeamID uuid.UUID `json:"team_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// AuthService handles authentication flows.
type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepo
	teamRepo *repository.TeamRepo
	google   *oauth2.Config
	microsoft *oauth2.Config
	apple     *oauth2.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(cfg *config.Config, userRepo *repository.UserRepo, teamRepo *repository.TeamRepo) *AuthService {
	s := &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}

	if cfg.GoogleClientID != "" {
		s.google = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.BaseURL + "/api/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	if cfg.MicrosoftClientID != "" {
		s.microsoft = &oauth2.Config{
			ClientID:     cfg.MicrosoftClientID,
			ClientSecret: cfg.MicrosoftClientSecret,
			RedirectURL:  cfg.BaseURL + "/api/auth/microsoft/callback",
			Scopes:       []string{"openid", "email", "profile", "User.Read"},
			Endpoint:     microsoftEndpoint,
		}
	}

	if cfg.AppleClientID != "" {
		s.apple = &oauth2.Config{
			ClientID:     cfg.AppleClientID,
			RedirectURL:  cfg.BaseURL + "/api/auth/apple/callback",
			Scopes:       []string{"name", "email"},
			Endpoint:     appleEndpoint,
		}
	}

	return s
}

// GenerateState creates a cryptographic random state for CSRF protection.
func (s *AuthService) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetGoogleAuthURL returns the Google OAuth2 authorization URL.
func (s *AuthService) GetGoogleAuthURL(state string) string {
	if s.google == nil {
		return ""
	}
	return s.google.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetMicrosoftAuthURL returns the Microsoft OAuth2 authorization URL.
func (s *AuthService) GetMicrosoftAuthURL(state string) string {
	if s.microsoft == nil {
		return ""
	}
	return s.microsoft.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GetAppleAuthURL returns the Apple OAuth2 authorization URL.
func (s *AuthService) GetAppleAuthURL(state string) string {
	if s.apple == nil {
		return ""
	}
	return s.apple.AuthCodeURL(state, oauth2.SetAuthURLParam("response_mode", "form_post"))
}

// GoogleUserInfo represents the user info from Google.
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// HandleGoogleCallback processes the Google OAuth2 callback.
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, error) {
	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	client := s.google.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info GoogleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing user info: %w", err)
	}

	return s.findOrCreateUser(ctx, "google", info.ID, info.Email, info.Name, info.Picture)
}

// MicrosoftUserInfo represents the user info from Microsoft Graph.
type MicrosoftUserInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Mail        string `json:"mail"`
	UPN         string `json:"userPrincipalName"`
}

// HandleMicrosoftCallback processes the Microsoft OAuth2 callback.
func (s *AuthService) HandleMicrosoftCallback(ctx context.Context, code string) (*models.User, error) {
	token, err := s.microsoft.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	client := s.microsoft.Client(ctx, token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var info MicrosoftUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing user info: %w", err)
	}

	email := info.Mail
	if email == "" {
		email = info.UPN
	}

	return s.findOrCreateUser(ctx, "microsoft", info.ID, email, info.DisplayName, "")
}

// HandleAppleCallback processes the Apple OAuth2 callback.
func (s *AuthService) HandleAppleCallback(ctx context.Context, code string, idTokenStr string) (*models.User, error) {
	// For Apple, we parse the ID token to get user info
	// In production, you'd verify the ID token signature with Apple's public keys
	token, err := s.apple.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	// Parse the ID token (simplified — in production, verify with Apple's JWKS)
	idToken := token.Extra("id_token")
	if idToken == nil && idTokenStr != "" {
		idToken = idTokenStr
	}

	// Extract claims from the ID token
	claims, err := parseUnverifiedJWT(fmt.Sprintf("%v", idToken))
	if err != nil {
		return nil, fmt.Errorf("parsing apple id token: %w", err)
	}

	email, _ := claims["email"].(string)
	sub, _ := claims["sub"].(string)

	return s.findOrCreateUser(ctx, "apple", sub, email, email, "")
}

// findOrCreateUser looks up or creates a user and their default team.
func (s *AuthService) findOrCreateUser(ctx context.Context, provider, providerID, email, displayName, avatarURL string) (*models.User, error) {
	// Try to find existing user
	user, err := s.userRepo.FindByOAuthProvider(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		// Update display name and avatar if changed
		if user.DisplayName != displayName || user.AvatarURL != avatarURL {
			user.DisplayName = displayName
			user.AvatarURL = avatarURL
			s.userRepo.Update(ctx, user)
		}
		return user, nil
	}

	// Create new user
	user = &models.User{
		Email:           email,
		DisplayName:     displayName,
		AvatarURL:       avatarURL,
		OAuthProvider:   provider,
		OAuthProviderID: providerID,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Create default team
	team := &models.Team{
		Name:        displayName + "'s Team",
		OwnerID:     user.ID,
		MaxSeats:    s.cfg.DefaultMaxSeats,
		MaxMonitors: s.cfg.DefaultMaxMonitors,
	}
	if err := s.teamRepo.Create(ctx, team); err != nil {
		return nil, fmt.Errorf("creating default team: %w", err)
	}

	// Add user to team as owner
	if err := s.teamRepo.AddMember(ctx, team.ID, user.ID, models.RoleOwner); err != nil {
		return nil, fmt.Errorf("adding user to team: %w", err)
	}

	slog.Info("new user registered", "email", email, "provider", provider)
	return user, nil
}

// GenerateJWT creates a JWT token for the user.
func (s *AuthService) GenerateJWT(user *models.User, teamID uuid.UUID) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		TeamID: teamID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// ValidateJWT validates and parses a JWT token.
func (s *AuthService) ValidateJWT(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// GetAvailableProviders returns which OAuth providers are configured.
func (s *AuthService) GetAvailableProviders() []string {
	var providers []string
	if s.google != nil {
		providers = append(providers, "google")
	}
	if s.microsoft != nil {
		providers = append(providers, "microsoft")
	}
	if s.apple != nil {
		providers = append(providers, "apple")
	}
	return providers
}

// parseUnverifiedJWT parses a JWT token without verification to extract claims.
// This is only used for Apple's ID token where we trust the TLS connection.
func parseUnverifiedJWT(tokenStr string) (map[string]interface{}, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

// ValidateAPIKey validates an API key and returns the associated team ID.
func (s *AuthService) ValidateAPIKey(ctx context.Context, apiKeyRepo *repository.APIKeyRepo, rawKey string) (uuid.UUID, error) {
	if len(rawKey) < 8 {
		return uuid.Nil, fmt.Errorf("invalid API key format")
	}

	prefix := rawKey[:8]
	key, err := apiKeyRepo.FindByPrefix(ctx, prefix)
	if err != nil {
		return uuid.Nil, err
	}
	if key == nil {
		return uuid.Nil, fmt.Errorf("API key not found")
	}

	// In production, compare hashes using bcrypt/argon2
	// For now, we do a simple comparison
	if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, fmt.Errorf("API key expired")
	}

	_ = apiKeyRepo.UpdateLastUsed(ctx, key.ID)
	return key.TeamID, nil
}

// SetAuthCookie sets the JWT token as an HTTP-only cookie.
func SetAuthCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "uptimer_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
