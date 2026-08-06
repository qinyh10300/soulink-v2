import { lazy, Suspense, type ReactNode } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'

import {
  RedirectIfAuthenticated,
  RequireAuth,
} from '@/components/auth/auth-route-guard'
import { ConnectionShell } from '@/components/connection/connection-shell'

const ConnectionHomePage = lazy(async () => {
  const module = await import('@/pages/connection-home-page')
  return { default: module.ConnectionHomePage }
})
const ConnectionProfilePage = lazy(async () => {
  const module = await import('@/pages/connection-profile-page')
  return { default: module.ConnectionProfilePage }
})
const FeedbackPage = lazy(async () => {
  const module = await import('@/pages/feedback-page')
  return { default: module.FeedbackPage }
})
const IntentComposerPage = lazy(async () => {
  const module = await import('@/pages/intent-composer-page')
  return { default: module.IntentComposerPage }
})
const IntentDetailPage = lazy(async () => {
  const module = await import('@/pages/intent-detail-page')
  return { default: module.IntentDetailPage }
})
const IntentFeedPage = lazy(async () => {
  const module = await import('@/pages/intent-feed-page')
  return { default: module.IntentFeedPage }
})
const LoginPage = lazy(async () => {
  const module = await import('@/pages/login-page')
  return { default: module.LoginPage }
})
const MatchRoomPage = lazy(async () => {
  const module = await import('@/pages/match-room-page')
  return { default: module.MatchRoomPage }
})
const MatchRoomsPage = lazy(async () => {
  const module = await import('@/pages/match-rooms-page')
  return { default: module.MatchRoomsPage }
})
const MyIntentsPage = lazy(async () => {
  const module = await import('@/pages/my-intents-page')
  return { default: module.MyIntentsPage }
})
const ProfilePage = lazy(async () => {
  const module = await import('@/pages/profile-page')
  return { default: module.ProfilePage }
})
const PlaygroundPage = lazy(async () => {
  const module = await import('@/pages/playground-page')
  return { default: module.PlaygroundPage }
})
const PublicProfilePage = lazy(async () => {
  const module = await import('@/pages/public-profile-page')
  return { default: module.PublicProfilePage }
})
const RegisterPage = lazy(async () => {
  const module = await import('@/pages/register-page')
  return { default: module.RegisterPage }
})
const SafetyPage = lazy(async () => {
  const module = await import('@/pages/safety-page')
  return { default: module.SafetyPage }
})
const SafetyAdminPage = lazy(async () => {
  const module = await import('@/pages/safety-admin-page')
  return { default: module.SafetyAdminPage }
})
const SurveyPage = lazy(async () => {
  const module = await import('@/pages/survey-page')
  return { default: module.SurveyPage }
})
const V2LandingPage = lazy(async () => {
  const module = await import('@/pages/v2-landing-page')
  return { default: module.V2LandingPage }
})

function withSuspense(element: ReactNode) {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center bg-[#f7f6f3] text-sm text-[#77736c]">
          正在打开 SouLink...
        </div>
      }
    >
      {element}
    </Suspense>
  )
}

export const router = createBrowserRouter([
  {
    element: <RequireAuth />,
    children: [
      {
        element: <ConnectionShell />,
        children: [
          {
            path: '/home',
            element: withSuspense(<ConnectionHomePage />),
          },
          {
            path: '/discover',
            element: withSuspense(<IntentFeedPage />),
          },
          {
            path: '/intents',
            element: withSuspense(<MyIntentsPage />),
          },
          {
            path: '/intents/new',
            element: withSuspense(<IntentComposerPage />),
          },
          {
            path: '/intents/:intentId',
            element: withSuspense(<IntentDetailPage />),
          },
          {
            path: '/matches',
            element: withSuspense(<MatchRoomsPage />),
          },
          {
            path: '/matches/:roomId',
            element: withSuspense(<MatchRoomPage />),
          },
          {
            path: '/profile',
            element: withSuspense(<ConnectionProfilePage />),
          },
          {
            path: '/account',
            element: withSuspense(<ProfilePage />),
          },
          {
            path: '/profile/:userId',
            element: withSuspense(<PublicProfilePage />),
          },
          {
            path: '/plaza',
            element: withSuspense(<PlaygroundPage />),
          },
          {
            path: '/playground',
            element: <Navigate replace to="/plaza" />,
          },
          {
            path: '/feedback',
            element: withSuspense(<FeedbackPage />),
          },
          {
            path: '/admin/safety',
            element: withSuspense(<SafetyAdminPage />),
          },
          {
            path: '/dashboard',
            element: <Navigate replace to="/home" />,
          },
          {
            path: '/match',
            element: <Navigate replace to="/intents" />,
          },
          {
            path: '/chat',
            element: <Navigate replace to="/matches" />,
          },
          {
            path: '/chat/:conversationId',
            element: <Navigate replace to="/matches" />,
          },
          {
            path: '/friends',
            element: <Navigate replace to="/discover" />,
          },
        ],
      },
      {
        path: '/survey',
        element: withSuspense(<SurveyPage />),
      },
    ],
  },
  {
    element: <RedirectIfAuthenticated />,
    children: [
      {
        path: '/login',
        element: withSuspense(<LoginPage />),
      },
      {
        path: '/register',
        element: withSuspense(<RegisterPage />),
      },
    ],
  },
  {
    path: '/',
    element: withSuspense(<V2LandingPage />),
  },
  {
    path: '/safety',
    element: withSuspense(<SafetyPage />),
  },
  {
    path: '/schools/*',
    element: <Navigate replace to="/" />,
  },
  {
    path: '/alliance',
    element: <Navigate replace to="/" />,
  },
  {
    path: '*',
    element: <Navigate replace to="/" />,
  },
])
