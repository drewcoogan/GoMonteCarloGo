import { SimulationRequestSettings } from "../models/simulation-request-settings";
import { SimulationResources } from "../models/simulation-resources";
import { SimulationResponse } from "../models/simulation-response";
import { SimulationRun } from "../models/simulation-run";
import { normalizeSimulationResponse } from "../utilities/simulation-response";
import { API_BASE, handleResponse } from "./controller-base";

export async function getSimulationResources(): Promise<SimulationResources> {
    const response = await fetch(`${API_BASE}/api/simulation/resources`);
    return handleResponse<SimulationResources>(response, 'Unable to load simulation resources');
}

export async function runSimulation(id: number, requestSettings: SimulationRequestSettings): Promise<SimulationResponse> {
    const response = await fetch(`${API_BASE}/api/simulation/run/${id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestSettings),
    });
    const data = await handleResponse<SimulationResponse>(response, 'Unable to run simulation');
    return normalizeSimulationResponse(data);
}

export async function getSimulationRunHistory(scenarioId: number): Promise<SimulationRun[]> {
    const response = await fetch(`${API_BASE}/api/simulation/run-history/${scenarioId}`);
    return handleResponse<SimulationRun[]>(response, 'Unable to load simulation run history');
}

export async function getSimulationResult(simulationRunId: number): Promise<SimulationResponse> {
    const response = await fetch(`${API_BASE}/api/simulation/result/${simulationRunId}`);
    const data = await handleResponse<SimulationResponse>(response, 'Unable to load simulation result');
    return normalizeSimulationResponse(data);
}