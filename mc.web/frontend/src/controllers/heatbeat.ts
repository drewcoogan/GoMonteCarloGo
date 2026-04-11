import { API_BASE, handleResponse } from "./controller-base";

export type HeartbeatPayload = {
  service: boolean;
  database: boolean;
  /** Go toolchain version of the running binary (e.g. go1.26.0), from runtime.Version(). */
  goVersion: string;
};

export async function getHeartbeatPayload(): Promise<HeartbeatPayload> {
  const response = await fetch(`${API_BASE}/api/heartbeat`);
  return handleResponse<HeartbeatPayload>(response, 'Unable to get heartbeat');
}

export async function getHearbeat(): Promise<Set<string>> {
  const heartbeats = await getHeartbeatPayload();
  const unhealthyServices = new Set<string>();
  if (!heartbeats.service) {
    unhealthyServices.add('service');
  }
  if (!heartbeats.database) {
    unhealthyServices.add('database');
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
