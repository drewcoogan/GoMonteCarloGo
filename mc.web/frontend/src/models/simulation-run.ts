export type SimulationRun = {
    id: number;
    name: string;
    floatedWeight: boolean;
    distributionType: string;
    simulationUnitOfTime: string;
    simulationDuration: number;
    maxLookback: number;
    iterations: number;
    seed: number;
    degreesOfFreedom: number;
    errorMessage: string;
    startTimeUtc: Date;
    endTimeUtc: Date;
    components: SimulationRunComponent[];
};

// TODO: need to add asset details here, like symbol, name, etc.
export type SimulationRunComponent = {
    assetId: number;
    weight: number;
};