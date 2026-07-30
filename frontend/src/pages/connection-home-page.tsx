import {
  ArrowRight,
  CalendarCheck2,
  Clock3,
  MapPin,
  Plus,
  Sparkles,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import thuCampusUrl from '@/assets/optimized/heroes/thu.webp'
import { IntentCard } from '@/components/connection/intent-card'
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

const shortcuts: Array<{
  category: IntentCategory
  prompt: string
}> = [
  { category: 'meal', prompt: '今晚想在紫荆找个饭搭子' },
  { category: 'study', prompt: '这周找一个安静高效的学习搭子' },
  { category: 'sport', prompt: '找一个工作日晚上的跑步搭子' },
  { category: 'conversation', prompt: '想找一个人散步认真聊聊' },
  { category: 'project', prompt: '找一个能一起做产品的项目伙伴' },
  { category: 'cofounder', prompt: '为我的 AI 项目找 CTO 或联合创始人' },
  { category: 'romantic', prompt: '想认真认识一个有好奇心的人' },
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
    return '等待双方确认时间'
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
      } else {
        setRecommendations([])
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

  const startComposer = (value = query) => {
    const text = value.trim()
    navigate('/intents/new', {
      state: text ? { text } : undefined,
    })
  }

  const activeIntents = myIntents.filter((intent) => intent.status === 'active')
  const upcomingRooms = rooms
    .filter((room) => room.status === 'open')
    .sort((left, right) => {
      const leftTime = left.scheduledAt
        ? new Date(left.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER
      const rightTime = right.scheduledAt
        ? new Date(right.scheduledAt).getTime()
        : Number.MAX_SAFE_INTEGER
      return leftTime - rightTime
    })

  return (
    <div className="mx-auto w-full max-w-[1180px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex items-end justify-between gap-5 border-b border-[#dfdcd6] pb-5">
        <div>
          <p className="text-xs font-medium text-[#77736c]">{formatCampusDate()} · 清华园</p>
          <h1 className="mt-1 text-2xl font-semibold text-[#23211f] sm:text-[28px]">
            {profile.nickname ? `${profile.nickname}，今天想认识谁？` : '今天想认识谁？'}
          </h1>
        </div>
        <img
          alt="清华校园"
          className="hidden h-16 w-40 rounded-md object-cover sm:block"
          src={thuCampusUrl}
        />
      </header>

      <section className="border-b border-[#dfdcd6] py-6">
        <div className="rounded-lg border border-[#d9d4cd] bg-white p-3 shadow-[0_10px_28px_rgba(35,33,31,0.05)] focus-within:border-[#7b5a91] focus-within:ring-3 focus-within:ring-[#5b3477]/10 sm:p-4">
          <textarea
            aria-label="描述你想认识的人"
            autoFocus
            className="min-h-24 w-full resize-none border-0 bg-transparent px-1 py-1 text-[17px] leading-7 text-[#292724] outline-none placeholder:text-[#9a958d]"
            maxLength={1200}
            onChange={(event) => setQuery(event.target.value)}
            onKeyDown={(event) => {
              if ((event.metaKey || event.ctrlKey) && event.key === 'Enter') {
                startComposer()
              }
            }}
            placeholder="例如：今晚想在紫荆找个饭搭子，最好也在做 AI 产品"
            value={query}
          />
          <div className="mt-2 flex items-center justify-between gap-3 border-t border-[#eeebe7] pt-3">
            <span className="inline-flex min-w-0 items-center gap-1.5 truncate text-xs text-[#77736c]">
              <Sparkles size={14} className="text-[#69447f]" />
              SouLink 会先整理意图，再给你确认
            </span>
            <button
              className="inline-flex h-10 shrink-0 items-center gap-2 rounded-md bg-[#5b3477] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#4c2b65] disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!query.trim()}
              onClick={() => startComposer()}
              type="button"
            >
              开始整理
              <ArrowRight size={16} />
            </button>
          </div>
        </div>

        <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
          {shortcuts.map(({ category, prompt }) => {
            const meta = intentCategoryMeta[category]
            return (
              <button
                className="inline-flex h-9 shrink-0 items-center gap-2 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 text-xs font-medium text-[#5f5b55] transition-colors hover:border-[#b9a8c4] hover:bg-white hover:text-[#4d2d66]"
                key={category}
                onClick={() => {
                  setQuery(prompt)
                  startComposer(prompt)
                }}
                type="button"
              >
                <meta.icon className={meta.iconClassName} size={15} />
                {meta.label}
              </button>
            )
          })}
        </div>
      </section>

      {notice ? (
        <div className="mt-5 rounded-md border border-[#ead6d1] bg-[#fff7f5] px-4 py-3 text-sm text-[#8e4941]">
          {notice}
        </div>
      ) : null}

      <div className="grid gap-7 py-7 lg:grid-cols-[minmax(0,1.55fr)_minmax(280px,0.75fr)]">
        <div className="min-w-0 space-y-8">
          <section>
            <div className="mb-4 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-[#292724]">今天在校园里</h2>
                <p className="mt-0.5 text-xs text-[#7b7770]">有限的、仍然有效的真实需求</p>
              </div>
              <Link
                className="inline-flex items-center gap-1 text-sm font-medium text-[#5b3477] hover:text-[#45265e]"
                to="/discover"
              >
                查看全部
                <ArrowRight size={15} />
              </Link>
            </div>

            <div className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-5">
              {loading ? (
                <p className="py-8 text-center text-sm text-[#817d76]">正在看看校园里发生什么...</p>
              ) : feed.length > 0 ? (
                feed.slice(0, 4).map((intent) => (
                  <IntentCard compact intent={intent} key={intent.id} />
                ))
              ) : (
                <div className="py-8 text-center">
                  <p className="text-sm font-medium text-[#4f4b46]">今天的公开意图还很安静</p>
                  <button
                    className="mt-3 inline-flex items-center gap-1.5 text-sm font-medium text-[#5b3477]"
                    onClick={() => startComposer()}
                    type="button"
                  >
                    <Plus size={15} />
                    发出第一个邀请
                  </button>
                </div>
              )}
            </div>
          </section>

          <section>
            <div className="mb-4 flex items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-[#292724]">可能值得聊聊</h2>
                <p className="mt-0.5 text-xs text-[#7b7770]">基于你最近一个活跃意图的三条推荐</p>
              </div>
              {activeIntents[0] ? (
                <Link
                  className="inline-flex items-center gap-1 text-sm font-medium text-[#5b3477]"
                  to={`/intents/${activeIntents[0].id}`}
                >
                  查看解释
                  <ArrowRight size={15} />
                </Link>
              ) : null}
            </div>

            {recommendations.length > 0 ? (
              <div className="grid gap-3 md:grid-cols-3">
                {recommendations.map((recommendation) => (
                  <Link
                    className="min-w-0 rounded-lg border border-[#dfdcd6] bg-white p-4 transition-colors hover:border-[#b9a8c4]"
                    key={recommendation.intent.id}
                    to={`/intents/${recommendation.intent.id}`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <IntentCategoryIcon
                        category={recommendation.intent.category}
                        size={16}
                      />
                      <span className="text-xs font-semibold text-[#5b3477]">
                        {recommendation.score}% 匹配线索
                      </span>
                    </div>
                    <h3 className="mt-3 line-clamp-2 text-sm font-semibold leading-5 text-[#302e2b]">
                      {recommendation.intent.title}
                    </h3>
                    <p className="mt-2 line-clamp-3 text-xs leading-5 text-[#706c65]">
                      {recommendation.fitReason}
                    </p>
                  </Link>
                ))}
              </div>
            ) : (
              <div className="rounded-lg border border-dashed border-[#d8d3cc] px-5 py-7 text-center">
                <p className="text-sm text-[#706c65]">
                  {activeIntents.length > 0
                    ? '还没有足够合适的候选，SouLink 不会为了凑数而推荐。'
                    : '发布一个具体意图后，这里会出现最多三条推荐。'}
                </p>
              </div>
            )}
          </section>
        </div>

        <aside className="min-w-0 space-y-7">
          <section>
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[#302e2b]">你的进行中</h2>
              <Link className="text-xs font-medium text-[#5b3477]" to="/intents">
                管理
              </Link>
            </div>
            <div className="rounded-lg border border-[#dfdcd6] bg-white">
              {activeIntents.length > 0 ? (
                activeIntents.slice(0, 3).map((intent) => (
                  <Link
                    className="flex items-center gap-3 border-b border-[#ebe8e3] px-4 py-3.5 last:border-b-0 hover:bg-[#faf9f7]"
                    key={intent.id}
                    to={`/intents/${intent.id}`}
                  >
                    <IntentCategoryIcon category={intent.category} size={15} />
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium text-[#36332f]">
                        {intent.title}
                      </span>
                      <span className="mt-0.5 block text-xs text-[#817d76]">
                        {intent.interestCount} 个回应
                      </span>
                    </span>
                    <ArrowRight size={15} className="text-[#918d86]" />
                  </Link>
                ))
              ) : (
                <div className="px-4 py-6 text-center">
                  <p className="text-sm text-[#77736c]">暂无活跃意图</p>
                  <Link
                    className="mt-2 inline-flex items-center gap-1 text-xs font-medium text-[#5b3477]"
                    to="/intents/new"
                  >
                    <Plus size={13} />
                    新建意图
                  </Link>
                </div>
              )}
            </div>
          </section>

          <section>
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-base font-semibold text-[#302e2b]">即将见面</h2>
              <Link className="text-xs font-medium text-[#5b3477]" to="/matches">
                匹配房间
              </Link>
            </div>
            <div className="rounded-lg border border-[#dfdcd6] bg-[#eef4f1] p-4">
              {upcomingRooms.length > 0 ? (
                <Link className="block" to={`/matches/${upcomingRooms[0].id}`}>
                  <div className="flex items-center gap-3">
                    <span className="inline-flex h-10 w-10 items-center justify-center rounded-md bg-white text-[#29634f]">
                      <CalendarCheck2 size={19} />
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-semibold text-[#2f4039]">
                        和 {upcomingRooms[0].other.nickname}
                      </span>
                      <span className="mt-0.5 block truncate text-xs text-[#65766f]">
                        {upcomingRooms[0].intentTitle}
                      </span>
                    </span>
                  </div>
                  <div className="mt-4 space-y-2 border-t border-[#d7e3dd] pt-3 text-xs text-[#53665e]">
                    <p className="flex items-center gap-2">
                      <Clock3 size={14} />
                      {formatMeetingTime(upcomingRooms[0].scheduledAt)}
                    </p>
                    <p className="flex items-center gap-2">
                      <MapPin size={14} />
                      {upcomingRooms[0].selectedPlace || '等待双方选择地点'}
                    </p>
                  </div>
                </Link>
              ) : (
                <div className="py-3 text-center text-sm text-[#65766f]">
                  双方都愿意后，见面安排会出现在这里。
                </div>
              )}
            </div>
          </section>
        </aside>
      </div>
    </div>
  )
}
