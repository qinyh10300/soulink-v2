import {
  AlertCircle,
  ArrowRight,
  Bookmark,
  Check,
  Compass,
  EyeOff,
  Filter,
  LoaderCircle,
  MessageSquarePlus,
  RefreshCw,
  Sparkles,
  X,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { IntentCard } from '@/components/connection/intent-card'
import { intentCategoryMeta } from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  dismissIntent,
  expressIntentInterest,
  fetchIntentFeed,
  fetchSavedIntents,
  setIntentSaved,
  type ConnectionIntent,
  type IntentCategory,
} from '@/lib/connection-api'
import { connectionPaths } from '@/lib/connection-routes'
import { cn } from '@/lib/utils'

const filters: Array<{ value: IntentCategory | ''; label: string }> = [
  { value: '', label: '全部' },
  { value: 'meal', label: '吃饭' },
  { value: 'study', label: '学习' },
  { value: 'sport', label: '运动' },
  { value: 'conversation', label: '聊天' },
  { value: 'event', label: '活动' },
  { value: 'project', label: '项目' },
  { value: 'cofounder', label: 'CTO / 联创' },
]

export function IntentFeedPage() {
  const [mode, setMode] = useState<'campus' | 'saved'>('campus')
  const [category, setCategory] = useState<IntentCategory | ''>('')
  const [intents, setIntents] = useState<ConnectionIntent[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [savingId, setSavingId] = useState<number | null>(null)
  const [explainedId, setExplainedId] = useState<number | null>(null)
  const [interestId, setInterestId] = useState<number | null>(null)
  const [interestMessage, setInterestMessage] = useState('')
  const [sendingInterest, setSendingInterest] = useState(false)
  const [dismissingId, setDismissingId] = useState<number | null>(null)

  const loadFeed = async (nextCategory = category, nextMode = mode) => {
    setLoading(true)
    setNotice('')
    try {
      setIntents(
        await (nextMode === 'saved'
          ? fetchSavedIntents(nextCategory)
          : fetchIntentFeed(nextCategory)),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '意图流暂时无法加载'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadFeed(category, mode)
    // loadFeed is scoped to this page and keyed by category and mode.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [category, mode])

  const handleSave = async (intent: ConnectionIntent) => {
    if (savingId) {
      return
    }
    setSavingId(intent.id)
    try {
      const updated = await setIntentSaved(intent.id, !intent.savedByMe)
      setIntents((current) => {
        if (mode === 'saved' && !updated.savedByMe) {
          return current.filter((item) => item.id !== updated.id)
        }
        return current.map((item) => (item.id === updated.id ? updated : item))
      })
    } catch (error) {
      setNotice(getErrorMessage(error, '收藏状态没有保存'))
    } finally {
      setSavingId(null)
    }
  }

  const handleDismiss = async (intentId: number) => {
    if (dismissingId) {
      return
    }
    setDismissingId(intentId)
    setNotice('')
    try {
      await dismissIntent(intentId)
      setIntents((current) => current.filter((intent) => intent.id !== intentId))
    } catch (error) {
      setNotice(getErrorMessage(error, '这条意图暂时无法忽略'))
    } finally {
      setDismissingId(null)
    }
  }

  const handleInterest = async (intentId: number) => {
    if (sendingInterest) {
      return
    }
    setSendingInterest(true)
    setNotice('')
    try {
      const updated = await expressIntentInterest(intentId, interestMessage)
      setIntents((current) =>
        current.map((item) => (item.id === updated.id ? updated : item)),
      )
      setInterestId(null)
      setInterestMessage('')
    } catch (error) {
      setNotice(getErrorMessage(error, '回应没有发出去'))
    } finally {
      setSendingInterest(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[980px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex flex-col gap-4 border-b border-[#dfdcd6] pb-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="flex items-center gap-1.5 text-xs font-medium text-[#477062]">
            <Filter size={13} />
            CAMPUS NOW
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">校园邀请</h1>
          <p className="mt-1 text-sm text-[#77736c]">
            清华园里，正在发生的具体邀请。
          </p>
        </div>
        <Link
          className="inline-flex h-10 items-center justify-center gap-2 self-start rounded-md bg-[#263b35] px-4 text-sm font-semibold text-white hover:bg-[#1d302a]"
          to={connectionPaths.newIntent}
        >
          <Sparkles size={16} />
          让 AI 帮我找
          <ArrowRight size={16} />
        </Link>
      </header>

      <div className="sticky top-15 z-20 -mx-4 overflow-x-auto border-b border-[#e1ded8] bg-[rgba(247,246,243,0.95)] px-4 py-3 backdrop-blur-lg sm:-mx-6 sm:px-6 lg:top-0 lg:-mx-8 lg:px-8">
        <div className="flex min-w-max gap-2">
          <button
            className={cn(
              'inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium transition-colors',
              mode === 'campus'
                ? 'bg-[#e8f1ed] text-[#315c4d]'
                : 'text-[#706c65] hover:bg-[#ece9e5]',
            )}
            onClick={() => setMode('campus')}
            type="button"
          >
            <Compass size={13} />
            校园此刻
          </button>
          <button
            className={cn(
              'inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium transition-colors',
              mode === 'saved'
                ? 'bg-[#eee8f2] text-[#5b3477]'
                : 'text-[#706c65] hover:bg-[#ece9e5]',
            )}
            onClick={() => setMode('saved')}
            type="button"
          >
            <Bookmark size={13} />
            已收藏
          </button>
          <span className="mx-1 h-8 w-px bg-[#d8d4cd]" />
          {filters.map((filter) => (
            <button
              className={cn(
                'h-8 rounded-md px-3 text-xs font-medium transition-colors',
                category === filter.value
                  ? 'bg-[#e8f1ed] text-[#315c4d]'
                  : 'text-[#706c65] hover:bg-[#ece9e5]',
              )}
              key={filter.value}
              onClick={() => setCategory(filter.value)}
              type="button"
            >
              {filter.label}
            </button>
          ))}
          <span className="ml-1 inline-flex h-8 items-center px-2 text-[11px] text-[#8a867f]">
            本轮 {intents.length} 条
          </span>
        </div>
      </div>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <section className="py-6">
        {loading ? (
          <div className="flex min-h-72 items-center justify-center text-sm text-[#77736c]">
            <LoaderCircle className="mr-2 animate-spin" size={17} />
            正在读取仍然有效的意图...
          </div>
        ) : intents.length > 0 ? (
          <div className="border-y border-[#dfdcd6] bg-[#fbfaf8] px-4 py-1 sm:px-6">
            {intents.map((intent) => {
              const meta = intentCategoryMeta[intent.category]
              const interestOpen = interestId === intent.id
              const explanationOpen = explainedId === intent.id

              return (
                <IntentCard
                  footer={
                    <div>
                      {explanationOpen ? (
                        <div className="mb-3 flex items-start gap-2 rounded-md bg-[#eaf2ee] px-3 py-2.5 text-xs leading-5 text-[#426456]">
                          <Sparkles className="mt-0.5 shrink-0" size={14} />
                          <span>
                            这条来自公开意图流，不是个性化推荐。你可以先核对
                            {intent.timeWindow ? `「${intent.timeWindow}」` : '时间'}
                            {intent.location ? `、地点「${intent.location}」` : '、地点'}
                            和目标「{meta.label}」是否与自己一致。
                          </span>
                        </div>
                      ) : null}

                      {interestOpen ? (
                        <div className="mb-3 rounded-md border border-[#d9d4cd] bg-white p-3">
                          <div className="flex items-center justify-between gap-3">
                            <p className="text-xs font-semibold text-[#4f4b46]">给对方一句低压力的附言</p>
                            <button
                              aria-label="关闭回应输入"
                              className="inline-flex h-7 w-7 items-center justify-center rounded-md text-[#817d76] hover:bg-[#efede9]"
                              onClick={() => {
                                setInterestId(null)
                                setInterestMessage('')
                              }}
                              type="button"
                            >
                              <X size={15} />
                            </button>
                          </div>
                          <textarea
                            autoFocus
                            className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#dedad4] bg-[#fcfbf9] p-2.5 text-sm leading-5 outline-none focus:border-[#7a5a8e]"
                            maxLength={500}
                            onChange={(event) => setInterestMessage(event.target.value)}
                            placeholder="例如：我今晚也有空，而且最近也在关注这个话题。"
                            value={interestMessage}
                          />
                          <div className="mt-2 flex justify-end">
                            <button
                              className="inline-flex h-9 items-center gap-2 rounded-md bg-[#263b35] px-3 text-xs font-semibold text-white disabled:opacity-50"
                              disabled={sendingInterest}
                              onClick={() => void handleInterest(intent.id)}
                              type="button"
                            >
                              {sendingInterest ? (
                                <LoaderCircle className="animate-spin" size={14} />
                              ) : (
                                <MessageSquarePlus size={14} />
                              )}
                              发送兴趣
                            </button>
                          </div>
                        </div>
                      ) : null}

                      <div className="flex flex-wrap items-center gap-2">
                        {intent.interestState === 'interested' ||
                        intent.interestState === 'accepted' ? (
                          <span className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#e8f3ed] px-3 text-xs font-semibold text-[#2d6a53]">
                            <Check size={14} />
                            {intent.interestState === 'accepted'
                              ? '已互相同意'
                              : '已表达兴趣'}
                          </span>
                        ) : (
                          <button
                            className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#263b35] px-3 text-xs font-semibold text-white hover:bg-[#1d302a]"
                            onClick={() => {
                              setInterestId(intent.id)
                              setExplainedId(null)
                            }}
                            type="button"
                          >
                            <MessageSquarePlus size={14} />
                            我感兴趣
                          </button>
                        )}
                        <button
                          className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium text-[#625e58] hover:bg-[#eeece8]"
                          onClick={() =>
                            setExplainedId(explanationOpen ? null : intent.id)
                          }
                          type="button"
                        >
                          <Sparkles size={14} />
                          问 AI 为什么
                        </button>
                        {intent.interestState !== 'interested' &&
                        intent.interestState !== 'accepted' ? (
                          <button
                            className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium text-[#77736c] hover:bg-[#eeece8] disabled:opacity-50"
                            disabled={dismissingId === intent.id}
                            onClick={() => void handleDismiss(intent.id)}
                            type="button"
                          >
                            {dismissingId === intent.id ? (
                              <LoaderCircle className="animate-spin" size={14} />
                            ) : (
                              <EyeOff size={14} />
                            )}
                            不适合我
                          </button>
                        ) : null}
                      </div>
                    </div>
                  }
                  intent={intent}
                  key={intent.id}
                  onSave={handleSave}
                  saving={savingId === intent.id}
                />
              )
            })}
            {mode === 'campus' ? (
              <div className="flex flex-col items-center justify-center border-t border-[#e6e2dc] py-6 text-center">
                <p className="text-xs font-semibold text-[#55514c]">本轮邀请已看完</p>
                <p className="mt-1 text-[11px] text-[#817d76]">
                  下一步留给表达兴趣、收藏，或发起自己的邀请。
                </p>
              </div>
            ) : null}
          </div>
        ) : (
          <div className="flex min-h-72 flex-col items-center justify-center rounded-lg border border-dashed border-[#d7d2cb] px-6 text-center">
            <p className="text-sm font-medium text-[#55514c]">
              {mode === 'saved' ? '这个分类还没有收藏' : '这个分类暂时没有活跃意图'}
            </p>
            <p className="mt-1 text-xs text-[#817d76]">
              {mode === 'saved' ? '在校园意图中收藏后，会留在这里。' : '过期需求会自动离开这里。'}
            </p>
            <button
              className="mt-4 inline-flex h-9 items-center gap-2 rounded-md border border-[#b9cbc4] px-3 text-xs font-semibold text-[#315c4d]"
              onClick={() => void loadFeed()}
              type="button"
            >
              <RefreshCw size={14} />
              刷新
            </button>
          </div>
        )}
      </section>
    </div>
  )
}
