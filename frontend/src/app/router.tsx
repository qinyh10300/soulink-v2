import { lazy, Suspense, type ReactNode } from 'react'
import { createBrowserRouter } from 'react-router-dom'

import {
  RedirectIfAuthenticated,
  RequireAuth,
} from '@/components/auth/auth-route-guard'
import { DashboardShell } from '@/components/dashboard/dashboard-shell'

const App = lazy(() => import('@/App'))
const AboutPage = lazy(async () => {
  const module = await import('@/pages/about-page')
  return { default: module.AboutPage }
})
const AlliancePage = lazy(async () => {
  const module = await import('@/pages/alliance-page')
  return { default: module.AlliancePage }
})
const ChatPage = lazy(async () => {
  const module = await import('@/pages/chat-page')
  return { default: module.ChatPage }
})
const DashboardPage = lazy(async () => {
  const module = await import('@/pages/dashboard-page')
  return { default: module.DashboardPage }
})
const FeedbackPage = lazy(async () => {
  const module = await import('@/pages/feedback-page')
  return { default: module.FeedbackPage }
})
const FriendsPage = lazy(async () => {
  const module = await import('@/pages/friends-page')
  return { default: module.FriendsPage }
})
const LoginPage = lazy(async () => {
  const module = await import('@/pages/login-page')
  return { default: module.LoginPage }
})
const MatchPage = lazy(async () => {
  const module = await import('@/pages/match-page')
  return { default: module.MatchPage }
})
const PlaygroundPage = lazy(async () => {
  const module = await import('@/pages/playground-page')
  return { default: module.PlaygroundPage }
})
const ProfilePage = lazy(async () => {
  const module = await import('@/pages/profile-page')
  return { default: module.ProfilePage }
})
const PublicProfilePage = lazy(async () => {
  const module = await import('@/pages/public-profile-page')
  return { default: module.PublicProfilePage }
})
const RegisterPage = lazy(async () => {
  const module = await import('@/pages/register-page')
  return { default: module.RegisterPage }
})
const SchoolPage = lazy(async () => {
  const module = await import('@/pages/school-page')
  return { default: module.SchoolPage }
})
const SingaporeSchoolPage = lazy(async () => {
  const module = await import('@/pages/singapore-school-page')
  return { default: module.SingaporeSchoolPage }
})
const SurveyPage = lazy(async () => {
  const module = await import('@/pages/survey-page')
  return { default: module.SurveyPage }
})

function withSuspense(element: ReactNode) {
  return <Suspense fallback={null}>{element}</Suspense>
}

export const router = createBrowserRouter([
  {
    element: <RequireAuth />,
    children: [
      {
        element: <DashboardShell />,
        children: [
          {
            path: '/dashboard',
            element: withSuspense(<DashboardPage />),
          },
          {
            path: '/match',
            element: withSuspense(<MatchPage />),
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
            path: '/friends',
            element: withSuspense(<FriendsPage />),
          },
          {
            path: '/playground',
            element: withSuspense(<PlaygroundPage />),
          },
          {
            path: '/profile',
            element: withSuspense(<ProfilePage />),
          },
          {
            path: '/profile/:userId',
            element: withSuspense(<PublicProfilePage />),
          },
          {
            path: '/feedback',
            element: withSuspense(<FeedbackPage />),
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
    element: withSuspense(<App />),
    children: [
      {
        index: true,
        element: withSuspense(<AlliancePage />),
      },
      {
        path: 'schools/nus',
        element: withSuspense(<SingaporeSchoolPage schoolKey="nus" />),
      },
      {
        path: 'schools/ntu',
        element: withSuspense(<SingaporeSchoolPage schoolKey="ntu" />),
      },
      {
        path: 'schools/fdu',
        element: withSuspense(<SingaporeSchoolPage schoolKey="fdu" />),
      },
      {
        path: 'schools/sjdu',
        element: withSuspense(<SingaporeSchoolPage schoolKey="sjdu" />),
      },
      {
        path: 'schools/:schoolKey',
        element: withSuspense(<SchoolPage />),
      },
      {
        path: 'about',
        element: withSuspense(<AboutPage />),
      },
    ],
  },
])
