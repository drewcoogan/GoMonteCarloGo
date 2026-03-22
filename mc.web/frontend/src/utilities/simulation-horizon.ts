export type HorizonUnit = 'months' | 'years';
export type LookbackUnit = 'months' | 'years';

const WEEKS_PER_YEAR = 52;

/** Matches server: weekly steps from user horizon (months or years). */
export function horizonToWeeklySteps(count: number, unit: HorizonUnit): number {
  if (unit === 'years') {
    return Math.round(count * WEEKS_PER_YEAR);
  }
  return Math.round((count * WEEKS_PER_YEAR) / 12);
}

export function maxHorizonCount(unit: HorizonUnit): number {
  return unit === 'years' ? 10 : 120;
}

/** Max lookback is 5 years (60 months or 5 years). */
export function maxLookbackCount(unit: LookbackUnit): number {
  return unit === 'years' ? 5 : 60;
}
