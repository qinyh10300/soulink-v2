import {
  ClipboardList,
  Compass,
  Clock3,
  type LucideIcon,
  LogOut,
  Menu,
  MessagesSquare,
  Plus,
  ShieldCheck,
  Sparkles,
  UserRound,
  X,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import {
  Link,
  NavLink,
  Outlet,
  useLocation,
  useNavigate,
} from 'react-router-dom'

import logoUrl from '@/assets/soulink-logo-icon.png'
import { logout } from '@/lib/auth-api'
import {
  fetchMyIntents,
  type ConnectionIntent,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'
import {
  clearAuthSession,
  clearUserProfile,
  getStoredUserProfile,
} from '@/lib/user-profile-storage'

interface NavItem {
  label: string
  mobileLabel: string
  to: string
  icon: LucideIcon
  match: (pathname: string) => boolean
}

const navItems: NavItem[] = [
  {
    label: 'SouLink 助手',
    mobileLabel: '助手',
    to: '/home',
    icon: Sparkles,
    match: (pathname) => pathname === '/home',
  },
  {
    label: '校园发现',
    mobileLabel: '发现',
    to: '/discover',
    icon: Compass,
    match: (pathname) => pathname === '/discover',
  },
  {
    label: '我的意图',
    mobileLabel: '意图',
    to: '/intents',
    icon: ClipboardList,
    match: (pathname) =>
      pathname === '/intents' || pathname.startsWith('/intents/'),
  },
  {
    label: '连接房间',
    mobileLabel: '连接',
    to: '/matches',
    icon: MessagesSquare,
    match: (pathname) => pathname.startsWith('/matches'),
  },
  {
    label: 'AI 档案',
    mobileLabel: '档案',
    to: '/profile',
    icon: UserRound,
    match: (pathname) => pathname === '/profile',
  },
]

function Brand() {
  return (
    <Link className="flex min-w-0 items-center gap-2.5" to="/home">
      <img
        alt=""
        className="h-8 w-8 rounded-md object-cover"
        src={logoUrl}
      />
      <span className="min-w-0">
        <span className="block truncate text-[17px] font-semibold text-[#211f1d]">
          SouLink
        </span>
        <span className="block truncate text-[10px] font-semibold text-[#6c776f]">
          TSINGHUA · BETA
        </span>
      </span>
    </Link>
  )
}

export function ConnectionShell() {
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)
  const [recentIntents, setRecentIntents] = useState<ConnectionIntent[]>([])
  const profile = useMemo(() => getStoredUserProfile(), [])

  useEffect(() => {
    let cancelled = false
    void fetchMyIntents()
      .then((intents) => {
        if (!cancelled) {
          setRecentIntents(
            [...intents]
              .sort(
                (left, right) =>
                  new Date(right.updatedAt).getTime() -
                  new Date(left.updatedAt).getTime(),
              )
              .slice(0, 4),
          )
        }
      })
      .catch(() => {
        if (!cancelled) {
          setRecentIntents([])
        }
      })

    return () => {
      cancelled = true
    }
  }, [location.pathname])

  const handleLogout = async () => {
    if (loggingOut) {
      return
    }
    setLoggingOut(true)
    try {
      await logout()
    } catch {
      // Local credentials are cleared even when the network request fails.
    } finally {
      clearAuthSession()
      clearUserProfile()
      navigate('/login')
    }
  }

  return (
    <div className="min-h-screen bg-[#f7f6f3] text-[#23211f]">
      <aside className="fixed inset-y-0 left-0 z-40 hidden w-[244px] border-r border-[#dfdcd6] bg-[#fbfaf8] px-4 py-5 lg:flex lg:flex-col">
        <Brand />

        <Link
          className="mt-7 inline-flex h-10 items-center justify-center gap-2 rounded-md bg-[#263b35] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#1d302a]"
          to="/intents/new"
        >
          <Plus size={17} />
          开始新对话
        </Link>

        <nav aria-label="产品导航" className="mt-5 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                className={({ isActive }) =>
                  cn(
                    'flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-[#e8f1ed] text-[#315c4d]'
                      : 'text-[#625e58] hover:bg-[#f0eeea] hover:text-[#23211f]',
                  )
                }
                key={item.to}
                to={item.to}
              >
                <Icon size={17} strokeWidth={1.8} />
                {item.label}
              </NavLink>
            )
          })}
        </nav>

        <section className="mt-6 min-h-0 border-t border-[#e4e1dc] pt-5">
          <div className="flex items-center justify-between px-2">
            <h2 className="text-[11px] font-semibold text-[#817d76]">最近意图</h2>
            <Link
              className="text-[11px] font-medium text-[#477062] hover:text-[#315c4d]"
              to="/intents"
            >
              全部
            </Link>
          </div>
          <div className="mt-2 space-y-0.5">
            {recentIntents.length > 0 ? (
              recentIntents.map((intent) => (
                <Link
                  className="group flex min-w-0 items-center gap-2.5 rounded-md px-2 py-2 text-xs text-[#6f6b65] hover:bg-[#f0eeea] hover:text-[#34312e]"
                  key={intent.id}
                  to={`/intents/${intent.id}`}
                >
                  <span
                    className={cn(
                      'h-1.5 w-1.5 shrink-0 rounded-full',
                      intent.status === 'active'
                        ? 'bg-[#4f846f]'
                        : 'bg-[#bbb6ae]',
                    )}
                  />
                  <span className="min-w-0 flex-1 truncate">{intent.title}</span>
                  <Clock3
                    className="shrink-0 text-[#a19c94] opacity-0 transition-opacity group-hover:opacity-100"
                    size={12}
                  />
                </Link>
              ))
            ) : (
              <p className="px-2 py-3 text-xs leading-5 text-[#9a958d]">
                暂无最近意图
              </p>
            )}
          </div>
        </section>

        <div className="mt-auto border-t border-[#e4e1dc] pt-4">
          <div className="flex items-center gap-2.5 px-2">
            <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-[#e9f1ed] text-sm font-semibold text-[#27614f]">
              {(profile.nickname || '同学').slice(0, 1)}
            </span>
            <span className="min-w-0 flex-1">
              <span className="block truncate text-sm font-medium text-[#302e2b]">
                {profile.nickname || '清华同学'}
              </span>
              <span className="flex items-center gap-1 text-[11px] text-[#77736c]">
                <ShieldCheck size={11} />
                身份已验证
              </span>
            </span>
            <button
              aria-label="退出登录"
              className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#817d76] transition-colors hover:bg-[#eeeae5] hover:text-[#8b3f3f]"
              disabled={loggingOut}
              onClick={() => void handleLogout()}
              title="退出登录"
              type="button"
            >
              <LogOut size={16} />
            </button>
          </div>
        </div>
      </aside>

      <header className="sticky top-0 z-40 flex h-15 items-center justify-between border-b border-[#dfdcd6] bg-[rgba(251,250,248,0.94)] px-4 backdrop-blur-lg lg:hidden">
        <Brand />
        <div className="flex items-center gap-1">
          <Link
            aria-label="发布新意图"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#263b35] text-white"
            title="开始新对话"
            to="/intents/new"
          >
            <Plus size={18} />
          </Link>
          <button
            aria-label={mobileMenuOpen ? '关闭菜单' : '打开菜单'}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-[#5f5b55] hover:bg-[#efede9]"
            onClick={() => setMobileMenuOpen((open) => !open)}
            type="button"
          >
            {mobileMenuOpen ? <X size={19} /> : <Menu size={19} />}
          </button>
        </div>
      </header>

      {mobileMenuOpen ? (
        <div className="fixed inset-x-0 top-15 z-30 border-b border-[#dfdcd6] bg-[#fbfaf8] p-3 shadow-[0_12px_32px_rgba(35,33,31,0.08)] lg:hidden">
          <Link
            className="flex h-10 items-center gap-3 rounded-md px-3 text-sm text-[#5f5b55] hover:bg-[#efede9]"
            onClick={() => setMobileMenuOpen(false)}
            to="/account"
          >
            <UserRound size={17} />
            账户设置
          </Link>
          <button
            className="flex h-10 w-full items-center gap-3 rounded-md px-3 text-sm text-[#8b3f3f] hover:bg-[#f5eaea]"
            disabled={loggingOut}
            onClick={() => void handleLogout()}
            type="button"
          >
            <LogOut size={17} />
            {loggingOut ? '正在退出...' : '退出登录'}
          </button>
        </div>
      ) : null}

      <main className="min-h-screen pb-20 lg:ml-[244px] lg:pb-0">
        <Outlet />
      </main>

      <nav
        aria-label="移动端导航"
        className="fixed inset-x-0 bottom-0 z-40 grid h-16 grid-cols-5 border-t border-[#dedbd5] bg-[rgba(251,250,248,0.96)] px-1 pb-[env(safe-area-inset-bottom)] backdrop-blur-lg lg:hidden"
      >
        {navItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              className={({ isActive }) =>
                cn(
                  'flex min-w-0 flex-col items-center justify-center gap-1 text-[10px] font-medium',
                  isActive ? 'text-[#315c4d]' : 'text-[#77736c]',
                )
              }
              key={item.to}
              to={item.to}
            >
              <Icon size={18} strokeWidth={1.8} />
              <span className="truncate">{item.mobileLabel}</span>
            </NavLink>
          )
        })}
      </nav>
    </div>
  )
}
