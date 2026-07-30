import {
  CircleAlert,
  Heart,
  Home,
  LogOut,
  MessageCircle,
  MessageSquareText,
  UserRound,
  Users,
} from 'lucide-react'
import {
  useEffect,
  useMemo,
  useState,
  type ComponentType,
  type CSSProperties,
} from 'react'
import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom'

import { BrandMark } from '@/components/shared/brand-mark'
import { WechatGroupModal } from '@/components/shared/wechat-group-modal'
import { useChat } from '@/contexts/chat-context'
import { logout } from '@/lib/auth-api'
import { getAvatarFallbackLabel } from '@/lib/avatar'
import { cn } from '@/lib/utils'
import { clearSurveyDraft } from '@/lib/survey-storage'
import { clearAuthSession, clearUserProfile, resolveDashboardUser } from '@/lib/user-profile-storage'
import { dashboardContent } from '@/mock/dashboard'
import type { DashboardUser } from '@/types/dashboard'

type DashboardNavItem = {
  badge?: number
  icon: ComponentType<{ className?: string; size?: number }>
  isActive: (pathname: string) => boolean
  label: string
  to: string
}

const text = {
  cancel: '\u53d6\u6d88',
  community: '\u52a0\u5165\u793e\u7fa4',
  confirm: '\u786e\u5b9a',
  confirmLogoutDescription: '\u786e\u5b9a\u540e\u5c06\u9000\u51fa\u5f53\u524d\u8d26\u53f7\uff0c\u9700\u8981\u91cd\u65b0\u767b\u5f55\u540e\u624d\u80fd\u7ee7\u7eed\u4f7f\u7528\u3002',
  confirmLogoutTitle: '\u786e\u8ba4\u9000\u51fa\u767b\u5f55\uff1f',
  feedback: '\u53cd\u9988\u610f\u89c1',
  friends: '\u597d\u53cb',
  home: '\u9996\u9875',
  logout: '\u9000\u51fa\u767b\u5f55',
  loggingOut: '\u9000\u51fa\u4e2d...',
  match: '\u5339\u914d\u7ed3\u679c',
  messages: '\u6d88\u606f',
  playground: '\u5e7f\u573a',
  profile: '\u4e2a\u4eba\u8d44\u6599',
  security: '\u8d26\u53f7\u5b89\u5168',
}

const dashboardShellStyle = {
  '--dashboard-sidebar-width': '15rem',
} as CSSProperties

function formatBadge(value: number) {
  return value > 99 ? '99+' : String(value)
}

function DashboardSidebar({ user }: { user: DashboardUser }) {
  const location = useLocation()
  const navigate = useNavigate()
  const { totalUnread, pendingRequests } = useChat()
  const [isWechatModalOpen, setIsWechatModalOpen] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const [isLogoutConfirmOpen, setIsLogoutConfirmOpen] = useState(false)

  const navItems: DashboardNavItem[] = [
    {
      icon: Home,
      isActive: (pathname) => pathname === '/dashboard',
      label: text.home,
      to: '/dashboard',
    },
    {
      icon: Heart,
      isActive: (pathname) => pathname === '/match',
      label: text.match,
      to: '/match',
    },
    {
      badge: totalUnread,
      icon: MessageCircle,
      isActive: (pathname) => pathname.startsWith('/chat'),
      label: text.messages,
      to: '/chat',
    },
    {
      badge: pendingRequests,
      icon: Users,
      isActive: (pathname) => pathname === '/friends',
      label: text.friends,
      to: '/friends',
    },
    // {
    //   icon: Compass,
    //   isActive: (pathname) => pathname.startsWith('/playground'),
    //   label: text.playground,
    //   to: '/playground',
    // },
  ]

  useEffect(() => {
    if (!isLogoutConfirmOpen) {
      return
    }

    const originalOverflow = document.body.style.overflow

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !isLoggingOut) {
        setIsLogoutConfirmOpen(false)
      }
    }

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = originalOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [isLoggingOut, isLogoutConfirmOpen])

  const handleLogout = async () => {
    if (isLoggingOut) {
      return
    }

    setIsLoggingOut(true)

    try {
      await logout()
    } catch {
      // Clear local session even if the revoke request fails.
    } finally {
      clearAuthSession()
      clearUserProfile()
      clearSurveyDraft()
      setIsLogoutConfirmOpen(false)
      setIsLoggingOut(false)
      navigate('/login')
    }
  }

  return (
    <>
      <aside className="z-40 border-b border-[#e7ddd4] bg-[rgba(251,246,241,0.94)] px-4 py-4 backdrop-blur-xl md:fixed md:inset-y-0 md:left-0 md:flex md:w-[var(--dashboard-sidebar-width)] md:flex-col md:border-b-0 md:border-r md:px-5 md:py-6">
        <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 md:mx-0 md:h-full md:max-w-none">
          <BrandMark
            iconClassName="border-white/70 shadow-none md:shadow-[0_14px_28px_rgba(39,76,155,0.18)]"
            subtitleClassName="text-muted-foreground"
            titleClassName="text-primary/75"
          />

          <nav className="flex flex-col gap-2" aria-label="Dashboard">
            {navItems.map((item) => {
              const Icon = item.icon
              const active = item.isActive(location.pathname)

              return (
                <Link
                  className={cn(
                    'group relative flex items-center gap-3 rounded-3xl px-4 py-3 text-sm font-medium transition-colors',
                    active
                      ? 'bg-[#f3ece4] text-[#2f221b] shadow-[0_12px_30px_rgba(66,45,31,0.06)]'
                      : 'text-[#75665d] hover:bg-[#f7f1ea] hover:text-[#2f221b]',
                  )}
                  key={item.to}
                  to={item.to}
                >
                  <Icon
                    className={cn(
                      'shrink-0 transition-colors',
                      active ? 'text-[#8f5c42]' : 'text-[#9e8a7c] group-hover:text-[#8f5c42]',
                    )}
                    size={18}
                  />
                  <span>{item.label}</span>
                  {item.badge && item.badge > 0 ? (
                    <span className="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#e74c3c] px-1.5 text-[10px] font-medium text-white">
                      {formatBadge(item.badge)}
                    </span>
                  ) : null}
                </Link>
              )
            })}

            <button
              className="flex items-center gap-3 rounded-3xl px-4 py-3 text-left text-sm font-medium text-[#75665d] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]"
              onClick={() => setIsWechatModalOpen(true)}
              type="button"
            >
              <Users className="shrink-0 text-[#9e8a7c]" size={18} />
              <span>{text.community}</span>
            </button>

            <Link
              className={cn(
                'group relative flex items-center gap-3 rounded-3xl px-4 py-3 text-sm font-medium transition-colors',
                location.pathname === '/profile'
                  ? 'bg-[#f3ece4] text-[#2f221b] shadow-[0_12px_30px_rgba(66,45,31,0.06)]'
                  : 'text-[#75665d] hover:bg-[#f7f1ea] hover:text-[#2f221b]',
              )}
              to="/profile"
            >
              <UserRound
                className={cn(
                  'shrink-0 transition-colors',
                  location.pathname === '/profile'
                    ? 'text-[#8f5c42]'
                    : 'text-[#9e8a7c] group-hover:text-[#8f5c42]',
                )}
                size={18}
              />
              <span>{text.profile}</span>
            </Link>

            <Link
              className={cn(
                'group relative flex items-center gap-3 rounded-3xl px-4 py-3 text-sm font-medium transition-colors',
                location.pathname === '/feedback'
                  ? 'bg-[#f3ece4] text-[#2f221b] shadow-[0_12px_30px_rgba(66,45,31,0.06)]'
                  : 'text-[#75665d] hover:bg-[#f7f1ea] hover:text-[#2f221b]',
              )}
              to="/feedback"
            >
              <MessageSquareText
                className={cn(
                  'shrink-0 transition-colors',
                  location.pathname === '/feedback'
                    ? 'text-[#8f5c42]'
                    : 'text-[#9e8a7c] group-hover:text-[#8f5c42]',
                )}
                size={18}
              />
              <span>{text.feedback}</span>
            </Link>

            <button
              className="group relative flex items-center gap-3 rounded-3xl px-4 py-3 text-left text-sm font-medium text-[#75665d] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b] disabled:cursor-not-allowed disabled:opacity-70"
              disabled={isLoggingOut}
              onClick={() => setIsLogoutConfirmOpen(true)}
              type="button"
            >
              <LogOut className="shrink-0 text-[#9e8a7c] transition-colors group-hover:text-[#8f5c42]" size={18} />
              <span>{isLoggingOut ? text.loggingOut : text.logout}</span>
            </button>
          </nav>

          <div className="mt-auto space-y-4">
            <Link
              className="flex w-full items-center gap-3 rounded-[28px] border border-[#ebe1d8] bg-white/78 px-3 py-3 text-left shadow-[0_10px_24px_rgba(66,45,31,0.06)] transition-colors hover:border-[#decfc1] hover:bg-white"
              to="/profile"
            >
              {user.avatarUrl ? (
                <img
                  alt={user.name}
                  className="h-10 w-10 rounded-full object-cover"
                  src={user.avatarUrl}
                />
              ) : (
                <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#f4e8dd] text-lg">
                  {getAvatarFallbackLabel(user.name, user.avatarEmoji)}
                </span>
              )}
              <span className="min-w-0 flex-1 truncate font-medium text-[#2f221b]">{user.name}</span>
            </Link>
          </div>
        </div>
      </aside>

      <WechatGroupModal
        onClose={() => setIsWechatModalOpen(false)}
        open={isWechatModalOpen}
      />

      {isLogoutConfirmOpen ? (
        <div className="fixed inset-0 z-[220] flex items-center justify-center p-4">
          <button
            aria-label={text.cancel}
            className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
            disabled={isLoggingOut}
            onClick={() => setIsLogoutConfirmOpen(false)}
            type="button"
          />

          <div className="relative z-10 w-full max-w-md rounded-[30px] border border-[#eadfd5] bg-[#fcf8f3] p-7 shadow-[0_26px_72px_rgba(61,40,29,0.18)]">
            <div className="flex items-center gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[linear-gradient(135deg,#f3e8dc,#ead9c8)] text-[#9e5e52] shadow-[0_14px_28px_rgba(111,84,61,0.14)]">
                <CircleAlert size={24} />
              </div>
              <div className="space-y-1">
                <p className="text-sm uppercase tracking-[0.22em] text-[#a4846d]">{text.security}</p>
                <h3 className="font-serif text-3xl text-[#2f221b]">{text.confirmLogoutTitle}</h3>
              </div>
            </div>

            <p className="mt-6 text-base leading-8 text-[#6f6056]">
              {text.confirmLogoutDescription}
            </p>

            <div className="mt-8 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                className="inline-flex h-11 items-center justify-center rounded-full border border-[#e7d8cb] bg-white/92 px-6 text-sm font-semibold text-[#7d6556] transition-colors hover:bg-white disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isLoggingOut}
                onClick={() => setIsLogoutConfirmOpen(false)}
                type="button"
              >
                {text.cancel}
              </button>
              <button
                className="inline-flex h-11 items-center justify-center rounded-full bg-[#9d433f] px-6 text-sm font-semibold text-white transition-colors hover:bg-[#8e3d39] disabled:cursor-not-allowed disabled:opacity-60"
                disabled={isLoggingOut}
                onClick={() => {
                  void handleLogout()
                }}
                type="button"
              >
                {isLoggingOut ? text.loggingOut : text.confirm}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  )
}

export function DashboardShell() {
  const user = useMemo(
    () => resolveDashboardUser(dashboardContent.user),
    [],
  )

  return (
    <div className="min-h-screen bg-[#f6efe8]" style={dashboardShellStyle}>
      <DashboardSidebar user={user} />
      <div className="min-w-0 md:pl-[var(--dashboard-sidebar-width)]">
        <Outlet />
      </div>
    </div>
  )
}
