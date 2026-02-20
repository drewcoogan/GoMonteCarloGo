export const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080';

export async function handleResponse<T>(response: Response, errorMessage: string): Promise<T> {
    const json = await response.json();
    if (!response.ok) {
      throw new Error((json as { error?: string }).error || errorMessage);
    }
    return json as T;
  }