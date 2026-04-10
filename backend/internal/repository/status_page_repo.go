package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// StatusPageRepo handles status page database operations.
type StatusPageRepo struct {
	pool *pgxpool.Pool
}

// NewStatusPageRepo creates a new StatusPageRepo.
func NewStatusPageRepo(pool *pgxpool.Pool) *StatusPageRepo {
	return &StatusPageRepo{pool: pool}
}

// Create creates a new status page.
func (r *StatusPageRepo) Create(ctx context.Context, sp *models.StatusPage) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO status_pages (team_id, name, slug, custom_domain, logo_url, primary_color,
		        is_password_protected, password_hash, hide_from_search, announcement, language)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id, created_at, updated_at`,
		sp.TeamID, sp.Name, sp.Slug, sp.CustomDomain, sp.LogoURL, sp.PrimaryColor,
		sp.IsPasswordProtected, sp.PasswordHash, sp.HideFromSearch, sp.Announcement, sp.Language,
	).Scan(&sp.ID, &sp.CreatedAt, &sp.UpdatedAt)
}

// FindByTeamID lists status pages for a team.
func (r *StatusPageRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.StatusPage, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, name, slug, custom_domain, logo_url, primary_color,
		        is_password_protected, hide_from_search, announcement, language, created_at, updated_at
		 FROM status_pages WHERE team_id = $1 ORDER BY name ASC`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []models.StatusPage
	for rows.Next() {
		var sp models.StatusPage
		if err := rows.Scan(&sp.ID, &sp.TeamID, &sp.Name, &sp.Slug, &sp.CustomDomain, &sp.LogoURL, &sp.PrimaryColor,
			&sp.IsPasswordProtected, &sp.HideFromSearch, &sp.Announcement, &sp.Language, &sp.CreatedAt, &sp.UpdatedAt); err != nil {
			return nil, err
		}
		pages = append(pages, sp)
	}
	return pages, nil
}

// FindBySlug finds a status page by its slug (for public access).
func (r *StatusPageRepo) FindBySlug(ctx context.Context, slug string) (*models.StatusPage, error) {
	var sp models.StatusPage
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, slug, custom_domain, logo_url, primary_color,
		        is_password_protected, password_hash, hide_from_search, announcement, language, created_at, updated_at
		 FROM status_pages WHERE slug = $1`, slug,
	).Scan(&sp.ID, &sp.TeamID, &sp.Name, &sp.Slug, &sp.CustomDomain, &sp.LogoURL, &sp.PrimaryColor,
		&sp.IsPasswordProtected, &sp.PasswordHash, &sp.HideFromSearch, &sp.Announcement, &sp.Language, &sp.CreatedAt, &sp.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

// FindByID finds a status page by ID.
func (r *StatusPageRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.StatusPage, error) {
	var sp models.StatusPage
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, slug, custom_domain, logo_url, primary_color,
		        is_password_protected, password_hash, hide_from_search, announcement, language, created_at, updated_at
		 FROM status_pages WHERE id = $1`, id,
	).Scan(&sp.ID, &sp.TeamID, &sp.Name, &sp.Slug, &sp.CustomDomain, &sp.LogoURL, &sp.PrimaryColor,
		&sp.IsPasswordProtected, &sp.PasswordHash, &sp.HideFromSearch, &sp.Announcement, &sp.Language, &sp.CreatedAt, &sp.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

// Update updates a status page.
func (r *StatusPageRepo) Update(ctx context.Context, sp *models.StatusPage) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE status_pages SET name=$2, slug=$3, custom_domain=$4, logo_url=$5, primary_color=$6,
		        is_password_protected=$7, password_hash=$8, hide_from_search=$9, announcement=$10, language=$11, updated_at=NOW()
		 WHERE id=$1`,
		sp.ID, sp.Name, sp.Slug, sp.CustomDomain, sp.LogoURL, sp.PrimaryColor,
		sp.IsPasswordProtected, sp.PasswordHash, sp.HideFromSearch, sp.Announcement, sp.Language,
	)
	return err
}

// Delete deletes a status page.
func (r *StatusPageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM status_pages WHERE id = $1`, id)
	return err
}

// SetMonitors sets the monitors for a status page.
func (r *StatusPageRepo) SetMonitors(ctx context.Context, statusPageID uuid.UUID, monitorIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM status_page_monitors WHERE status_page_id = $1`, statusPageID)
	if err != nil {
		return err
	}

	for i, mid := range monitorIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO status_page_monitors (status_page_id, monitor_id, sort_order) VALUES ($1, $2, $3)`,
			statusPageID, mid, i,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindMonitorsByStatusPageID lists monitors for a status page.
func (r *StatusPageRepo) FindMonitorsByStatusPageID(ctx context.Context, statusPageID uuid.UUID) ([]models.Monitor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.team_id, m.name, m.type, m.url, m.status, m.last_checked_at, m.last_response_ms,
		        m.uptime_percentage, m.created_at, m.updated_at
		 FROM monitors m
		 JOIN status_page_monitors spm ON m.id = spm.monitor_id
		 WHERE spm.status_page_id = $1
		 ORDER BY spm.sort_order ASC`, statusPageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var monitors []models.Monitor
	for rows.Next() {
		var m models.Monitor
		if err := rows.Scan(&m.ID, &m.TeamID, &m.Name, &m.Type, &m.URL, &m.Status, &m.LastCheckedAt, &m.LastResponseMs,
			&m.UptimePercentage, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}
