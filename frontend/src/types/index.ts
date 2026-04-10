export interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url?: string;
  oauth_provider: string;
  created_at: string;
}

export interface Team {
  id: string;
  name: string;
  owner_id: string;
  max_seats: number;
  max_monitors: number;
  created_at: string;
}

export interface TeamMember {
  team_id: string;
  user_id: string;
  role: string;
  joined_at: string;
  user?: User;
}

export type MonitorType = 'http' | 'ping' | 'port' | 'keyword' | 'api' | 'udp' | 'ssl' | 'dns' | 'domain' | 'heartbeat';
export type MonitorStatus = 'up' | 'down' | 'degraded' | 'paused' | 'pending';
export type IncidentStatus = 'ongoing' | 'resolved' | 'acknowledged';

export interface Monitor {
  id: string;
  team_id: string;
  name: string;
  type: MonitorType;
  url?: string;
  ip_address?: string;
  port?: number;
  interval_seconds: number;
  timeout_seconds: number;
  http_method?: string;
  http_headers?: Record<string, string>;
  http_body?: string;
  expected_status_codes?: number[];
  follow_redirects?: boolean;
  keyword?: string;
  keyword_type?: 'exists' | 'not_exists';
  api_assertions?: APIAssertion[];
  udp_data?: string;
  udp_expected?: string;
  ssl_expiry_reminder?: number;
  dns_record_type?: string;
  dns_expected_value?: string;
  domain_expiry_reminder?: number;
  monitoring_regions?: string[];
  slow_threshold_ms?: number;
  heartbeat_token?: string;
  heartbeat_grace_sec?: number;
  status: MonitorStatus;
  last_checked_at?: string;
  last_response_ms?: number;
  uptime_percentage: number;
  total_checks: number;
  created_at: string;
  updated_at: string;
}

export interface APIAssertion {
  source: 'status_code' | 'response_time' | 'header' | 'body';
  property: string;
  comparison: 'equals' | 'contains' | 'greater_than' | 'less_than';
  value: string;
}

export interface MonitorResult {
  id: string;
  monitor_id: string;
  status: MonitorStatus;
  response_time_ms?: number;
  status_code?: number;
  error_message?: string;
  region?: string;
  checked_at: string;
}

export interface Incident {
  id: string;
  monitor_id: string;
  started_at: string;
  resolved_at?: string;
  duration_seconds?: number;
  reason?: string;
  root_cause?: string;
  status: IncidentStatus;
  monitor_name?: string;
}

export interface AlertContact {
  id: string;
  team_id: string;
  type: string;
  name: string;
  value: string;
  config?: Record<string, unknown>;
  is_active: boolean;
  created_at: string;
}

export interface StatusPage {
  id: string;
  team_id: string;
  name: string;
  slug: string;
  custom_domain?: string;
  logo_url?: string;
  primary_color: string;
  is_password_protected: boolean;
  hide_from_search: boolean;
  announcement?: string;
  language: string;
  monitors?: StatusPageMonitor[];
  created_at: string;
}

export interface StatusPageMonitor {
  status_page_id: string;
  monitor_id: string;
  sort_order: number;
  monitor?: Monitor;
}

export interface MaintenanceWindow {
  id: string;
  team_id: string;
  name: string;
  start_time: string;
  end_time: string;
  recurring: boolean;
  recurrence_rule?: string;
  monitor_ids?: string[];
}

export interface APIKey {
  id: string;
  team_id: string;
  name: string;
  prefix: string;
  scopes: string[];
  last_used?: string;
  expires_at?: string;
  created_at: string;
  key?: string; // Only on creation
}

export interface CreateMonitorRequest {
  name: string;
  type: MonitorType;
  url?: string;
  ip_address?: string;
  port?: number;
  interval_seconds?: number;
  timeout_seconds?: number;
  http_method?: string;
  http_headers?: Record<string, string>;
  http_body?: string;
  expected_status_codes?: number[];
  follow_redirects?: boolean;
  keyword?: string;
  keyword_type?: string;
  api_assertions?: APIAssertion[];
  udp_data?: string;
  udp_expected?: string;
  ssl_expiry_reminder?: number;
  dns_record_type?: string;
  dns_expected_value?: string;
  domain_expiry_reminder?: number;
  monitoring_regions?: string[];
  slow_threshold_ms?: number;
  heartbeat_grace_sec?: number;
  alert_contact_ids?: string[];
}

export const MONITOR_TYPE_LABELS: Record<MonitorType, string> = {
  http: 'HTTP(S)',
  ping: 'Ping',
  port: 'Port',
  keyword: 'Keyword',
  api: 'API',
  udp: 'UDP',
  ssl: 'SSL Certificate',
  dns: 'DNS',
  domain: 'Domain Expiry',
  heartbeat: 'Heartbeat',
};

export const STATUS_LABELS: Record<MonitorStatus, string> = {
  up: 'Up',
  down: 'Down',
  degraded: 'Degraded',
  paused: 'Paused',
  pending: 'Pending',
};

export const INTEGRATION_TYPES = [
  { type: 'email', name: 'Email', description: 'Get alerts via email', icon: 'Mail' },
  { type: 'slack', name: 'Slack', description: 'Post alerts to Slack channels', icon: 'Hash' },
  { type: 'teams', name: 'Microsoft Teams', description: 'Send alerts to Teams channels', icon: 'Users' },
  { type: 'discord', name: 'Discord', description: 'Post alerts to Discord channels', icon: 'MessageCircle' },
  { type: 'telegram', name: 'Telegram', description: 'Send alerts via Telegram bot', icon: 'Send' },
  { type: 'pagerduty', name: 'PagerDuty', description: 'Trigger PagerDuty incidents', icon: 'AlertTriangle' },
  { type: 'webhook', name: 'Webhook', description: 'Send alerts to custom webhooks', icon: 'Globe' },
  { type: 'googlechat', name: 'Google Chat', description: 'Post alerts to Google Chat', icon: 'MessageSquare' },
  { type: 'pushbullet', name: 'Pushbullet', description: 'Push notifications via Pushbullet', icon: 'Smartphone' },
  { type: 'pushover', name: 'Pushover', description: 'Push notifications via Pushover', icon: 'Bell' },
  { type: 'mattermost', name: 'Mattermost', description: 'Post alerts to Mattermost', icon: 'MessageCircle' },
  { type: 'zapier', name: 'Zapier', description: 'Trigger Zapier automations', icon: 'Zap' },
] as const;
