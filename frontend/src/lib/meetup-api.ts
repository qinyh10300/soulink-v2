import { apiRequest } from '@/lib/api'

export type MeetupContextType = 'match_room' | 'direct' | 'group'
export type MeetupStatus =
  | 'proposed'
  | 'confirmed'
  | 'declined'
  | 'reschedule_requested'
  | 'cancelled'
  | 'completed'
  | 'no_show'

export interface MeetupContext {
  type: MeetupContextType
  id: number
}

export interface AvailabilityWindow {
  id: number
  windowType: 'one_time' | 'weekly'
  timezone: string
  startsAt?: string
  endsAt?: string
  weekday?: number
  startMinute?: number
  endMinute?: number
  validFrom?: string
  validUntil?: string
}

export type AvailabilityWindowInput = Omit<AvailabilityWindow, 'id'>

export interface TimeSuggestion {
  startsAt: string
  endsAt: string
  label: string
  explanation: string
}

export interface CampusPlace {
  name: string
  area: string
  bestFor: string[]
}

export interface MeetupSuggestions {
  participantCount: number
  missingAvailabilityCount: number
  candidates: TimeSuggestion[]
  places: CampusPlace[]
}

export interface MeetupParticipant {
  userId: number
  nickname: string
  department: string
  grade: string
  status: 'pending' | 'accepted' | 'declined'
  isMe: boolean
  isOrganizer: boolean
}

export interface MeetupFeedback {
  attendance: 'attended' | 'no_show' | 'cancelled'
  rating?: number
  wouldMeetAgain?: boolean
  safetyConcern: boolean
  privateNotes: string
  submittedAt: string
}

export interface MeetupReminder {
  stage:
    | 'awaiting_response'
    | 'awaiting_others'
    | 'confirmed'
    | 'day_before'
    | 'soon'
    | 'underway'
    | 'feedback_due'
    | 'complete'
    | 'closed'
    | 'none'
  text: string
}

export interface Meetup {
  id: number
  context: MeetupContext
  organizerId: number
  title: string
  category: string
  startsAt: string
  endsAt: string
  location: string
  status: MeetupStatus
  version: number
  calendarEventId?: number
  participants: MeetupParticipant[]
  myStatus: 'pending' | 'accepted' | 'declined'
  canRespond: boolean
  canReschedule: boolean
  canCancel: boolean
  canComplete: boolean
  reminder: MeetupReminder
  suggestedPlaces: CampusPlace[]
  myFeedback?: MeetupFeedback
  createdAt: string
  updatedAt: string
}

export interface CreateMeetupInput {
  context: MeetupContext
  title: string
  category: string
  startsAt: string
  endsAt: string
  location: string
}

export interface FeedbackInput {
  attendance: 'attended' | 'no_show' | 'cancelled'
  rating: number | null
  wouldMeetAgain: boolean | null
  safetyConcern: boolean
  privateNotes: string
}

export interface TrustedContact {
  id: number
  label: string
  contactType: 'email' | 'phone' | 'wechat' | 'other'
  contactValue: string
  updatedAt: string
}

export interface MeetupSafetyShare {
  id: number
  trustedContact: TrustedContact
  summary: {
    meetupId: number
    startsAt: string
    endsAt: string
    location: string
    participantCount: number
    status: string
  }
  deliveryStatus: 'sent' | 'recorded' | 'failed'
  manualText?: string
  createdAt: string
}

export function fetchAvailability() {
  return apiRequest<AvailabilityWindow[]>('/users/me/availability', { auth: true })
}

export function replaceAvailability(windows: AvailabilityWindowInput[]) {
  return apiRequest<AvailabilityWindow[]>('/users/me/availability', {
    method: 'PUT',
    auth: true,
    body: { windows },
  })
}

export function suggestMeetupTimes(
  context: MeetupContext,
  input: { from: string; to: string; durationMinutes: number },
) {
  return apiRequest<MeetupSuggestions>('/meetup-suggestions', {
    method: 'POST',
    auth: true,
    body: { context, ...input },
  })
}

export function fetchMeetups(context?: MeetupContext, from?: Date, to?: Date) {
  const params = new URLSearchParams()
  if (context) {
    params.set('contextType', context.type)
    params.set('contextId', String(context.id))
  }
  if (from) params.set('from', from.toISOString())
  if (to) params.set('to', to.toISOString())
  const query = params.size > 0 ? `?${params.toString()}` : ''
  return apiRequest<Meetup[]>(`/meetups/${query}`, { auth: true })
}

export function fetchMeetup(meetupId: number) {
  return apiRequest<Meetup>(`/meetups/${meetupId}`, { auth: true })
}

export function createMeetup(input: CreateMeetupInput) {
  return apiRequest<Meetup>('/meetups/', {
    method: 'POST',
    auth: true,
    body: { ...input },
  })
}

export function respondMeetup(
  meetupId: number,
  action: 'accept' | 'decline',
  expectedVersion: number,
) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/response`, {
    method: 'PUT',
    auth: true,
    body: { action, expectedVersion },
  })
}

export function rescheduleMeetup(
  meetupId: number,
  input: {
    startsAt: string
    endsAt: string
    location: string
    expectedVersion: number
  },
) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/reschedule`, {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function cancelMeetup(meetupId: number, expectedVersion: number, reason: string) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/cancel`, {
    method: 'PUT',
    auth: true,
    body: { expectedVersion, reason },
  })
}

export function completeMeetup(
  meetupId: number,
  expectedVersion: number,
  outcome: 'completed' | 'no_show',
) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/complete`, {
    method: 'PUT',
    auth: true,
    body: { expectedVersion, outcome },
  })
}

export function submitMeetupFeedback(meetupId: number, input: FeedbackInput) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/feedback`, {
    method: 'POST',
    auth: true,
    body: { ...input },
  })
}

export function revokeMeetupFeedback(meetupId: number) {
  return apiRequest<Meetup>(`/meetups/${meetupId}/feedback`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchTrustedContacts() {
  return apiRequest<TrustedContact[]>('/users/me/trusted-contacts', { auth: true })
}

export function createTrustedContact(input: {
  label: string
  contactType: TrustedContact['contactType']
  contactValue: string
}) {
  return apiRequest<TrustedContact>('/users/me/trusted-contacts', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function deleteTrustedContact(contactId: number) {
  return apiRequest<null>(`/users/me/trusted-contacts/${contactId}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function shareMeetupSafetySummary(meetupId: number, trustedContactId: number) {
  return apiRequest<MeetupSafetyShare>(`/meetups/${meetupId}/safety-share`, {
    method: 'POST',
    auth: true,
    body: { trustedContactId },
  })
}
