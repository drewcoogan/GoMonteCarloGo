import { API_BASE, handleResponse } from "./controller-base";
import { NewScenarioRequest, Scenario } from "../models/scenario";

/**
 * GET /api/scenarios
 * 
 * Makes a request to get all scenarios available for running
 * 
 *  @returns {Scenario[]} The list of scenarios
 */
export async function getScenarios(): Promise<Scenario[]> {
  const response = await fetch(`${API_BASE}/api/scenarios`);
  return handleResponse<Scenario[]>(response, 'Unable to load scenarios');
}

/**
 * POST /api/scenarios
 * 
 * Makes a request to create a new scenario
 * 
 *  @param {NewScenarioRequest} newScenarioRequest - The new scenario request
 *  @returns {Scenario} The created scenario
 */
export async function createScenario(newScenarioRequest: NewScenarioRequest): Promise<Scenario> {
  const response = await fetch(`${API_BASE}/api/scenarios`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(newScenarioRequest),
  });
  return handleResponse<Scenario>(response, 'Unable to create scenario.');
}
