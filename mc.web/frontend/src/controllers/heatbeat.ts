import { API_BASE, handleResponse } from "./controller-base";

export async function getHearbeat(): Promise<Set<string>> {
    const response = await fetch(`${API_BASE}/api/heartbeat`);
    const heartbeats = await handleResponse<Map<string, boolean>>(response, 'Unable to get heartbeat');

    const unhealthyServices = new Set<string>();
    for (const [key, value] of Object.entries(heartbeats)) {
        if (!value) {
            unhealthyServices.add(key);
        }
    }

    return unhealthyServices;
}