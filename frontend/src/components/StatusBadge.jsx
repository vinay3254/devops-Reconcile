import React from 'react';

export default function StatusBadge({ status }) {
  const normalized = (status || 'UNKNOWN').toUpperCase();
  let className = 'status-pill unknown';
  let label = 'Unknown';

  if (normalized === 'UP') {
    className = 'status-pill up';
    label = 'Operational';
  } else if (normalized === 'DOWN') {
    className = 'status-pill down';
    label = 'Downtime';
  }

  return (
    <div className={className}>
      <span className="pulse-dot"></span>
      <span>{label}</span>
    </div>
  );
}
