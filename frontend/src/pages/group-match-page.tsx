import {
  ArrowRight,
  CalendarDays,
  Check,
  Clock3,
  LoaderCircle,
  MapPin,
  MessageCircleMore,
  Plus,
  RefreshCw,
  ShieldCheck,
  Sparkles,
  UserRoundCheck,
  UsersRound,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import { connectionPaths } from '@/lib/connection-routes'
import {
  createGroupMatchIntent,
  fetchGroupMatchIntent,
  fetchGroupMatchRecommendations,
  fetchMyGroupMatchIntents,
  fetchMyGroupMatchInvitations,
  formGroupMatchIntent,
  inviteGroupMatchCandidate,
  leaveGroupMatchIntent,
  respondGroupMatchInvitation,
  type GroupMatchCategory,
  type GroupMatchInboxItem,
  type GroupMatchIntent,
  type GroupMatchIntentInput,
  type GroupMatchRecommendation,
  type GroupTimeSuggestion,
} from '@/lib/group-match-api'
import { cn } from '@/lib/utils'

const categories: Array<{
  value: GroupMatchCategory
  label: string
  hint: string
}> = [
  { value: 'meal', label: '饭局', hint: '一起吃饭' },
  { value: 'study', label: '自习', hint: '稳定学习' },
  { value: 'sport', label: '运动', hint: '固定节奏' },
  { value: 'event', label: '活动', hint: '共同参与' },
  { value: 'project', label: '项目', hint: '角色补位' },
]

const categoryLabel = Object.fromEntries(
  categories.map((item) => [item.value, item.label]),
) as Record<GroupMatchCategory, string>

const statusLabels: Record<GroupMatchIntent['status'], string> = {
  open: '组队中',
  joined: '已加入其他组',
  formed: '已成组',
  expired: '已到期',
  cancelled: '已取消',
}

const invitationStatusLabels = {
  invited: '待确认',
  accepted: '已接受',
  declined: '已拒绝',
  waitlisted: '候补中',
  left: '已退出',
} as const

function splitTags(value: string) {
  return value
    .split(/[，,、\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function formatDeadline(value: string) {
  const deadline = new Date(value)
  const diff = deadline.getTime() - Date.now()
  if (diff <= 0) return '已截止'
  const days = Math.ceil(diff / 86_400_000)
  return days <= 1 ? '24 小时内截止' : `${days} 天后截止`
}

function formatSchedule(value?: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    weekday: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function InferenceLabel({ intent }: { intent: GroupMatchIntent }) {
  return (
    <span className="inline-flex items-center gap-1 text-[11px] text-[#77736c]">
      <Sparkles size={12} />
      {intent.inference.source === 'model' ? '模型整理' : '本地规则'} ·{' '}
      {Math.round(intent.inference.confidence * 100)}%
    </span>
  )
}

function CommonTimes({ items }: { items: GroupTimeSuggestion[] }) {
  if (items.length === 0) {
    return <p className="text-xs text-[#8a857e]">共同空闲仍待双方补充或确认</p>
  }
  return (
    <div className="flex flex-wrap gap-2">
      {items.slice(0, 3).map((item) => (
        <span
          className="inline-flex items-center gap-1.5 rounded-md bg-[#edf3ef] px-2.5 py-1.5 text-xs text-[#315c4d]"
          key={item.startsAt}
          title={item.explanation}
        >
          <Clock3 size={13} />
          {item.label}
        </span>
      ))}
    </div>
  )
}

interface ComposerProps {
  onClose: () => void
  onCreated: (intent: GroupMatchIntent) => void
}

function GroupIntentComposer({ onClose, onCreated }: ComposerProps) {
  const [category, setCategory] = useState<GroupMatchCategory>('meal')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [roleOffer, setRoleOffer] = useState('')
  const [rolesNeeded, setRolesNeeded] = useState('')
  const [constraints, setConstraints] = useState('')
  const [timeWindow, setTimeWindow] = useState('')
  const [location, setLocation] = useState('')
  const [minMembers, setMinMembers] = useState(3)
  const [maxMembers, setMaxMembers] = useState(5)
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [createCalendar, setCreateCalendar] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [notice, setNotice] = useState('')

  const handleSubmit = async () => {
    if (!title.trim()) {
      setNotice('请先写清这次组队要做什么')
      return
    }
    if (maxMembers < minMembers) {
      setNotice('目标人数不能低于最低成组人数')
      return
    }
    if ((startsAt && !endsAt) || (!startsAt && endsAt)) {
      setNotice('请同时填写开始和结束时间')
      return
    }
    const input: GroupMatchIntentInput = {
      category,
      title: title.trim(),
      description: description.trim(),
      roleOffer: roleOffer.trim(),
      rolesNeeded: splitTags(rolesNeeded),
      constraints: splitTags(constraints),
      timeWindow: timeWindow.trim(),
      location: location.trim(),
      minMembers,
      maxMembers,
      createCalendarEvent: createCalendar,
      ...(startsAt && endsAt
        ? {
            startsAt: new Date(startsAt).toISOString(),
            endsAt: new Date(endsAt).toISOString(),
          }
        : {}),
    }
    try {
      setSubmitting(true)
      setNotice('')
      onCreated(await createGroupMatchIntent(input))
    } catch (error) {
      setNotice(getErrorMessage(error, '组队意图创建失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-60 flex items-end justify-center bg-black/45 sm:items-center sm:p-5">
      <section
        aria-modal="true"
        className="flex max-h-[92dvh] w-full max-w-2xl flex-col rounded-t-lg bg-[#fbfaf8] shadow-2xl sm:rounded-lg"
        role="dialog"
      >
        <header className="flex h-15 shrink-0 items-center border-b border-[#dfdcd6] px-4 sm:px-5">
          <div>
            <h2 className="text-base font-semibold text-[#292724]">发起多人组队</h2>
            <p className="text-[11px] text-[#817d76]">仅匹配主动发布过同类意图的同学</p>
          </div>
          <button
            aria-label="关闭"
            className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-[#efede9]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X size={16} />
          </button>
        </header>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto px-4 py-5 sm:px-5">
          <fieldset>
            <legend className="text-xs font-semibold text-[#625e58]">场景</legend>
            <div className="mt-2 grid grid-cols-5 gap-1 rounded-md bg-[#efede9] p-1">
              {categories.map((item) => (
                <button
                  className={cn(
                    'min-w-0 rounded px-1 py-2 text-xs font-medium',
                    category === item.value
                      ? 'bg-white text-[#315c4d] shadow-sm'
                      : 'text-[#77736c] hover:text-[#34312e]',
                  )}
                  key={item.value}
                  onClick={() => setCategory(item.value)}
                  title={item.hint}
                  type="button"
                >
                  {item.label}
                </button>
              ))}
            </div>
          </fieldset>

          <label className="block">
            <span className="text-xs font-semibold text-[#625e58]">组队目标</span>
            <input
              className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
              maxLength={100}
              onChange={(event) => setTitle(event.target.value)}
              placeholder="例如：周四晚组一个三人 AI 产品验证小队"
              value={title}
            />
          </label>

          <label className="block">
            <span className="text-xs font-semibold text-[#625e58]">补充说明</span>
            <textarea
              className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#d4d0c9] bg-white px-3 py-2 text-sm leading-6 outline-none focus:border-[#527b6d]"
              maxLength={1200}
              onChange={(event) => setDescription(event.target.value)}
              placeholder="目标、已有进展和这次希望共同完成的结果"
              value={description}
            />
          </label>

          <div className="grid gap-4 sm:grid-cols-2">
            <label>
              <span className="text-xs font-semibold text-[#625e58]">我能负责</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                maxLength={80}
                onChange={(event) => setRoleOffer(event.target.value)}
                placeholder="产品、组织、后端..."
                value={roleOffer}
              />
            </label>
            <label>
              <span className="text-xs font-semibold text-[#625e58]">还需要</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                onChange={(event) => setRolesNeeded(event.target.value)}
                placeholder="前端、设计，用逗号分隔"
                value={rolesNeeded}
              />
            </label>
          </div>

          <label className="block">
            <span className="text-xs font-semibold text-[#625e58]">共同约束</span>
            <input
              className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
              onChange={(event) => setConstraints(event.target.value)}
              placeholder="守时、每周同步、四周交付，用逗号分隔"
              value={constraints}
            />
          </label>

          <div className="grid gap-4 sm:grid-cols-2">
            <label>
              <span className="text-xs font-semibold text-[#625e58]">时间倾向</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                maxLength={200}
                onChange={(event) => setTimeWindow(event.target.value)}
                placeholder="周四晚间"
                value={timeWindow}
              />
            </label>
            <label>
              <span className="text-xs font-semibold text-[#625e58]">地点倾向</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                maxLength={160}
                onChange={(event) => setLocation(event.target.value)}
                placeholder="FIT 楼、紫荆园..."
                value={location}
              />
            </label>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <label>
              <span className="text-xs font-semibold text-[#625e58]">最低成组</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                max={12}
                min={2}
                onChange={(event) => setMinMembers(Number(event.target.value))}
                type="number"
                value={minMembers}
              />
            </label>
            <label>
              <span className="text-xs font-semibold text-[#625e58]">目标人数</span>
              <input
                className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                max={12}
                min={2}
                onChange={(event) => setMaxMembers(Number(event.target.value))}
                type="number"
                value={maxMembers}
              />
            </label>
          </div>

          <div className="border-t border-[#e4e1dc] pt-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <label>
                <span className="text-xs font-semibold text-[#625e58]">确定开始时间</span>
                <input
                  className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                  onChange={(event) => setStartsAt(event.target.value)}
                  type="datetime-local"
                  value={startsAt}
                />
              </label>
              <label>
                <span className="text-xs font-semibold text-[#625e58]">确定结束时间</span>
                <input
                  className="mt-2 h-10 w-full rounded-md border border-[#d4d0c9] bg-white px-3 text-sm outline-none focus:border-[#527b6d]"
                  onChange={(event) => setEndsAt(event.target.value)}
                  type="datetime-local"
                  value={endsAt}
                />
              </label>
            </div>
            <label className="mt-3 flex min-h-10 items-center gap-3 text-sm text-[#5f5b55]">
              <input
                checked={createCalendar}
                className="h-4 w-4 accent-[#315c4d]"
                disabled={!startsAt || !endsAt}
                onChange={(event) => setCreateCalendar(event.target.checked)}
                type="checkbox"
              />
              成组后为所有成员创建忙碌日程
            </label>
          </div>

          {notice ? <p className="text-sm text-[#9b4b43]">{notice}</p> : null}
        </div>

        <footer className="flex shrink-0 justify-end gap-2 border-t border-[#dfdcd6] px-4 py-3 sm:px-5">
          <button
            className="h-9 rounded-md px-4 text-sm text-[#625e58] hover:bg-[#efede9]"
            onClick={onClose}
            type="button"
          >
            取消
          </button>
          <button
            className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-4 text-sm font-semibold text-white hover:bg-[#274f42] disabled:opacity-50"
            disabled={submitting}
            onClick={() => void handleSubmit()}
            type="button"
          >
            {submitting ? <LoaderCircle className="animate-spin" size={15} /> : <Plus size={15} />}
            发布组队
          </button>
        </footer>
      </section>
    </div>
  )
}

interface RecommendationListProps {
  busyId: number | null
  items: GroupMatchRecommendation[]
  onInvite: (candidateIntentId: number) => void
}

function RecommendationList({ busyId, items, onInvite }: RecommendationListProps) {
  if (items.length === 0) {
    return (
      <div className="border-y border-dashed border-[#d8d4cd] py-8 text-center">
        <p className="text-sm font-medium text-[#5f5b55]">本轮没有新的合适候选</p>
        <p className="mt-1 text-xs text-[#8a857e]">新的同类组队意图出现后会进入下一轮</p>
      </div>
    )
  }
  return (
    <div className="divide-y divide-[#e4e1dc] border-y border-[#e4e1dc]">
      {items.map((item) => (
        <article className="py-5" key={item.candidateIntentId}>
          <div className="flex items-start gap-3">
            <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#e8f1ed] text-[#315c4d]">
              <UsersRound size={18} />
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="text-sm font-semibold text-[#302e2b]">{item.alias}</h3>
                <span className="rounded bg-[#eee9f1] px-2 py-1 text-[11px] font-semibold text-[#674d75]">
                  {item.suggestedRole}
                </span>
                <span className="ml-auto font-mono text-sm font-semibold text-[#315c4d]">
                  {item.score}
                </span>
              </div>
              <p className="mt-2 text-sm leading-6 text-[#5f5b55]">{item.fitReason}</p>
              <p className="mt-1 text-xs leading-5 text-[#8a6b55]">待确认：{item.uncertainty}</p>
              {item.remainingRoleGaps.length > 0 ? (
                <p className="mt-2 text-xs text-[#77736c]">
                  加入后仍缺：{item.remainingRoleGaps.join('、')}
                </p>
              ) : (
                <p className="mt-2 inline-flex items-center gap-1 text-xs text-[#39705a]">
                  <Check size={13} />
                  加入后关键角色齐备
                </p>
              )}
              <div className="mt-3">
                <CommonTimes items={item.commonFreeTimes} />
              </div>
              <div className="mt-4 flex items-center justify-between gap-3">
                <span className="text-[11px] text-[#8a857e]">
                  {item.inference.source === 'model' ? '模型解释' : '本地规则'} ·{' '}
                  {Math.round(item.inference.confidence * 100)}%
                </span>
                <button
                  className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white hover:bg-[#274f42] disabled:opacity-50"
                  disabled={busyId === item.candidateIntentId}
                  onClick={() => onInvite(item.candidateIntentId)}
                  type="button"
                >
                  {busyId === item.candidateIntentId ? (
                    <LoaderCircle className="animate-spin" size={14} />
                  ) : (
                    <UserRoundCheck size={14} />
                  )}
                  发出邀请
                </button>
              </div>
            </div>
          </div>
        </article>
      ))}
    </div>
  )
}

export function GroupMatchPage() {
  const [mode, setMode] = useState<'mine' | 'inbox'>('mine')
  const [intents, setIntents] = useState<GroupMatchIntent[]>([])
  const [inbox, setInbox] = useState<GroupMatchInboxItem[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [selected, setSelected] = useState<GroupMatchIntent | null>(null)
  const [recommendations, setRecommendations] = useState<GroupMatchRecommendation[]>([])
  const [loading, setLoading] = useState(true)
  const [detailLoading, setDetailLoading] = useState(false)
  const [composerOpen, setComposerOpen] = useState(false)
  const [notice, setNotice] = useState('')
  const [busyKey, setBusyKey] = useState('')
  const [readyGroup, setReadyGroup] = useState<GroupMatchIntent | null>(null)

  const pendingCount = useMemo(
    () => inbox.filter((item) => item.invitation.status === 'invited').length,
    [inbox],
  )

  const loadOverview = useCallback(async () => {
    try {
      setLoading(true)
      const [nextIntents, nextInbox] = await Promise.all([
        fetchMyGroupMatchIntents(),
        fetchMyGroupMatchInvitations(),
      ])
      setIntents(nextIntents)
      setInbox(nextInbox)
      setSelectedId((current) =>
        current && nextIntents.some((item) => item.id === current)
          ? current
          : (nextIntents[0]?.id ?? null),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '组队工作台暂时无法加载'))
    } finally {
      setLoading(false)
    }
  }, [])

  const loadDetail = useCallback(async (intentId: number) => {
    try {
      setDetailLoading(true)
      const detail = await fetchGroupMatchIntent(intentId)
      setSelected(detail)
      setRecommendations(
        detail.status === 'open'
          ? await fetchGroupMatchRecommendations(intentId)
          : [],
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '组队详情暂时无法加载'))
      setSelected(null)
      setRecommendations([])
    } finally {
      setDetailLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadOverview()
  }, [loadOverview])

  useEffect(() => {
    if (mode === 'mine' && selectedId) {
      void loadDetail(selectedId)
    }
  }, [loadDetail, mode, selectedId])

  const refresh = async () => {
    setNotice('')
    await loadOverview()
    if (mode === 'mine' && selectedId) await loadDetail(selectedId)
  }

  const handleInvite = async (candidateIntentId: number) => {
    if (!selected) return
    try {
      setBusyKey(`invite-${candidateIntentId}`)
      setNotice('')
      await inviteGroupMatchCandidate(selected.id, candidateIntentId)
      setNotice('邀请已发出，对方同意前不会展示真实身份。')
      await Promise.all([loadOverview(), loadDetail(selected.id)])
    } catch (error) {
      setNotice(getErrorMessage(error, '邀请发送失败'))
    } finally {
      setBusyKey('')
    }
  }

  const handleRespond = async (invitationId: number, status: 'accepted' | 'declined') => {
    try {
      setBusyKey(`response-${invitationId}-${status}`)
      setNotice('')
      const result = await respondGroupMatchInvitation(invitationId, status)
      if (result.status === 'formed') setReadyGroup(result)
      setNotice(status === 'accepted' ? '已接受邀请，系统会在条件满足时自动建群。' : '已拒绝邀请。')
      await loadOverview()
    } catch (error) {
      setNotice(getErrorMessage(error, '邀请处理失败'))
    } finally {
      setBusyKey('')
    }
  }

  const handleForm = async () => {
    if (!selected) return
    try {
      setBusyKey('form')
      const result = await formGroupMatchIntent(selected.id)
      setSelected(result)
      setReadyGroup(result)
      setNotice('小组已成立，成员身份与群聊现已开放。')
      await loadOverview()
    } catch (error) {
      setNotice(getErrorMessage(error, '暂时无法成组'))
    } finally {
      setBusyKey('')
    }
  }

  const handleLeave = async () => {
    if (!selected) return
    try {
      setBusyKey('leave')
      await leaveGroupMatchIntent(selected.id)
      setNotice(selected.status === 'open' ? '组队意图已取消。' : '已退出这个组队。')
      await loadOverview()
    } catch (error) {
      setNotice(getErrorMessage(error, '暂时无法退出'))
    } finally {
      setBusyKey('')
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1180px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex flex-col gap-4 border-b border-[#dfdcd6] pb-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-medium text-[#315c4d]">GROUP MATCH</p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">多人组队</h1>
          <p className="mt-1 text-sm text-[#77736c]">围绕一件具体的事，找到真正能互补的一小组人。</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            aria-label="刷新组队"
            className="inline-flex h-10 w-10 items-center justify-center rounded-md border border-[#d7d3cc] text-[#625e58] hover:bg-white"
            onClick={() => void refresh()}
            title="刷新"
            type="button"
          >
            <RefreshCw className={cn(loading && 'animate-spin')} size={16} />
          </button>
          <button
            className="inline-flex h-10 items-center gap-2 rounded-md bg-[#315c4d] px-4 text-sm font-semibold text-white hover:bg-[#274f42]"
            onClick={() => setComposerOpen(true)}
            type="button"
          >
            <Plus size={16} />
            发起组队
          </button>
        </div>
      </header>

      <div className="mt-5 flex items-center gap-1 border-b border-[#dfdcd6]">
        <button
          className={cn(
            'relative h-11 px-4 text-sm font-medium',
            mode === 'mine' ? 'text-[#315c4d]' : 'text-[#77736c]',
          )}
          onClick={() => setMode('mine')}
          type="button"
        >
          我的组队
          {mode === 'mine' ? <span className="absolute inset-x-3 bottom-0 h-0.5 bg-[#315c4d]" /> : null}
        </button>
        <button
          className={cn(
            'relative h-11 px-4 text-sm font-medium',
            mode === 'inbox' ? 'text-[#315c4d]' : 'text-[#77736c]',
          )}
          onClick={() => setMode('inbox')}
          type="button"
        >
          待我确认
          {pendingCount > 0 ? (
            <span className="ml-2 inline-flex h-5 min-w-5 items-center justify-center rounded-full bg-[#315c4d] px-1 text-[10px] text-white">
              {pendingCount}
            </span>
          ) : null}
          {mode === 'inbox' ? <span className="absolute inset-x-3 bottom-0 h-0.5 bg-[#315c4d]" /> : null}
        </button>
      </div>

      {notice ? (
        <div className="mt-4 flex items-start gap-2 border-l-2 border-[#789b8d] bg-[#edf3ef] px-4 py-3 text-sm text-[#365d4e]">
          <ShieldCheck className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      {readyGroup?.chatGroupId ? (
        <div className="mt-4 flex flex-wrap items-center gap-3 border-y border-[#cdd8d3] bg-[#f0f5f2] px-4 py-4">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#315c4d] text-white">
            <UsersRound size={17} />
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-semibold text-[#2e4f43]">{readyGroup.title} 已成组</p>
            <p className="mt-0.5 text-xs text-[#688076]">{readyGroup.acceptedMembers} 位成员已完成同意</p>
          </div>
          <Link
            className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white"
            to={`/chat/group/${readyGroup.chatGroupId}`}
          >
            <MessageCircleMore size={14} />
            进入群聊
          </Link>
        </div>
      ) : null}

      {mode === 'mine' ? (
        <div className="mt-6 grid min-h-[560px] gap-6 lg:grid-cols-[280px_minmax(0,1fr)]">
          <aside className="border-r-0 border-[#dfdcd6] lg:border-r lg:pr-5">
            {loading ? (
              <div className="flex min-h-40 items-center justify-center text-sm text-[#77736c]">
                <LoaderCircle className="mr-2 animate-spin" size={16} />
                正在同步
              </div>
            ) : intents.length > 0 ? (
              <div className="space-y-1">
                {intents.map((intent) => (
                  <button
                    className={cn(
                      'w-full rounded-md px-3 py-3 text-left transition-colors',
                      selectedId === intent.id
                        ? 'bg-[#e8f1ed] text-[#294f41]'
                        : 'text-[#5f5b55] hover:bg-[#efede9]',
                    )}
                    key={intent.id}
                    onClick={() => setSelectedId(intent.id)}
                    type="button"
                  >
                    <span className="flex items-center justify-between gap-3">
                      <span className="text-[11px] font-semibold">
                        {categoryLabel[intent.category]} · {statusLabels[intent.status]}
                      </span>
                      <span className="text-[10px] text-[#8a857e]">{intent.acceptedMembers}/{intent.maxMembers}</span>
                    </span>
                    <span className="mt-1.5 block line-clamp-2 text-sm font-semibold leading-5">
                      {intent.title}
                    </span>
                    <span className="mt-1 block text-[11px] text-[#817d76]">
                      {formatDeadline(intent.expiresAt)}
                    </span>
                  </button>
                ))}
              </div>
            ) : (
              <div className="py-12 text-center">
                <UsersRound className="mx-auto text-[#a19c94]" size={24} />
                <p className="mt-3 text-sm text-[#625e58]">还没有组队意图</p>
                <button
                  className="mt-4 inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white"
                  onClick={() => setComposerOpen(true)}
                  type="button"
                >
                  <Plus size={14} />
                  发起第一组
                </button>
              </div>
            )}
          </aside>

          <main className="min-w-0">
            {detailLoading ? (
              <div className="flex min-h-80 items-center justify-center text-sm text-[#77736c]">
                <LoaderCircle className="mr-2 animate-spin" size={17} />
                正在计算组队状态
              </div>
            ) : selected ? (
              <div>
                <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="rounded bg-[#e8f1ed] px-2 py-1 text-[11px] font-semibold text-[#315c4d]">
                        {statusLabels[selected.status]}
                      </span>
                      <span className="text-xs text-[#817d76]">{categoryLabel[selected.category]}</span>
                      <InferenceLabel intent={selected} />
                    </div>
                    <h2 className="mt-3 text-xl font-semibold leading-7 text-[#292724]">{selected.title}</h2>
                    {selected.description ? (
                      <p className="mt-2 max-w-3xl text-sm leading-6 text-[#67635d]">{selected.description}</p>
                    ) : null}
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    {selected.canForm ? (
                      <button
                        className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white disabled:opacity-50"
                        disabled={busyKey === 'form'}
                        onClick={() => void handleForm()}
                        type="button"
                      >
                        {busyKey === 'form' ? <LoaderCircle className="animate-spin" size={14} /> : <UsersRound size={14} />}
                        立即成组
                      </button>
                    ) : null}
                    {selected.canLeave ? (
                      <button
                        className="h-9 rounded-md border border-[#d7d3cc] px-3 text-xs font-medium text-[#765a50] hover:bg-[#f6ece8] disabled:opacity-50"
                        disabled={busyKey === 'leave'}
                        onClick={() => void handleLeave()}
                        type="button"
                      >
                        {selected.status === 'open' ? '取消组队' : '退出'}
                      </button>
                    ) : null}
                  </div>
                </div>

                <div className="mt-6 grid grid-cols-3 divide-x divide-[#dfdcd6] border-y border-[#dfdcd6] py-4 text-center">
                  <div>
                    <strong className="block text-xl text-[#315c4d]">{selected.acceptedMembers}</strong>
                    <span className="mt-1 block text-[11px] text-[#817d76]">已同意</span>
                  </div>
                  <div>
                    <strong className="block text-xl text-[#5f5b55]">{selected.pendingInvitations}</strong>
                    <span className="mt-1 block text-[11px] text-[#817d76]">待回应</span>
                  </div>
                  <div>
                    <strong className="block text-xl text-[#8a6b55]">{selected.waitlistedMembers}</strong>
                    <span className="mt-1 block text-[11px] text-[#817d76]">候补</span>
                  </div>
                </div>

                <section className="grid gap-5 border-b border-[#e4e1dc] py-5 sm:grid-cols-2">
                  <div>
                    <h3 className="text-xs font-semibold text-[#5f5b55]">角色拼图</h3>
                    <div className="mt-3 flex flex-wrap gap-2">
                      <span className="rounded bg-[#e8f1ed] px-2.5 py-1.5 text-xs text-[#315c4d]">
                        我：{selected.roleOffer}
                      </span>
                      {selected.roleGaps.map((role) => (
                        <span className="rounded bg-[#f3ece5] px-2.5 py-1.5 text-xs text-[#805e43]" key={role}>
                          仍缺：{role}
                        </span>
                      ))}
                      {selected.roleGaps.length === 0 ? (
                        <span className="inline-flex items-center gap-1 text-xs text-[#39705a]">
                          <Check size={13} />
                          关键角色已齐
                        </span>
                      ) : null}
                    </div>
                  </div>
                  <div>
                    <h3 className="text-xs font-semibold text-[#5f5b55]">时间与地点</h3>
                    <div className="mt-3 space-y-2 text-xs text-[#706c65]">
                      <p className="flex items-center gap-2"><Clock3 size={13} />{selected.timeWindow || '时间待确认'}</p>
                      <p className="flex items-center gap-2"><MapPin size={13} />{selected.location || '地点待确认'}</p>
                      {selected.startsAt ? <p className="flex items-center gap-2"><CalendarDays size={13} />{formatSchedule(selected.startsAt)}</p> : null}
                    </div>
                  </div>
                </section>

                {selected.status === 'formed' ? (
                  <section className="py-6">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <h3 className="text-base font-semibold text-[#302e2b]">小组已经就绪</h3>
                      <div className="flex gap-2">
                        {selected.calendarEventId ? (
                          <Link className="inline-flex h-9 items-center gap-2 rounded-md border border-[#cfd7d3] px-3 text-xs font-semibold text-[#315c4d]" to="/calendar">
                            <CalendarDays size={14} />
                            查看日程
                          </Link>
                        ) : null}
                        {selected.chatGroupId ? (
                          <Link className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white" to={`/chat/group/${selected.chatGroupId}`}>
                            <MessageCircleMore size={14} />
                            进入群聊
                          </Link>
                        ) : null}
                      </div>
                    </div>
                    <p className="mt-4 border-l-2 border-[#789b8d] pl-4 text-sm leading-7 text-[#5f5b55]">{selected.groupIntro}</p>
                    <div className="mt-5">
                      <h4 className="text-xs font-semibold text-[#5f5b55]">开场建议</h4>
                      <ol className="mt-3 space-y-2 text-sm leading-6 text-[#67635d]">
                        {selected.conversationStarters.map((starter, index) => (
                          <li className="flex gap-3" key={starter}>
                            <span className="font-mono text-xs text-[#789b8d]">0{index + 1}</span>
                            <span>{starter}</span>
                          </li>
                        ))}
                      </ol>
                    </div>
                  </section>
                ) : (
                  <>
                    <section className="py-6">
                      <div className="flex items-end justify-between gap-4">
                        <div>
                          <h3 className="text-base font-semibold text-[#302e2b]">补位候选</h3>
                          <p className="mt-1 text-xs text-[#817d76]">仅展示有限、可解释且通过安全边界的候选</p>
                        </div>
                        <span className="text-xs text-[#817d76]">{recommendations.length} 位</span>
                      </div>
                      <div className="mt-4">
                        <RecommendationList
                          busyId={busyKey.startsWith('invite-') ? Number(busyKey.slice(7)) : null}
                          items={recommendations}
                          onInvite={(candidateIntentId) => void handleInvite(candidateIntentId)}
                        />
                      </div>
                    </section>
                  </>
                )}

                {selected.invitations.length > 0 ? (
                  <section className="border-t border-[#dfdcd6] py-6">
                    <h3 className="text-sm font-semibold text-[#302e2b]">邀请状态</h3>
                    <div className="mt-3 divide-y divide-[#e4e1dc]">
                      {selected.invitations.map((invitation) => (
                        <div className="flex items-center gap-3 py-3" key={invitation.id}>
                          <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-[#efede9] text-[#625e58]">
                            <UsersRound size={15} />
                          </span>
                          <span className="min-w-0 flex-1">
                            <span className="block truncate text-sm font-medium text-[#45413d]">{invitation.candidate.nickname}</span>
                            <span className="mt-0.5 block text-[11px] text-[#817d76]">{invitation.suggestedRole}</span>
                          </span>
                          <span className="text-xs text-[#706c65]">{invitationStatusLabels[invitation.status]}</span>
                        </div>
                      ))}
                    </div>
                  </section>
                ) : null}
              </div>
            ) : (
              <div className="flex min-h-80 items-center justify-center text-sm text-[#817d76]">选择一个组队查看进展</div>
            )}
          </main>
        </div>
      ) : (
        <main className="mx-auto mt-6 max-w-3xl">
          {loading ? (
            <div className="flex min-h-64 items-center justify-center text-sm text-[#77736c]">
              <LoaderCircle className="mr-2 animate-spin" size={16} />
              正在读取邀请
            </div>
          ) : inbox.length > 0 ? (
            <div className="divide-y divide-[#dfdcd6] border-y border-[#dfdcd6]">
              {inbox.map((item) => (
                <article className="py-5" key={item.invitation.id}>
                  <div className="flex items-start gap-3">
                    <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#e8f1ed] text-[#315c4d]">
                      <UsersRound size={18} />
                    </span>
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="rounded bg-[#efede9] px-2 py-1 text-[11px] font-semibold text-[#625e58]">{categoryLabel[item.category]}</span>
                        <span className="text-xs text-[#817d76]">{item.organizer.nickname} 发起</span>
                        <span className="ml-auto text-xs font-medium text-[#315c4d]">{invitationStatusLabels[item.invitation.status]}</span>
                      </div>
                      <h2 className="mt-2 text-base font-semibold leading-6 text-[#302e2b]">{item.title}</h2>
                      {item.description ? <p className="mt-1 text-sm leading-6 text-[#67635d]">{item.description}</p> : null}
                      <div className="mt-3 flex flex-wrap gap-x-4 gap-y-2 text-xs text-[#706c65]">
                        <span className="inline-flex items-center gap-1.5"><UserRoundCheck size={13} />邀请你补位 {item.invitation.suggestedRole}</span>
                        <span className="inline-flex items-center gap-1.5"><Clock3 size={13} />{item.timeWindow || '时间待确认'}</span>
                        <span className="inline-flex items-center gap-1.5"><MapPin size={13} />{item.location || '地点待确认'}</span>
                      </div>
                      <div className="mt-4 border-l-2 border-[#d1d9d5] pl-3">
                        <p className="text-sm leading-6 text-[#5f5b55]">{item.invitation.fitReason}</p>
                        <p className="mt-1 text-xs text-[#8a6b55]">待确认：{item.invitation.uncertainty}</p>
                      </div>
                      <div className="mt-4"><CommonTimes items={item.invitation.commonFreeTimes} /></div>
                      {item.invitation.status === 'invited' ? (
                        <div className="mt-5 flex justify-end gap-2">
                          <button
                            className="h-9 rounded-md border border-[#d7d3cc] px-4 text-xs font-semibold text-[#6e625b] hover:bg-[#f4efeb] disabled:opacity-50"
                            disabled={busyKey.startsWith(`response-${item.invitation.id}`)}
                            onClick={() => void handleRespond(item.invitation.id, 'declined')}
                            type="button"
                          >
                            拒绝
                          </button>
                          <button
                            className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-4 text-xs font-semibold text-white hover:bg-[#274f42] disabled:opacity-50"
                            disabled={busyKey.startsWith(`response-${item.invitation.id}`)}
                            onClick={() => void handleRespond(item.invitation.id, 'accepted')}
                            type="button"
                          >
                            {busyKey.startsWith(`response-${item.invitation.id}`) ? <LoaderCircle className="animate-spin" size={14} /> : <Check size={14} />}
                            接受邀请
                          </button>
                        </div>
                      ) : null}
                    </div>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <div className="flex min-h-64 flex-col items-center justify-center border-y border-dashed border-[#d8d4cd] text-center">
              <ShieldCheck className="text-[#8ca095]" size={24} />
              <p className="mt-3 text-sm font-medium text-[#5f5b55]">目前没有待确认邀请</p>
              <p className="mt-1 text-xs text-[#8a857e]">只有你主动发布过的同类意图才会进入候选池</p>
            </div>
          )}
        </main>
      )}

      <footer className="mt-8 flex flex-wrap items-center justify-between gap-3 border-t border-[#dfdcd6] pt-4 text-xs text-[#817d76]">
        <span className="inline-flex items-center gap-1.5"><ShieldCheck size={13} />未同意成员之间不展示私密资料</span>
        <Link className="inline-flex items-center gap-1 font-medium text-[#315c4d]" to={connectionPaths.intents}>
          返回我的寻找
          <ArrowRight size={13} />
        </Link>
      </footer>

      {composerOpen ? (
        <GroupIntentComposer
          onClose={() => setComposerOpen(false)}
          onCreated={(intent) => {
            setComposerOpen(false)
            setMode('mine')
            setSelectedId(intent.id)
            setNotice('组队意图已发布，正在寻找互补候选。')
            void loadOverview()
          }}
        />
      ) : null}
    </div>
  )
}
