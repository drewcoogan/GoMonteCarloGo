import { API_BASE, handleResponse } from "./controller-base";

export async function getHearbeat(): Promise<Set<string>> {
    const response = await fetch(`${API_BASE}/api/heartbeat`);
    const heartbeats = await handleResponse<Record<string, boolean>>(response, 'Unable to get heartbeat');

    const unhealthyServices = new Set<string>();
    for (const [key, value] of Object.entries(heartbeats)) {
        if (!value) {
            unhealthyServices.add(key);
        }
    }

    return unhealthyServices;
}

/** True when the API responds and every dependency in the heartbeat payload is healthy. */
export async function isServiceHealthy(): Promise<boolean> {
    try {
        const unhealthy = await getHearbeat();
        return unhealthy.size === 0;
    } catch {
        return false;
    }
}