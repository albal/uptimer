import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { alertContactsAPI } from '../api/client';
import { INTEGRATION_TYPES, AlertContact } from '../types';
import { Plus, Trash2, Mail, Hash, Users, MessageCircle, Send, AlertTriangle, Globe, MessageSquare, Smartphone, Bell, Zap } from 'lucide-react';

const iconMap: Record<string, React.ReactNode> = {
  Mail: <Mail size={20} />, Hash: <Hash size={20} />, Users: <Users size={20} />,
  MessageCircle: <MessageCircle size={20} />, Send: <Send size={20} />,
  AlertTriangle: <AlertTriangle size={20} />, Globe: <Globe size={20} />,
  MessageSquare: <MessageSquare size={20} />, Smartphone: <Smartphone size={20} />,
  Bell: <Bell size={20} />, Zap: <Zap size={20} />,
};

export function Integrations() {
  const queryClient = useQueryClient();
  const { data: contacts = [] } = useQuery({ queryKey: ['alert-contacts'], queryFn: alertContactsAPI.list });
  const [showAdd, setShowAdd] = useState(false);
  const [selectedType, setSelectedType] = useState('');
  const [form, setForm] = useState({ name: '', value: '' });

  const createContact = useMutation({
    mutationFn: (data: Partial<AlertContact>) => alertContactsAPI.create(data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['alert-contacts'] }); setShowAdd(false); setSelectedType(''); },
  });

  const deleteContact = useMutation({
    mutationFn: (id: string) => alertContactsAPI.delete(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['alert-contacts'] }),
  });

  return (
    <div>
      <h1 className="page-title">Integrations</h1>
      <p className="page-subtitle">Configure notification channels to receive alerts when monitors go down.</p>

      {/* Active Integrations */}
      {contacts.length > 0 && (
        <div style={{ marginBottom: 'var(--space-8)' }}>
          <h2 style={{ fontSize: 'var(--font-size-lg)', fontWeight: 600, marginBottom: 'var(--space-4)' }}>Active Channels</h2>
          <div className="grid-2">
            {contacts.map((c: AlertContact) => (
              <div key={c.id} className="integration-card">
                <div className="integration-icon">
                  {iconMap[INTEGRATION_TYPES.find(t => t.type === c.type)?.icon || 'Globe'] || <Globe size={20} />}
                </div>
                <div style={{ flex: 1 }}>
                  <div className="integration-name">{c.name}</div>
                  <div className="integration-desc">{c.type} · {c.value.substring(0, 40)}{c.value.length > 40 ? '...' : ''}</div>
                </div>
                <button className="btn btn-ghost btn-sm" onClick={() => { if (confirm('Delete?')) deleteContact.mutate(c.id); }}>
                  <Trash2 size={14} />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Available Integrations */}
      <h2 style={{ fontSize: 'var(--font-size-lg)', fontWeight: 600, marginBottom: 'var(--space-4)' }}>Available Integrations</h2>
      <div className="grid-3">
        {INTEGRATION_TYPES.map((integration, i) => (
          <motion.div
            key={integration.type}
            className="integration-card"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.03 }}
            onClick={() => { setSelectedType(integration.type); setShowAdd(true); setForm({ name: integration.name, value: '' }); }}
          >
            <div className="integration-icon">
              {iconMap[integration.icon] || <Globe size={20} />}
            </div>
            <div>
              <div className="integration-name">{integration.name}</div>
              <div className="integration-desc">{integration.description}</div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Add Modal */}
      {showAdd && (
        <div className="modal-overlay" onClick={() => setShowAdd(false)}>
          <motion.div className="modal" onClick={e => e.stopPropagation()} initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}>
            <div className="modal-header">
              <h2 className="modal-title">Add {INTEGRATION_TYPES.find(t => t.type === selectedType)?.name}</h2>
              <button className="modal-close" onClick={() => setShowAdd(false)}>✕</button>
            </div>
            <form onSubmit={e => { e.preventDefault(); createContact.mutate({ type: selectedType, name: form.name, value: form.value }); }}>
              <div className="form-group">
                <label className="form-label">Name</label>
                <input className="form-input" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} required />
              </div>
              <div className="form-group">
                <label className="form-label">
                  {selectedType === 'email' ? 'Email Address' :
                   selectedType === 'telegram' ? 'Chat ID' :
                   selectedType === 'pagerduty' ? 'Routing Key' :
                   selectedType === 'pushbullet' || selectedType === 'pushover' ? 'API Key / User Key' :
                   'Webhook URL'}
                </label>
                <input className="form-input" value={form.value} onChange={e => setForm({ ...form, value: e.target.value })} required
                  placeholder={selectedType === 'email' ? 'alerts@example.com' : 'https://hooks.example.com/...'}
                />
              </div>
              <div style={{ display: 'flex', gap: 'var(--space-3)', justifyContent: 'flex-end', marginTop: 'var(--space-6)' }}>
                <button type="button" className="btn btn-secondary" onClick={() => setShowAdd(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={createContact.isPending}>Add Integration</button>
              </div>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  );
}
