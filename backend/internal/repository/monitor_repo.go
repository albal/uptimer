package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/albal/uptimer/internal/models"
)

// MonitorRepo handles monitor database operations.
type MonitorRepo struct {
	pool *pgxpool.Pool
}

// NewMonitorRepo creates a new MonitorRepo.
func NewMonitorRepo(pool *pgxpool.Pool) *MonitorRepo {
	return &MonitorRepo{pool: pool}
}

// FindByID finds a monitor by ID.
func (r *MonitorRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Monitor, error) {
	var m models.Monitor
	var headersJSON, assertionsJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, type, url, ip_address, port, interval_seconds, timeout_seconds,
		        http_method, http_headers, http_body, http_auth_type, http_username, http_password_enc,
		        expected_status_codes, follow_redirects, keyword, keyword_type, api_assertions,
		        udp_data, udp_expected, ssl_expiry_reminder, dns_record_type, dns_expected_value,
		        domain_expiry_reminder, monitoring_regions, slow_threshold_ms,
		        heartbeat_token, heartbeat_grace_sec, heartbeat_last_ping,
		        status, last_checked_at, last_response_ms, uptime_percentage, total_checks, total_downtime_sec,
		        created_by, created_at, updated_at
		 FROM monitors WHERE id = $1`, id,
	).Scan(
		&m.ID, &m.TeamID, &m.Name, &m.Type, &m.URL, &m.IPAddress, &m.Port, &m.IntervalSeconds, &m.TimeoutSeconds,
		&m.HTTPMethod, &headersJSON, &m.HTTPBody, &m.HTTPAuthType, &m.HTTPUsername, &m.HTTPPasswordEnc,
		&m.ExpectedStatusCodes, &m.FollowRedirects, &m.Keyword, &m.KeywordType, &assertionsJSON,
		&m.UDPData, &m.UDPExpected, &m.SSLExpiryReminder, &m.DNSRecordType, &m.DNSExpectedValue,
		&m.DomainExpiryReminder, &m.MonitoringRegions, &m.SlowThresholdMs,
		&m.HeartbeatToken, &m.HeartbeatGraceSec, &m.HeartbeatLastPing,
		&m.Status, &m.LastCheckedAt, &m.LastResponseMs, &m.UptimePercentage, &m.TotalChecks, &m.TotalDowntimeSec,
		&m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("finding monitor by id: %w", err)
	}
	if headersJSON != nil {
		json.Unmarshal(headersJSON, &m.HTTPHeaders)
	}
	if assertionsJSON != nil {
		json.Unmarshal(assertionsJSON, &m.APIAssertions)
	}
	return &m, nil
}

// FindByTeamID lists all monitors for a team.
func (r *MonitorRepo) FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]models.Monitor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, name, type, url, ip_address, port, interval_seconds, timeout_seconds,
		        status, last_checked_at, last_response_ms, uptime_percentage, total_checks,
		        monitoring_regions, slow_threshold_ms, created_at, updated_at
		 FROM monitors WHERE team_id = $1 ORDER BY name ASC`, teamID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing monitors: %w", err)
	}
	defer rows.Close()

	var monitors []models.Monitor
	for rows.Next() {
		var m models.Monitor
		if err := rows.Scan(
			&m.ID, &m.TeamID, &m.Name, &m.Type, &m.URL, &m.IPAddress, &m.Port, &m.IntervalSeconds, &m.TimeoutSeconds,
			&m.Status, &m.LastCheckedAt, &m.LastResponseMs, &m.UptimePercentage, &m.TotalChecks,
			&m.MonitoringRegions, &m.SlowThresholdMs, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning monitor row: %w", err)
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}

// FindAllActive returns all active monitors for the engine.
func (r *MonitorRepo) FindAllActive(ctx context.Context) ([]models.Monitor, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, team_id, name, type, url, ip_address, port, interval_seconds, timeout_seconds,
		        http_method, http_headers, http_body, http_auth_type, http_username, http_password_enc,
		        expected_status_codes, follow_redirects, keyword, keyword_type, api_assertions,
		        udp_data, udp_expected, ssl_expiry_reminder, dns_record_type, dns_expected_value,
		        domain_expiry_reminder, monitoring_regions, slow_threshold_ms,
		        heartbeat_token, heartbeat_grace_sec, heartbeat_last_ping,
		        status, last_checked_at, last_response_ms, uptime_percentage, total_checks, total_downtime_sec,
		        created_by, created_at, updated_at
		 FROM monitors WHERE status != 'paused'`,
	)
	if err != nil {
		return nil, fmt.Errorf("finding active monitors: %w", err)
	}
	defer rows.Close()

	var monitors []models.Monitor
	for rows.Next() {
		var m models.Monitor
		var headersJSON, assertionsJSON []byte
		if err := rows.Scan(
			&m.ID, &m.TeamID, &m.Name, &m.Type, &m.URL, &m.IPAddress, &m.Port, &m.IntervalSeconds, &m.TimeoutSeconds,
			&m.HTTPMethod, &headersJSON, &m.HTTPBody, &m.HTTPAuthType, &m.HTTPUsername, &m.HTTPPasswordEnc,
			&m.ExpectedStatusCodes, &m.FollowRedirects, &m.Keyword, &m.KeywordType, &assertionsJSON,
			&m.UDPData, &m.UDPExpected, &m.SSLExpiryReminder, &m.DNSRecordType, &m.DNSExpectedValue,
			&m.DomainExpiryReminder, &m.MonitoringRegions, &m.SlowThresholdMs,
			&m.HeartbeatToken, &m.HeartbeatGraceSec, &m.HeartbeatLastPing,
			&m.Status, &m.LastCheckedAt, &m.LastResponseMs, &m.UptimePercentage, &m.TotalChecks, &m.TotalDowntimeSec,
			&m.CreatedBy, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning active monitor: %w", err)
		}
		if headersJSON != nil {
			json.Unmarshal(headersJSON, &m.HTTPHeaders)
		}
		if assertionsJSON != nil {
			json.Unmarshal(assertionsJSON, &m.APIAssertions)
		}
		monitors = append(monitors, m)
	}
	return monitors, nil
}

// Create creates a new monitor.
func (r *MonitorRepo) Create(ctx context.Context, m *models.Monitor) error {
	headersJSON, _ := json.Marshal(m.HTTPHeaders)
	assertionsJSON, _ := json.Marshal(m.APIAssertions)

	return r.pool.QueryRow(ctx,
		`INSERT INTO monitors (team_id, name, type, url, ip_address, port, interval_seconds, timeout_seconds,
		        http_method, http_headers, http_body, http_auth_type, http_username, http_password_enc,
		        expected_status_codes, follow_redirects, keyword, keyword_type, api_assertions,
		        udp_data, udp_expected, ssl_expiry_reminder, dns_record_type, dns_expected_value,
		        domain_expiry_reminder, monitoring_regions, slow_threshold_ms,
		        heartbeat_token, heartbeat_grace_sec, status, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31)
		 RETURNING id, created_at, updated_at`,
		m.TeamID, m.Name, m.Type, m.URL, m.IPAddress, m.Port, m.IntervalSeconds, m.TimeoutSeconds,
		m.HTTPMethod, headersJSON, m.HTTPBody, m.HTTPAuthType, m.HTTPUsername, m.HTTPPasswordEnc,
		m.ExpectedStatusCodes, m.FollowRedirects, m.Keyword, m.KeywordType, assertionsJSON,
		m.UDPData, m.UDPExpected, m.SSLExpiryReminder, m.DNSRecordType, m.DNSExpectedValue,
		m.DomainExpiryReminder, m.MonitoringRegions, m.SlowThresholdMs,
		m.HeartbeatToken, m.HeartbeatGraceSec, m.Status, m.CreatedBy,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// Update updates a monitor.
func (r *MonitorRepo) Update(ctx context.Context, m *models.Monitor) error {
	headersJSON, _ := json.Marshal(m.HTTPHeaders)
	assertionsJSON, _ := json.Marshal(m.APIAssertions)

	_, err := r.pool.Exec(ctx,
		`UPDATE monitors SET name=$2, type=$3, url=$4, ip_address=$5, port=$6, interval_seconds=$7, timeout_seconds=$8,
		        http_method=$9, http_headers=$10, http_body=$11, http_auth_type=$12, http_username=$13, http_password_enc=$14,
		        expected_status_codes=$15, follow_redirects=$16, keyword=$17, keyword_type=$18, api_assertions=$19,
		        udp_data=$20, udp_expected=$21, ssl_expiry_reminder=$22, dns_record_type=$23, dns_expected_value=$24,
		        domain_expiry_reminder=$25, monitoring_regions=$26, slow_threshold_ms=$27,
		        heartbeat_grace_sec=$28, updated_at=NOW()
		 WHERE id = $1`,
		m.ID, m.Name, m.Type, m.URL, m.IPAddress, m.Port, m.IntervalSeconds, m.TimeoutSeconds,
		m.HTTPMethod, headersJSON, m.HTTPBody, m.HTTPAuthType, m.HTTPUsername, m.HTTPPasswordEnc,
		m.ExpectedStatusCodes, m.FollowRedirects, m.Keyword, m.KeywordType, assertionsJSON,
		m.UDPData, m.UDPExpected, m.SSLExpiryReminder, m.DNSRecordType, m.DNSExpectedValue,
		m.DomainExpiryReminder, m.MonitoringRegions, m.SlowThresholdMs,
		m.HeartbeatGraceSec,
	)
	return err
}

// Delete deletes a monitor.
func (r *MonitorRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM monitors WHERE id = $1`, id)
	return err
}

// UpdateStatus updates a monitor's status and check results.
func (r *MonitorRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, responseMs int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE monitors SET status = $2, last_checked_at = NOW(), last_response_ms = $3,
		        total_checks = total_checks + 1, updated_at = NOW()
		 WHERE id = $1`,
		id, status, responseMs,
	)
	return err
}

// UpdateUptimePercentage recalculates uptime percentage.
func (r *MonitorRepo) UpdateUptimePercentage(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE monitors SET uptime_percentage = (
			SELECT COALESCE(
				(COUNT(*) FILTER (WHERE status = 'up') * 100.0 / NULLIF(COUNT(*), 0)),
				100.0
			) FROM monitor_results WHERE monitor_id = $1
		), updated_at = NOW() WHERE id = $1`, id,
	)
	return err
}

// CountByTeamID returns the number of monitors for a team.
func (r *MonitorRepo) CountByTeamID(ctx context.Context, teamID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM monitors WHERE team_id = $1`, teamID).Scan(&count)
	return count, err
}

// InsertResult records a monitoring check result.
func (r *MonitorRepo) InsertResult(ctx context.Context, result *models.MonitorResult) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO monitor_results (monitor_id, status, response_time_ms, status_code, error_message, region, checked_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		result.MonitorID, result.Status, result.ResponseTimeMs, result.StatusCode, result.ErrorMessage, result.Region, result.CheckedAt,
	)
	return err
}

// FindResults returns paginated results for a monitor.
func (r *MonitorRepo) FindResults(ctx context.Context, monitorID uuid.UUID, limit int, offset int) ([]models.MonitorResult, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM monitor_results WHERE monitor_id = $1`, monitorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, monitor_id, status, response_time_ms, status_code, error_message, region, checked_at
		 FROM monitor_results WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT $2 OFFSET $3`,
		monitorID, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []models.MonitorResult
	for rows.Next() {
		var r models.MonitorResult
		if err := rows.Scan(&r.ID, &r.MonitorID, &r.Status, &r.ResponseTimeMs, &r.StatusCode, &r.ErrorMessage, &r.Region, &r.CheckedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}
	return results, total, nil
}

// FindByHeartbeatToken finds a monitor by heartbeat token.
func (r *MonitorRepo) FindByHeartbeatToken(ctx context.Context, token string) (*models.Monitor, error) {
	var m models.Monitor
	err := r.pool.QueryRow(ctx,
		`SELECT id, team_id, name, type, heartbeat_grace_sec, heartbeat_last_ping, status
		 FROM monitors WHERE heartbeat_token = $1 AND type = 'heartbeat'`, token,
	).Scan(&m.ID, &m.TeamID, &m.Name, &m.Type, &m.HeartbeatGraceSec, &m.HeartbeatLastPing, &m.Status)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateHeartbeatPing updates the last heartbeat ping time.
func (r *MonitorRepo) UpdateHeartbeatPing(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE monitors SET heartbeat_last_ping = NOW(), status = 'up', last_checked_at = NOW(), updated_at = NOW() WHERE id = $1`, id,
	)
	return err
}
