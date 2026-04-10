import React from 'react';
import { useAuth } from '../../hooks/useAuth';
import { LogOut, Bell, Search } from 'lucide-react';

export function Header() {
  const { logout } = useAuth();

  return (
    <header style={{
      position: 'fixed',
      top: 0,
      left: 'var(--sidebar-width)',
      right: 0,
      height: 'var(--header-height)',
      background: 'rgba(10, 14, 26, 0.8)',
      backdropFilter: 'blur(12px)',
      borderBottom: '1px solid var(--color-border)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      padding: '0 var(--space-6)',
      zIndex: 99,
    }}>
      {/* Search */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 'var(--space-2)',
        background: 'var(--color-bg-tertiary)',
        borderRadius: 'var(--radius-md)',
        padding: 'var(--space-2) var(--space-3)',
        border: '1px solid var(--color-border)',
        minWidth: 280,
      }}>
        <Search size={16} style={{ color: 'var(--color-text-muted)' }} />
        <input
          type="text"
          placeholder="Search monitors..."
          style={{
            background: 'transparent',
            border: 'none',
            outline: 'none',
            color: 'var(--color-text-primary)',
            fontSize: 'var(--font-size-sm)',
            width: '100%',
          }}
        />
        <kbd style={{
          padding: '2px 6px',
          fontSize: '10px',
          background: 'var(--color-bg-primary)',
          borderRadius: 'var(--radius-sm)',
          color: 'var(--color-text-muted)',
          border: '1px solid var(--color-border)',
        }}>⌘K</kbd>
      </div>

      {/* Actions */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
        <button className="btn btn-ghost" style={{ padding: 'var(--space-2)' }}>
          <Bell size={18} />
        </button>
        <button className="btn btn-ghost" onClick={logout} style={{ padding: 'var(--space-2)' }}>
          <LogOut size={18} />
        </button>
      </div>
    </header>
  );
}
