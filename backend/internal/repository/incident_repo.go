package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// IncidentRepo handles incident database operations.
type IncidentRepo struct {
	pool *pgxpool.Pool
}

// NewIncidentRepo creates a new IncidentRepo.
func NewIncidentRepo(pool *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{pool: pool}
}

// Create creates a new incident.
func (r *IncidentRepo) Create(ctx context.Context, i *models.Incident) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO incidents (monitor_id, reason, status)
		 VALUES ($1, $2, $3) RETURNING id, started_at, created_at`,
		i.MonitorID, i.Reason, i.Status,
	).Scan(&i.ID, &i.StartedAt, &i.CreatedAt)
}

// Resolve marks an incident as resolved.
func (r *IncidentRepo) Resolve(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE incidents SET resolved_at = NOW(), status = 'resolved',
		        duration_seconds = EXTRACT(EPOCH FROM (NOW() - started_at))::INTEGER
		 WHERE id = $1`, id,
	)
	return err
}

// FindOngoingByMonitorID finds any ongoing incident for a monitor.
func (r *IncidentRepo) FindOngoingByMonitorID(ctx context.Context, monitorID uuid.UUID) (*models.Incident, error) {
	var i models.Incident
	err := r.pool.QueryRow(ctx,
		`SELECT id, monitor_id, started_at, resolved_at, duration_seconds, reason, root_cause, status, created_at
		 FROM incidents WHERE monitor_id = $1 AND status = 'ongoing' ORDER BY started_at DESC LIMIT 1`, monitorID,
	).Scan(&i.ID, &i.MonitorID, &i.StartedAt, &i.ResolvedAt, &i.DurationSeconds, &i.Reason, &i.RootCause, &i.Status, &i.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding ongoing incident: %w", err)
	}
	return &i, nil
}

// FindByTeamID lists incidents for all monitors in a team.
func (r *IncidentRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID, limit, offset int) ([]models.Incident, int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM incidents i JOIN monitors m ON i.monitor_id = m.id WHERE m.team_id = $1`, teamID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT i.id, i.monitor_id, i.started_at, i.resolved_at, i.duration_seconds, i.reason, i.root_cause, i.status, i.created_at, m.name
		 FROM incidents i JOIN monitors m ON i.monitor_id = m.id
		 WHERE m.team_id = $1 ORDER BY i.started_at DESC LIMIT $2 OFFSET $3`, teamID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.StartedAt, &inc.ResolvedAt, &inc.DurationSeconds,
			&inc.Reason, &inc.RootCause, &inc.Status, &inc.CreatedAt, &inc.MonitorName); err != nil {
			return nil, 0, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, total, nil
}

// FindByMonitorID lists incidents for a specific monitor.
func (r *IncidentRepo) FindByMonitorID(ctx context.Context, monitorID uuid.UUID, limit, offset int) ([]models.Incident, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, monitor_id, started_at, resolved_at, duration_seconds, reason, root_cause, status, created_at
		 FROM incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT $2 OFFSET $3`, monitorID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.StartedAt, &inc.ResolvedAt, &inc.DurationSeconds,
			&inc.Reason, &inc.RootCause, &inc.Status, &inc.CreatedAt); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}
