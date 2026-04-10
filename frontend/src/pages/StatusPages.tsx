import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Plus, Trash2, Globe, ExternalLink } from 'lucide-react';
import { statusPagesAPI } from '../api/client';
import { StatusPage } from '../types';

export function StatusPages() {
  const queryClient = useQueryClient();
  const { data: pages = [], isLoading } = useQuery({ queryKey: ['status-pages'], queryFn: statusPagesAPI.list });
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', slug: '', primary_color: '#10B981', language: 'en' });

  const createPage = useMutation({
    mutationFn: (data: Partial<StatusPage>) => statusPagesAPI.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['status-pages'] }); setShowCreate(false); },
  });

  const deletePage = useMutation({
    mutationFn: (id: string) => statusPagesAPI.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['status-pages'] }),
  });

  if (isLoading) return <div className="loading-center"><div className="spinner" /></div>;

  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-6)' }}>
        <h1 className="page-title" style={{ marginBottom: 0 }}>Status Pages</h1>
        <button className="btn btn-primary" onClick={() => setShowCreate(true)}><Plus size={16} /> Create Status Page</button>
      </div>

      {pages.length === 0 ? (
        <div className="empty-state">
          <Globe size={48} />
          <h3>No status pages</h3>
          <p>Create a public status page to keep your users informed about service health.</p>
        </div>
      ) : (
        <div className="grid-2">
          {pages.map((page: StatusPage, i: number) => (
            <motion.div key={page.id} className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
              <div className="card-header">
                <div>
                  <div className="card-title">{page.name}</div>
                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginTop: 2 }}>/{page.slug}</div>
                </div>
                <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
                  <a href={`/status/${page.slug}`} target="_blank" rel="noopener" className="btn btn-ghost btn-sm"><ExternalLink size={14} /></a>
                  <button className="btn btn-ghost btn-sm" onClick={() => { if (confirm('Delete this status page?')) deletePage.mutate(page.id); }}><Trash2 size={14} /></button>
                </div>
              </div>
              <div style={{ display: 'flex', gap: 'var(--space-3)', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                <span>Color: <span style={{ color: page.primary_color }}>●</span> {page.primary_color}</span>
                <span>Language: {page.language}</span>
                {page.is_password_protected && <span className="badge badge-paused">Password Protected</span>}
              </div>
            </motion.div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="modal-overlay" onClick={() => setShowCreate(false)}>
          <motion.div className="modal" onClick={e => e.stopPropagation()} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}>
            <div className="modal-header">
              <h2 className="modal-title">Create Status Page</h2>
              <button className="modal-close" onClick={() => setShowCreate(false)}>✕</button>
            </div>
            <form onSubmit={e => { e.preventDefault(); createPage.mutate(form); }}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <input className="form-input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} required placeholder="My Service Status" />
              </div>
              <div className="form-group">
                <label className="form-label">Slug (URL path)</label>
                <input className="form-input" value={form.slug} onChange={e => setForm({ ...form, slug: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '-') })} required placeholder="my-service" />
              </div>
              <div className="form-group">
                <label className="form-label">Primary Color</label>
                <input type="color" className="form-input" value={form.primary_color} onChange={e => setForm({ ...form, primary_color: e.target.value })} style={{ height: 40, padding: 4 }} />
              </div>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'flex-end', marginTop: 'var(--space-6)' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreate(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={createPage.isPending}>Create</button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  );
}
