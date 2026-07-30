import { apiRequest } from '@/lib/api'

export interface MatchContact {
  type: string
  label: string
  value: string
}

export interface MatchPartner {
  userId?: number
  nickname?: string
  schoolName?: string
  department?: string
  grade?: string
  contacts: MatchContact[]
}

export interface MatchReasonJson {
  text?: string
  error?: string
  match_summary?: string
  match_reasons?: string[]
  date_topics?: string[]
  cautions?: string[]
}

export interface MatchResult {
  matchId: number
  cycleId: number
  cycleKey: string
  cycleTitle: string
  revealStatus: string
  revealAt: string
  resultStatus: 'matched' | 'failed'
  message: string
  score?: number
  reason?: string
  reasonJson?: MatchReasonJson
  partner?: MatchPartner
}

export interface MatchHistoryResult {
  items: MatchResult[]
}

export interface MatchCycleStatus {
  hasCurrentCycle: boolean
  cycleId?: number
  cycleKey?: string
  cycleTitle?: string
  revealStatus?: string
  revealAt?: string
  participantCount?: number
  matchedUserCount?: number
  successRatePercent?: number
  hasUserResult: boolean
  userResultStatus?: 'matched' | 'failed'
  userRevealAt?: string
  canViewUserResult?: boolean
}

export interface MatchContactResult {
  action: 'chat_ready' | 'request_sent'
  friendshipId?: number
  conversationId?: number
  message?: string
}

export function fetchCurrentMatch() {
  return apiRequest<MatchResult>('/matches/current', {
    auth: true,
  })
}

export function fetchMatchHistory(limit = 10) {
  return apiRequest<MatchHistoryResult>(`/matches/history?limit=${limit}`, {
    auth: true,
  })
}

export function fetchMatchCycleStatus() {
  return apiRequest<MatchCycleStatus>('/matches/cycle/status', {
    auth: true,
  })
}

export function fetchMatchDetail(matchId: number) {
  return apiRequest<MatchResult>(`/matches/${matchId}`, {
    auth: true,
  })
}

export function contactMatchedPartner(matchId: number) {
  return apiRequest<MatchContactResult>(`/matches/${matchId}/contact`, {
    method: 'POST',
    auth: true,
  })
}
