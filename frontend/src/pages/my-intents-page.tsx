import {
  AlertCircle,
  Archive,
  ArrowRight,
  Clock3,
  FileEdit,
  LoaderCircle,
  Plus,
  Radio,
  UsersRound,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { IntentCategoryIcon } from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  expireIntent,
  fetchMyIntents,
  type ConnectionIntent,
  type IntentStatus,
} from '@/lib/connection-api'
import { connectionIntentPath, connectionPaths } from '@/lib/connection-routes'
import { cn } from '@/lib/utils'

type IntentFilter = 'open' | 'draft' | 'closed'

const filterMeta: Array<{ value: IntentFilter; label: string }> = [
  { value: 'open', label: '进行中' },
  { value: 'draft', label: '草稿' },
  { value: 'closed', label: '已结束' },
]

function statusLabel(status: IntentStatus) {
  const labels: Record<IntentStatus, string> = {
    draft: '草稿',
    active: '匹配中',
    paused: '已暂停',
    fulfilled: '已完成',
    expired: '已结束',
  }
  return labels[status]
}

function formatExpiry(value: string | null) {
  if (!value) {
    return '长期有效'
  }
  const diff = new Date(value).getTime() - Date.now()
  if (diff <= 0) {
    return '已到期'
  }
  const days = Math.ceil(diff / 86_400_000)
  return days <= 1 ? '24 小时内到期' : `${days} 天后到期`
}

export function MyIntentsPage() {
  const [filter, setFilter] = useState<IntentFilter>('open')
  const [intents, setIntents] = useState<ConnectionIntent[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [expiringId, setExpiringId] = useState<number | null>(null)

  const load = async () => {
    setLoading(true)
    setNotice('')
    try {
      setIntents(await fetchMyIntents())
    } catch (error) {
      setNotice(getErrorMessage(error, '我的意图暂时无法加载'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const filtered = useMemo(() => {
    return intents.filter((intent) => {
      if (filter === 'open') {
        return intent.status === 'active' || intent.status === 'paused'
      }
      if (filter === 'draft') {
        return intent.status === 'draft'
      }
      return intent.status === 'fulfilled' || intent.status === 'expired'
    })
  }, [filter, intents])

  const counts = useMemo(
    () => ({
      open: intents.filter(
        (intent) => intent.status === 'active' || intent.status === 'paused',
      ).length,
      draft: intents.filter((intent) => intent.status === 'draft').length,
      closed: intents.filter(
        (intent) =>
          intent.status === 'fulfilled' || intent.status === 'expired',
      ).length,
    }),
    [intents],
  )

  const handleExpire = async (intentId: number) => {
    if (expiringId) {
      return
    }
    setExpiringId(intentId)
    setNotice('')
    try {
      await expireIntent(intentId)
      setIntents((current) =>
        current.map((intent) =>
          intent.id === intentId
            ? { ...intent, status: 'expired' as const }
            : intent,
        ),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '这个意图暂时无法结束'))
    } finally {
      setExpiringId(null)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1040px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex flex-col gap-4 border-b border-[#dfdcd6] pb-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-medium text-[#6c4a80]">MY INTENTS</p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">我的意图</h1>
          <p className="mt-1 text-sm text-[#77736c]">每一次寻找，都有清楚的状态和边界。</p>
        </div>
        <Link
          className="inline-flex h-10 items-center justify-center gap-2 self-start rounded-md bg-[#5b3477] px-4 text-sm font-semibold text-white hover:bg-[#4c2b65]"
          to={connectionPaths.newIntent}
        >
          <Plus size={16} />
          新建意图
        </Link>
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <div className="mt-6 grid grid-cols-3 gap-2">
        {filterMeta.map((item) => (
          <button
            className={cn(
              'rounded-md border px-3 py-3 text-left transition-colors',
              filter === item.value
                ? 'border-[#bba9c5] bg-[#f0ebf3]'
                : 'border-[#dedad4] bg-[#fbfaf8] hover:bg-white',
            )}
            key={item.value}
            onClick={() => setFilter(item.value)}
            type="button"
          >
            <span className="block text-lg font-semibold text-[#37332f]">
              {counts[item.value]}
            </span>
            <span className="mt-0.5 block text-xs text-[#77736c]">{item.label}</span>
          </button>
        ))}
      </div>

      <section className="py-6">
        {loading ? (
          <div className="flex min-h-64 items-center justify-center text-sm text-[#77736c]">
            <LoaderCircle className="mr-2 animate-spin" size={17} />
            正在整理你的意图...
          </div>
        ) : filtered.length > 0 ? (
          <div className="space-y-3">
            {filtered.map((intent) => (
              <article
                className="rounded-lg border border-[#dfdcd6] bg-white p-4 sm:p-5"
                key={intent.id}
              >
                <div className="flex items-start gap-3">
                  <IntentCategoryIcon category={intent.category} size={18} />
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <span
                            className={cn(
                              'rounded-md px-2 py-1 text-[11px] font-semibold',
                              intent.status === 'active'
                                ? 'bg-[#e8f3ed] text-[#28664f]'
                                : intent.status === 'draft'
                                  ? 'bg-[#eeeae5] text-[#6d6963]'
                                  : 'bg-[#f3edf0] text-[#765867]',
                            )}
                          >
                            {statusLabel(intent.status)}
                          </span>
                          <span className="text-xs text-[#817d76]">
                            {formatExpiry(intent.expiresAt)}
                          </span>
                        </div>
                        <Link
                          className="mt-2 block text-base font-semibold leading-6 text-[#302e2b] hover:text-[#5b3477]"
                          to={connectionIntentPath(intent.id)}
                        >
                          {intent.title}
                        </Link>
                        <p className="mt-1 line-clamp-2 text-sm leading-6 text-[#706c65]">
                          {intent.description}
                        </p>
                      </div>
                      <Link
                        aria-label="查看意图"
                        className="inline-flex h-9 shrink-0 items-center justify-center gap-1.5 self-start rounded-md border border-[#d8d3cc] px-3 text-xs font-semibold text-[#5f5b55] hover:border-[#bba9c5] hover:text-[#5b3477]"
                        to={connectionIntentPath(intent.id)}
                      >
                        查看
                        <ArrowRight size={14} />
                      </Link>
                    </div>

                    <div className="mt-4 grid gap-2 border-t border-[#ece9e4] pt-3 text-xs text-[#706c65] sm:grid-cols-3">
                      <span className="inline-flex items-center gap-1.5">
                        <UsersRound size={14} />
                        {intent.interestCount} 个回应
                      </span>
                      <span className="inline-flex items-center gap-1.5">
                        <Clock3 size={14} />
                        {intent.timeWindow || '时间待定'}
                      </span>
                      <span className="inline-flex items-center gap-1.5">
                        {intent.status === 'draft' ? (
                          <FileEdit size={14} />
                        ) : (
                          <Radio size={14} />
                        )}
                        {intent.visibility === 'feed'
                          ? '公开意图流'
                          : intent.visibility === 'candidates'
                            ? '仅候选可见'
                            : '仅 AI 搜索'}
                      </span>
                    </div>

                    {intent.status === 'active' ? (
                      <div className="mt-4 flex justify-end">
                        <button
                          className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium text-[#80665c] hover:bg-[#f4ece8] hover:text-[#8b4438] disabled:opacity-50"
                          disabled={expiringId === intent.id}
                          onClick={() => void handleExpire(intent.id)}
                          type="button"
                        >
                          {expiringId === intent.id ? (
                            <LoaderCircle className="animate-spin" size={14} />
                          ) : (
                            <Archive size={14} />
                          )}
                          结束意图
                        </button>
                      </div>
                    ) : null}
                  </div>
                </div>
              </article>
            ))}
          </div>
        ) : (
          <div className="flex min-h-64 flex-col items-center justify-center rounded-lg border border-dashed border-[#d7d2cb] px-6 text-center">
            <p className="text-sm font-medium text-[#55514c]">
              {filter === 'open'
                ? '目前没有进行中的寻找'
                : filter === 'draft'
                  ? '没有未完成的草稿'
                  : '还没有结束的意图'}
            </p>
            {filter !== 'closed' ? (
              <Link
                className="mt-4 inline-flex h-9 items-center gap-2 rounded-md bg-[#5b3477] px-3 text-xs font-semibold text-white"
                to={connectionPaths.newIntent}
              >
                <Plus size={14} />
                新建意图
              </Link>
            ) : null}
          </div>
        )}
      </section>
    </div>
  )
}
