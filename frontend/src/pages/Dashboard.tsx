import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { monitorsAPI, incidentsAPI } from '../api/client';
import { Monitor, Incident } from '../types';
import { Activity, ArrowDown, ArrowUp, AlertTriangle, Clock } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, Area, AreaChart } from 'recharts';

export function Dashboard() {
  const { data: monitors = [] } = useQuery({ queryKey: ['monitors'], queryFn: monitorsAPI.list, refetchInterval: 30000 });
  const { data: incidentData } = useQuery({ queryKey: ['incidents'], queryFn: () => incidentsAPI.list(10) });
  const incidents = incidentData?.incidents || [];

  const stats = {
    total: monitors.length,
    up: monitors.filter((m: Monitor) => m.status === 'up').length,
    down: monitors.filter((m: Monitor) => m.status === 'down').length,
    degraded: monitors.filter((m: Monitor) => m.status === 'degraded').length,
    paused: monitors.filter((m: Monitor) => m.status === 'paused').length,
    avgUptime: monitors.length > 0
      ? (monitors.reduce((acc: number, m: Monitor) => acc + m.uptime_percentage, 0) / monitors.length).toFixed(2)
      : '100.00',
  };

  const container = {
    hidden: { opacity: 0 },
    show: { opacity: 1, transition: { staggerChildren: 0.05 } },
  };

  const item = {
    hidden: { opacity: 0, y: 20 },
    show: { opacity: 1, y: 0 },
  };

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>

      {/* Stats Grid */}
      <motion.div className="stats-grid" variants={container} initial="hidden" animate="show">
        <motion.div className="stat-card total" variants={item}>
          <div className="stat-label">Total Monitors</div>
          <div className="stat-value">{stats.total}</div>
        </motion.div>
        <motion.div className="stat-card up" variants={item}>
          <div className="stat-label flex items-center gap-2">
            <span className="pulse-dot up" /> Up
          </div>
          <div className="stat-value">{stats.up}</div>
        </motion.div>
        <motion.div className="stat-card down" variants={item}>
          <div className="stat-label flex items-center gap-2">
            <span className="pulse-dot down" /> Down
          </div>
          <div className="stat-value">{stats.down}</div>
        </motion.div>
        <motion.div className="stat-card degraded" variants={item}>
          <div className="stat-label">Degraded</div>
          <div className="stat-value">{stats.degraded}</div>
        </motion.div>
      </motion.div>

      {/* Avg Uptime Banner */}
      <motion.div
        className="card"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        style={{
          marginBottom: 'var(--space-8)',
          background: 'linear-gradient(135deg, rgba(16, 185, 129, 0.1), rgba(99, 102, 241, 0.05))',
          borderColor: 'var(--color-up-border)',
          display: 'flex',
          alignItems: 'center',
          gap: 'var(--space-6)',
        }}
      >
        <div style={{
          width: 64,
          height: 64,
          borderRadius: 'var(--radius-lg)',
          background: 'var(--color-up-bg)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          border: '1px solid var(--color-up-border)',
        }}>
          <Activity size={28} style={{ color: 'var(--color-up)' }} />
        </div>
        <div>
          <div style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', marginBottom: 4 }}>
            Average Uptime
          </div>
          <div style={{ fontSize: 'var(--font-size-4xl)', fontWeight: 800, color: 'var(--color-up)', letterSpacing: '-0.025em' }}>
            {stats.avgUptime}%
          </div>
        </div>
      </motion.div>

      <div className="grid-2">
        {/* Recent Incidents */}
        <motion.div
          className="card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <div className="card-header">
            <h2 className="card-title">Recent Incidents</h2>
            <AlertTriangle size={18} style={{ color: 'var(--color-text-muted)' }} />
          </div>
          {incidents.length === 0 ? (
            <div className="empty-state">
              <p>No recent incidents 🎉</p>
            </div>
          ) : (
            <div className="incident-timeline">
              {incidents.slice(0, 5).map((inc: Incident) => (
                <div key={inc.id} className={`incident-item ${inc.status === 'resolved' ? 'resolved' : ''}`}>
                  <div className="incident-header">
                    <span className="incident-monitor">{inc.monitor_name || 'Unknown'}</span>
                    <span className={`badge badge-${inc.status === 'resolved' ? 'up' : 'down'}`}>
                      {inc.status}
                    </span>
                  </div>
                  <div className="incident-reason">{inc.reason}</div>
                  <div className="incident-time">
                    {new Date(inc.started_at).toLocaleString()}
                    {inc.duration_seconds && ` · ${formatDuration(inc.duration_seconds)}`}
                  </div>
                </div>
              ))}
            </div>
          )}
        </motion.div>

        {/* Monitor Status Overview */}
        <motion.div
          className="card"
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
        >
          <div className="card-header">
            <h2 className="card-title">Monitor Overview</h2>
            <Clock size={18} style={{ color: 'var(--color-text-muted)' }} />
          </div>
          {monitors.length === 0 ? (
            <div className="empty-state">
              <p>Add your first monitor to get started</p>
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
              {monitors.slice(0, 8).map((m: Monitor) => (
                <div key={m.id} style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: 'var(--space-2) var(--space-3)',
                  borderRadius: 'var(--radius-md)',
                  background: 'var(--color-bg-primary)',
                }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    <span className={`pulse-dot ${m.status}`} />
                    <span style={{ fontSize: 'var(--font-size-sm)', fontWeight: 500 }}>{m.name}</span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                    <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
                      {m.last_response_ms != null ? `${m.last_response_ms}ms` : '—'}
                    </span>
                    <span className={`badge badge-${m.status}`}>{m.uptime_percentage.toFixed(1)}%</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>
    </div>
  );
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  return `${Math.floor(seconds / 86400)}d ${Math.floor((seconds % 86400) / 3600)}h`;
}
