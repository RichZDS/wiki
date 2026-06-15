import { post } from './client'
import type { RAGIngestRequest, RAGIngestResult, RAGSearchRequest, RAGSearchResult } from '../types/rag'

export function ingest(req: RAGIngestRequest): Promise<RAGIngestResult> {
  return post<RAGIngestResult>('/rag/ingest', req)
}

export function search(req: RAGSearchRequest): Promise<RAGSearchResult> {
  return post<RAGSearchResult>('/rag/search', req)
}
