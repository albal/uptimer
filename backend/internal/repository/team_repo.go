package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

func (r *TeamRepo) Create(ctx context.Context, t *models.Team) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO teams (name, owner_id, max_seats, max_monitors)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		t.Name, t.OwnerID, t.MaxSeats, t.MaxMonitors,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TeamRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	var t models.Team
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, max_seats, max_monitors, created_at, updated_at FROM teams WHERE id = $1`, id,
	).Scan(&t.ID, &t.Name, &t.OwnerID, &t.MaxSeats, &t.MaxMonitors, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TeamRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.name, t.owner_id, t.max_seats, t.max_monitors, t.created_at, t.updated_at
		 FROM teams t JOIN team_members tm ON t.id = tm.team_id WHERE tm.user_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var t models.Team
		if err := rows.Scan(&t.ID, &t.Name, &t.OwnerID, &t.MaxSeats, &t.MaxMonitors, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r *TeamRepo) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO team_members (team_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		teamID, userID, role,
	)
	return err
}

func (r *TeamRepo) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID,
	)
	return err
}

func (r *TeamRepo) FindMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT tm.team_id, tm.user_id, tm.role, tm.joined_at, u.email, u.display_name, u.avatar_url
		 FROM team_members tm JOIN users u ON tm.user_id = u.id WHERE tm.team_id = $1`, teamID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var m models.TeamMember
		u := &models.User{}
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role, &m.JoinedAt, &u.Email, &u.DisplayName, &u.AvatarURL); err != nil {
			return nil, err
		}
		u.ID = m.UserID
		m.User = u
		members = append(members, m)
	}
	return members, nil
}

func (r *TeamRepo) CountMembers(ctx context.Context, teamID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE team_id = $1`, teamID).Scan(&count)
	return count, err
}
