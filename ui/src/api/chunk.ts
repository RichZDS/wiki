import { post } from './client'
import type {
  ChunkRequest,
  ChunkResult,
  ChunkCompareRequest,
  ChunkCompareResult,
} from '../types/chunk'

export function chunk(req: ChunkRequest): Promise<ChunkResult> {
  return post<ChunkResult>('/chunk', req)
}

export function compare(req: ChunkCompareRequest): Promise<ChunkCompareResult> {
  return post<ChunkCompareResult>('/chunk/compare', req)
}
