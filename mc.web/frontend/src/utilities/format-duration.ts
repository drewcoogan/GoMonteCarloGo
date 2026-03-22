import type { SimulationRun } from '../models/simulation-run';
import type { SimulationResponse } from '../models/simulation-response';

const US = 1_000;
const MS = 1_000_000;
const S = 1_000_000_000;
const MIN = 60 * S;

function trimFixed(x: number, digits: number): string {
  return x.toFixed(digits).replace(/\.?0+$/, '');
}

/**
 * Formats a non-negative duration in nanoseconds using ns, µs, ms, s, or min (+ s remainder).
 */
export function formatDurationFromNanos(ns: number): string {
  if (!Number.isFinite(ns) || ns < 0) {
    return '—';
  }
  if (ns === 0) {
    return '0 ns';
  }
  if (ns < US) {
    return `${Math.round(ns)} ns`;
  }
  if (ns < MS) {
    const v = ns / US;
    return `${trimFixed(v, v < 10 ? 2 : 1)} µs`;
  }
  if (ns < S) {
    const v = ns / MS;
    return `${trimFixed(v, v < 10 ? 2 : 1)} ms`;
  }
  if (ns < MIN) {
    const v = ns / S;
    return `${trimFixed(v, v < 10 ? 2 : 2)} s`;
  }
  const minutes = Math.floor(ns / MIN);
  const remNs = ns % MIN;
  if (remNs < S * 0.05) {
    return `${minutes} min`;
  }
  const secs = remNs / S;
  return `${minutes} min ${trimFixed(secs, 1)} s`;
}

/** Elapsed time from run row timestamps (DB resolution; may differ slightly from wallTimeNanos). */
export function wallClockNsFromRun(run: SimulationRun): number | null {
  if (!run.endTimeUtc || !run.startTimeUtc) {
    return null;
  }
  const ms = new Date(run.endTimeUtc).getTime() - new Date(run.startTimeUtc).getTime();
  return ms > 0 ? Math.round(ms * 1e6) : null;
}

/** Prefer persisted server timing from the result payload; else fall back to run history timestamps. */
export function simulationWallTimeSettingsLine(
  result: SimulationResponse,
  run?: SimulationRun
): { label: string; value: string } | null {
  if (result.wallTimeNanos != null && result.wallTimeNanos > 0) {
    return { label: 'Simulation duration', value: formatDurationFromNanos(result.wallTimeNanos) };
  }
  const ns = run ? wallClockNsFromRun(run) : null;
  if (ns != null) {
    return { label: 'Simulation duration', value: formatDurationFromNanos(ns) };
  }
  return null;
}
