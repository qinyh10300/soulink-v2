import type { DashboardUser } from '@/types/dashboard'
import type { AuthSession, UserProfile } from '@/types/profile'
import { createDefaultProfile } from '@/mock/profile'

const USER_PROFILE_KEY = 'soulink:user-profile'
const AUTH_SESSION_KEY = 'soulink:auth-session'

interface StoredAuthUser {
  id: number
  email: string
  nickname: string
  schoolName: string
  department: string
  grade: string
  contactType?: string
  contactValue?: string
}

function hasWindow() {
  return typeof window !== 'undefined'
}

function decodeJwtPayload(token: string) {
  const segments = token.split('.')

  if (segments.length < 2 || !hasWindow()) {
    return null
  }

  try {
    const base64 = segments[1].replace(/-/g, '+').replace(/_/g, '/')
    const padded = base64.padEnd(Math.ceil(base64.length / 4) * 4, '=')
    const decoded = window.atob(padded)

    return JSON.parse(decoded) as { exp?: number }
  } catch {
    return null
  }
}

export function getAuthSession(): AuthSession | null {
  if (!hasWindow()) {
    return null
  }

  const raw = window.localStorage.getItem(AUTH_SESSION_KEY)

  if (!raw) {
    return null
  }

  try {
    const parsed = JSON.parse(raw) as Partial<AuthSession>

    if (
      typeof parsed.token !== 'string' ||
      typeof parsed.userId !== 'number' ||
      typeof parsed.email !== 'string' ||
      typeof parsed.nickname !== 'string' ||
      typeof parsed.loggedInAt !== 'string'
    ) {
      return null
    }

    return parsed as AuthSession
  } catch {
    return null
  }
}

export function hasActiveAuthSession() {
  const session = getAuthSession()

  if (!session?.token) {
    return false
  }

  const payload = decodeJwtPayload(session.token)

  if (typeof payload?.exp === 'number') {
    return payload.exp * 1000 > Date.now()
  }

  return true
}

export function setAuthSession(session: AuthSession) {
  if (!hasWindow()) {
    return
  }

  window.localStorage.setItem(AUTH_SESSION_KEY, JSON.stringify(session))
}

export function clearAuthSession() {
  if (!hasWindow()) {
    return
  }

  window.localStorage.removeItem(AUTH_SESSION_KEY)
}

export function saveAuthenticatedUser(token: string, user: StoredAuthUser) {
  syncAuthenticatedUser(user, token)
}

export function syncAuthenticatedUser(user: StoredAuthUser, token?: string) {
  const existingProfile = getStoredUserProfile(user.email)
  const currentSession = getAuthSession()

  if (token || currentSession) {
    setAuthSession({
      token: token ?? currentSession?.token ?? '',
      userId: user.id,
      email: user.email,
      nickname: user.nickname,
      loggedInAt: currentSession?.loggedInAt ?? new Date().toISOString(),
    })
  }

  saveUserProfile({
    ...createDefaultProfile(user.email),
    ...existingProfile,
    email: user.email,
    nickname: user.nickname,
    schoolName: user.schoolName,
    department: user.department,
    grade: user.grade,
    contactType: (user.contactType ?? existingProfile.contactType ?? '') as UserProfile['contactType'],
    contactValue: user.contactValue ?? existingProfile.contactValue ?? '',
  })
}

export function getStoredUserProfile(fallbackEmail?: string): UserProfile {
  const email = fallbackEmail || getAuthSession()?.email || 'yourname@pku.edu.cn'

  if (!hasWindow()) {
    return createDefaultProfile(email)
  }

  const raw = window.localStorage.getItem(USER_PROFILE_KEY)

  if (!raw) {
    return createDefaultProfile(email)
  }

  try {
    const parsed = JSON.parse(raw) as Partial<UserProfile>

    return {
      ...createDefaultProfile(email),
      ...parsed,
      email,
    }
  } catch {
    return createDefaultProfile(email)
  }
}

export function saveUserProfile(profile: UserProfile) {
  if (!hasWindow()) {
    return
  }

  window.localStorage.setItem(USER_PROFILE_KEY, JSON.stringify(profile))
}

export function clearUserProfile() {
  if (!hasWindow()) {
    return
  }

  window.localStorage.removeItem(USER_PROFILE_KEY)
}

export function resolveDashboardUser(defaultUser: DashboardUser): DashboardUser {
  const profile = getStoredUserProfile()

  return {
    ...defaultUser,
    name: profile.nickname || defaultUser.name,
    avatarUrl: profile.avatar || null,
  }
}
