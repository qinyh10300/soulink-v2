import { ChevronDown, Menu, MessageCircle, Users } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import { UserMenuDropdown } from '@/components/dashboard/user-menu-dropdown'
import { BrandMark } from '@/components/shared/brand-mark'
import { WechatGroupModal } from '@/components/shared/wechat-group-modal'
import { useChat } from '@/contexts/chat-context'
import { logout } from '@/lib/auth-api'
import { getAvatarFallbackLabel } from '@/lib/avatar'
import { clearSurveyDraft } from '@/lib/survey-storage'
import { clearAuthSession, clearUserProfile } from '@/lib/user-profile-storage'
import type { DashboardUser } from '@/types/dashboard'

interface DashboardHeaderProps {
  user: DashboardUser
}

export function DashboardHeader({ user }: DashboardHeaderProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { totalUnread, pendingRequests } = useChat()
  const [menuOpen, setMenuOpen] = useState(false)
  const [mobileNavOpen, setMobileNavOpen] = useState(false)
  const [isWechatModalOpen, setIsWechatModalOpen] = useState(false)
  const [isLoggingOut, setIsLoggingOut] = useState(false)
  const menuRef = useRef<HTMLDivElement | null>(null)
  const mobileNavRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!menuOpen && !mobileNavOpen) {
      return
    }

    const handlePointerDown = (event: MouseEvent) => {
      const target = event.target as Node

      if (menuRef.current?.contains(target) || mobileNavRef.current?.contains(target)) {
        return
      }

      setMenuOpen(false)
      setMobileNavOpen(false)
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setMenuOpen(false)
        setMobileNavOpen(false)
      }
    }

    window.addEventListener('mousedown', handlePointerDown)
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('mousedown', handlePointerDown)
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [menuOpen, mobileNavOpen])

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
      setMenuOpen(false)
      setMobileNavOpen(false)
      setIsLoggingOut(false)
      navigate('/login')
    }
  }

  const handleMatchResultsClick = () => {
    setMobileNavOpen(false)
    navigate('/match')
  }

  return (
    <>
      <header className="sticky top-0 z-40 border-b border-[#e7ddd4] bg-[rgba(251,246,241,0.92)] backdrop-blur-xl">
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-8 px-6 py-2 lg:px-8">
          <BrandMark
            iconClassName="border-white/70 shadow-none md:shadow-[0_14px_28px_rgba(39,76,155,0.18)]"
            subtitleClassName="text-muted-foreground"
            titleClassName="text-primary/75"
          />

          <div className="flex items-center gap-2 sm:gap-3">
            <div className="relative md:hidden" ref={mobileNavRef}>
              <button
                aria-label="打开导航菜单"
                className="flex h-12 w-12 items-center justify-center rounded-full border border-[#ebe1d8] bg-white/82 shadow-[0_10px_24px_rgba(66,45,31,0.06)] transition-colors hover:border-[#decfc1] hover:bg-white"
                onClick={() => {
                  setMobileNavOpen((current) => !current)
                  setMenuOpen(false)
                }}
                type="button"
              >
                <Menu className="text-[#3f2e24]" size={22} />
              </button>

              {mobileNavOpen ? (
                <div className="absolute right-0 top-[calc(100%+0.85rem)] z-50 w-[10rem] rounded-[28px] border border-[#eadfd5] bg-[#fcf8f3] p-3 shadow-[0_22px_50px_rgba(59,39,28,0.14)]">
                  <div className="space-y-1">
                    <Link
                      className={
                        location.pathname === '/dashboard'
                          ? 'block rounded-2xl bg-[#f3ece4] px-4 py-3 text-sm font-medium text-[#2f221b]'
                          : 'block rounded-2xl px-4 py-3 text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]'
                      }
                      onClick={() => setMobileNavOpen(false)}
                      to="/dashboard"
                    >
                      首页
                    </Link>
                    <button
                      className={
                        location.pathname === '/match'
                          ? 'block w-full rounded-2xl bg-[#f3ece4] px-4 py-3 text-left text-sm font-medium text-[#2f221b]'
                          : 'block w-full rounded-2xl px-4 py-3 text-left text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]'
                      }
                      onClick={handleMatchResultsClick}
                      type="button"
                    >
                      匹配结果
                    </button>
                    <Link
                      className="flex items-center gap-2 rounded-2xl px-4 py-3 text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]"
                      onClick={() => setMobileNavOpen(false)}
                      to="/chat"
                    >
                      <MessageCircle size={16} />
                      消息
                      {totalUnread > 0 ? (
                        <span className="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#e74c3c] px-1.5 text-[10px] font-medium text-white">
                          {totalUnread > 99 ? '99+' : totalUnread}
                        </span>
                      ) : null}
                    </Link>
                    <Link
                      className="flex items-center gap-2 rounded-2xl px-4 py-3 text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]"
                      onClick={() => setMobileNavOpen(false)}
                      to="/friends"
                    >
                      <Users size={16} />
                      好友
                      {pendingRequests > 0 ? (
                        <span className="ml-auto inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#e74c3c] px-1.5 text-[10px] font-medium text-white">
                          {pendingRequests > 99 ? '99+' : pendingRequests}
                        </span>
                      ) : null}
                    </Link>
                    <button
                      className="block w-full rounded-2xl px-4 py-3 text-left text-sm text-[#5f5148] transition-colors hover:bg-[#f7f1ea] hover:text-[#2f221b]"
                      onClick={() => {
                        setMobileNavOpen(false)
                        setIsWechatModalOpen(true)
                      }}
                      type="button"
                    >
                      加入社群
                    </button>
                  </div>
                </div>
              ) : null}
            </div>

            <div className="relative" ref={menuRef}>
              <div className="md:hidden">
                <button
                  className="flex items-center gap-2 rounded-full border border-[#ebe1d8] bg-white/82 px-2.5 py-2 shadow-[0_10px_24px_rgba(66,45,31,0.06)] transition-colors hover:border-[#decfc1] hover:bg-white"
                  onClick={() => {
                    setMenuOpen((current) => !current)
                    setMobileNavOpen(false)
                  }}
                  type="button"
                >
                  {user.avatarUrl ? (
                    <img
                      alt={user.name}
                      className="h-8 w-8 rounded-full object-cover"
                      src={user.avatarUrl}
                    />
                  ) : (
                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[#f4e8dd] text-lg">
                      {getAvatarFallbackLabel(user.name, user.avatarEmoji)}
                    </span>
                  )}
                  <ChevronDown className="text-[#9e8a7c]" size={16} />
                </button>
              </div>

              <div className="hidden items-center gap-8 text-[15px] text-[#75665d] md:flex">
                <Link
                  className={
                    location.pathname === '/dashboard'
                      ? 'font-semibold text-[#2f221b]'
                      : 'transition-colors hover:text-[#2f221b]'
                  }
                  to="/dashboard"
                >
                  首页
                </Link>
                <button
                  className={
                    location.pathname === '/match'
                      ? 'font-semibold text-[#2f221b]'
                      : 'transition-colors hover:text-[#2f221b]'
                  }
                  onClick={handleMatchResultsClick}
                  type="button"
                >
                  匹配结果
                </button>

                <Link
                  className={`relative flex items-center gap-1 transition-colors hover:text-[#2f221b] ${
                    location.pathname.startsWith('/chat')
                      ? 'font-semibold text-[#2f221b]'
                      : ''
                  }`}
                  to="/chat"
                >
                  <MessageCircle size={16} />
                  消息
                  {totalUnread > 0 ? (
                    <span className="absolute -right-3 -top-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-[#e74c3c] px-1 text-[9px] font-medium text-white">
                      {totalUnread > 99 ? '99+' : totalUnread}
                    </span>
                  ) : null}
                </Link>
                <Link
                  className={`relative flex items-center gap-1 transition-colors hover:text-[#2f221b] ${
                    location.pathname === '/friends'
                      ? 'font-semibold text-[#2f221b]'
                      : ''
                  }`}
                  to="/friends"
                >
                  <Users size={16} />
                  好友
                  {pendingRequests > 0 ? (
                    <span className="absolute -right-3 -top-1.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-[#e74c3c] px-1 text-[9px] font-medium text-white">
                      {pendingRequests > 99 ? '99+' : pendingRequests}
                    </span>
                  ) : null}
                </Link>

                <button
                  className="transition-colors hover:text-[#2f221b]"
                  onClick={() => setIsWechatModalOpen(true)}
                  type="button"
                >
                  加入社群
                </button>

                <button
                  className="flex items-center gap-3 rounded-full border border-[#ebe1d8] bg-white/78 px-3 py-2 shadow-[0_10px_24px_rgba(66,45,31,0.06)] transition-colors hover:border-[#decfc1] hover:bg-white"
                  onClick={() => setMenuOpen((current) => !current)}
                  type="button"
                >
                  {user.avatarUrl ? (
                    <img
                      alt={user.name}
                      className="h-8 w-8 rounded-full object-cover"
                      src={user.avatarUrl}
                    />
                  ) : (
                    <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[#f4e8dd] text-lg">
                      {getAvatarFallbackLabel(user.name, user.avatarEmoji)}
                    </span>
                  )}
                  <span className="font-medium text-[#2f221b]">{user.name}</span>
                  <ChevronDown className="text-[#9e8a7c]" size={16} />
                </button>
              </div>

              {menuOpen ? (
                <UserMenuDropdown
                  currentPath={location.pathname}
                  isLoggingOut={isLoggingOut}
                  onClose={() => setMenuOpen(false)}
                  onLogout={() => {
                    void handleLogout()
                  }}
                />
              ) : null}
            </div>
          </div>
        </div>
      </header>

      <WechatGroupModal
        onClose={() => setIsWechatModalOpen(false)}
        open={isWechatModalOpen}
      />
    </>
  )
}
