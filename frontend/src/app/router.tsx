import { lazy, Suspense, type ReactNode } from 'react'
import {
  createBrowserRouter,
  Navigate,
  useLocation,
  useParams,
} from 'react-router-dom'

import {
  RedirectIfAuthenticated,
  RequireAuth,
} from '@/components/auth/auth-route-guard'
import { ConnectionShell } from '@/components/connection/connection-shell'
import {
  connectionIntentPath,
  connectionPaths,
  connectionRoomPath,
} from '@/lib/connection-routes'

const ConnectionHomePage = lazy(async () => {
  const module = await import('@/pages/connection-home-page')
  return { default: module.ConnectionHomePage }
})
const ConnectionProfilePage = lazy(async () => {
  const module = await import('@/pages/connection-profile-page')
  return { default: module.ConnectionProfilePage }
})
const ChatPage = lazy(async () => {
  const module = await import('@/pages/chat-page')
  return { default: module.ChatPage }
})
const CalendarPage = lazy(async () => {
  const module = await import('@/pages/calendar-page')
  return { default: module.CalendarPage }
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

function LegacyParameterizedRedirect({ type }: { type: 'intent' | 'room' }) {
  const params = useParams()
  const location = useLocation()
  const rawId = type === 'intent' ? params.intentId : params.roomId
  const target =
    type === 'intent'
      ? connectionIntentPath(rawId ?? '')
      : connectionRoomPath(rawId ?? '')
  return <Navigate replace to={`${target}${location.search}${location.hash}`} />
}

export const router = createBrowserRouter([
  {
    element: <RequireAuth />,
    children: [
      {
        element: <ConnectionShell />,
        children: [
          {
            path: connectionPaths.root,
            element: withSuspense(<ConnectionHomePage />),
          },
          {
            path: connectionPaths.discover,
            element: withSuspense(<IntentFeedPage />),
          },
          {
            path: connectionPaths.intents,
            element: withSuspense(<MyIntentsPage />),
          },
          {
            path: connectionPaths.newIntent,
            element: withSuspense(<IntentComposerPage />),
          },
          {
            path: '/connect/intents/:intentId',
            element: withSuspense(<IntentDetailPage />),
          },
          {
            path: connectionPaths.rooms,
            element: withSuspense(<MatchRoomsPage />),
          },
          {
            path: '/connect/rooms/:roomId',
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
            path: '/calendar',
            element: withSuspense(<CalendarPage />),
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
            element: <Navigate replace to={connectionPaths.root} />,
          },
          {
            path: '/match',
            element: <Navigate replace to={connectionPaths.intents} />,
          },
          {
            path: '/chat',
            element: withSuspense(<ChatPage />),
          },
          {
            path: '/chat/:conversationId',
            element: withSuspense(<ChatPage />),
          },
          {
            path: '/chat/direct/:conversationId',
            element: withSuspense(<ChatPage />),
          },
          {
            path: '/chat/group/:groupId',
            element: withSuspense(<ChatPage />),
          },
          {
            path: '/friends',
            element: <Navigate replace to={connectionPaths.discover} />,
          },
          {
            path: '/home',
            element: <Navigate replace to={connectionPaths.root} />,
          },
          {
            path: '/discover',
            element: <Navigate replace to={connectionPaths.discover} />,
          },
          {
            path: '/intents',
            element: <Navigate replace to={connectionPaths.intents} />,
          },
          {
            path: '/intents/new',
            element: <Navigate replace to={connectionPaths.newIntent} />,
          },
          {
            path: '/intents/:intentId',
            element: <LegacyParameterizedRedirect type="intent" />,
          },
          {
            path: '/matches',
            element: <Navigate replace to={connectionPaths.rooms} />,
          },
          {
            path: '/matches/:roomId',
            element: <LegacyParameterizedRedirect type="room" />,
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
