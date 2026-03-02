import React, { useCallback, useEffect, useState } from 'react';
import { getScenarios } from '../controllers/scenario';
import { SimulationResources } from '../models/simulation-resources';
import { getSimulationResources } from '../controllers/simulation';
import { Scenario } from '../models/scenario';

const SimulationPage: React.FC = () => {
    const [simulationResources, setSimulationResources] = useState<SimulationResources>();
    const [scenarios, setScenarios] = useState<Scenario[]>([]);
    const [loadingSimulationResources, setLoadingSimulationResources] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const fetchSimulationResources = useCallback(async () => {
        setLoadingSimulationResources(true);
        try {

          const [simulationResources, scenarios] = await Promise.all([
            getSimulationResources(),
            getScenarios(),
          ]);

          setSimulationResources(simulationResources);
          setScenarios(scenarios);
        } catch (err: any) {
          setError(`Error loading simulation resources: ${err.message}`);
          throw err;
        } finally {
          setLoadingSimulationResources(false);
        }
      }, []);

      useEffect(() => {
        fetchSimulationResources();
      }, [fetchSimulationResources]);

      return (
        <div>
          <h1>Simulation Page</h1>
        </div>
      );
    }

export default SimulationPage;