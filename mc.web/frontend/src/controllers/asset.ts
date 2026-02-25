import { API_BASE, handleResponse } from "./controller-base";
import { Asset } from "../models/asset";

/**
 * GET /api/assets
 * 
 * Makes a request to get all assets available for selection
 * 
 *  @returns {Asset[]} The list of assets
 */
export async function getAssets(): Promise<Asset[]> {
  const response = await fetch(`${API_BASE}/api/assets`);
  return handleResponse<Asset[]>(response, 'Unable to load assets');
}

/**
 * POST /api/assets/sync
 * 
 * Makes a request to sync the most recent data for a given symbol
 * 
 *  @param {string} symbol - The symbol of the asset to sync
 *  @returns {Date} The last refreshed date for the asset
 */
export async function syncAsset(symbol: string): Promise<Date> {
  const response = await fetch(`${API_BASE}/api/assets/sync`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ symbol }),
  });
  return await handleResponse<Date>(response, 'Unable to sync stock data');
}