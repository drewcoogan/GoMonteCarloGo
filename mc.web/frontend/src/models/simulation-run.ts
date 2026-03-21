export type SimulationRun = {
    id: number;
    name: string;
    floatedWeight: boolean;
    distributionType: string;
    simulationUnitOfTime: string;
    simulationDuration: number;
    /** ISO-8601 from the API (`time.Time`). */
    maxLookback: string;
    iterations: number;
    seed: number;
    degreesOfFreedom: number;
    errorMessage?: string;
    startTimeUtc: string;
    endTimeUtc?: string;
    components: SimulationRunComponent[];
};

// TODO: need to add asset details here, like symbol, name, etc.
export type SimulationRunComponent = {
    assetId: number;
    weight: number;
};