package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// AlertContactRepo handles alert contact database operations.
type AlertContactRepo struct {
	pool *pgxpool.Pool
}

// NewAlertContactRepo creates a new AlertContactRepo.
func NewAlertContactRepo(pool *pgxpool.Pool) *AlertContactRepo {
	return &AlertContactRepo{pool: pool}
}

// Create creates a new alert contact.
func (r *AlertContactRepo) Create(ctx context.Context, ac *models.AlertContact) error {
	configJSON, _ := json.Marshal(ac.Config)
	return r.pool.QueryRow(ctx,
		`INSERT INTO alert_contacts (team_id, type, name, value, config, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`,
		ac.TeamID, ac.Type, ac.Name, ac.Value, configJSON, ac.IsActive,
	).Scan(&ac.ID, &ac.CreatedAt, &ac.UpdatedAt)
}

// FindByTeamID lists alert contacts for a team.
func (r *AlertContactRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.AlertContact, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, type, name, value, config, is_active, created_at, updated_at
		 FROM alert_contacts WHERE team_id = $1 ORDER BY name ASC`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.AlertContact
	for rows.Next() {
		var ac models.AlertContact
		var configJSON []byte
		if err := rows.Scan(&ac.ID, &ac.TeamID, &ac.Type, &ac.Name, &ac.Value, &configJSON, &ac.IsActive, &ac.CreatedAt, &ac.UpdatedAt); err != nil {
			return nil, err
		}
		if configJSON != nil {
			json.Unmarshal(configJSON, &ac.Config)
		}
		contacts = append(contacts, ac)
	}
	return contacts, nil
}

// FindByID finds an alert contact by ID.
func (r *AlertContactRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.AlertContact, error) {
	var ac models.AlertContact
	var configJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, type, name, value, config, is_active, created_at, updated_at
		 FROM alert_contacts WHERE id = $1`, id,
	).Scan(&ac.ID, &ac.TeamID, &ac.Type, &ac.Name, &ac.Value, &configJSON, &ac.IsActive, &ac.CreatedAt, &ac.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if configJSON != nil {
		json.Unmarshal(configJSON, &ac.Config)
	}
	return &ac, nil
}

// Update updates an alert contact.
func (r *AlertContactRepo) Update(ctx context.Context, ac *models.AlertContact) error {
	configJSON, _ := json.Marshal(ac.Config)
	_, err := r.pool.Exec(ctx,
		`UPDATE alert_contacts SET type=$2, name=$3, value=$4, config=$5, is_active=$6, updated_at=NOW() WHERE id=$1`,
		ac.ID, ac.Type, ac.Name, ac.Value, configJSON, ac.IsActive,
	)
	return err
}

// Delete deletes an alert contact.
func (r *AlertContactRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM alert_contacts WHERE id = $1`, id)
	return err
}

// FindByMonitorID returns all alert contacts linked to a monitor.
func (r *AlertContactRepo) FindByMonitorID(ctx context.Context, monitorID uuid.UUID) ([]models.AlertContact, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ac.id, ac.team_id, ac.type, ac.name, ac.value, ac.config, ac.is_active, ac.created_at, ac.updated_at
		 FROM alert_contacts ac
		 JOIN monitor_alert_contacts mac ON ac.id = mac.alert_contact_id
		 WHERE mac.monitor_id = $1 AND ac.is_active = true`, monitorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.AlertContact
	for rows.Next() {
		var ac models.AlertContact
		var configJSON []byte
		if err := rows.Scan(&ac.ID, &ac.TeamID, &ac.Type, &ac.Name, &ac.Value, &configJSON, &ac.IsActive, &ac.CreatedAt, &ac.UpdatedAt); err != nil {
			return nil, err
		}
		if configJSON != nil {
			json.Unmarshal(configJSON, &ac.Config)
		}
		contacts = append(contacts, ac)
	}
	return contacts, nil
}

// LinkToMonitor links an alert contact to a monitor.
func (r *AlertContactRepo) LinkToMonitor(ctx context.Context, monitorID, alertContactID uuid.UUID, thresholdSeconds int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO monitor_alert_contacts (monitor_id, alert_contact_id, threshold_seconds)
		 VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, monitorID, alertContactID, thresholdSeconds,
	)
	return err
}

// UnlinkFromMonitor removes an alert contact from a monitor.
func (r *AlertContactRepo) UnlinkFromMonitor(ctx context.Context, monitorID, alertContactID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM monitor_alert_contacts WHERE monitor_id = $1 AND alert_contact_id = $2`, monitorID, alertContactID,
	)
	return err
}
