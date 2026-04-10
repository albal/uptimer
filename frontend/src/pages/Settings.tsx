import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { teamAPI, apiKeysAPI } from '../api/client';
import { TeamMember, APIKey } from '../types';
import { Users, Key, Plus, Trash2, Copy, Shield } from 'lucide-react';
import { useState } from 'react';

export function Settings() {
  const queryClient = useQueryClient();
  const { data: team } = useQuery({ queryKey: ['team'], queryFn: teamAPI.get });
  const { data: members = [] } = useQuery({ queryKey: ['team-members'], queryFn: teamAPI.listMembers });
  const { data: apiKeys = [] } = useQuery({ queryKey: ['api-keys'], queryFn: apiKeysAPI.list });

  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('member');
  const [apiKeyName, setApiKeyName] = useState('');
  const [newKey, setNewKey] = useState<string | null>(null);

  const inviteMember = useMutation({
    mutationFn: () => teamAPI.inviteMember(inviteEmail, inviteRole),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['team-members'] }); setInviteEmail(''); },
  });

  const removeMember = useMutation({
    mutationFn: (userId: string) => teamAPI.removeMember(userId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['team-members'] }),
  });

  const createKey = useMutation({
    mutationFn: () => apiKeysAPI.create(apiKeyName, ['read', 'write']),
    onSuccess: (data) => { queryClient.invalidateQueries({ queryKey: ['api-keys'] }); setNewKey(data.key); setApiKeyName(''); },
  });

  const deleteKey = useMutation({
    mutationFn: (id: string) => apiKeysAPI.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }),
  });

  return (
    <div>
      <h1 className="page-title">Settings</h1>

      {/* Team Info */}
      <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} style={{ marginBottom: 'var(--space-6)' }}>
        <div className="card-header">
          <h2 className="card-title"><Shield size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />Team</h2>
        </div>
        {team && (
          <div style={{ display: 'flex', gap: 'var(--space-8)' }}>
            <div><span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>Name:</span> <strong>{team.name}</strong></div>
            <div><span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>Seats:</span> <strong>{members.length}/{team.max_seats}</strong></div>
            <div><span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>Max Monitors:</span> <strong>{team.max_monitors}</strong></div>
          </div>
        )}
      </motion.div>

      {/* Team Members */}
      <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} style={{ marginBottom: 'var(--space-6)' }}>
        <div className="card-header">
          <h2 className="card-title"><Users size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />Members</h2>
        </div>
        <div className="table-container" style={{ marginBottom: 'var(--space-4)' }}>
          <table>
            <thead><tr><th>User</th><th>Email</th><th>Role</th><th>Joined</th><th></th></tr></thead>
            <tbody>
              {members.map((m: TeamMember) => (
                <tr key={m.user_id}>
                  <td style={{ fontWeight: 500 }}>{m.user?.display_name || '—'}</td>
                  <td>{m.user?.email || '—'}</td>
                  <td><span className={`badge badge-${m.role === 'owner' ? 'up' : 'pending'}`}>{m.role}</span></td>
                  <td style={{ fontSize: 'var(--font-size-xs)' }}>{new Date(m.joined_at).toLocaleDateString()}</td>
                  <td>
                    {m.role !== 'owner' && (
                      <button className="btn btn-ghost btn-sm" onClick={() => { if (confirm('Remove member?')) removeMember.mutate(m.user_id); }}>
                        <Trash2 size={14} />
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <form onSubmit={e => { e.preventDefault(); inviteMember.mutate(); }} style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'flex-end' }}>
          <div className="form-group" style={{ flex: 1, marginBottom: 0 }}>
            <label className="form-label">Invite by email</label>
            <input className="form-input" value={inviteEmail} onChange={e => setInviteEmail(e.target.value)} placeholder="user@example.com" required />
          </div>
          <select className="form-select" value={inviteRole} onChange={e => setInviteRole(e.target.value)} style={{ width: 'auto' }}>
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
          <button type="submit" className="btn btn-primary" disabled={inviteMember.isPending}><Plus size={14} /> Invite</button>
        </form>
      </motion.div>

      {/* API Keys */}
      <motion.div className="card" initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }}>
        <div className="card-header">
          <h2 className="card-title"><Key size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />API Keys</h2>
        </div>

        {newKey && (
          <div style={{ padding: 'var(--space-4)', background: 'var(--color-up-bg)', border: '1px solid var(--color-up-border)', borderRadius: 'var(--radius-md)', marginBottom: 'var(--space-4)' }}>
            <div style={{ fontWeight: 600, marginBottom: 'var(--space-2)', color: 'var(--color-up)' }}>🔑 API Key Created</div>
            <div style={{ fontSize: 'var(--font-size-sm)', marginBottom: 'var(--space-2)' }}>Copy this key now — it won't be shown again:</div>
            <code style={{ display: 'block', padding: 'var(--space-2)', background: 'var(--color-bg-primary)', borderRadius: 'var(--radius-sm)', wordBreak: 'break-all', fontSize: 'var(--font-size-sm)' }}>{newKey}</code>
            <button className="btn btn-secondary btn-sm" style={{ marginTop: 'var(--space-2)' }}
              onClick={() => { navigator.clipboard.writeText(newKey); }}><Copy size={12} /> Copy</button>
          </div>
        )}

        {apiKeys.length > 0 && (
          <div className="table-container" style={{ marginBottom: 'var(--space-4)' }}>
            <table>
              <thead><tr><th>Name</th><th>Prefix</th><th>Scopes</th><th>Last Used</th><th></th></tr></thead>
              <tbody>
                {apiKeys.map((k: APIKey) => (
                  <tr key={k.id}>
                    <td style={{ fontWeight: 500 }}>{k.name}</td>
                    <td><code>{k.prefix}...</code></td>
                    <td>{k.scopes.join(', ')}</td>
                    <td style={{ fontSize: 'var(--font-size-xs)' }}>{k.last_used ? new Date(k.last_used).toLocaleString() : 'Never'}</td>
                    <td><button className="btn btn-ghost btn-sm" onClick={() => { if (confirm('Delete this API key?')) deleteKey.mutate(k.id); }}><Trash2 size={14} /></button></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <form onSubmit={e => { e.preventDefault(); createKey.mutate(); }} style={{ display: 'flex', gap: 'var(--space-3)', alignItems: 'flex-end' }}>
          <div className="form-group" style={{ flex: 1, marginBottom: 0 }}>
            <label className="form-label">Create API Key</label>
            <input className="form-input" value={apiKeyName} onChange={e => setApiKeyName(e.target.value)} placeholder="My API Key" required />
          </div>
          <button type="submit" className="btn btn-primary" disabled={createKey.isPending}><Key size={14} /> Create</button>
        </form>
      </motion.div>
    </div>
  );
}
