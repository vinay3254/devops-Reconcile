import React, { useState } from 'react';
import { X, Plus, Activity } from 'lucide-react';

export default function CreateMonitorModal({ isOpen, onClose, onCreate }) {
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [interval, setInterval] = useState(60);
  const [expectedStatus, setExpectedStatus] = useState(200);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await onCreate({
        name,
        url,
        interval_seconds: Number(interval),
        expected_status: Number(expectedStatus),
        timeout_seconds: 10,
      });
      setName('');
      setUrl('');
      setInterval(60);
      setExpectedStatus(200);
      onClose();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="glass-card" style={{ width: '100%', maxWidth: '480px', padding: '28px', position: 'relative' }}>
        <button
          onClick={onClose}
          style={{ position: 'absolute', top: '20px', right: '20px', background: 'none', border: 'none', color: 'var(--text-muted)', cursor: 'pointer' }}
        >
          <X size={20} />
        </button>

        <div style={{ display: 'flex', alignItems: 'center', gap: '10px', marginBottom: '20px' }}>
          <div style={{ padding: '8px', borderRadius: '8px', background: 'rgba(56,189,248,0.15)', color: 'var(--primary)' }}>
            <Activity size={20} />
          </div>
          <div>
            <h3 style={{ fontSize: '1.25rem', fontWeight: 700 }}>Add New Monitor</h3>
            <p style={{ fontSize: '0.82rem', color: 'var(--text-muted)' }}>Provisions a Kubernetes Monitor CRD & worker</p>
          </div>
        </div>

        {error && (
          <div style={{ background: 'rgba(244,63,94,0.12)', border: '1px solid rgba(244,63,94,0.3)', color: 'var(--danger)', padding: '10px 14px', borderRadius: 'var(--radius-sm)', fontSize: '0.85rem', marginBottom: '16px' }}>
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Monitor Name</label>
            <input
              type="text"
              className="form-input"
              placeholder="e.g. Production API"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="form-group">
            <label>Target HTTP/HTTPS URL</label>
            <input
              type="url"
              className="form-input"
              placeholder="https://api.mycompany.com/health"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              required
            />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px' }}>
            <div className="form-group">
              <label>Check Interval</label>
              <select className="form-input" value={interval} onChange={(e) => setInterval(e.target.value)}>
                <option value={30}>Every 30s</option>
                <option value={60}>Every 60s</option>
                <option value={120}>Every 2m</option>
                <option value={300}>Every 5m</option>
              </select>
            </div>

            <div className="form-group">
              <label>Expected HTTP Status</label>
              <input
                type="number"
                className="form-input"
                value={expectedStatus}
                onChange={(e) => setExpectedStatus(e.target.value)}
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '24px' }}>
            <button type="button" onClick={onClose} className="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" disabled={loading} className="btn btn-primary">
              <Plus size={16} />
              {loading ? 'Deploying CR...' : 'Create Monitor'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
