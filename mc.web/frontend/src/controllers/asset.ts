import { API_BASE, handleResponse } from "./controller-base";
import { Asset } from "../models/asset";
import { ServiceResponse } from "../models/service-response";

export async function getAssets(): Promise<Asset[]> {
  const response = await fetch(`${API_BASE}/api/assets`);
  return handleResponse<Asset[]>(response, 'Unable to load assets');
}

export async function syncStockData(symbol: string): Promise<ServiceResponse<Date>> {
  const response = await fetch(`${API_BASE}/api/syncStockData`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbol }),
  });

  const data = await handleResponse<{ date?: string; error?: string; message?: string }>(response, 'Unable to sync stock data');
  if (!response.ok) {
    const message = data?.error ?? data?.message ?? `Request failed (${response.status})`;
    return { data: null, error: message };
  }
  return { data: new Date(data?.date ?? 0), error: '' };
}