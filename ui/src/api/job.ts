import { get, post } from './client'
import type { JobLogListResult, JobSnapshot } from '../types/job'

export function listJobs(): Promise<JobSnapshot[]> {
  return get<JobSnapshot[]>('/jobs')
}

export function getJob(name: string): Promise<JobSnapshot> {
  return get<JobSnapshot>(`/jobs/${encodeURIComponent(name)}`)
}

export function startJob(name: string): Promise<JobSnapshot> {
  return post<JobSnapshot>(`/jobs/${encodeURIComponent(name)}/start`, {})
}

export function stopJob(name: string): Promise<JobSnapshot> {
  return post<JobSnapshot>(`/jobs/${encodeURIComponent(name)}/stop`, {})
}

export function runJobNow(name: string): Promise<JobSnapshot> {
  return post<JobSnapshot>(`/jobs/${encodeURIComponent(name)}/run`, {})
}

export function listJobLogs(name: string, level = '', page = 1, size = 20): Promise<JobLogListResult> {
  const params = new URLSearchParams({ page: String(page), size: String(size) })
  if (level) {
    params.set('level', level)
  }
  return get<JobLogListResult>(`/jobs/${encodeURIComponent(name)}/logs?${params}`)
}
