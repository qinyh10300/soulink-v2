import { apiBlobRequest, apiRequest } from '@/lib/api'

export type SafetyContextType = 'intent' | 'match_room' | 'direct' | 'group'
export type SafetyReportCategory =
  | 'harassment'
  | 'privacy'
  | 'impersonation'
  | 'unsafe'
  | 'spam'
  | 'other'
export type SafetyReportStatus = 'open' | 'reviewing' | 'resolved' | 'dismissed'
export type PhotoVerificationStatus = 'pending' | 'approved' | 'rejected'

export interface TrustProfile {
  emailVerified: boolean
  profileComplete: boolean
  photoStatus: PhotoVerificationStatus | 'not_submitted' | 'not_verified'
  recentlyActive: boolean
  photoRequestId?: number
  photoReviewNotes?: string
  photoAppealReason?: string
  photoRequestedAt?: string
  photoUpdatedAt?: string
}

export interface PhotoVerification {
  id: number
  userId: number
  userNickname: string
  userEmail: string
  evidenceMime: string
  status: PhotoVerificationStatus
  reviewNotes: string
  appealReason?: string
  createdAt: string
  updatedAt: string
}

export interface SafetyReportResult {
  id: number
  ticketId: string
  blocked: boolean
  createdAt: string
}

export interface SafetyReportAppeal {
  id: number
  reason: string
  status: 'pending' | 'accepted' | 'rejected'
  reviewNotes?: string
  createdAt: string
  updatedAt: string
}

export interface SafetyCase {
  id: number
  ticketId: string
  category: SafetyReportCategory
  contextType: SafetyContextType
  contextTitle: string
  status: 'resolved' | 'dismissed'
  createdAt: string
  updatedAt: string
  appeal?: SafetyReportAppeal
}

export interface SafetySnapshotMessage {
  id: number
  senderRole: 'reporter' | 'reported_user' | 'other'
  content: string
  createdAt: string
}

export interface SafetyContextSnapshot {
  version?: number
  contextType?: SafetyContextType
  contextTitle?: string
  capturedAt?: string
  content?: {
    category?: string
    title?: string
    description?: string
  }
  messages?: SafetySnapshotMessage[]
}

export interface SafetyReviewEvent {
  id: number
  actorNickname: string
  action: string
  previousStatus: string
  nextStatus: string
  note: string
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
  assignedTo?: number
  assignedToNickname?: string
  reviewNotes: string
  contextSnapshot: SafetyContextSnapshot
  riskSignals: string[]
  appeal?: SafetyReportAppeal
  createdAt: string
  updatedAt: string
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
  reportedUserId?: number
}) {
  return apiRequest<SafetyReportResult>('/safety/reports', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function fetchMySafetyCases() {
  return apiRequest<SafetyCase[]>('/safety/cases', { auth: true })
}

export function appealSafetyReport(reportId: number, reason: string) {
  return apiRequest<SafetyReportAppeal>(`/safety/reports/${reportId}/appeals`, {
    method: 'POST',
    auth: true,
    body: { reason },
  })
}

export function fetchSafetyReports(status: SafetyReportStatus | '') {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiRequest<SafetyReport[]>(`/admin/safety-reports${query}`, {
    auth: true,
  })
}

export function updateSafetyReport(
  reportId: number,
  input: {
    status: SafetyReportStatus
    reviewNotes: string
    assignToMe: boolean
  },
) {
  return apiRequest<SafetyReport>(`/admin/safety-reports/${reportId}`, {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function reviewSafetyReportAppeal(
  appealId: number,
  input: { status: 'accepted' | 'rejected'; reviewNotes: string },
) {
  return apiRequest<SafetyReportAppeal>(`/admin/safety-appeals/${appealId}`, {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function fetchSafetyReviewEvents(entityType: 'report' | 'report_appeal' | 'photo_verification', entityId: number) {
  const query = new URLSearchParams({ entityType, entityId: String(entityId) })
  return apiRequest<SafetyReviewEvent[]>(`/admin/safety-review-events?${query}`, { auth: true })
}

export function fetchMyTrustProfile() {
  return apiRequest<TrustProfile>('/users/me/trust', { auth: true })
}

export function fetchPublicTrustProfile(userId: number) {
  return apiRequest<TrustProfile>(`/users/${userId}/trust`, { auth: true })
}

export function submitPhotoVerification(file: File) {
  const body = new FormData()
  body.set('photo', file)
  return apiRequest<TrustProfile>('/users/me/photo-verifications', {
    method: 'POST',
    auth: true,
    body,
  })
}

export function appealPhotoVerification(requestId: number, reason: string) {
  return apiRequest<TrustProfile>(`/users/me/photo-verifications/${requestId}/appeal`, {
    method: 'POST',
    auth: true,
    body: { reason },
  })
}

export function fetchPhotoVerifications(status: PhotoVerificationStatus | '' = 'pending') {
  const query = status ? `?status=${encodeURIComponent(status)}` : ''
  return apiRequest<PhotoVerification[]>(`/admin/photo-verifications${query}`, { auth: true })
}

export function fetchPhotoVerificationEvidence(requestId: number) {
  return apiBlobRequest(`/admin/photo-verifications/${requestId}/evidence`)
}

export function reviewPhotoVerification(
  requestId: number,
  input: { status: 'approved' | 'rejected'; notes: string },
) {
  return apiRequest<PhotoVerification>(`/admin/photo-verifications/${requestId}`, {
    method: 'PUT',
    auth: true,
    body: input,
  })
}
