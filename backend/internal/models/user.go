package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an authenticated user (OAuth only, no passwords stored).
type User struct {
	ID              uuid.UUID `json:"id" db:"id"`
	Email           string    `json:"email" db:"email"`
	DisplayName     string    `json:"display_name" db:"display_name"`
	AvatarURL       string    `json:"avatar_url,omitempty" db:"avatar_url"`
	OAuthProvider   string    `json:"oauth_provider" db:"oauth_provider"`
	OAuthProviderID string    `json:"-" db:"oauth_provider_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Team represents a team/workspace.
type Team struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	OwnerID     uuid.UUID `json:"owner_id" db:"owner_id"`
	MaxSeats    int       `json:"max_seats" db:"max_seats"`
	MaxMonitors int       `json:"max_monitors" db:"max_monitors"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TeamMember represents a user's membership in a team.
type TeamMember struct {
	TeamID   uuid.UUID `json:"team_id" db:"team_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	User     *User     `json:"user,omitempty"`
}

// TeamRole constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)
