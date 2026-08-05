import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { Activity, ExternalLink, LogOut, ShieldCheck } from 'lucide-react';

export default function Navbar() {
  const { user, tenant, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="glass-card" style={{ borderRadius: 0, borderTop: 'none', borderLeft: 'none', borderRight: 'none', marginBottom: '32px' }}>
      <div style={{ maxWidth: '1200px', margin: '0 auto', padding: '16px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Link to="/dashboard" style={{ display: 'flex', alignItems: 'center', gap: '10px', textDecoration: 'none', color: 'var(--text-main)' }}>
          <div style={{ width: '36px', height: '36px', borderRadius: '10px', background: 'linear-gradient(135deg, #38bdf8 0%, #10b981 100%)', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 0 16px rgba(56, 189, 248, 0.4)' }}>
            <Activity size={20} color="#000" strokeWidth={2.5} />
          </div>
          <div>
            <span style={{ fontFamily: 'var(--font-heading)', fontWeight: 700, fontSize: '1.2rem', letterSpacing: '-0.02em' }}>StatusPage</span>
            <span style={{ display: 'block', fontSize: '0.7rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', marginTop: '-2px' }}>Kubernetes Native</span>
          </div>
        </Link>

        {user && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '20px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '6px 14px', borderRadius: 'var(--radius-full)', background: 'rgba(255,255,255,0.04)', border: '1px solid var(--border-color)', fontSize: '0.85rem' }}>
              <ShieldCheck size={16} color="var(--primary)" />
              <span style={{ fontWeight: 600 }}>{tenant?.name}</span>
              <span style={{ color: 'var(--text-dim)', fontSize: '0.75rem', textTransform: 'uppercase', background: 'rgba(56,189,248,0.15)', padding: '2px 6px', borderRadius: '4px' }}>{tenant?.tier}</span>
            </div>

            <Link
              to={`/status/${tenant?.id}`}
              target="_blank"
              className="btn btn-secondary"
              style={{ fontSize: '0.85rem', padding: '8px 14px' }}
            >
              <ExternalLink size={15} />
              Public Status Page
            </Link>

            <button onClick={handleLogout} className="btn btn-danger" style={{ padding: '8px 12px' }} title="Logout">
              <LogOut size={16} />
            </button>
          </div>
        )}
      </div>
    </header>
  );
}
