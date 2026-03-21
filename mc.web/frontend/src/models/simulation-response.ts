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
    /** Mean annualized return (decimal); JSON field name is historical. */
    meanFinalValue: number;
    /** Median annualized return (decimal). */
    medianFinalValue: number;
};

/** Mirrors backend `mc.data/models` roles: percentile, maxDrawdown, maxVolatility, sample. */
export const SAMPLE_PATH_ROLES = {
    percentile: 'percentile',
    maxDrawdown: 'maxDrawdown',
    maxVolatility: 'maxVolatility',
    sample: 'sample',
} as const;

export type SamplePath = {
    role: string;
    percentile: number;
    label: string;
    values: number[];
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
