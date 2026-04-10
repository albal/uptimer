import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Plus, Search, Filter } from 'lucide-react';
import { useMonitors, useCreateMonitor } from '../hooks/useMonitors';
import type { Monitor, CreateMonitorRequest } from '../types';
import { MonitorType, MONITOR_TYPE_LABELS, STATUS_LABELS } from '../types';

export function Monitors() {
  const navigate = useNavigate();
  const { data: monitors = [], isLoading } = useMonitors();
  const createMonitor = useCreateMonitor();
  const [showCreate, setShowCreate] = useState(false);
  const [search, setSearch] = useState('');
  const [filterType, setFilterType] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  const filtered = monitors.filter((m: Monitor) => {
    if (search && !m.name.toLowerCase().includes(search.toLowerCase())) return false;
    if (filterType !== 'all' && m.type !== filterType) return false;
    if (filterStatus !== 'all' && m.status !== filterStatus) return false;
    return true;
  });

  const [form, setForm] = useState<CreateMonitorRequest>({
    name: '', type: 'http', url: '', interval_seconds: 300, timeout_seconds: 30,
    follow_redirects: true, expected_status_codes: [200],
  });

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    try {
      await createMonitor.mutateAsync(form);
      setShowCreate(false);
      setForm({ name: '', type: 'http', url: '', interval_seconds: 300, timeout_seconds: 30, follow_redirects: true, expected_status_codes: [200] });
    } catch (err: any) {
      alert(err.message);
    }
  }

  if (isLoading) {
    return <div className="loading-center"><div className="spinner" /></div>;
  }

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-6)' }}>
        <h1 className="page-title" style={{ marginBottom: 0 }}>Monitors</h1>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}>
          <Plus size={16} /> Add Monitor
        </button>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', gap: 'var(--space-3)', marginBottom: 'var(--space-6)', flexWrap: 'wrap' }}>
        <div style={{ flex: 1, minWidth: 200, display: 'flex', alignItems: 'center', gap: 'var(--space-2)', background: 'var(--color-bg-tertiary)', borderRadius: 'var(--radius-md)', padding: '0 var(--space-3)', border: '1px solid var(--color-border)' }}>
          <Search size={16} style={{ color: 'var(--color-text-muted)' }} />
          <input type="text" placeholder="Search monitors..." value={search} onChange={e => setSearch(e.target.value)} className="form-input" style={{ border: 'none', background: 'transparent', padding: 'var(--space-2) 0' }} />
        </div>
        <select className="form-select" value={filterType} onChange={e => setFilterType(e.target.value)} style={{ width: 'auto' }}>
          <option value="all">All Types</option>
          {Object.entries(MONITOR_TYPE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <select className="form-select" value={filterStatus} onChange={e => setFilterStatus(e.target.value)} style={{ width: 'auto' }}>
          <option value="all">All Statuses</option>
          {Object.entries(STATUS_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
      </div>

      {/* Monitor Grid */}
      {filtered.length === 0 ? (
        <div className="empty-state">
          <MonitorIcon size={48} />
          <h3>No monitors found</h3>
          <p>Create your first monitor to start tracking uptime</p>
        </div>
      ) : (
        <motion.div className="grid-2" initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
          {filtered.map((m: Monitor, i: number) => (
            <motion.div
              key={m.id}
              className="monitor-card"
              onClick={() => navigate(`/monitors/${m.id}`)}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.03 }}
            >
              <div className="monitor-card-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                  <span className={`pulse-dot ${m.status}`} />
                  <span className="monitor-card-name">{m.name}</span>
                </div>
                <span className={`badge badge-${m.status}`}>{STATUS_LABELS[m.status]}</span>
              </div>
              <div className="monitor-card-url">{m.url || m.ip_address || '—'}</div>

              {/* Simple uptime bar */}
              <div className="uptime-bar" title={`${m.uptime_percentage.toFixed(2)}% uptime`}>
                {Array.from({ length: 30 }, (_, i) => (
                  <div key={i} className={`bar ${m.status === 'paused' ? 'no-data' : m.status}`} style={{ height: '100%' }} />
                ))}
              </div>

              <div className="monitor-card-stats">
                <span>{MONITOR_TYPE_LABELS[m.type]}</span>
                <span>{m.uptime_percentage.toFixed(2)}% uptime</span>
                <span>{m.last_response_ms != null ? `${m.last_response_ms}ms` : '—'}</span>
              </div>
            </motion.div>
          ))}
        </motion.div>
      )}

      {/* Create Modal */}
      {showCreate && (
        <div className="modal-overlay" onClick={() => setShowCreate(false)}>
          <motion.div
            className="modal"
            onClick={e => e.stopPropagation()}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
          >
            <div className="modal-header">
              <h2 className="modal-title">Add Monitor</h2>
              <button className="modal-close" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <input className="form-input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} required placeholder="My Website" />
              </div>
              <div className="form-group">
                <label className="form-label">Type</label>
                <select className="form-select" value={form.type} onChange={e => setForm({ ...form, type: e.target.value as MonitorType })}>
                  {Object.entries(MONITOR_TYPE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
              </div>
              <div className="form-group">
                <label className="form-label">{form.type === 'ping' ? 'Host / IP' : 'URL'}</label>
                <input className="form-input" value={form.url || ''} onChange={e => setForm({ ...form, url: e.target.value })} placeholder={form.type === 'ping' ? '192.168.1.1' : 'https://example.com'} />
              </div>
              {(form.type === 'port' || form.type === 'udp') && (
                <div className="form-group">
                  <label className="form-label">Port</label>
                  <input type="number" className="form-input" value={form.port || ''} onChange={e => setForm({ ...form, port: parseInt(e.target.value) || undefined })} placeholder="80" />
                </div>
              )}
              {form.type === 'keyword' && (
                <>
                  <div className="form-group">
                    <label className="form-label">Keyword</label>
                    <input className="form-input" value={form.keyword || ''} onChange={e => setForm({ ...form, keyword: e.target.value })} placeholder="Expected text on page" />
                  </div>
                  <div className="form-group">
                    <label className="form-label">Keyword Type</label>
                    <select className="form-select" value={form.keyword_type || 'exists'} onChange={e => setForm({ ...form, keyword_type: e.target.value })}>
                      <option value="exists">Should Exist</option>
                      <option value="not_exists">Should Not Exist</option>
                    </select>
                  </div>
                </>
              )}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-4)' }}>
                <div className="form-group">
                  <label className="form-label">Check Interval (sec)</label>
                  <input type="number" className="form-input" value={form.interval_seconds || 300} onChange={e => setForm({ ...form, interval_seconds: parseInt(e.target.value) })} min={30} />
                </div>
                <div className="form-group">
                  <label className="form-label">Timeout (sec)</label>
                  <input type="number" className="form-input" value={form.timeout_seconds || 30} onChange={e => setForm({ ...form, timeout_seconds: parseInt(e.target.value) })} min={1} max={120} />
                </div>
              </div>
              {form.type === 'ssl' && (
                <div className="form-group">
                  <label className="form-label">Remind before expiry (days)</label>
                  <input type="number" className="form-input" value={form.ssl_expiry_reminder || 30} onChange={e => setForm({ ...form, ssl_expiry_reminder: parseInt(e.target.value) })} />
                </div>
              )}
              <div className="form-group">
                <label className="form-label">Slow Response Threshold (ms, optional)</label>
                <input type="number" className="form-input" value={form.slow_threshold_ms || ''} onChange={e => setForm({ ...form, slow_threshold_ms: e.target.value ? parseInt(e.target.value) : undefined })} placeholder="e.g. 2000" />
              </div>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'flex-end', marginTop: 'var(--space-6)' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreate(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={createMonitor.isPending}>
                  {createMonitor.isPending ? 'Creating...' : 'Create Monitor'}
                </button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  );
}

function MonitorIcon({ size }: { size: number }) {
  return <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="2" y="3" width="20" height="14" rx="2" /><line x1="8" y1="21" x2="16" y2="21" /><line x1="12" y1="17" x2="12" y2="21" /></svg>;
}
