package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// MaintenanceWindowRepo handles maintenance window database operations.
type MaintenanceWindowRepo struct {
	pool *pgxpool.Pool
}

// NewMaintenanceWindowRepo creates a new MaintenanceWindowRepo.
func NewMaintenanceWindowRepo(pool *pgxpool.Pool) *MaintenanceWindowRepo {
	return &MaintenanceWindowRepo{pool: pool}
}

// Create creates a new maintenance window.
func (r *MaintenanceWindowRepo) Create(ctx context.Context, mw *models.MaintenanceWindow) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO maintenance_windows (team_id, name, start_time, end_time, recurring, recurrence_rule)
		 VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at, updated_at`,
		mw.TeamID, mw.Name, mw.StartTime, mw.EndTime, mw.Recurring, mw.RecurrenceRule,
	).Scan(&mw.ID, &mw.CreatedAt, &mw.UpdatedAt)
	if err != nil {
		return err
	}

	for _, mid := range mw.MonitorIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO maintenance_window_monitors (maintenance_window_id, monitor_id) VALUES ($1, $2)`,
			mw.ID, mid,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindByTeamID lists maintenance windows for a team.
func (r *MaintenanceWindowRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.MaintenanceWindow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, name, start_time, end_time, recurring, recurrence_rule, created_at, updated_at
		 FROM maintenance_windows WHERE team_id = $1 ORDER BY start_time DESC`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var windows []models.MaintenanceWindow
	for rows.Next() {
		var mw models.MaintenanceWindow
		if err := rows.Scan(&mw.ID, &mw.TeamID, &mw.Name, &mw.StartTime, &mw.EndTime,
			&mw.Recurring, &mw.RecurrenceRule, &mw.CreatedAt, &mw.UpdatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, mw)
	}
	return windows, nil
}

// FindByID finds a maintenance window by ID.
func (r *MaintenanceWindowRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.MaintenanceWindow, error) {
	var mw models.MaintenanceWindow
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, start_time, end_time, recurring, recurrence_rule, created_at, updated_at
		 FROM maintenance_windows WHERE id = $1`, id,
	).Scan(&mw.ID, &mw.TeamID, &mw.Name, &mw.StartTime, &mw.EndTime,
		&mw.Recurring, &mw.RecurrenceRule, &mw.CreatedAt, &mw.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load associated monitor IDs
	midRows, err := r.pool.Query(ctx,
		`SELECT monitor_id FROM maintenance_window_monitors WHERE maintenance_window_id = $1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer midRows.Close()

	for midRows.Next() {
		var mid uuid.UUID
		if err := midRows.Scan(&mid); err != nil {
			return nil, err
		}
		mw.MonitorIDs = append(mw.MonitorIDs, mid)
	}
	return &mw, nil
}

// Delete deletes a maintenance window.
func (r *MaintenanceWindowRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM maintenance_windows WHERE id = $1`, id)
	return err
}

// IsMonitorInMaintenance checks if a monitor is currently in a maintenance window.
func (r *MaintenanceWindowRepo) IsMonitorInMaintenance(ctx context.Context, monitorID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM maintenance_window_monitors mwm
			JOIN maintenance_windows mw ON mwm.maintenance_window_id = mw.id
			WHERE mwm.monitor_id = $1 AND mw.start_time <= NOW() AND mw.end_time >= NOW()
		)`, monitorID,
	).Scan(&exists)
	return exists, err
}
