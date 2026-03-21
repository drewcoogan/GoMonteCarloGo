import React from 'react';
import { RiskMetrics } from '../../models/simulation-response';

const cellStyle: React.CSSProperties = {
  padding: 12,
  background: '#f5f5f5',
  borderRadius: 4,
};

type Props = {
  metrics: RiskMetrics;
};

/** Portfolio return / risk figures from a completed Monte Carlo run. */
const RiskMetricsGrid: React.FC<Props> = ({ metrics }) => {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(140px, 1fr))', gap: 12 }}>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>VaR 95%</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.var95 * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>VaR 99%</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.var99 * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>CVaR 95%</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.cvar95 * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>CVaR 99%</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.cvar99 * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>Prob. of loss</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.probabilityOfLoss * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>Max drawdown P95</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.maxDrawdownP95 * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>Mean annualized return</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.meanFinalValue * 100).toFixed(2)}%</div>
      </div>
      <div style={cellStyle}>
        <div style={{ fontSize: 12, color: '#666' }}>Median annualized return</div>
        <div style={{ fontWeight: 'bold' }}>{(metrics.medianFinalValue * 100).toFixed(2)}%</div>
      </div>
    </div>
  );
};

export default RiskMetricsGrid;
