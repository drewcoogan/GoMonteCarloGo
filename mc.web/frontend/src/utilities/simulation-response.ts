import { RiskMetrics, SimulationResponse, SimulationStats } from '../models/simulation-response';

const emptyStats = (): SimulationStats => ({
  mean: [],
  stdDev: [],
  p5: [],
  p25: [],
  p50: [],
  p75: [],
  p95: [],
});

const emptyRiskMetrics = (): RiskMetrics => ({
  var95: 0,
  var99: 0,
  cvar95: 0,
  cvar99: 0,
  probabilityOfLoss: 0,
  maxDrawdownP95: 0,
  meanFinalValue: 0,
  medianFinalValue: 0,
});

/** Fills in missing arrays/objects so chart and metrics UI never read undefined. */
export function normalizeSimulationResponse(raw: SimulationResponse | null | undefined): SimulationResponse {
  if (!raw || typeof raw !== 'object') {
    return {
      riskMetrics: emptyRiskMetrics(),
      samplePaths: [],
      simulationStats: emptyStats(),
    };
  }

  const samplePaths = Array.isArray(raw.samplePaths) ? raw.samplePaths : [];

  return {
    riskMetrics: raw.riskMetrics ?? emptyRiskMetrics(),
    samplePaths: samplePaths.map(p => ({
      role: p.role ?? 'sample',
      percentile: typeof p.percentile === 'number' ? p.percentile : -1,
      label: p.label ?? '',
      values: Array.isArray(p.values) ? p.values : [],
    })),
    simulationStats: raw.simulationStats ?? emptyStats(),
    wallTimeNanos: typeof raw.wallTimeNanos === 'number' && raw.wallTimeNanos > 0 ? raw.wallTimeNanos : undefined,
  };
}
