import { apiRequest } from '@/lib/api'

export type EngagementScene = 'meal' | 'study' | 'sport' | 'romantic' | 'project'
export type NotificationPermissionState =
  | 'default'
  | 'granted'
  | 'denied'
  | 'unavailable'

export interface EngagementPreferences {
  notificationsEnabled: boolean
  newMessages: boolean
  meetupInvites: boolean
  meetupReminders: boolean
  dailyPicks: boolean
  permissionState: NotificationPermissionState
  timezone: string
  selectedScenes: EngagementScene[]
  onboardingCompleted: boolean
  onboardingCompletedAt?: string
  updatedAt?: string
}

export interface NotificationPreferencesInput {
  notificationsEnabled: boolean
  newMessages: boolean
  meetupInvites: boolean
  meetupReminders: boolean
  dailyPicks: boolean
  permissionState: NotificationPermissionState
  timezone: string
}

export interface CampusTheme {
  key: string
  title: string
  description: string
  category: EngagementScene | 'event'
  defaultPrompt: string
  scheduleHint: string
  iconName: string
  isActive: boolean
  displayOrder: number
  selected: boolean
  updatedAt: string
}

export type NotificationKind =
  | 'new_message'
  | 'meetup_invite'
  | 'meetup_reminder'
  | 'daily_picks'

export interface UserNotification {
  id: number
  kind: NotificationKind
  title: string
  body: string
  actionPath: string
  scheduledFor: string
  deliveredAt?: string
  readAt?: string
  createdAt: string
}

export interface NotificationPage {
  items: UserNotification[]
  unreadCount: number
  nextCursor?: number
}

export interface CampusThemeMetric {
  key: string
  title: string
  category: string
  supplyUsers: number
  demandUsers: number
  groupIntents: number
  groupsFormed: number
  groupFormationRate: number
  meetupsEligible: number
  meetupsCompleted: number
  meetupCompletionRate: number
}

export interface CampusThemeMetrics {
  days: number
  from: string
  to: string
  themes: CampusThemeMetric[]
  generatedAt: string
}

export interface UpdateCampusThemeInput {
  title: string
  description: string
  defaultPrompt: string
  scheduleHint: string
  iconName: string
  isActive: boolean
  displayOrder: number
}

export function fetchEngagementPreferences() {
  return apiRequest<EngagementPreferences>('/users/me/engagement', { auth: true })
}

export function updateNotificationPreferences(input: NotificationPreferencesInput) {
  return apiRequest<EngagementPreferences>('/users/me/notification-preferences', {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}

export function completeOnboarding(selectedScenes: EngagementScene[]) {
  return apiRequest<EngagementPreferences>('/users/me/onboarding', {
    method: 'PUT',
    auth: true,
    body: { selectedScenes },
  })
}

export function fetchCampusThemes() {
  return apiRequest<CampusTheme[]>('/campus-themes', { auth: true })
}

export function replaceCampusThemes(themeKeys: string[]) {
  return apiRequest<CampusTheme[]>('/users/me/campus-themes', {
    method: 'PUT',
    auth: true,
    body: { themeKeys },
  })
}

export function fetchNotifications(before?: number, limit = 20) {
  const params = new URLSearchParams({ limit: String(limit) })
  if (before) params.set('before', String(before))
  return apiRequest<NotificationPage>(`/notifications/?${params.toString()}`, {
    auth: true,
  })
}

export function markNotificationRead(notificationId: number) {
  return apiRequest<null>(`/notifications/${notificationId}/read`, {
    method: 'POST',
    auth: true,
  })
}

export function markNotificationDelivered(notificationId: number) {
  return apiRequest<null>(`/notifications/${notificationId}/delivered`, {
    method: 'POST',
    auth: true,
  })
}

export function markAllNotificationsRead() {
  return apiRequest<null>('/notifications/read-all', {
    method: 'POST',
    auth: true,
  })
}

export function fetchCampusThemeMetrics(days: 7 | 28 | 90) {
  return apiRequest<CampusThemeMetrics>(`/admin/campus-themes/metrics?days=${days}`, {
    auth: true,
  })
}

export function fetchAdminCampusThemes() {
  return apiRequest<CampusTheme[]>('/admin/campus-themes', { auth: true })
}

export function updateCampusTheme(themeKey: string, input: UpdateCampusThemeInput) {
  return apiRequest<CampusTheme>(`/admin/campus-themes/${themeKey}`, {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}
