import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import StatusBadge from '../components/StatusBadge';
import UptimeBarChart from '../components/UptimeBarChart';
import { monitorApi } from '../api/client';
import { Activity, AlertTriangle, CheckCircle2, RefreshCw, ShieldAlert, ShieldCheck } from 'lucide-react';

export default function PublicStatusPage() {
  const { tenant_id } = useParams();
  const [statusData, setStatusData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchStatus = async () => {
    try {
      const data = await monitorApi.getPublicStatus(tenant_id);
      setStatusData(data);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 15000);
    return () => clearInterval(interval);
  }, [tenant_id]);

  if (loading) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-muted)' }}>
        Loading System Status...
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '20px' }}>
        <div className="glass-card" style={{ padding: '36px', textAlign: 'center', maxWidth: '440px' }}>
          <ShieldAlert size={48} color="var(--danger)" style={{ marginBottom: '16px' }} />
          <h2 style={{ fontSize: '1.4rem', marginBottom: '8px' }}>Status Page Not Found</h2>
          <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>{error}</p>
        </div>
      </div>
    );
  }

  const { tenant_name, system_status, monitors = [] } = statusData;

  let bannerBg = 'rgba(16, 185, 129, 0.12)';
  let bannerBorder = 'rgba(16, 185, 129, 0.3)';
  let bannerColor = 'var(--success)';
  let bannerIcon = <CheckCircle2 size={24} />;
  let bannerText = 'All Systems Operational';

  if (system_status === 'DEGRADED') {
    bannerBg = 'rgba(245, 158, 11, 0.12)';
    bannerBorder = 'rgba(245, 158, 11, 0.3)';
    bannerColor = 'var(--warning)';
    bannerIcon = <AlertTriangle size={24} />;
    bannerText = 'Partial System Degradation';
  } else if (system_status === 'OUTAGE') {
    bannerBg = 'rgba(244, 63, 94, 0.12)';
    bannerBorder = 'rgba(244, 63, 94, 0.3)';
    bannerColor = 'var(--danger)';
    bannerIcon = <ShieldAlert size={24} />;
    bannerText = 'Major System Outage';
  }

  return (
    <div style={{ minHeight: '100vh', padding: '40px 20px' }}>
      <div style={{ maxWidth: '840px', margin: '0 auto' }}>
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '32px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div style={{ width: '40px', height: '40px', borderRadius: '12px', background: 'linear-gradient(135deg, #38bdf8 0%, #10b981 100%)', display: 'flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 0 20px rgba(56, 189, 248, 0.4)' }}>
              <Activity size={22} color="#000" strokeWidth={2.5} />
            </div>
            <div>
              <h1 style={{ fontSize: '1.6rem', fontWeight: 800 }}>{tenant_name}</h1>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>System Status & Uptime</span>
            </div>
          </div>

          <button onClick={fetchStatus} className="btn btn-secondary" style={{ padding: '8px 12px' }}>
            <RefreshCw size={14} />
            Refresh
          </button>
        </div>

        {/* Global Banner */}
        <div style={{ background: bannerBg, border: `1px solid ${bannerBorder}`, color: bannerColor, padding: '20px 24px', borderRadius: 'var(--radius-md)', display: 'flex', alignItems: 'center', gap: '14px', marginBottom: '36px', boxShadow: `0 0 20px -5px ${bannerBg}` }}>
          {bannerIcon}
          <div>
            <h3 style={{ fontSize: '1.15rem', fontWeight: 700 }}>{bannerText}</h3>
            <p style={{ fontSize: '0.85rem', opacity: 0.9, marginTop: '2px' }}>
              Real-time monitor status updated automatically every 15 seconds.
            </p>
          </div>
        </div>

        {/* Services List */}
        <h2 style={{ fontSize: '1.2rem', fontWeight: 700, marginBottom: '16px' }}>Services Availability</h2>

        <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
          {monitors.map((m) => (
            <div key={m.id} className="glass-card" style={{ padding: '20px 24px' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '8px' }}>
                <div>
                  <h4 style={{ fontSize: '1.05rem', fontWeight: 600 }}>{m.name}</h4>
                  <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>{m.url}</span>
                </div>
                <StatusBadge status={m.status} />
              </div>

              <UptimeBarChart checkResults={m.check_results} />
            </div>
          ))}
        </div>

        {/* Footer */}
        <footer style={{ textAlign: 'center', marginTop: '48px', color: 'var(--text-dim)', fontSize: '0.8rem' }}>
          Powered by <strong style={{ color: 'var(--text-muted)' }}>StatusPage Operator</strong> — Kubernetes-Native Uptime Monitoring
        </footer>
      </div>
    </div>
  );
}
