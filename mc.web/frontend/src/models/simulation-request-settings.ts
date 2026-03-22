import type { HorizonUnit, LookbackUnit } from '../utilities/simulation-horizon';

export type SimulationRequestSettings = {
    distributionType: number;
    maxLookbackCount: number;
    maxLookbackUnit: LookbackUnit;
    simulationHorizonCount: number;
    simulationHorizonUnit: HorizonUnit;
    iterations: number;
    seed: number;
    degreesOfFreedom: number;
};

/** Defaults; distribution type is overwritten when resources load. */
export const DEFAULT_SIMULATION_REQUEST_SETTINGS: SimulationRequestSettings = {
    distributionType: 0,
    maxLookbackCount: 2,
    maxLookbackUnit: 'years',
    simulationHorizonCount: 1,
    simulationHorizonUnit: 'years',
    iterations: 1000,
    seed: 42,
    degreesOfFreedom: 5,
};
