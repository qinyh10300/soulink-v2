import { apiRequest } from '@/lib/api'

export type IntentCategory =
  | 'meal'
  | 'study'
  | 'sport'
  | 'event'
  | 'conversation'
  | 'romantic'
  | 'project'
  | 'cofounder'
  | 'other'

export type IntentVisibility = 'private' | 'candidates' | 'feed'
export type IntentStatus = 'draft' | 'active' | 'paused' | 'fulfilled' | 'expired'

export interface ClarifyingQuestion {
  id: string
  prompt: string
  options: string[]
}

export interface IntentInterpretation {
  category: IntentCategory
  title: string
  description: string
  goal: string
  timeWindow: string
  location: string
  visibility: IntentVisibility
  constraints: string[]
  desiredNextStep: string
  confidence: number
  questions: ClarifyingQuestion[]
  inferences: string[]
  projectMemo?: ProjectMemo
  safety: {
    level: 'clear' | 'review'
    notice: string
  }
  engine: string
	 source: 'model' | 'deterministic'
	 model?: string
	 promptVersion: string
	 editable: boolean
	 fallbackReason?: string
}

export interface InferenceMetadata {
	 source: 'model' | 'deterministic'
	 model?: string
	 confidence: number
	 promptVersion: string
	 editable: boolean
}

export interface MatchingProfile {
  bio: string
  skills: string[]
  interests: string[]
  availability: string
  preferredScenes: string[]
  collaborationStyle: string
  boundaries: string
  aiPortrait: string
  aiPortraitApproved: boolean
  romanticOptIn: boolean
  matchingPaused: boolean
  hiddenCategories: IntentCategory[]
  updatedAt?: string
}

export interface ProjectMemo {
  building: string
  stage: string
  weeklyTime: string
  compensation: string
  requiredSkills: string[]
  niceToHaveSkills: string[]
  workingStyle: string
  riskAppetite: string
}

export interface IntentInput {
  category: IntentCategory
  title: string
  description: string
  goal: string
  timeWindow: string
  location: string
  visibility: IntentVisibility
  status: IntentStatus
  constraints: string[]
  dealBreakers: string[]
  desiredNextStep: string
  projectMemo: ProjectMemo | null
  expiresAt: string | null
}

export interface IntentAuthor {
  nickname: string
  department: string
  grade: string
  verified: boolean
}

export interface ConnectionIntent extends IntentInput {
  id: number
  author: IntentAuthor
  interestCount: number
  interestState: '' | 'interested' | 'accepted' | 'declined' | 'withdrawn'
  savedByMe: boolean
  isOwner: boolean
  publishedAt?: string
  createdAt: string
  updatedAt: string
}

export interface IntentRecommendation {
  intent: ConnectionIntent
  score: number
  fitReason: string
  uncertainty: string
  sharedContext: string[]
	 inference: InferenceMetadata
}

export interface DailyRecommendations {
	 sourceIntentId?: number
	 generatedAt?: string
	 refreshAt?: string
	 algorithm: string
	 items: IntentRecommendation[]
}

export interface RecommendationAffinity {
	 category: IntentCategory
	 weight: number
	 signals: number
	 summary: string
}

export interface RecommendationSignal {
	 type: string
	 category: IntentCategory
	 impact: number
	 source: 'interaction' | 'meetup'
	 occurredAt: string
}

export interface RecommendationPreferences {
	 learningEnabled: boolean
	 affinities: RecommendationAffinity[]
	 recentSignals: RecommendationSignal[]
	 learningResetAt?: string
	 updatedAt?: string
}

export interface IntentInterest {
  userId: number
  nickname: string
  department: string
  grade: string
  bio: string
  skills: string[]
  collaborationStyle: string
  status: 'interested' | 'accepted' | 'declined'
  message: string
  createdAt: string
  updatedAt: string
}

export interface RoomParticipant {
  userId: number
  nickname: string
  department: string
  grade: string
  contact?: {
    type: string
    value: string
  }
}

export interface MatchRoom {
  id: number
  intentId: number
  intentTitle: string
  intentCategory: IntentCategory
  me: RoomParticipant
  other: RoomParticipant
  status: 'open' | 'met' | 'closed'
  fitReason: string
  uncertainty: string
  sharedContext: string[]
  conversationStarters: string[]
	 assistant: InferenceMetadata
  suggestedPlaces: string[]
  scheduledAt?: string
  selectedPlace: string
  myContactShared: boolean
  otherContactShared: boolean
  createdAt: string
  updatedAt: string
}

export interface MatchRoomMessage {
  id: number
  roomId: number
  senderId: number
  senderNickname: string
  content: string
  createdAt: string
}

export function interpretIntent(text: string) {
  return apiRequest<IntentInterpretation>('/intent-interpretations', {
    method: 'POST',
    auth: true,
    body: { text },
  })
}

export function fetchMatchingProfile() {
  return apiRequest<MatchingProfile>('/users/me/matching-profile', {
    auth: true,
  })
}

export function updateMatchingProfile(input: MatchingProfile) {
  return apiRequest<MatchingProfile>('/users/me/matching-profile', {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}

export function fetchIntentFeed(category?: IntentCategory | '') {
  const query = category ? `?category=${encodeURIComponent(category)}` : ''
  return apiRequest<ConnectionIntent[]>(`/intents/feed${query}`, {
    auth: true,
  })
}

export function fetchSavedIntents(category?: IntentCategory | '') {
  const query = category ? `?category=${encodeURIComponent(category)}` : ''
  return apiRequest<ConnectionIntent[]>(`/intents/saved${query}`, {
    auth: true,
  })
}

export function fetchMyIntents() {
  return apiRequest<ConnectionIntent[]>('/intents/mine', {
    auth: true,
  })
}

export function fetchIntent(intentId: number) {
  return apiRequest<ConnectionIntent>(`/intents/${intentId}`, {
    auth: true,
  })
}

export function createIntent(input: IntentInput) {
  return apiRequest<ConnectionIntent>('/intents/', {
    method: 'POST',
    auth: true,
    body: { ...input },
  })
}

export function updateIntent(intentId: number, input: IntentInput) {
  return apiRequest<ConnectionIntent>(`/intents/${intentId}`, {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}

export function expireIntent(intentId: number) {
  return apiRequest<{ expired: boolean }>(`/intents/${intentId}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function setIntentSaved(intentId: number, saved: boolean) {
  return apiRequest<ConnectionIntent>(`/intents/${intentId}/save`, {
    method: saved ? 'POST' : 'DELETE',
    auth: true,
  })
}

export function dismissIntent(intentId: number) {
  return apiRequest<{ dismissed: boolean }>(`/intents/${intentId}/dismiss`, {
    method: 'POST',
    auth: true,
  })
}

export function expressIntentInterest(intentId: number, message: string) {
  return apiRequest<ConnectionIntent>(`/intents/${intentId}/interest`, {
    method: 'POST',
    auth: true,
    body: { message },
  })
}

export function withdrawIntentInterest(intentId: number) {
  return apiRequest<ConnectionIntent>(`/intents/${intentId}/interest`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchIntentInterests(intentId: number) {
  return apiRequest<IntentInterest[]>(`/intents/${intentId}/interests`, {
    auth: true,
  })
}

export function respondToIntentInterest(
  intentId: number,
  candidateId: number,
  accepted: boolean,
) {
  return apiRequest<{ accepted: boolean; room: MatchRoom | null }>(
    `/intents/${intentId}/interests/${candidateId}/${accepted ? 'accept' : 'decline'}`,
    {
      method: 'POST',
      auth: true,
    },
  )
}

export function fetchIntentRecommendations(intentId: number) {
  return apiRequest<IntentRecommendation[]>(
    `/intents/${intentId}/recommendations`,
    { auth: true },
  )
}

export function fetchDailyRecommendations() {
	return apiRequest<DailyRecommendations>('/recommendations/daily', {
		auth: true,
	})
}

export function fetchRecommendationPreferences() {
	return apiRequest<RecommendationPreferences>(
		'/users/me/recommendation-preferences',
		{ auth: true },
	)
}

export function updateRecommendationPreferences(learningEnabled: boolean) {
	return apiRequest<RecommendationPreferences>(
		'/users/me/recommendation-preferences',
		{
			method: 'PUT',
			auth: true,
			body: { learningEnabled },
		},
	)
}

export function resetRecommendationPreferences() {
	return apiRequest<RecommendationPreferences>(
		'/users/me/recommendation-preferences/reset',
		{
			method: 'POST',
			auth: true,
		},
	)
}

export function fetchMatchRooms() {
  return apiRequest<MatchRoom[]>('/match-rooms/', {
    auth: true,
  })
}

export function fetchMatchRoom(roomId: number) {
  return apiRequest<MatchRoom>(`/match-rooms/${roomId}`, {
    auth: true,
  })
}

export function updateMatchRoomPlan(
  roomId: number,
  input: { scheduledAt: string | null; place: string },
) {
  return apiRequest<MatchRoom>(`/match-rooms/${roomId}/plan`, {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function updateMatchRoomStatus(roomId: number, status: 'met' | 'closed') {
  return apiRequest<MatchRoom>(`/match-rooms/${roomId}/status`, {
    method: 'PUT',
    auth: true,
    body: { status },
  })
}

export function setMatchRoomContactSharing(roomId: number, shared: boolean) {
  return apiRequest<MatchRoom>(`/match-rooms/${roomId}/contact-sharing`, {
    method: 'PUT',
    auth: true,
    body: { shared },
  })
}

export function fetchMatchRoomMessages(roomId: number) {
  return apiRequest<MatchRoomMessage[]>(`/match-rooms/${roomId}/messages`, {
    auth: true,
  })
}

export function createMatchRoomMessage(roomId: number, content: string) {
  return apiRequest<MatchRoomMessage>(`/match-rooms/${roomId}/messages`, {
    method: 'POST',
    auth: true,
    body: { content },
  })
}
