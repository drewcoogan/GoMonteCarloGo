/** Matches API response - keys are lowercase */
export type SimulationResources = {
  distributionType: Record<string, number>;
  simulationUnitOfTime: Record<string, number>;
  simulationDuration: Record<string, number>;
};