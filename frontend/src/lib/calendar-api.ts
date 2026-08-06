import { apiRequest } from '@/lib/api'

export type CalendarEventType =
  | 'meetup'
  | 'study'
  | 'sports'
  | 'meal'
  | 'date'
  | 'project'
  | 'other'
  | 'busy'

export type CalendarVisibility = 'public' | 'busy_only' | 'private'
export type CalendarRSVP = 'pending' | 'accepted' | 'declined'

export interface CalendarAttendee {
  userId: number
  nickname: string
  department: string
  grade: string
  status: CalendarRSVP
}

export interface CalendarEvent {
  id: number
  organizerId?: number
  organizerNickname?: string
  organizerDepartment?: string
  eventType: CalendarEventType
  title: string
  description: string
  location: string
  startsAt: string
  endsAt: string
  visibility: CalendarVisibility
  isOrganizer: boolean
  attendeeStatus?: CalendarRSVP
  isBusyOnly: boolean
  attendees: CalendarAttendee[]
  createdAt: string
}

export interface CalendarEventInput {
  eventType: Exclude<CalendarEventType, 'busy'>
  title: string
  description: string
  location: string
  startsAt: string
  endsAt: string
  visibility: CalendarVisibility
  inviteeIds: number[]
}

function rangeQuery(from: Date, to: Date) {
  const params = new URLSearchParams({
    from: from.toISOString(),
    to: to.toISOString(),
  })
  return params.toString()
}

export function fetchCalendarEvents(from: Date, to: Date) {
  return apiRequest<CalendarEvent[]>(`/calendar/events?${rangeQuery(from, to)}`, {
    auth: true,
  })
}

export function fetchUserSchedule(userId: number, from: Date, to: Date) {
  return apiRequest<CalendarEvent[]>(
    `/users/${userId}/calendar?${rangeQuery(from, to)}`,
    { auth: true },
  )
}

export function createCalendarEvent(input: CalendarEventInput) {
  return apiRequest<CalendarEvent>('/calendar/events', {
    method: 'POST',
    auth: true,
    body: { ...input },
  })
}

export function updateCalendarEvent(eventId: number, input: CalendarEventInput) {
  return apiRequest<CalendarEvent>(`/calendar/events/${eventId}`, {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}

export function cancelCalendarEvent(eventId: number) {
  return apiRequest<null>(`/calendar/events/${eventId}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function respondCalendarEvent(eventId: number, status: 'accepted' | 'declined') {
  return apiRequest<CalendarEvent>(`/calendar/events/${eventId}/rsvp`, {
    method: 'PUT',
    auth: true,
    body: { status },
  })
}
