import type {
  ApiResponse,
  ChunkRequest,
  ChunkResult,
  ChunkCompareRequest,
  ChunkCompareResult,
} from '../types/chunk'

const BASE = '/api/v1'

async function post<T>(path: string, body: unknown): Promise<T> {
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

// chunk 调用单策略切块接口。
export function chunk(req: ChunkRequest): Promise<ChunkResult> {
  return post<ChunkResult>('/chunk', req)
}

// compare 调用多策略对比切块接口。
export function compare(req: ChunkCompareRequest): Promise<ChunkCompareResult> {
  return post<ChunkCompareResult>('/chunk/compare', req)
}
