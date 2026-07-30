import { apiRequest } from '@/lib/api'

export type SafetyContextType = 'intent' | 'match_room'
export type SafetyReportCategory =
  | 'harassment'
  | 'privacy'
  | 'impersonation'
  | 'unsafe'
  | 'spam'
  | 'other'
export type SafetyReportStatus = 'open' | 'reviewing' | 'resolved' | 'dismissed'

export interface SafetyReportResult {
  id: number
  ticketId: string
  blocked: boolean
  createdAt: string
}

export interface SafetyReport {
  id: number
  ticketId: string
  reporterId: number
  reporterNickname: string
  reporterEmail: string
  reportedUserId: number
  reportedUserNickname: string
  reportedUserEmail: string
  contextType: SafetyContextType
  contextId: number
  contextTitle: string
  category: SafetyReportCategory
  details: string
  status: SafetyReportStatus
  createdAt: string
  resolvedAt?: string
  resolvedBy?: number
}

export function fetchSafetyCapabilities() {
  return apiRequest<{ isAdmin: boolean }>('/safety/capabilities', {
    auth: true,
  })
}

export function submitSafetyReport(input: {
  contextType: SafetyContextType
  contextId: number
  category: SafetyReportCategory
  details: string
  blockUser: boolean
}) {
  return apiRequest<SafetyReportResult>('/safety/reports', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function fetchSafetyReports(status: SafetyReportStatus | '') {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiRequest<SafetyReport[]>(`/admin/safety-reports${query}`, {
    auth: true,
  })
}

export function updateSafetyReportStatus(
  reportId: number,
  status: SafetyReportStatus,
) {
  return apiRequest<SafetyReport>(`/admin/safety-reports/${reportId}`, {
    method: 'PUT',
    auth: true,
    body: { status },
  })
}
