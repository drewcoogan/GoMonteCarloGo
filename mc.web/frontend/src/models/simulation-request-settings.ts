export type SimulationRequestSettings = {
    distributionType: number;
    simulationUnitOfTime: number;
    simulationDuration: number;
    maxLookback: number;
    iterations: number;
    seed: number;
    degreesOfFreedom: number;
};

/** Default values for the form; resource-driven fields (distributionType, simulationUnitOfTime, simulationDuration) are overwritten when resources load. */
export const DEFAULT_SIMULATION_REQUEST_SETTINGS: SimulationRequestSettings = {
    distributionType: 0,
    simulationUnitOfTime: 52,
    simulationDuration: 52,
    maxLookback: 730 * 24 * 60 * 60 * 1e9, // 730 days in nanoseconds
    iterations: 1000,
    seed: 42,
    degreesOfFreedom: 5,
};