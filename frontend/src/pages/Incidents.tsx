import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { incidentsAPI } from '../api/client';
import { Incident } from '../types';

export function Incidents() {
  const { data, isLoading } = useQuery({ queryKey: ['incidents'], queryFn: () => incidentsAPI.list(100) });
  const incidents = data?.incidents || [];

  if (isLoading) return <div className="loading-center"><div className="spinner" /></div>;

  return (
    <div>
      <h1 className="page-title">Incidents</h1>
      <p className="page-subtitle">View all downtime incidents across your monitors.</p>

      {incidents.length === 0 ? (
        <div className="empty-state">
          <h3>No incidents recorded</h3>
          <p>All your monitors are running smoothly 🎉</p>
        </div>
      ) : (
        <div className="incident-timeline">
          {incidents.map((inc: Incident, i: number) => (
            <motion.div
              key={inc.id}
              className={`incident-item ${inc.status === 'resolved' ? 'resolved' : ''}`}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: i * 0.03 }}
            >
              <div className="incident-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                  <span className={`pulse-dot ${inc.status === 'ongoing' ? 'down' : 'up'}`} />
                  <span className="incident-monitor">{inc.monitor_name || 'Unknown Monitor'}</span>
                </div>
                <span className={`badge badge-${inc.status === 'resolved' ? 'up' : 'down'}`}>{inc.status}</span>
              </div>
              <div className="incident-reason">{inc.reason || 'No details available'}</div>
              <div className="incident-time">
                Started: {new Date(inc.started_at).toLocaleString()}
                {inc.resolved_at && ` · Resolved: ${new Date(inc.resolved_at).toLocaleString()}`}
                {inc.duration_seconds != null && ` · Duration: ${formatDuration(inc.duration_seconds)}`}
              </div>
            </motion.div>
          ))}
        </div>
      )}
    </div>
  );
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}
