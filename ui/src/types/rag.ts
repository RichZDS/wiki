import type { ChunkStrategy } from './chunk'

export type { ChunkStrategy }

export interface RAGIngestRequest {
  content: string
  strategy?: ChunkStrategy
  chunk_size?: number
  chunk_overlap?: number
  separators?: string[]
  doc_id_prefix?: string
}

export interface RAGIngestResult {
  stored_ids: string[]
  chunk_count: number
  total_chars: number
}

export interface RAGSearchRequest {
  query: string
  top_k?: number
}

export interface RAGSearchItem {
  id: string
  content: string
  score?: number
  metadata: Record<string, unknown>
}

export interface RAGSearchResult {
  results: RAGSearchItem[]
}
