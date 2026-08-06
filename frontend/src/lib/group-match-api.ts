import { apiRequest } from '@/lib/api'
import type { InferenceMetadata } from '@/lib/connection-api'

export type GroupMatchCategory = 'meal' | 'study' | 'sport' | 'event' | 'project'
export type GroupMatchStatus = 'open' | 'joined' | 'formed' | 'expired' | 'cancelled'
export type GroupInvitationStatus =
  | 'invited'
  | 'accepted'
  | 'declined'
  | 'waitlisted'
  | 'left'

export interface GroupTimeSuggestion {
  startsAt: string
  endsAt: string
  label: string
  explanation: string
}

export interface GroupMatchOrganizer {
  userId: number
  nickname: string
  department: string
  grade: string
}

export interface GroupMatchCandidateIdentity {
  userId?: number
  nickname: string
  department?: string
  grade?: string
  revealed: boolean
}

export interface GroupMatchInvitation {
  id: number
  groupIntentId: number
  candidateIntentId: number
  candidate: GroupMatchCandidateIdentity
  suggestedRole: string
  score: number
  fitReason: string
  uncertainty: string
  sharedConstraints: string[]
  status: GroupInvitationStatus
  commonFreeTimes: GroupTimeSuggestion[]
  createdAt: string
  updatedAt: string
}

export interface GroupMatchIntent {
  id: number
  organizer: GroupMatchOrganizer
  category: GroupMatchCategory
  title: string
  description: string
  roleOffer: string
  rolesNeeded: string[]
  roleGaps: string[]
  constraints: string[]
  timeWindow: string
  location: string
  minMembers: number
  maxMembers: number
  acceptedMembers: number
  pendingInvitations: number
  waitlistedMembers: number
  startsAt?: string
  endsAt?: string
  createCalendarEvent: boolean
  status: GroupMatchStatus
  chatGroupId?: number
  calendarEventId?: number
  groupIntro?: string
  conversationStarters: string[]
  inference: InferenceMetadata
  commonFreeTimes: GroupTimeSuggestion[]
  invitations: GroupMatchInvitation[]
  canForm: boolean
  canLeave: boolean
  expiresAt: string
  createdAt: string
  updatedAt: string
}

export interface GroupMatchRecommendation {
  candidateIntentId: number
  alias: string
  roleOffer: string
  suggestedRole: string
  score: number
  fitReason: string
  uncertainty: string
  sharedConstraints: string[]
  remainingRoleGaps: string[]
  commonFreeTimes: GroupTimeSuggestion[]
  inference: InferenceMetadata
}

export interface GroupMatchInboxItem {
  invitation: GroupMatchInvitation
  organizer: GroupMatchOrganizer
  category: GroupMatchCategory
  title: string
  description: string
  timeWindow: string
  location: string
  minMembers: number
  maxMembers: number
  status: GroupMatchStatus
  expiresAt: string
}

export interface GroupMatchIntentInput {
  category: GroupMatchCategory
  title: string
  description: string
  roleOffer: string
  rolesNeeded: string[]
  constraints: string[]
  timeWindow: string
  location: string
  minMembers: number
  maxMembers: number
  startsAt?: string
  endsAt?: string
  createCalendarEvent: boolean
  expiresAt?: string
}

export function fetchMyGroupMatchIntents() {
  return apiRequest<GroupMatchIntent[]>('/group-intents/', { auth: true })
}

export function fetchGroupMatchIntent(intentId: number) {
  return apiRequest<GroupMatchIntent>(`/group-intents/${intentId}`, { auth: true })
}

export function createGroupMatchIntent(input: GroupMatchIntentInput) {
  return apiRequest<GroupMatchIntent>('/group-intents/', {
    method: 'POST',
    auth: true,
    body: { ...input },
  })
}

export function fetchGroupMatchRecommendations(intentId: number) {
  return apiRequest<GroupMatchRecommendation[]>(
    `/group-intents/${intentId}/recommendations`,
    { auth: true },
  )
}

export function inviteGroupMatchCandidate(intentId: number, candidateIntentId: number) {
  return apiRequest<GroupMatchInvitation>(`/group-intents/${intentId}/invitations`, {
    method: 'POST',
    auth: true,
    body: { candidateIntentId },
  })
}

export function fetchMyGroupMatchInvitations() {
  return apiRequest<GroupMatchInboxItem[]>('/group-invitations', { auth: true })
}

export function respondGroupMatchInvitation(
  invitationId: number,
  status: 'accepted' | 'declined',
) {
  return apiRequest<GroupMatchIntent>(`/group-invitations/${invitationId}/response`, {
    method: 'PUT',
    auth: true,
    body: { status },
  })
}

export function formGroupMatchIntent(intentId: number) {
  return apiRequest<GroupMatchIntent>(`/group-intents/${intentId}/form`, {
    method: 'POST',
    auth: true,
  })
}

export function leaveGroupMatchIntent(intentId: number) {
  return apiRequest<GroupMatchIntent>(`/group-intents/${intentId}/leave`, {
    method: 'POST',
    auth: true,
  })
}
