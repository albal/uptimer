import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, Pause, Play, Trash2, ExternalLink } from 'lucide-react';
import { useMonitor, useMonitorResults, usePauseMonitor, useResumeMonitor, useDeleteMonitor } from '../hooks/useMonitors';
import { MONITOR_TYPE_LABELS, STATUS_LABELS, MonitorResult } from '../types';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';

export function MonitorDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: monitor, isLoading } = useMonitor(id!);
  const { data: resultsData } = useMonitorResults(id!, 200);
  const pauseMonitor = usePauseMonitor();
  const resumeMonitor = useResumeMonitor();
  const deleteMonitor = useDeleteMonitor();

  const results = resultsData?.results || [];

  if (isLoading || !monitor) {
    return <div className="loading-center"><div className="spinner" /></div>;
  }

  const chartData = [...results].reverse().map((r: MonitorResult) => ({
    time: new Date(r.checked_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
    responseTime: r.response_time_ms || 0,
    status: r.status,
  }));

  async function handlePause() {
    await pauseMonitor.mutateAsync(monitor!.id);
  }

  async function handleResume() {
    await resumeMonitor.mutateAsync(monitor!.id);
  }

  async function handleDelete() {
    if (!confirm('Are you sure you want to delete this monitor?')) return;
    await deleteMonitor.mutateAsync(monitor!.id);
    navigate('/monitors');
  }

  return (
    <div>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)', marginBottom: 'var(--space-6)' }}>
        <button className="btn btn-ghost" onClick={() => navigate('/monitors')}>
          <ArrowLeft size={18} />
        </button>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
            <span className={`pulse-dot ${monitor.status}`} />
            <h1 className="page-title" style={{ marginBottom: 0 }}>{monitor.name}</h1>
            <span className={`badge badge-${monitor.status}`}>{STATUS_LABELS[monitor.status]}</span>
          </div>
          <div style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-muted)', marginTop: 4 }}>
            {monitor.url || monitor.ip_address} · {MONITOR_TYPE_LABELS[monitor.type]} · Every {monitor.interval_seconds}s
          </div>
        </div>
        <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
          {monitor.status === 'paused' ? (
            <button className="btn btn-secondary" onClick={handleResume}><Play size={14} /> Resume</button>
          ) : (
            <button className="btn btn-secondary" onClick={handlePause}><Pause size={14} /> Pause</button>
          )}
          <button className="btn btn-danger" onClick={handleDelete}><Trash2 size={14} /></button>
        </div>
      </div>

      {/* Stats */}
      <div className="stats-grid" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="stat-card up">
          <div className="stat-label">Uptime</div>
          <div className="stat-value">{monitor.uptime_percentage.toFixed(2)}%</div>
        </div>
        <div className="stat-card total">
          <div className="stat-label">Last Response</div>
          <div className="stat-value">{monitor.last_response_ms != null ? `${monitor.last_response_ms}ms` : '—'}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Total Checks</div>
          <div className="stat-value" style={{ color: 'var(--color-text-primary)' }}>{monitor.total_checks.toLocaleString()}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Last Checked</div>
          <div className="stat-value" style={{ fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
            {monitor.last_checked_at ? new Date(monitor.last_checked_at).toLocaleString() : 'Never'}
          </div>
        </div>
      </div>

      {/* Response Time Chart */}
      <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} style={{ marginBottom: 'var(--space-6)' }}>
        <h2 className="card-title" style={{ marginBottom: 'var(--space-4)' }}>Response Time</h2>
        {chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={chartData}>
              <defs>
                <linearGradient id="colorResp" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#6366f1" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="time" stroke="#6b7280" fontSize={12} tickLine={false} axisLine={false} />
              <YAxis stroke="#6b7280" fontSize={12} tickLine={false} axisLine={false} tickFormatter={v => `${v}ms`} />
              <Tooltip
                contentStyle={{ background: '#1f2937', border: '1px solid #374151', borderRadius: 8, fontSize: 13 }}
                labelStyle={{ color: '#9ca3af' }}
                formatter={(v: number) => [`${v}ms`, 'Response Time']}
              />
              <Area type="monotone" dataKey="responseTime" stroke="#6366f1" fill="url(#colorResp)" strokeWidth={2} />
            </AreaChart>
          </ResponsiveContainer>
        ) : (
          <div className="empty-state"><p>No data yet</p></div>
        )}
      </motion.div>

      {/* Recent Results */}
      <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }}>
        <h2 className="card-title" style={{ marginBottom: 'var(--space-4)' }}>Recent Checks</h2>
        {results.length === 0 ? (
          <div className="empty-state"><p>No check results yet</p></div>
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Response Time</th>
                  <th>Status Code</th>
                  <th>Region</th>
                  <th>Checked At</th>
                  <th>Error</th>
                </tr>
              </thead>
              <tbody>
                {results.slice(0, 20).map((r: MonitorResult) => (
                  <tr key={r.id}>
                    <td><span className={`badge badge-${r.status}`}>{r.status}</span></td>
                    <td>{r.response_time_ms != null ? `${r.response_time_ms}ms` : '—'}</td>
                    <td>{r.status_code || '—'}</td>
                    <td>{r.region || '—'}</td>
                    <td style={{ fontSize: 'var(--font-size-xs)' }}>{new Date(r.checked_at).toLocaleString()}</td>
                    <td style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-down)', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.error_message || ''}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </motion.div>

      {/* Heartbeat Token */}
      {monitor.type === 'heartbeat' && monitor.heartbeat_token && (
        <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} style={{ marginTop: 'var(--space-6)' }}>
          <h2 className="card-title" style={{ marginBottom: 'var(--space-2)' }}>Heartbeat Endpoint</h2>
          <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', marginBottom: 'var(--space-3)' }}>
            Send a GET request to this URL from your cron job or scheduler:
          </p>
          <code style={{
            display: 'block',
            padding: 'var(--space-3)',
            background: 'var(--color-bg-primary)',
            borderRadius: 'var(--radius-md)',
            fontSize: 'var(--font-size-sm)',
            color: 'var(--color-up)',
            wordBreak: 'break-all',
          }}>
            {window.location.origin}/api/heartbeat/{monitor.heartbeat_token}
          </code>
        </motion.div>
      )}
    </div>
  );
}
