export type SimulationResponse = {
    riskMetrics: RiskMetrics;
    samplePaths: SamplePath[];
    simulationStats: SimulationStats;
};

export type RiskMetrics = {
    var95: number;
    var99: number;
    cvar95: number;
    cvar99: number;
    probabilityOfLoss: number;
    maxDrawdownP95: number;
    meanFinalValue: number;
    medianFinalValue: number;
};

export type SamplePath = {
    percentile: number;
    values: number[];
    label: string;
};

export type SimulationStats = {
    mean: number[];
    stdDev: number[];
    p5: number[];
    p25: number[];
    p50: number[];
    p75: number[];
    p95: number[];
};