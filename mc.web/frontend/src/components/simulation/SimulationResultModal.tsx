import React, { useEffect } from 'react';
import { SimulationResponse } from '../../models/simulation-response';
import { SettingsLine, WeightAtRunRow } from '../../utilities/simulation-result-view';
import RiskMetricsGrid from './RiskMetricsGrid';
import SamplePathsChart from './SamplePathsChart';

const backdropStyle: React.CSSProperties = {
  position: 'fixed',
  inset: 0,
  background: 'rgba(0,0,0,0.45)',
  zIndex: 1000,
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  padding: 16,
  boxSizing: 'border-box',
};

const panelStyle: React.CSSProperties = {
  background: '#fff',
  borderRadius: 8,
  maxWidth: 960,
  width: '100%',
  maxHeight: '90vh',
  overflow: 'auto',
  boxShadow: '0 8px 32px rgba(0,0,0,0.2)',
  padding: 24,
};

type Props = {
  result: SimulationResponse;
  settingsLines: SettingsLine[];
  weightsAtRun: WeightAtRunRow[];
  onClose: () => void;
};

const SimulationResultModal: React.FC<Props> = ({ result, settingsLines, weightsAtRun, onClose }) => {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div
      style={backdropStyle}
      role="dialog"
      aria-modal="true"
      aria-labelledby="sim-result-title"
      onClick={e => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div style={panelStyle} onClick={e => e.stopPropagation()}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 16 }}>
          <h2 id="sim-result-title" style={{ marginTop: 0, marginBottom: 0 }}>
            Simulation results
          </h2>
          <button type="button" className="mc-btn mc-btn--outline" onClick={onClose}>
            Close
          </button>
        </div>

        <section style={{ marginTop: 20 }}>
          <h3 style={{ marginTop: 0, marginBottom: 12, fontSize: 16 }}>Settings used</h3>
          <dl
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
              gap: '8px 16px',
              margin: 0,
            }}
          >
            {settingsLines.map(row => (
              <div key={row.label} style={{ margin: 0 }}>
                <dt style={{ margin: 0, fontSize: 12, color: '#666' }}>{row.label}</dt>
                <dd style={{ margin: '4px 0 0', fontWeight: 600 }}>{row.value}</dd>
              </div>
            ))}
          </dl>
        </section>

        <section style={{ marginTop: 24 }}>
          <h3 style={{ marginTop: 0, marginBottom: 12, fontSize: 16 }}>Risk metrics</h3>
          <RiskMetricsGrid metrics={result.riskMetrics} />
        </section>

        <section style={{ marginTop: 24 }}>
          <h3 style={{ marginTop: 0, marginBottom: 12, fontSize: 16 }}>Sample paths</h3>
          <p style={{ marginTop: 0, marginBottom: 12, color: '#666', fontSize: 14 }}>
            Exemplar paths use role <code>percentile</code>, <code>maxDrawdown</code>, or <code>maxVolatility</code>.
            Extra draws use role <code>sample</code> (lighter, dashed)—picked with the run seed for reproducibility.
          </p>
          <SamplePathsChart samplePaths={result.samplePaths} />
        </section>

        {weightsAtRun.length > 0 && (
          <section style={{ marginTop: 24 }}>
            <h3 style={{ marginTop: 0, marginBottom: 12, fontSize: 16 }}>Weights at run</h3>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
              <thead>
                <tr>
                  <th style={{ textAlign: 'left', borderBottom: '1px solid #ddd', padding: '8px 8px 8px 0' }}>
                    Ticker
                  </th>
                  <th style={{ textAlign: 'right', borderBottom: '1px solid #ddd', padding: '8px 0 8px 8px' }}>
                    Weight
                  </th>
                </tr>
              </thead>
              <tbody>
                {weightsAtRun.map(row => (
                  <tr key={row.assetId}>
                    <td style={{ padding: '8px 8px 8px 0', borderBottom: '1px solid #eee' }}>{row.ticker}</td>
                    <td style={{ textAlign: 'right', padding: '8px 0 8px 8px', borderBottom: '1px solid #eee' }}>
                      {(row.weight * 100).toFixed(2)}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        )}
      </div>
    </div>
  );
};

export default SimulationResultModal;
