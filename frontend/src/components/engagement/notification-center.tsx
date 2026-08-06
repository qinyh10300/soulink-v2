import {
  Bell,
  BellOff,
  CalendarClock,
  CheckCheck,
  ChevronRight,
  Download,
  LoaderCircle,
  MessageCircleMore,
  Settings,
  Sparkles,
  X,
  type LucideIcon,
} from 'lucide-react'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { useEngagement } from '@/contexts/engagement-context'
import type { NotificationKind } from '@/lib/engagement-api'
import { cn } from '@/lib/utils'

const notificationIcons: Record<NotificationKind, LucideIcon> = {
  new_message: MessageCircleMore,
  meetup_invite: CalendarClock,
  meetup_reminder: CalendarClock,
  daily_picks: Sparkles,
}

function formatNotificationTime(value: string) {
  const date = new Date(value)
  const now = new Date()
  const sameDay = date.toDateString() === now.toDateString()
  return new Intl.DateTimeFormat('zh-CN', {
    ...(sameDay ? {} : { month: 'numeric', day: 'numeric' }),
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

export function NotificationCenter({ mobile = false }: { mobile?: boolean }) {
  const navigate = useNavigate()
  const {
    notifications,
    unreadNotifications,
    nextNotificationCursor,
    installAvailable,
    installApp,
    loadMoreNotifications,
    markAllNotificationsRead,
    markNotificationRead,
  } = useEngagement()
  const [open, setOpen] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)

  const openNotification = async (notificationId: number, path: string, unread: boolean) => {
    if (unread) {
      await markNotificationRead(notificationId).catch(() => {})
    }
    setOpen(false)
    navigate(path)
  }

  const loadMore = async () => {
    if (loadingMore) return
    setLoadingMore(true)
    try {
      await loadMoreNotifications()
    } finally {
      setLoadingMore(false)
    }
  }

  return (
    <>
      <button
        aria-expanded={open}
        aria-label="通知中心"
        className={cn(
          'relative inline-flex items-center justify-center rounded-md text-[#6f6b65] transition-colors hover:bg-[#efede9] hover:text-[#315c4d] focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-[#527b6d]',
          mobile ? 'h-10 w-10' : 'h-8 w-8',
        )}
        onClick={() => setOpen(true)}
        title="通知中心"
        type="button"
      >
        <Bell size={mobile ? 18 : 16} />
        {unreadNotifications > 0 ? (
          <span className="absolute -right-0.5 -top-0.5 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-[#d95656] px-1 text-[9px] font-semibold text-white">
            {unreadNotifications > 9 ? '9+' : unreadNotifications}
          </span>
        ) : null}
      </button>

      {open ? (
        <>
          <button
            aria-label="关闭通知中心"
            className="fixed inset-0 z-[70] cursor-default bg-black/15"
            onClick={() => setOpen(false)}
            type="button"
          />
          <section
            aria-label="通知中心"
            className="fixed inset-y-0 right-0 z-[80] flex w-full max-w-[390px] flex-col border-l border-[#d9d5cf] bg-[#fbfaf8] shadow-[-20px_0_60px_rgba(35,33,31,0.16)] sm:inset-y-auto sm:right-5 sm:top-16 sm:max-h-[min(680px,calc(100dvh-5rem))] sm:rounded-lg sm:border"
          >
            <header className="flex h-15 shrink-0 items-center justify-between border-b border-[#e2dfda] px-4">
              <div className="flex items-center gap-2">
                <h2 className="text-sm font-semibold text-[#302e2b]">通知</h2>
                {unreadNotifications > 0 ? (
                  <span className="text-[11px] font-medium text-[#527164]">
                    {unreadNotifications} 条未读
                  </span>
                ) : null}
              </div>
              <div className="flex items-center gap-0.5">
                {unreadNotifications > 0 ? (
                  <button
                    aria-label="全部标为已读"
                    className="inline-flex h-9 w-9 items-center justify-center rounded-md text-[#77736c] hover:bg-[#efede9] hover:text-[#315c4d]"
                    onClick={() => void markAllNotificationsRead()}
                    title="全部已读"
                    type="button"
                  >
                    <CheckCheck size={16} />
                  </button>
                ) : null}
                <button
                  aria-label="关闭通知中心"
                  className="inline-flex h-9 w-9 items-center justify-center rounded-md text-[#77736c] hover:bg-[#efede9]"
                  onClick={() => setOpen(false)}
                  type="button"
                >
                  <X size={17} />
                </button>
              </div>
            </header>

            <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
              {notifications.length > 0 ? (
                <div className="divide-y divide-[#e8e5e0]">
                  {notifications.map((notification) => {
                    const Icon = notificationIcons[notification.kind]
                    const unread = !notification.readAt
                    return (
                      <button
                        className={cn(
                          'flex w-full min-w-0 gap-3 px-4 py-4 text-left transition-colors hover:bg-white',
                          unread && 'bg-[#f0f5f2]',
                        )}
                        key={notification.id}
                        onClick={() =>
                          void openNotification(
                            notification.id,
                            notification.actionPath,
                            unread,
                          )
                        }
                        type="button"
                      >
                        <span
                          className={cn(
                            'inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md',
                            notification.kind === 'new_message'
                              ? 'bg-[#e8f0f4] text-[#416879]'
                              : notification.kind === 'daily_picks'
                                ? 'bg-[#eee9f2] text-[#664c76]'
                                : 'bg-[#edf3e9] text-[#4e6f43]',
                          )}
                        >
                          <Icon size={16} />
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="flex items-start justify-between gap-3">
                            <span className="text-sm font-semibold leading-5 text-[#3f3b37]">
                              {notification.title}
                            </span>
                            <span className="shrink-0 text-[10px] text-[#918d86]">
                              {formatNotificationTime(notification.scheduledFor)}
                            </span>
                          </span>
                          <span className="mt-1 block text-xs leading-5 text-[#77736c]">
                            {notification.body}
                          </span>
                        </span>
                        {unread ? (
                          <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-[#527b6d]" />
                        ) : (
                          <ChevronRight className="mt-1 shrink-0 text-[#aaa59d]" size={14} />
                        )}
                      </button>
                    )
                  })}
                  {nextNotificationCursor ? (
                    <button
                      className="flex h-12 w-full items-center justify-center gap-2 text-xs font-semibold text-[#527164] hover:bg-white"
                      disabled={loadingMore}
                      onClick={() => void loadMore()}
                      type="button"
                    >
                      {loadingMore ? <LoaderCircle className="animate-spin" size={14} /> : null}
                      {loadingMore ? '加载中...' : '查看更多'}
                    </button>
                  ) : null}
                </div>
              ) : (
                <div className="flex min-h-72 flex-col items-center justify-center px-8 text-center">
                  <span className="inline-flex h-12 w-12 items-center justify-center rounded-md bg-[#efede9] text-[#817d76]">
                    <BellOff size={20} />
                  </span>
                  <h2 className="mt-4 text-sm font-semibold text-[#45413d]">暂时没有新提醒</h2>
                  <p className="mt-1 text-xs leading-5 text-[#817d76]">
                    新消息、约见和每日精选会出现在这里。
                  </p>
                </div>
              )}
            </div>

            <footer className="flex shrink-0 items-center gap-2 border-t border-[#e2dfda] p-3">
              <Link
                className="inline-flex h-10 min-w-0 flex-1 items-center justify-center gap-2 rounded-md border border-[#d8d4cd] bg-white px-3 text-xs font-semibold text-[#5f5b55] hover:bg-[#f2f0ec]"
                onClick={() => setOpen(false)}
                to="/profile#notification-settings"
              >
                <Settings size={14} />
                提醒设置
              </Link>
              {installAvailable ? (
                <button
                  className="inline-flex h-10 min-w-0 flex-1 items-center justify-center gap-2 rounded-md bg-[#263b35] px-3 text-xs font-semibold text-white hover:bg-[#1d302a]"
                  onClick={() => void installApp()}
                  type="button"
                >
                  <Download size={14} />
                  安装应用
                </button>
              ) : null}
            </footer>
          </section>
        </>
      ) : null}
    </>
  )
}
