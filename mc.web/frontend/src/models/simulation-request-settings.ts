export type SimulationRequestSettings = {
    distType: number;
    simulationUnitOfTime: number;
    simulationDuration: number;
    maxLookback: number;
    iterations: number;
    seed: number;
    degreesOfFreedom: number;
};