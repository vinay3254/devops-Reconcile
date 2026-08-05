import React from 'react';

export default function UptimeBarChart({ checkResults = [] }) {
  const totalBars = 30;
  const bars = Array.from({ length: totalBars });

  // Fill from right to left with actual check results, fallback to empty/unknown
  const filledBars = bars.map((_, i) => {
    const reverseIdx = totalBars - 1 - i;
    if (checkResults && checkResults[reverseIdx]) {
      return checkResults[reverseIdx].result.toLowerCase();
    }
    return 'unknown';
  }).reverse();

  return (
    <div style={{ marginTop: '12px' }}>
      <div className="uptime-matrix">
        {filledBars.map((res, idx) => (
          <div
            key={idx}
            className={`uptime-bar ${res}`}
            title={`Check ${idx + 1}: ${res.toUpperCase()}`}
          />
        ))}
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.72rem', color: 'var(--text-dim)', marginTop: '6px' }}>
        <span>History (Recent Checks)</span>
        <span>100% Uptime</span>
      </div>
    </div>
  );
}
