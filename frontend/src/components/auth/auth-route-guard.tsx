import { useEffect, useState } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { ScrollManager } from '@/components/shared/scroll-manager'
import { ChatProvider } from '@/contexts/chat-context'
import { ApiError } from '@/lib/api'
import { fetchCurrentUser } from '@/lib/auth-api'
import { connectionPaths } from '@/lib/connection-routes'
import { clearSurveyDraft } from '@/lib/survey-storage'
import {
  clearAuthSession,
  clearUserProfile,
  getAuthSession,
  syncAuthenticatedUser,
} from '@/lib/user-profile-storage'

function clearStoredAuthentication() {
  clearAuthSession()
  clearUserProfile()
  clearSurveyDraft()
}

export function RequireAuth() {
  const location = useLocation()
  const session = getAuthSession()
  const [status, setStatus] = useState<
    'checking' | 'ready' | 'unauthorized' | 'unavailable'
  >(session?.token ? 'checking' : 'unauthorized')
  const [retryKey, setRetryKey] = useState(0)

  useEffect(() => {
    if (!session?.token) {
      setStatus('unauthorized')
      return
    }

    let cancelled = false

    const validateSession = async () => {
      setStatus('checking')

      try {
        const user = await fetchCurrentUser()

        if (cancelled) {
          return
        }

        syncAuthenticatedUser(user)
        setStatus('ready')
      } catch (error) {
        if (cancelled) {
          return
        }

        if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
          clearStoredAuthentication()
          setStatus('unauthorized')
          return
        }

        setStatus('unavailable')
      }
    }

    void validateSession()

    return () => {
      cancelled = true
    }
  }, [retryKey, session?.token])

  if (!session?.token || status === 'unauthorized') {
    return <Navigate replace state={{ from: location }} to="/login" />
  }

  if (status === 'checking') {
    return (
      <div className="min-h-screen bg-[#fafafa] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-lg border border-[#e4e4e7] bg-white p-10 text-center text-base text-[#71717a] shadow-[0_18px_54px_rgba(9,9,11,0.06)]">
          正在验证登录状态...
        </div>
      </div>
    )
  }

  if (status === 'unavailable') {
    return (
      <div className="min-h-screen bg-[#fafafa] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl space-y-6 rounded-lg border border-[#e4e4e7] bg-white p-10 text-center shadow-[0_18px_54px_rgba(9,9,11,0.06)]">
          <div className="space-y-2">
            <h1 className="text-3xl font-semibold text-[#09090b]">暂时无法验证登录状态</h1>
            <p className="text-base leading-7 text-[#71717a]">
              当前无法连接到认证服务。你可以稍后重试，或重新登录后再继续。
            </p>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-3">
            <button
              className="inline-flex rounded-md bg-[#1447e6] px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-[#103bc4]"
              onClick={() => setRetryKey((value) => value + 1)}
              type="button"
            >
              重新验证
            </button>
            <button
              className="inline-flex rounded-md border border-[#d4d4d8] bg-white px-5 py-3 text-sm font-medium text-[#52525b] transition-colors hover:bg-[#f4f4f5] hover:text-[#18181b]"
              onClick={() => {
                clearStoredAuthentication()
                setStatus('unauthorized')
              }}
              type="button"
            >
              退出登录
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <ChatProvider>
      <ScrollManager />
      <Outlet />
    </ChatProvider>
  )
}

export function RedirectIfAuthenticated() {
  const session = getAuthSession()
  const [status, setStatus] = useState<'checking' | 'authenticated' | 'unauthenticated'>(
    session?.token ? 'checking' : 'unauthenticated',
  )

  useEffect(() => {
    if (!session?.token) {
      setStatus('unauthenticated')
      return
    }

    let cancelled = false

    const validateSession = async () => {
      setStatus('checking')

      try {
        const user = await fetchCurrentUser()

        if (cancelled) {
          return
        }

        syncAuthenticatedUser(user)
        setStatus('authenticated')
      } catch (error) {
        if (cancelled) {
          return
        }

        if (error instanceof ApiError && (error.status === 401 || error.status === 403)) {
          clearStoredAuthentication()
        }

        setStatus('unauthenticated')
      }
    }

    void validateSession()

    return () => {
      cancelled = true
    }
  }, [session?.token])

  if (status === 'checking') {
    return (
      <div className="min-h-screen bg-[#fafafa] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-lg border border-[#e4e4e7] bg-white p-10 text-center text-base text-[#71717a] shadow-[0_18px_54px_rgba(9,9,11,0.06)]">
          正在验证登录状态...
        </div>
      </div>
    )
  }

  if (status === 'authenticated') {
    return <Navigate replace to={connectionPaths.root} />
  }

  return <Outlet />
}
