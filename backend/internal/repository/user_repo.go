package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// UserRepo handles user database operations.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// FindByID finds a user by ID.
func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, oauth_provider, oauth_provider_id, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.OAuthProvider, &u.OAuthProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return &u, nil
}

// FindByEmail finds a user by email.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, oauth_provider, oauth_provider_id, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.OAuthProvider, &u.OAuthProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return &u, nil
}

// FindByOAuthProvider finds a user by OAuth provider and provider ID.
func (r *UserRepo) FindByOAuthProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, display_name, avatar_url, oauth_provider, oauth_provider_id, created_at, updated_at
		 FROM users WHERE oauth_provider = $1 AND oauth_provider_id = $2`, provider, providerID,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.AvatarURL, &u.OAuthProvider, &u.OAuthProviderID, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding user by oauth: %w", err)
	}
	return &u, nil
}

// Create inserts a new user.
func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users (email, display_name, avatar_url, oauth_provider, oauth_provider_id)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`,
		u.Email, u.DisplayName, u.AvatarURL, u.OAuthProvider, u.OAuthProviderID,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// Update updates user profile data.
func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET display_name = $2, avatar_url = $3, updated_at = NOW() WHERE id = $1`,
		u.ID, u.DisplayName, u.AvatarURL,
	)
	return err
}
