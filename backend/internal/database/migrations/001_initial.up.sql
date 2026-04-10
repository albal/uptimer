-- Uptimer Initial Schema
-- Covers all tables for the full enterprise-grade monitoring platform.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

------------------------------------------------------------
-- Users (OAuth only, no passwords stored)
------------------------------------------------------------
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    display_name    VARCHAR(255) NOT NULL,
    avatar_url      TEXT,
    oauth_provider  VARCHAR(50) NOT NULL,
    oauth_provider_id VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(oauth_provider, oauth_provider_id)
);

------------------------------------------------------------
-- Teams (multi-seat support)
------------------------------------------------------------
CREATE TABLE teams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    max_seats   INTEGER NOT NULL DEFAULT 5,
    max_monitors INTEGER NOT NULL DEFAULT 1000,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE team_members (
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id)
);

------------------------------------------------------------
-- Monitors
------------------------------------------------------------
CREATE TABLE monitors (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id             UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    type                VARCHAR(50) NOT NULL,
    url                 TEXT,
    ip_address          VARCHAR(255),
    port                INTEGER,
    interval_seconds    INTEGER NOT NULL DEFAULT 300,
    timeout_seconds     INTEGER NOT NULL DEFAULT 30,

    -- HTTP options
    http_method         VARCHAR(10) DEFAULT 'GET',
    http_headers        JSONB DEFAULT '{}',
    http_body           TEXT,
    http_auth_type      VARCHAR(20),
    http_username       VARCHAR(255),
    http_password_enc   TEXT,
    expected_status_codes INTEGER[] DEFAULT '{200}',
    follow_redirects    BOOLEAN DEFAULT true,

    -- Keyword options
    keyword             TEXT,
    keyword_type        VARCHAR(20),

    -- API monitoring options
    api_assertions      JSONB DEFAULT '[]',

    -- UDP options
    udp_data            TEXT,
    udp_expected        TEXT,

    -- SSL options
    ssl_expiry_reminder INTEGER DEFAULT 30,

    -- DNS options
    dns_record_type     VARCHAR(10),
    dns_expected_value  TEXT,

    -- Domain options
    domain_expiry_reminder INTEGER DEFAULT 30,

    -- Location-specific
    monitoring_regions  TEXT[] DEFAULT '{na,eu,apac,au}',

    -- Slow response alert
    slow_threshold_ms   INTEGER,

    -- Heartbeat
    heartbeat_token     VARCHAR(64),
    heartbeat_grace_sec INTEGER DEFAULT 300,
    heartbeat_last_ping TIMESTAMPTZ,

    -- Status tracking
    status              VARCHAR(20) NOT NULL DEFAULT 'paused',
    last_checked_at     TIMESTAMPTZ,
    last_response_ms    INTEGER,
    uptime_percentage   DECIMAL(7,4) DEFAULT 100.0000,
    total_checks        BIGINT DEFAULT 0,
    total_downtime_sec  BIGINT DEFAULT 0,

    created_by          UUID REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitors_team_id ON monitors(team_id);
CREATE INDEX idx_monitors_status ON monitors(status);
CREATE INDEX idx_monitors_type ON monitors(type);
CREATE INDEX idx_monitors_heartbeat_token ON monitors(heartbeat_token) WHERE heartbeat_token IS NOT NULL;

------------------------------------------------------------
-- Monitor Results (time-series data)
------------------------------------------------------------
CREATE TABLE monitor_results (
    id              UUID DEFAULT gen_random_uuid(),
    monitor_id      UUID NOT NULL,
    status          VARCHAR(20) NOT NULL,
    response_time_ms INTEGER,
    status_code     INTEGER,
    error_message   TEXT,
    region          VARCHAR(10),
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, checked_at)
) PARTITION BY RANGE (checked_at);

-- Create initial partitions (current month and next 3 months)
DO $$
DECLARE
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR i IN 0..3 LOOP
        start_date := date_trunc('month', CURRENT_DATE) + (i || ' months')::INTERVAL;
        end_date := start_date + '1 month'::INTERVAL;
        partition_name := 'monitor_results_' || to_char(start_date, 'YYYY_MM');
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF monitor_results FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;

CREATE INDEX idx_monitor_results_monitor_id ON monitor_results(monitor_id, checked_at DESC);

------------------------------------------------------------
-- Incidents
------------------------------------------------------------
CREATE TABLE incidents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id      UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at     TIMESTAMPTZ,
    duration_seconds INTEGER,
    reason          TEXT,
    root_cause      TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'ongoing',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_incidents_monitor_id ON incidents(monitor_id, started_at DESC);
CREATE INDEX idx_incidents_status ON incidents(status);

------------------------------------------------------------
-- Alert Contacts
------------------------------------------------------------
CREATE TABLE alert_contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    name        VARCHAR(255) NOT NULL,
    value       TEXT NOT NULL,
    config      JSONB DEFAULT '{}',
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_contacts_team_id ON alert_contacts(team_id);

------------------------------------------------------------
-- Monitor <-> Alert Contact association
------------------------------------------------------------
CREATE TABLE monitor_alert_contacts (
    monitor_id       UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    alert_contact_id UUID NOT NULL REFERENCES alert_contacts(id) ON DELETE CASCADE,
    threshold_seconds INTEGER DEFAULT 0,
    PRIMARY KEY (monitor_id, alert_contact_id)
);

------------------------------------------------------------
-- Status Pages
------------------------------------------------------------
CREATE TABLE status_pages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id         UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) UNIQUE NOT NULL,
    custom_domain   VARCHAR(255),
    logo_url        TEXT,
    primary_color   VARCHAR(7) DEFAULT '#10B981',
    is_password_protected BOOLEAN NOT NULL DEFAULT false,
    password_hash   VARCHAR(255),
    hide_from_search BOOLEAN NOT NULL DEFAULT false,
    announcement    TEXT,
    language        VARCHAR(10) DEFAULT 'en',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_pages_team_id ON status_pages(team_id);
CREATE INDEX idx_status_pages_slug ON status_pages(slug);

CREATE TABLE status_page_monitors (
    status_page_id  UUID NOT NULL REFERENCES status_pages(id) ON DELETE CASCADE,
    monitor_id      UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    sort_order      INTEGER DEFAULT 0,
    PRIMARY KEY (status_page_id, monitor_id)
);

------------------------------------------------------------
-- Maintenance Windows
------------------------------------------------------------
CREATE TABLE maintenance_windows (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id         UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    recurring       BOOLEAN NOT NULL DEFAULT false,
    recurrence_rule TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maintenance_windows_team_id ON maintenance_windows(team_id);

CREATE TABLE maintenance_window_monitors (
    maintenance_window_id UUID NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
    monitor_id            UUID NOT NULL REFERENCES monitors(id) ON DELETE CASCADE,
    PRIMARY KEY (maintenance_window_id, monitor_id)
);

------------------------------------------------------------
-- API Keys
------------------------------------------------------------
CREATE TABLE api_keys (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    key_hash    VARCHAR(255) NOT NULL,
    prefix      VARCHAR(10) NOT NULL,
    scopes      TEXT[] DEFAULT '{read}',
    last_used   TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_team_id ON api_keys(team_id);
CREATE INDEX idx_api_keys_prefix ON api_keys(prefix);

------------------------------------------------------------
-- Notification Log (for deduplication and audit)
------------------------------------------------------------
CREATE TABLE notification_log (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id       UUID REFERENCES incidents(id) ON DELETE CASCADE,
    alert_contact_id  UUID REFERENCES alert_contacts(id) ON DELETE SET NULL,
    type              VARCHAR(50) NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'sent',
    error_message     TEXT,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_log_incident_id ON notification_log(incident_id);
