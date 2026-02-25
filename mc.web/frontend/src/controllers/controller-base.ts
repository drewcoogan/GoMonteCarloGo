import { ServiceResponse } from "../models/service-response";

export const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080';

/**
 * Handles the response from the API and returns the data
 * 
 * @param {Response} response - The response from the API
 * @param {string} errorMessage - The error message to throw if the request fails
 * @returns {T} The data from the API, parsed from service response
 */
export async function handleResponse<T>(response: Response, errorMessage: string): Promise<T> {
    const json = await response.json();
    
    if (!response.ok) {
      throw new Error(errorMessage);
    }
    
    const res = json as ServiceResponse<T>;
    if (isNotNullOrEmpty(res.error)) {
      throw new Error(res.error);
    }
    
    return res.data as T;
  }

  function isNotNullOrEmpty(value: string | null | undefined): value is string {
    return typeof value === 'string' && value.trim().length > 0;
  }  