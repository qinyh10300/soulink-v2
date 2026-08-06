import { apiRequest } from '@/lib/api'

export interface OperationalMetrics {
  startedAt: string
  uptimeSeconds: number
  requests: {
    total: number
    clientErrors: number
    serverErrors: number
    averageDurationMs: number
    maxDurationMs: number
  }
  routes: Array<{
    method: string
    route: string
    count: number
    clientErrors: number
    serverErrors: number
    averageDurationMs: number
    maxDurationMs: number
  }>
  tasks: Array<{
    name: string
    count: number
    failures: number
    averageDurationMs: number
    maxDurationMs: number
  }>
  generatedAt: string
}

export function fetchOperationalMetrics() {
  return apiRequest<OperationalMetrics>('/admin/operational-metrics', { auth: true })
}
