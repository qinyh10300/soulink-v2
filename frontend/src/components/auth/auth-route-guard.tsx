import { useEffect, useState } from 'react'
import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { ScrollManager } from '@/components/shared/scroll-manager'
import { ChatProvider } from '@/contexts/chat-context'
import { ApiError } from '@/lib/api'
import { fetchCurrentUser } from '@/lib/auth-api'
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
      <div className="min-h-screen bg-[#f7f2ec] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-10 text-center text-lg text-[#7e6d61] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          正在验证登录状态...
        </div>
      </div>
    )
  }

  if (status === 'unavailable') {
    return (
      <div className="min-h-screen bg-[#f7f2ec] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl space-y-6 rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-10 text-center shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          <div className="space-y-2">
            <h1 className="font-serif text-4xl text-[#2f221b]">暂时无法验证登录状态</h1>
            <p className="text-base leading-7 text-[#7e6d61]">
              当前无法连接到认证服务。你可以稍后重试，或重新登录后再继续。
            </p>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-3">
            <button
              className="inline-flex rounded-full bg-[#8f6548] px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-[#7f5a40]"
              onClick={() => setRetryKey((value) => value + 1)}
              type="button"
            >
              重新验证
            </button>
            <button
              className="inline-flex rounded-full border border-[#d9c8bb] bg-white px-5 py-3 text-sm font-medium text-[#5f4d42] transition-colors hover:border-[#c7b09e] hover:text-[#2f221b]"
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
      <div className="min-h-screen bg-[#f7f2ec] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-10 text-center text-lg text-[#7e6d61] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          正在验证登录状态...
        </div>
      </div>
    )
  }

  if (status === 'authenticated') {
    return <Navigate replace to="/home" />
  }

  return <Outlet />
}
