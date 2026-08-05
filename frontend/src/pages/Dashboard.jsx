import React, { useEffect, useState } from 'react';
import Navbar from '../components/Navbar';
import StatusBadge from '../components/StatusBadge';
import UptimeBarChart from '../components/UptimeBarChart';
import CreateMonitorModal from '../components/CreateMonitorModal';
import { monitorApi } from '../api/client';
import { Activity, Clock, Globe, Plus, RefreshCw, Trash2, Server, ShieldCheck } from 'lucide-react';

export default function Dashboard() {
  const [monitors, setMonitors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const fetchMonitors = async () => {
    try {
      setRefreshing(true);
      const res = await monitorApi.list();
      setMonitors(res.monitors || []);
      setError(null);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchMonitors();
    const interval = setInterval(fetchMonitors, 10000); // Auto refresh every 10s
    return () => clearInterval(interval);
  }, []);

  const handleCreateMonitor = async (data) => {
    await monitorApi.create(data);
    await fetchMonitors();
  };

  const handleDeleteMonitor = async (id, name) => {
    if (!window.confirm(`Are you sure you want to delete monitor "${name}" and its Kubernetes CR?`)) return;
    try {
      await monitorApi.delete(id);
      setMonitors((prev) => prev.filter((m) => m.id !== id));
    } catch (err) {
      alert(`Delete failed: ${err.message}`);
    }
  };

  const totalMonitors = monitors.length;
  const upMonitors = monitors.filter((m) => m.status === 'UP').length;
  const uptimePct = totalMonitors > 0 ? ((upMonitors / totalMonitors) * 100).toFixed(1) : '100.0';

  return (
    <div style={{ minHeight: '100vh', paddingBottom: '60px' }}>
      <Navbar />

      <main style={{ maxWidth: '1200px', margin: '0 auto', padding: '0 24px' }}>
        {/* Header Stats Banner */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '20px', marginBottom: '32px' }}>
          <div className="glass-card" style={{ padding: '24px', display: 'flex', alignItems: 'center', gap: '16px' }}>
            <div style={{ padding: '12px', borderRadius: '12px', background: 'rgba(56,189,248,0.15)', color: 'var(--primary)' }}>
              <Activity size={24} />
            </div>
            <div>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Total Monitors</span>
              <h3 style={{ fontSize: '1.8rem', fontWeight: 700, lineHeight: 1.2 }}>{totalMonitors}</h3>
            </div>
          </div>

          <div className="glass-card" style={{ padding: '24px', display: 'flex', alignItems: 'center', gap: '16px' }}>
            <div style={{ padding: '12px', borderRadius: '12px', background: 'rgba(16,185,129,0.15)', color: 'var(--success)' }}>
              <ShieldCheck size={24} />
            </div>
            <div>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Overall Uptime</span>
              <h3 style={{ fontSize: '1.8rem', fontWeight: 700, lineHeight: 1.2, color: 'var(--success)' }}>{uptimePct}%</h3>
            </div>
          </div>

          <div className="glass-card" style={{ padding: '24px', display: 'flex', alignItems: 'center', gap: '16px' }}>
            <div style={{ padding: '12px', borderRadius: '12px', background: 'rgba(245,158,11,0.15)', color: 'var(--warning)' }}>
              <Server size={24} />
            </div>
            <div>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>Kubernetes CRDs</span>
              <h3 style={{ fontSize: '1.8rem', fontWeight: 700, lineHeight: 1.2 }}>{totalMonitors} Active</h3>
            </div>
          </div>
        </div>

        {/* Section Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '24px' }}>
          <div>
            <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Monitors Overview</h2>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>Real-time HTTP checks driven by Custom Resource controller</p>
          </div>

          <div style={{ display: 'flex', gap: '12px' }}>
            <button onClick={fetchMonitors} disabled={refreshing} className="btn btn-secondary" style={{ padding: '8px 12px' }}>
              <RefreshCw size={16} className={refreshing ? 'spin' : ''} />
            </button>

            <button onClick={() => setIsModalOpen(true)} className="btn btn-primary">
              <Plus size={18} />
              Add Monitor
            </button>
          </div>
        </div>

        {/* Error Alert */}
        {error && (
          <div style={{ background: 'rgba(244,63,94,0.12)', border: '1px solid rgba(244,63,94,0.3)', color: 'var(--danger)', padding: '14px 18px', borderRadius: 'var(--radius-sm)', marginBottom: '24px' }}>
            Failed to load monitors: {error}
          </div>
        )}

        {/* Monitors List Grid */}
        {loading ? (
          <div className="glass-card" style={{ padding: '40px', textAlign: 'center', color: 'var(--text-muted)' }}>
            Loading Kubernetes Monitor CRs...
          </div>
        ) : monitors.length === 0 ? (
          <div className="glass-card" style={{ padding: '60px', textAlign: 'center' }}>
            <Globe size={48} color="var(--text-dim)" style={{ marginBottom: '16px' }} />
            <h3 style={{ fontSize: '1.2rem', marginBottom: '8px' }}>No Monitors Configured</h3>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.9rem', marginBottom: '20px' }}>
              Add your first endpoint to start tracking availability & response latency.
            </p>
            <button onClick={() => setIsModalOpen(true)} className="btn btn-primary">
              <Plus size={18} />
              Create First Monitor
            </button>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(360px, 1fr))', gap: '24px' }}>
            {monitors.map((m) => (
              <div key={m.id} className="glass-card" style={{ padding: '24px', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
                <div>
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: '12px' }}>
                    <div>
                      <h4 style={{ fontSize: '1.15rem', fontWeight: 700, marginBottom: '4px' }}>{m.name}</h4>
                      <a
                        href={m.url}
                        target="_blank"
                        rel="noreferrer"
                        style={{ fontSize: '0.82rem', color: 'var(--primary)', textDecoration: 'none', fontFamily: 'var(--font-mono)' }}
                      >
                        {m.url}
                      </a>
                    </div>
                    <StatusBadge status={m.status} />
                  </div>

                  <div style={{ display: 'flex', gap: '16px', fontSize: '0.8rem', color: 'var(--text-muted)', margin: '12px 0 16px 0' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <Clock size={14} />
                      <span>{m.interval_seconds}s check</span>
                    </div>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <Server size={14} />
                      <span style={{ fontFamily: 'var(--font-mono)' }}>{m.k8s_name}</span>
                    </div>
                  </div>

                  <UptimeBarChart checkResults={m.check_results} />
                </div>

                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '20px', paddingTop: '16px', borderTop: '1px solid var(--border-color)' }}>
                  <button
                    onClick={() => handleDeleteMonitor(m.id, m.name)}
                    className="btn btn-danger"
                    style={{ fontSize: '0.8rem', padding: '6px 12px' }}
                  >
                    <Trash2 size={14} />
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        <CreateMonitorModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onCreate={handleCreateMonitor}
        />
      </main>
    </div>
  );
}
