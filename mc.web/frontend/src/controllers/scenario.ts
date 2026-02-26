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

/**
 * GET /api/scenarios/{id}
 * 
 * Makes a request to get a scenario by its ID
 * 
 *  @param {number} id - The ID of the scenario to get
 *  @returns {Scenario} The scenario
 */
export async function getScenario(id: number): Promise<Scenario> {
  const response = await fetch(`${API_BASE}/api/scenarios/${id}`);
  return handleResponse<Scenario>(response, 'Unable to get scenario.');
}

/**
 * PUT /api/scenarios/{id}
 * 
 * Makes a request to update a scenario by its ID
 * 
 *  @param {number} id - The ID of the scenario to update
 *  @param {Scenario} scenario - The scenario to update
 *  @returns {Scenario} The updated scenario
 */
export async function updateScenario(id: number, scenario: Scenario): Promise<Scenario> {
  const response = await fetch(`${API_BASE}/api/scenarios/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(scenario),
  });
  return handleResponse<Scenario>(response, 'Unable to update scenario.');
}

/**
 * DELETE /api/scenarios/{id}
 * 
 * Makes a request to delete a scenario by its ID
 * 
 *  @param {number} id - The ID of the scenario to delete
 *  @returns {boolean} If delete is successful, will be true, otherwise false
 */
export async function deleteScenario(id: number): Promise<boolean> {
  const response = await fetch(`${API_BASE}/api/scenarios/${id}`, {
    method: 'DELETE',
  });
  return handleResponse<boolean>(response, 'Unable to delete scenario.');
}


