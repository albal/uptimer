import React from 'react';
import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Activity, CheckCircle, AlertCircle, Clock } from 'lucide-react';
import { StatusPage, Monitor } from '../types';

export function PublicStatusPage() {
  const { slug } = useParams<{ slug: string }>();
  
  // Custom fetcher for public API
  const { data: page, isLoading, error } = useQuery<StatusPage & { monitors_data: Monitor[] }>({
    queryKey: ['public-status', slug],
    queryFn: async () => {
      const res = await fetch(`/api/status/${slug}`);
      if (!res.ok) throw new Error('Status page not found');
      return res.json();
    },
  });

  if (isLoading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyItems: 'center', height: '100vh', background: 'var(--color-bg-primary)' }}>
        <div className="spinner" style={{ margin: '0 auto' }} />
      </div>
    );
  }

  if (error || !page) {
    return (
      <div className="login-page">
        <div className="login-card">
          <AlertCircle size={48} color="var(--color-down)" style={{ marginBottom: 16 }} />
          <h1>Status Page Not Found</h1>
          <p>The status page you are looking for does not exist or has been removed.</p>
          <a href="/" className="btn btn-primary" style={{ marginTop: 24 }}>Back to Home</a>
        </div>
      </div>
    );
  }

  const allUp = page.monitors_data?.every(m => m.status === 'up');
  const someDown = page.monitors_data?.some(m => m.status === 'down');

  return (
    <div className="status-page-public" style={{ background: 'var(--color-bg-primary)', minHeight: '100vh', color: 'var(--color-text-primary)' }}>
      <header className="status-page-header">
        {page.logo_url && <img src={page.logo_url} alt={page.name} style={{ height: 48, marginBottom: 16 }} />}
        <h1 style={{ fontSize: 'var(--font-size-3xl)', fontWeight: 800 }}>{page.name}</h1>
      </header>

      <div className={`status-page-overall ${allUp ? 'all-up' : someDown ? 'has-issues' : ''}`} 
           style={{ borderColor: allUp ? 'var(--color-up-border)' : 'var(--color-down-border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
          {allUp ? <CheckCircle size={24} /> : <AlertCircle size={24} />}
          <span>{allUp ? 'All Systems Operational' : someDown ? 'Partial System Outage' : 'System Issues Detected'}</span>
        </div>
      </div>

      {page.announcement && (
        <motion.div 
          className="card" 
          initial={{ opacity: 0, y: 10 }} 
          animate={{ opacity: 1, y: 0 }}
          style={{ marginBottom: 32, background: 'var(--color-accent-bg)', borderColor: 'var(--color-accent)' }}
        >
          <div style={{ fontWeight: 600, marginBottom: 4 }}>Announcement</div>
          <div style={{ fontSize: 'var(--font-size-sm)' }}>{page.announcement}</div>
        </motion.div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        {page.monitors_data?.map((m, i) => (
          <motion.div 
            key={m.id} 
            className="card"
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: i * 0.05 }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div style={{ fontWeight: 600 }}>{m.name}</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span className={`badge badge-${m.status}`}>{m.status.toUpperCase()}</span>
              </div>
            </div>
            
            {/* 90-day uptime bars (simulated for now) */}
            <div style={{ marginTop: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginBottom: 4 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  <Activity size={12} />
                  <span>Last 90 days</span>
                </div>
                <span>{m.uptime_percentage.toFixed(2)}% uptime</span>
              </div>
              <div className="uptime-bar">
                {Array.from({ length: 90 }, (_, i) => (
                  <div key={i} className={`bar ${m.status === 'up' ? 'up' : 'down'}`} style={{ height: 24, opacity: 0.2 + (Math.random() * 0.8) }} />
                ))}
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      <footer style={{ marginTop: 64, textAlign: 'center', paddingBottom: 32, opacity: 0.5 }}>
        <div style={{ fontSize: 'var(--font-size-xs)', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          <Clock size={12} />
          <span>Last updated: {new Date().toLocaleTimeString()}</span>
        </div>
        <div style={{ marginTop: 8, fontSize: 'var(--font-size-xs)' }}>
          Powered by <a href="/" style={{ color: 'var(--color-accent)', fontWeight: 600 }}>Uptimer</a>
        </div>
      </footer>
    </div>
  );
}
