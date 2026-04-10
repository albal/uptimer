import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { maintenanceAPI } from '../api/client';
import { MaintenanceWindow } from '../types';
import { Calendar, Plus, Trash2 } from 'lucide-react';
import { useState } from 'react';

export function Maintenance() {
  const queryClient = useQueryClient();
  const { data: windows = [], isLoading } = useQuery({ queryKey: ['maintenance'], queryFn: maintenanceAPI.list });
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', start_time: '', end_time: '' });

  const createWindow = useMutation({
    mutationFn: (data: Partial<MaintenanceWindow>) => maintenanceAPI.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['maintenance'] }); setShowCreate(false); },
  });

  const deleteWindow = useMutation({
    mutationFn: (id: string) => maintenanceAPI.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['maintenance'] }),
  });

  if (isLoading) return <div className="loading-center"><div className="spinner" /></div>;

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-6)' }}>
        <h1 className="page-title" style={{ marginBottom: 0 }}>Maintenance Windows</h1>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}><Plus size={16} /> Schedule Maintenance</button>
      </div>

      {windows.length === 0 ? (
        <div className="empty-state">
          <Calendar size={48} />
          <h3>No maintenance windows</h3>
          <p>Schedule maintenance windows to suppress alerts during planned downtime.</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
          {windows.map((mw: MaintenanceWindow, i: number) => {
            const isActive = new Date(mw.start_time) <= new Date() && new Date(mw.end_time) >= new Date();
            const isPast = new Date(mw.end_time) < new Date();
            return (
              <motion.div key={mw.id} className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}
                style={{ borderLeft: `3px solid ${isActive ? 'var(--color-degraded)' : isPast ? 'var(--color-text-muted)' : 'var(--color-accent)'}`, opacity: isPast ? 0.6 : 1 }}>
                <div className="card-header">
                  <div>
                    <div className="card-title">{mw.name}</div>
                    <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginTop: 4 }}>
                      {new Date(mw.start_time).toLocaleString()} → {new Date(mw.end_time).toLocaleString()}
                    </div>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                    {isActive && <span className="badge badge-degraded">Active</span>}
                    {isPast && <span className="badge badge-paused">Completed</span>}
                    {!isActive && !isPast && <span className="badge badge-pending">Upcoming</span>}
                    <button className="btn btn-ghost btn-sm" onClick={() => { if (confirm('Delete?')) deleteWindow.mutate(mw.id); }}><Trash2 size={14} /></button>
                  </div>
                </div>
              </motion.div>
            );
          })}
        </div>
      )}

      {showCreate && (
        <div className="modal-overlay" onClick={() => setShowCreate(false)}>
          <motion.div className="modal" onClick={e => e.stopPropagation()} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}>
            <div className="modal-header">
              <h2 className="modal-title">Schedule Maintenance</h2>
              <button className="modal-close" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <form onSubmit={e => { e.preventDefault(); createWindow.mutate({ name: form.name, start_time: new Date(form.start_time).toISOString(), end_time: new Date(form.end_time).toISOString() }); }}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <input className="form-input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} required placeholder="Database Migration" />
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-4)' }}>
                <div className="form-group">
                  <label className="form-label">Start Time</label>
                  <input type="datetime-local" className="form-input" value={form.start_time} onChange={e => setForm({ ...form, start_time: e.target.value })} required />
                </div>
                <div className="form-group">
                  <label className="form-label">End Time</label>
                  <input type="datetime-local" className="form-input" value={form.end_time} onChange={e => setForm({ ...form, end_time: e.target.value })} required />
                </div>
              </div>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'flex-end', marginTop: 'var(--space-6)' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreate(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={createWindow.isPending}>Schedule</button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  );
}
