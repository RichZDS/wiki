export interface JobSnapshot {
  name: string
  interval_ms: number
  enabled: boolean
  running: boolean
  next_run?: string
  last_started_at?: string
  last_finished_at?: string
  last_status: string
  last_error: string
}

export interface JobLog {
  id: number
  job_name: string
  run_id: string
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  created_at: string
}

export interface JobLogListResult {
  total: number
  list: JobLog[]
}
