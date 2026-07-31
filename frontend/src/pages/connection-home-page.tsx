import {
  ArrowRight,
  CalendarCheck2,
  ChevronRight,
  Clock3,
  MapPin,
  Send,
  Sparkles,
  UsersRound,
} from 'lucide-react'
import { useEffect, useMemo, useState, type KeyboardEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import {
  IntentCategoryIcon,
  intentCategoryMeta,
} from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  fetchIntentFeed,
  fetchIntentRecommendations,
  fetchMatchRooms,
  fetchMyIntents,
  type ConnectionIntent,
  type IntentCategory,
  type IntentRecommendation,
  type MatchRoom,
} from '@/lib/connection-api'
import { getStoredUserProfile } from '@/lib/user-profile-storage'

const starters: Array<{
  category: IntentCategory
  label: string
  prompt: string
}> = [
  {
    category: 'meal',
    label: '找饭搭子',
    prompt: '今晚想在紫荆找个饭搭子，最好也在做 AI 产品',
  },
  {
    category: 'study',
    label: '找学习搭子',
    prompt: '这周找一个安静高效的学习搭子，一起准备期末',
  },
  {
    category: 'project',
    label: '找项目伙伴',
    prompt: '找一个能一起做产品的项目伙伴，先用两周做出原型',
  },
  {
    category: 'cofounder',
    label: '找 CTO / 联创',
    prompt: '为我的 AI 项目找 CTO 或联合创始人，希望能长期投入',
  },
]

function formatCampusDate() {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'short',
  }).format(new Date())
}

function formatMeetingTime(value?: string) {
  if (!value) {
    return '待确认时间'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

export function ConnectionHomePage() {
  const navigate = useNavigate()
  const profile = useMemo(() => getStoredUserProfile(), [])
  const [query, setQuery] = useState('')
  const [feed, setFeed] = useState<ConnectionIntent[]>([])
  const [myIntents, setMyIntents] = useState<ConnectionIntent[]>([])
  const [rooms, setRooms] = useState<MatchRoom[]>([])
  const [recommendations, setRecommendations] = useState<IntentRecommendation[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    let cancelled = false

    const load = async () => {
      setLoading(true)
      setNotice('')
      const [feedResult, intentsResult, roomsResult] = await Promise.allSettled([
        fetchIntentFeed(),
        fetchMyIntents(),
        fetchMatchRooms(),
      ])

      if (cancelled) {
        return
      }

      const nextFeed = feedResult.status === 'fulfilled' ? feedResult.value : []
      const nextIntents =
        intentsResult.status === 'fulfilled' ? intentsResult.value : []
      const nextRooms = roomsResult.status === 'fulfilled' ? roomsResult.value : []

      setFeed(nextFeed)
      setMyIntents(nextIntents)
      setRooms(nextRooms)

      const activeIntent = nextIntents.find((intent) => intent.status === 'active')
      if (activeIntent) {
        try {
          const result = await fetchIntentRecommendations(activeIntent.id)
          if (!cancelled) {
            setRecommendations(result)
          }
        } catch {
          if (!cancelled) {
            setRecommendations([])
          }
        }
      }

      if (
        feedResult.status === 'rejected' &&
        intentsResult.status === 'rejected' &&
        roomsResult.status === 'rejected'
      ) {
        setNotice(getErrorMessage(feedResult.reason, '首页暂时无法加载'))
      }
      setLoading(false)
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const startConversation = (value = query) => {
    const text = value.trim()
    navigate('/intents/new', {
      state: text ? { text } : undefined,
    })
  }

  const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (
      event.key === 'Enter' &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing
    ) {
      event.preventDefault()
      if (query.trim()) {
        startConversation()
      }
    }
  }

  const activeIntents = myIntents.filter((intent) => intent.status === 'active')
  const upcomingRoom = rooms
    .filter((room) => room.status === 'open')
    .sort((left, right) => {
      const leftTime = left.scheduledAt
        ? new Date(left.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER
      const rightTime = right.scheduledAt
        ? new Date(right.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER
      return leftTime - rightTime
    })[0]

  return (
    <div className="h-[calc(100dvh-8.75rem)] overflow-hidden bg-[#f7f6f3] lg:h-screen xl:grid xl:grid-cols-[minmax(0,1fr)_320px]">
      <section className="flex h-full min-h-0 min-w-0 flex-col">
        <header className="flex h-16 shrink-0 items-center justify-between border-b border-[#dfdcd6] px-4 sm:px-6">
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-sm font-semibold text-[#292724]">
                SouLink 助手
              </h1>
              <span className="h-1.5 w-1.5 rounded-full bg-[#4f846f]" />
            </div>
            <p className="mt-0.5 text-[11px] text-[#817d76]">
              {formatCampusDate()} · 清华园
            </p>
          </div>

          <div className="flex items-center gap-3 text-[11px] text-[#77736c] xl:hidden">
            <Link className="hover:text-[#315c4d]" to="/intents">
              {activeIntents.length} 个意图
            </Link>
            <span className="h-3 w-px bg-[#d6d2cb]" />
            <Link className="hover:text-[#315c4d]" to="/matches">
              {rooms.filter((room) => room.status === 'open').length} 个连接
            </Link>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="mx-auto flex min-h-full w-full max-w-3xl flex-col items-center justify-center px-5 py-8 text-center sm:px-8">
            <span className="inline-flex h-11 w-11 items-center justify-center rounded-md bg-[#263b35] text-white">
              <Sparkles size={19} strokeWidth={1.8} />
            </span>
            <h2 className="mt-5 max-w-xl text-2xl font-semibold leading-tight text-[#262421] sm:text-[32px]">
              {profile.nickname
                ? `${profile.nickname}，今天想认识谁？`
                : '今天想认识谁？'}
            </h2>
            <p className="mt-3 max-w-md text-sm leading-6 text-[#77736c]">
              从一件具体的事开始，说一句就够了。
            </p>

            <div className="mt-8 grid w-full max-w-2xl gap-2 sm:grid-cols-2">
              {starters.map(({ category, label, prompt }) => {
                const meta = intentCategoryMeta[category]
                const Icon = meta.icon
                return (
                  <button
                    className="group flex min-h-14 items-center gap-3 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 py-2.5 text-left transition-colors hover:border-[#9eaea8] hover:bg-white"
                    key={category}
                    onClick={() => startConversation(prompt)}
                    type="button"
                  >
                    <span
                      className={`inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md ${meta.surfaceClassName}`}
                    >
                      <Icon className={meta.iconClassName} size={15} />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-xs font-semibold text-[#45413d]">
                        {label}
                      </span>
                      <span className="mt-0.5 block truncate text-[11px] text-[#817d76]">
                        {prompt}
                      </span>
                    </span>
                    <ArrowRight
                      className="shrink-0 text-[#a19c94] transition-transform group-hover:translate-x-0.5 group-hover:text-[#477062]"
                      size={15}
                    />
                  </button>
                )
              })}
            </div>

            {notice ? (
              <p className="mt-5 text-xs text-[#8c463d]">{notice}</p>
            ) : null}
          </div>
        </div>

        <footer className="shrink-0 border-t border-[#dfdcd6] bg-[rgba(247,246,243,0.96)] px-3 py-3 backdrop-blur-lg sm:px-6 sm:py-4">
          <div className="mx-auto max-w-3xl">
            <div className="flex items-end gap-2 rounded-lg border border-[#d6d2cb] bg-white p-2 shadow-[0_8px_24px_rgba(35,33,31,0.06)] focus-within:border-[#6f9185] focus-within:ring-3 focus-within:ring-[#55786c]/10">
              <textarea
                aria-label="描述你想认识的人"
                autoFocus
                className="max-h-28 min-h-10 min-w-0 flex-1 resize-none border-0 bg-transparent px-2 py-2 text-sm leading-6 text-[#292724] outline-none placeholder:text-[#9a958d]"
                maxLength={1200}
                onChange={(event) => setQuery(event.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="描述你想找的人，以及想一起做什么"
                rows={1}
                value={query}
              />
              <button
                aria-label="发送"
                className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#263b35] text-white transition-colors hover:bg-[#1d302a] disabled:cursor-not-allowed disabled:bg-[#cbc8c2]"
                disabled={!query.trim()}
                onClick={() => startConversation()}
                title="发送"
                type="button"
              >
                <Send size={17} />
              </button>
            </div>
          </div>
        </footer>
      </section>

      <aside className="hidden min-h-0 overflow-y-auto border-l border-[#dfdcd6] bg-[#fbfaf8] xl:block">
        <header className="flex h-16 items-center justify-between border-b border-[#dfdcd6] px-5">
          <h2 className="text-sm font-semibold text-[#302e2b]">你的上下文</h2>
          <span className="text-[11px] text-[#817d76]">
            {loading ? '同步中' : '刚刚同步'}
          </span>
        </header>

        <section className="border-b border-[#e2dfd9] px-5 py-5">
          <div className="flex items-center justify-between gap-3">
            <h3 className="text-xs font-semibold text-[#5f5b55]">进行中的意图</h3>
            <Link className="text-[11px] font-medium text-[#477062]" to="/intents">
              管理
            </Link>
          </div>
          {activeIntents.length > 0 ? (
            <div className="mt-3">
              {activeIntents.slice(0, 2).map((intent) => (
                <Link
                  className="flex items-center gap-3 border-b border-[#ece9e4] py-3 last:border-b-0"
                  key={intent.id}
                  to={`/intents/${intent.id}`}
                >
                  <IntentCategoryIcon category={intent.category} size={14} />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-xs font-semibold text-[#45413d]">
                      {intent.title}
                    </span>
                    <span className="mt-0.5 block text-[11px] text-[#817d76]">
                      {intent.interestCount} 个回应
                    </span>
                  </span>
                  <ChevronRight className="text-[#a19c94]" size={14} />
                </Link>
              ))}
            </div>
          ) : (
            <p className="mt-3 text-xs leading-5 text-[#817d76]">暂无活跃意图</p>
          )}
        </section>

        {recommendations.length > 0 ? (
          <section className="border-b border-[#e2dfd9] px-5 py-5">
            <h3 className="text-xs font-semibold text-[#5f5b55]">值得聊聊</h3>
            <div className="mt-3 space-y-3">
              {recommendations.slice(0, 2).map((recommendation) => (
                <Link
                  className="block border-l-2 border-[#b9cbc4] pl-3"
                  key={recommendation.intent.id}
                  to={`/intents/${recommendation.intent.id}`}
                >
                  <span className="block truncate text-xs font-semibold text-[#45413d]">
                    {recommendation.intent.title}
                  </span>
                  <span className="mt-1 line-clamp-2 text-[11px] leading-4 text-[#817d76]">
                    {recommendation.fitReason}
                  </span>
                </Link>
              ))}
            </div>
          </section>
        ) : null}

        <section className="border-b border-[#e2dfd9] px-5 py-5">
          <div className="flex items-center justify-between gap-3">
            <h3 className="text-xs font-semibold text-[#5f5b55]">下一次连接</h3>
            <Link className="text-[11px] font-medium text-[#477062]" to="/matches">
              全部
            </Link>
          </div>
          {upcomingRoom ? (
            <Link className="mt-3 block" to={`/matches/${upcomingRoom.id}`}>
              <div className="flex items-center gap-3">
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#e8f1ed] text-[#356653]">
                  <CalendarCheck2 size={17} />
                </span>
                <span className="min-w-0">
                  <span className="block truncate text-xs font-semibold text-[#45413d]">
                    和 {upcomingRoom.other.nickname}
                  </span>
                  <span className="mt-0.5 block truncate text-[11px] text-[#817d76]">
                    {upcomingRoom.intentTitle}
                  </span>
                </span>
              </div>
              <div className="mt-3 space-y-1.5 text-[11px] text-[#706c65]">
                <p className="flex items-center gap-2">
                  <Clock3 size={13} />
                  {formatMeetingTime(upcomingRoom.scheduledAt)}
                </p>
                <p className="flex items-center gap-2">
                  <MapPin size={13} />
                  {upcomingRoom.selectedPlace || '待确认地点'}
                </p>
              </div>
            </Link>
          ) : (
            <p className="mt-3 text-xs leading-5 text-[#817d76]">暂无待安排的连接</p>
          )}
        </section>

        <section className="px-5 py-5">
          <div className="flex items-center justify-between gap-3">
            <h3 className="text-xs font-semibold text-[#5f5b55]">校园此刻</h3>
            <Link className="text-[11px] font-medium text-[#477062]" to="/discover">
              发现
            </Link>
          </div>
          {feed.length > 0 ? (
            <div className="mt-3">
              {feed.slice(0, 3).map((intent) => (
                <Link
                  className="flex gap-2.5 border-b border-[#ece9e4] py-2.5 last:border-b-0"
                  key={intent.id}
                  to={`/intents/${intent.id}`}
                >
                  <UsersRound
                    className="mt-0.5 shrink-0 text-[#817d76]"
                    size={13}
                  />
                  <span className="line-clamp-2 text-xs leading-5 text-[#5f5b55]">
                    {intent.title}
                  </span>
                </Link>
              ))}
            </div>
          ) : (
            <p className="mt-3 text-xs leading-5 text-[#817d76]">暂时没有公开意图</p>
          )}
        </section>
      </aside>
    </div>
  )
}
