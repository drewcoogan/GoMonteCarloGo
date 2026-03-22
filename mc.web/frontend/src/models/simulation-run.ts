export type SimulationRun = {
    id: number;
    name: string;
    floatedWeight: boolean;
    distributionType: string;
    simulationUnitOfTime: string;
    simulationDuration: number;
    /** ISO-8601 cutoff date from the API (`time.Time` / DB date). */
    maxLookback: string;
    maxLookbackCount?: number;
    maxLookbackUnit?: string;
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