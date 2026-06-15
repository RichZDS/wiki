import type { ApiResponse } from '../types/chunk'

const BASE = '/api/v1'

export async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const json: ApiResponse<T> = await res.json()
  if (!res.ok || json.err) {
    throw new Error(json.err || `HTTP ${res.status}`)
  }
  return json.data
}
