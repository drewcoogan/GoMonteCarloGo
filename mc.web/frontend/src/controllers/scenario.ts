import { API_BASE, handleResponse } from "./controller-base";
import { NewScenarioRequest, Scenario } from "../models/scenario";

export async function getScenarios(): Promise<Scenario[]> {
  const response = await fetch(`${API_BASE}/api/scenarios`);
  return handleResponse<Scenario[]>(response, 'Unable to load scenarios');
}

export async function createScenario(payload: NewScenarioRequest): Promise<Scenario> {
  const response = await fetch(`${API_BASE}/api/scenarios`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  return handleResponse<Scenario>(response, 'Unable to create scenario.');
}
