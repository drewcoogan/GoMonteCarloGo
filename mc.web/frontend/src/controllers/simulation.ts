import { SimulationRequestSettings } from "../models/simulation-request-settings";
import { SimulationResources } from "../models/simulation-resources";
import { SimulationResponse } from "../models/simulation-response";
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
    return handleResponse<SimulationResponse>(response, 'Unable to run simulation');
}