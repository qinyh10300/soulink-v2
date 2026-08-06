import { apiRequest } from '@/lib/api'

export interface ProductMetricsDashboard {
  period: {
    days: number
    from: string
    to: string
  }
  northStar: {
    meetupsCompleted: number
    meetupsConfirmed: number
    matchToMeetupRate: number
    repeatIntentRate: number
  }
  activity: {
    registeredUsers: number
    activeUsers: number
    intentCreators: number
    repeatIntentUsers: number
  }
  safety: {
    openReports: number
    reportRate: number
  }
  funnel: Array<{
    key: string
    label: string
    count: number
    conversionFromPrevious: number
  }>
  events: Array<{
    eventName: string
    count: number
    uniqueActors: number
  }>
  daily: Array<{
    date: string
    activeUsers: number
    intentsCreated: number
    recommendationsViewed: number
    interestsExpressed: number
    matchesCreated: number
    meetupsConfirmed: number
    meetupsCompleted: number
    safetyReports: number
  }>
  generatedAt: string
}

export function fetchProductMetrics(days: 7 | 28 | 90) {
  return apiRequest<ProductMetricsDashboard>(`/admin/product-metrics?days=${days}`, { auth: true })
}
