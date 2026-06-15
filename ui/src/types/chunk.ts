export type ChunkStrategy = 'free' | 'md' | 'eino' | 'hierarchical'

export interface ChunkConfigItem {
  strategy: ChunkStrategy
  chunk_size: number
  chunk_overlap: number
  separators: string[]
}

export interface ChunkRequest {
  content: string
  strategy: ChunkStrategy
  chunk_size: number
  chunk_overlap: number
  separators: string[]
}

export interface ChunkCompareRequest {
  content: string
  configs: ChunkConfigItem[]
}

export interface ChunkItem {
  id: string
  content: string
  length: number
  metadata: Record<string, unknown>
}

export interface ChunkStats {
  total_chunks: number
  total_characters: number
  avg_chunk_length: number
  min_chunk_length: number
  max_chunk_length: number
}

export interface ChunkResult {
  chunks: ChunkItem[]
  stats: ChunkStats
}

export interface ChunkStrategyResult {
  strategy: ChunkStrategy
  chunks: ChunkItem[]
  stats: ChunkStats
}

export interface ChunkCompareResult {
  results: ChunkStrategyResult[]
  errors?: Record<string, string>
}

export interface ApiResponse<T> {
  code: number
  data: T
  err?: string
}
