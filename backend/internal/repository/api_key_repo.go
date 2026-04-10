package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

type APIKeyRepo struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{pool: pool}
}

func (r *APIKeyRepo) Create(ctx context.Context, k *models.APIKey) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO api_keys (team_id, name, key_hash, prefix, scopes, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`,
		k.TeamID, k.Name, k.KeyHash, k.Prefix, k.Scopes, k.ExpiresAt,
	).Scan(&k.ID, &k.CreatedAt)
}

func (r *APIKeyRepo) FindByPrefix(ctx context.Context, prefix string) (*models.APIKey, error) {
	var k models.APIKey
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, key_hash, prefix, scopes, last_used, expires_at, created_at
		 FROM api_keys WHERE prefix = $1`, prefix,
	).Scan(&k.ID, &k.TeamID, &k.Name, &k.KeyHash, &k.Prefix, &k.Scopes, &k.LastUsed, &k.ExpiresAt, &k.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &k, err
}

func (r *APIKeyRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, name, key_hash, prefix, scopes, last_used, expires_at, created_at
		 FROM api_keys WHERE team_id = $1 ORDER BY created_at DESC`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.TeamID, &k.Name, &k.KeyHash, &k.Prefix, &k.Scopes, &k.LastUsed, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *APIKeyRepo) Delete(ctx context.Context, id uuid.UUID, teamID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM api_keys WHERE id = $1 AND team_id = $2`, id, teamID)
	return err
}

func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET last_used = NOW() WHERE id = $1`, id)
	return err
}
