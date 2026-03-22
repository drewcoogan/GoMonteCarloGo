import { SimulationResources } from '../models/simulation-resources';
import { SimulationRequestSettings } from '../models/simulation-request-settings';
import { SimulationRun } from '../models/simulation-run';
import { horizonToWeeklySteps } from './simulation-horizon';

export type SettingsLine = { label: string; value: string };

export type WeightAtRunRow = { assetId: number; ticker: string; weight: number };

export function buildWeightAtRunRows(
  components: { assetId: number; weight: number }[] | undefined,
  symbolByAssetId: Map<number, string>
): WeightAtRunRow[] {
  if (!components?.length) return [];
  return components.map(c => ({
    assetId: c.assetId,
    ticker: symbolByAssetId.get(c.assetId) ?? `Asset ${c.assetId}`,
    weight: c.weight,
  }));
}

function isStudentTDistributionType(dist: string): boolean {
  return dist === 'studentT';
}

function humanizeKey(key: string): string {
  return key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase());
}

const UNIT_PLURALS: Record<string, [string, string]> = {
  days: ['day', 'days'],
  weeks: ['week', 'weeks'],
  months: ['month', 'months'],
  years: ['year', 'years'],
};

/** Human-readable "3 months", "1 year", etc. */
export function formatCountAndUnit(count: number, unit: string): string {
  const u = unit.toLowerCase().trim();
  const pair = UNIT_PLURALS[u];
  if (!pair) {
    return `${count} ${u}`;
  }
  const label = count === 1 ? pair[0] : pair[1];
  return `${count} ${label}`;
}

function approximateLookbackDaysFromRun(run: SimulationRun): number | null {
  const start = new Date(run.startTimeUtc).getTime();
  const cutoff = new Date(run.maxLookback).getTime();
  if (!Number.isFinite(start) || !Number.isFinite(cutoff)) {
    return null;
  }
  const days = Math.round((start - cutoff) / (24 * 60 * 60 * 1000));
  return days > 0 ? days : null;
}

const WEEKS_PER_YEAR = 52;

/**
 * Uses persisted `simulationDuration` + `simulationUnitOfTime` only (same fields the engine uses).
 * For weekly horizons, maps whole-year step counts to years and reverses our month→weeks rounding when possible.
 */
function formatRunHorizon(run: SimulationRun): string {
  const weeks = run.simulationDuration;
  const u = run.simulationUnitOfTime.toLowerCase();
  if (weeks <= 0) {
    return '—';
  }
  if (u !== 'weekly' && u !== 'weeks') {
    return `${weeks} ${run.simulationUnitOfTime}`;
  }
  if (weeks % WEEKS_PER_YEAR === 0) {
    return formatCountAndUnit(weeks / WEEKS_PER_YEAR, 'years');
  }
  for (let m = 1; m <= 120; m += 1) {
    if (Math.round((m * WEEKS_PER_YEAR) / 12) === weeks) {
      return formatCountAndUnit(m, 'months');
    }
  }
  return `${weeks} weeks`;
}

function formatRunMaxLookback(run: SimulationRun): string {
  if (run.maxLookbackCount != null && run.maxLookbackCount > 0 && run.maxLookbackUnit) {
    return formatCountAndUnit(run.maxLookbackCount, run.maxLookbackUnit);
  }
  const days = approximateLookbackDaysFromRun(run);
  if (days != null) {
    return formatCountAndUnit(days, 'days');
  }
  return '—';
}

/** Labels for the run form (numeric codes → resource labels). */
export function settingsLinesFromLive(
  scenarioName: string,
  settings: SimulationRequestSettings,
  resources: SimulationResources | null
): SettingsLine[] {
  const distEntry = resources
    ? Object.entries(resources.distributionType).find(([, v]) => v === settings.distributionType)
    : undefined;
  const distLabel = distEntry ? humanizeKey(distEntry[0]) : String(settings.distributionType);
  const weeklySteps = horizonToWeeklySteps(settings.simulationHorizonCount, settings.simulationHorizonUnit);

  const lines: SettingsLine[] = [
    { label: 'Scenario', value: scenarioName },
    { label: 'Distribution', value: distLabel },
    {
      label: 'Horizon',
      value: formatCountAndUnit(settings.simulationHorizonCount, settings.simulationHorizonUnit),
    },
    { label: 'Path step', value: `Weekly (${weeklySteps} steps)` },
    {
      label: 'Max lookback',
      value: formatCountAndUnit(settings.maxLookbackCount, settings.maxLookbackUnit),
    },
    { label: 'Iterations', value: String(settings.iterations) },
    { label: 'Seed', value: String(settings.seed) },
  ];

  const studentTCode = resources?.distributionType?.studentT;
  if (studentTCode !== undefined && settings.distributionType === studentTCode) {
    lines.push({ label: 'Degrees of freedom', value: String(settings.degreesOfFreedom) });
  }

  return lines;
}

/** Labels from a persisted run row (already string enums / dates from the API). */
export function settingsLinesFromRun(run: SimulationRun): SettingsLine[] {
  const lines: SettingsLine[] = [
    { label: 'Scenario', value: run.name },
    { label: 'Floated weights', value: run.floatedWeight ? 'Yes' : 'No' },
    { label: 'Distribution', value: humanizeKey(run.distributionType) },
    { label: 'Horizon', value: formatRunHorizon(run) },
    { label: 'Path step', value: `Weekly (${run.simulationDuration} steps)` },
    { label: 'Max lookback', value: formatRunMaxLookback(run) },
    { label: 'Iterations', value: String(run.iterations) },
    { label: 'Seed', value: String(run.seed) },
  ];

  if (isStudentTDistributionType(run.distributionType)) {
    lines.push({ label: 'Degrees of freedom', value: String(run.degreesOfFreedom) });
  }

  return lines;
}

export function runHasViewableResult(run: SimulationRun): boolean {
  return !run.errorMessage && Boolean(run.endTimeUtc);
}
