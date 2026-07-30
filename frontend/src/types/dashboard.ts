export interface DashboardUser {
  name: string
  avatarEmoji: string
  avatarUrl?: string | null
}

export interface CountdownInfo {
  label: string
  description: string
  targetAt: string
}

export interface DashboardStatusItem {
  id: string
  title: string
  value: string
  tone: 'success' | 'active' | 'neutral'
  description?: string
  progress?: number
  actionLabel?: string
  to?: string
}

export interface DashboardFeatureItem {
  id: string
  title: string
  description: string
  meta: string
  icon: string
  emphasized?: boolean
  disabled?: boolean
  to?: string
}

export interface MatchHistoryEmptyState {
  title: string
  description: string
  icon: string
}

export interface MatchHistoryListItem {
  id: number
  title: string
  subtitle: string
  message: string
  statusLabel: string
  statusTone: 'success' | 'neutral'
  scoreLabel?: string
  canContact?: boolean
}

export interface WeeklyMatchMockResult {
  cycleLabel: string
  avatarEmoji: string
  name: string
  department: string
  grade: string
  score: number
  matchReasons: string[]
  suggestedTopics: string[]
  cautions: string[]
  contactLabel: string
  contactValue: string
}

export interface DashboardContent {
  user: DashboardUser
  countdown: CountdownInfo
  statuses: DashboardStatusItem[]
  features: DashboardFeatureItem[]
  weeklyMockResult: WeeklyMatchMockResult
  history: MatchHistoryEmptyState
}
