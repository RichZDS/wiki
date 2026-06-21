import type { ApiResponse } from '../types/chunk'

const BASE = '/api/v1'

async function parseResponse<T>(res: Response): Promise<T> {
  const json: ApiResponse<T> = await res.json()
  if (!res.ok || json.err) {
    throw new Error(json.err || `HTTP ${res.status}`)
  }
  return json.data
}

export async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  return parseResponse<T>(res)
}

export async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return parseResponse<T>(res)
}
