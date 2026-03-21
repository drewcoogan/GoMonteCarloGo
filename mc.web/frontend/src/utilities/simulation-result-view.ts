import { SimulationResources } from '../models/simulation-resources';
import { SimulationRequestSettings } from '../models/simulation-request-settings';
import { SimulationRun } from '../models/simulation-run';
import { NanosecondsToDays } from './time';

export type SettingsLine = { label: string; value: string };

function capitalizeLabel(s: string): string {
  const t = s.trim();
  if (!t) return t;
  return t.charAt(0).toUpperCase() + t.slice(1);
}

function humanizeKey(key: string): string {
  return key.replace(/([A-Z])/g, ' $1').replace(/^./, s => s.toUpperCase());
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
  const unitEntry = resources
    ? Object.entries(resources.simulationUnitOfTime).find(([, v]) => v === settings.simulationUnitOfTime)
    : undefined;
  const durEntry = resources
    ? Object.entries(resources.simulationDuration).find(([, v]) => v === settings.simulationDuration)
    : undefined;

  const distLabel = distEntry ? humanizeKey(distEntry[0]) : String(settings.distributionType);
  const unitLabel = capitalizeLabel(unitEntry ? unitEntry[0] : String(settings.simulationUnitOfTime));
  const durationSuffix = durEntry ? durEntry[0] : 'periods';

  return [
    { label: 'Scenario', value: scenarioName },
    { label: 'Distribution', value: distLabel },
    { label: 'Unit of time', value: unitLabel },
    { label: 'Duration', value: `${settings.simulationDuration} ${durationSuffix}` },
    { label: 'Max lookback', value: `${NanosecondsToDays(settings.maxLookback)} days` },
    { label: 'Iterations', value: String(settings.iterations) },
    { label: 'Seed', value: String(settings.seed) },
    { label: 'Degrees of freedom', value: String(settings.degreesOfFreedom) },
  ];
}

/** Labels from a persisted run row (already string enums / dates from the API). */
export function settingsLinesFromRun(run: SimulationRun): SettingsLine[] {
  const maxLookback = new Date(run.maxLookback).toLocaleString();

  const lines: SettingsLine[] = [
    { label: 'Scenario', value: run.name },
    { label: 'Floated weights', value: run.floatedWeight ? 'Yes' : 'No' },
    { label: 'Distribution', value: humanizeKey(run.distributionType) },
    { label: 'Unit of time', value: capitalizeLabel(run.simulationUnitOfTime) },
    { label: 'Duration (periods)', value: String(run.simulationDuration) },
    { label: 'Max lookback (cutoff)', value: maxLookback },
    { label: 'Iterations', value: String(run.iterations) },
    { label: 'Seed', value: String(run.seed) },
    { label: 'Degrees of freedom', value: String(run.degreesOfFreedom) },
  ];

  if (run.components?.length) {
    lines.push({
      label: 'Weights at run',
      value: run.components.map(c => `${(c.weight * 100).toFixed(2)}% (asset ${c.assetId})`).join(', '),
    });
  }

  return lines;
}

export function runHasViewableResult(run: SimulationRun): boolean {
  return !run.errorMessage && Boolean(run.endTimeUtc);
}
