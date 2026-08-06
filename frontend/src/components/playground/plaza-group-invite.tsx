import {
  CalendarPlus,
  Check,
  Clock3,
  Eye,
  EyeOff,
  LoaderCircle,
  MapPin,
  MessageCircleMore,
  Sparkles,
  UserRoundCheck,
  Users,
  X,
} from 'lucide-react'
import { useEffect, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { Button, buttonVariants } from '@/components/ui/button'
import { getErrorMessage } from '@/lib/api'
import {
  createPlazaGroupInvite,
  fetchPlazaGroupInvite,
  formPlazaGroupInvite,
  keepPlazaGroupInvite,
  respondToPlazaGroupApplication,
  type PlazaGroupApplicationStatus,
  type PlazaGroupInvite,
  type PlazaGroupInviteSummary,
  type PlazaGroupInviteType,
  type PlazaGroupRevealStrategy,
} from '@/lib/playground-api'
import { cn } from '@/lib/utils'

const statusLabel: Record<PlazaGroupApplicationStatus, string> = {
  applied: '待洞主确认',
  accepted: '已同意',
  declined: '未入选',
  waitlisted: '候补中',
  left: '已离开',
}

const inviteStatusLabel: Record<PlazaGroupInviteSummary['status'], string> = {
  open: '申请中',
  formed: '临时群已成立',
  expired: '申请已截止',
  archived: '临时群已归档',
  kept: '群聊已保留',
  cancelled: '已取消',
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function ModalFrame({
  children,
  onClose,
  title,
}: {
  children: ReactNode
  onClose: () => void
  title: string
}) {
  return (
    <div
      aria-label={title}
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-end justify-center bg-black/35 sm:items-center sm:p-5"
      role="dialog"
    >
      <div className="flex max-h-[92dvh] w-full flex-col overflow-hidden rounded-t-lg border border-[#d5d7d4] bg-[#fafaf8] shadow-2xl sm:max-w-[640px] sm:rounded-lg">
        <header className="flex h-14 shrink-0 items-center border-b border-[#deded9] px-4 sm:px-5">
          <h2 className="min-w-0 flex-1 truncate text-base font-semibold text-[#262a27]">{title}</h2>
          <button
            aria-label="关闭"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded text-[#6f746f] hover:bg-[#e9ebe8]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </header>
        {children}
      </div>
    </div>
  )
}

export function PlazaGroupInviteStrip({
  busy,
  invite,
  onApply,
  onOpen,
}: {
  busy: boolean
  invite: PlazaGroupInviteSummary
  onApply: () => void
  onOpen: () => void
}) {
  const actionLabel = invite.isOwner
    ? `管理 ${invite.applicationCount} 个申请`
    : invite.viewerApplicationStatus
      ? statusLabel[invite.viewerApplicationStatus]
      : invite.canApply
        ? '申请加入'
        : '先回复再申请'
  const canView = invite.isOwner || Boolean(invite.viewerApplicationStatus)

  return (
    <section className="border-t border-[#d9ddd9] bg-[#eef3f0] px-4 py-3">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            {invite.inviteType === 'event' ? (
              <CalendarPlus className="h-4 w-4 text-[#41695a]" />
            ) : (
              <Users className="h-4 w-4 text-[#41695a]" />
            )}
            <strong className="min-w-0 truncate text-sm text-[#29352f]">{invite.title}</strong>
            <span className="rounded bg-white/75 px-2 py-0.5 text-[11px] font-medium text-[#557064]">
              {inviteStatusLabel[invite.status]}
            </span>
          </div>
          <div className="mt-1.5 flex flex-wrap gap-x-4 gap-y-1 text-xs text-[#68716b]">
            <span>{invite.memberCount}/{invite.maxMembers} 人</span>
            <span>截止 {formatDateTime(invite.deadline)}</span>
            {invite.extractedTime ? <span>{invite.extractedTime}</span> : null}
            {invite.extractedLocation ? <span>{invite.extractedLocation}</span> : null}
          </div>
        </div>
        {invite.status === 'open' ? (
          <Button
            className="w-full shrink-0 sm:w-auto"
            disabled={busy || (!invite.isOwner && !invite.canApply && !invite.viewerApplicationStatus)}
            onClick={invite.canApply ? onApply : onOpen}
            size="sm"
            variant={invite.canApply ? 'default' : 'secondary'}
          >
            {busy ? <LoaderCircle className="mr-2 h-4 w-4 animate-spin" /> : <UserRoundCheck className="mr-2 h-4 w-4" />}
            {actionLabel}
          </Button>
        ) : (
          <Button
            className="w-full shrink-0 sm:w-auto"
            disabled={!canView}
            onClick={onOpen}
            size="sm"
            variant="secondary"
          >
            <Eye className="mr-2 h-4 w-4" />
            {canView ? '查看进度' : inviteStatusLabel[invite.status]}
          </Button>
        )}
      </div>
    </section>
  )
}

export function CreatePlazaGroupInviteDialog({
  onClose,
  onCreated,
  postId,
}: {
  onClose: () => void
  onCreated: (invite: PlazaGroupInvite) => void
  postId: number
}) {
  const [inviteType, setInviteType] = useState<PlazaGroupInviteType>('event')
  const [title, setTitle] = useState('')
  const [summary, setSummary] = useState('')
  const [minMembers, setMinMembers] = useState(2)
  const [maxMembers, setMaxMembers] = useState(6)
  const [revealStrategy, setRevealStrategy] = useState<PlazaGroupRevealStrategy>('on_form')
  const [deadline, setDeadline] = useState('')
  const [groupExpiresAt, setGroupExpiresAt] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async () => {
    try {
      setSubmitting(true)
      setError('')
      const invite = await createPlazaGroupInvite({
        postId,
        inviteType,
        title: title.trim(),
        summary: summary.trim(),
        minMembers,
        maxMembers,
        revealStrategy,
        deadline: deadline ? new Date(deadline).toISOString() : '',
        groupExpiresAt: groupExpiresAt ? new Date(groupExpiresAt).toISOString() : '',
      })
      onCreated(invite)
    } catch (requestError) {
      setError(getErrorMessage(requestError, '创建失败，请检查设置'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <ModalFrame onClose={onClose} title={`树洞 #${postId} · 创建连接`}>
      <div className="overflow-y-auto px-4 py-4 sm:px-5">
        <div className="grid grid-cols-2 rounded border border-[#ced3cf] bg-white p-1">
          {([
            ['event', '临时活动', CalendarPlus],
            ['group', '同频群组', Users],
          ] as const).map(([value, label, Icon]) => (
            <button
              className={cn(
                'inline-flex h-9 items-center justify-center gap-2 rounded text-sm',
                inviteType === value ? 'bg-[#315c4d] font-semibold text-white' : 'text-[#616761] hover:bg-[#eef0ed]',
              )}
              key={value}
              onClick={() => setInviteType(value)}
              type="button"
            >
              <Icon className="h-4 w-4" />
              {label}
            </button>
          ))}
        </div>

        <div className="mt-4 space-y-4">
          <label className="block text-sm font-medium text-[#3e433f]">
            标题
            <input
              className="mt-1.5 h-10 w-full rounded border border-[#c9cec9] bg-white px-3 font-normal outline-none focus:border-[#527465]"
              maxLength={100}
              onChange={(event) => setTitle(event.target.value)}
              placeholder="留空则从讨论生成"
              value={title}
            />
          </label>
          <label className="block text-sm font-medium text-[#3e433f]">
            说明
            <textarea
              className="mt-1.5 min-h-20 w-full resize-y rounded border border-[#c9cec9] bg-white px-3 py-2 font-normal leading-6 outline-none focus:border-[#527465]"
              maxLength={1000}
              onChange={(event) => setSummary(event.target.value)}
              placeholder="留空则整理当前讨论"
              value={summary}
            />
          </label>
          <div className="grid grid-cols-2 gap-3">
            <label className="text-sm font-medium text-[#3e433f]">
              最少人数
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#c9cec9] bg-white px-3 font-normal outline-none focus:border-[#527465]"
                max={maxMembers}
                min={2}
                onChange={(event) => setMinMembers(Number(event.target.value))}
                type="number"
                value={minMembers}
              />
            </label>
            <label className="text-sm font-medium text-[#3e433f]">
              人数上限
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#c9cec9] bg-white px-3 font-normal outline-none focus:border-[#527465]"
                max={12}
                min={minMembers}
                onChange={(event) => setMaxMembers(Number(event.target.value))}
                type="number"
                value={maxMembers}
              />
            </label>
          </div>
          <div className="grid gap-3 sm:grid-cols-2">
            <label className="text-sm font-medium text-[#3e433f]">
              申请截止（默认 3 天）
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#c9cec9] bg-white px-3 font-normal outline-none focus:border-[#527465]"
                onChange={(event) => setDeadline(event.target.value)}
                type="datetime-local"
                value={deadline}
              />
            </label>
            <label className="text-sm font-medium text-[#3e433f]">
              临时群到期（默认截止后 7 天）
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#c9cec9] bg-white px-3 font-normal outline-none focus:border-[#527465]"
                onChange={(event) => setGroupExpiresAt(event.target.value)}
                type="datetime-local"
                value={groupExpiresAt}
              />
            </label>
          </div>
          <fieldset>
            <legend className="text-sm font-medium text-[#3e433f]">身份揭示</legend>
            <div className="mt-1.5 grid gap-2 sm:grid-cols-2">
              {([
                ['on_form', '成组后揭示', EyeOff],
                ['on_accept', '接受后向洞主揭示', Eye],
              ] as const).map(([value, label, Icon]) => (
                <button
                  className={cn(
                    'flex h-10 items-center gap-2 rounded border px-3 text-left text-sm',
                    revealStrategy === value
                      ? 'border-[#527465] bg-[#e8efeb] text-[#315c4d]'
                      : 'border-[#cfd3cf] bg-white text-[#656a65]',
                  )}
                  key={value}
                  onClick={() => setRevealStrategy(value)}
                  type="button"
                >
                  <Icon className="h-4 w-4 shrink-0" />
                  {label}
                </button>
              ))}
            </div>
          </fieldset>
        </div>
        {error ? <p className="mt-4 rounded bg-[#f5e7e6] px-3 py-2 text-sm text-[#873f43]">{error}</p> : null}
      </div>
      <footer className="flex shrink-0 justify-end gap-2 border-t border-[#deded9] px-4 py-3 sm:px-5">
        <Button disabled={submitting} onClick={onClose} variant="secondary">取消</Button>
        <Button disabled={submitting} onClick={() => void handleSubmit()}>
          {submitting ? <LoaderCircle className="mr-2 h-4 w-4 animate-spin" /> : <Sparkles className="mr-2 h-4 w-4" />}
          创建邀请
        </Button>
      </footer>
    </ModalFrame>
  )
}

export function PlazaGroupInviteDialog({
  inviteId,
  onChanged,
  onClose,
}: {
  inviteId: number
  onChanged: () => void
  onClose: () => void
}) {
  const [invite, setInvite] = useState<PlazaGroupInvite | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionKey, setActionKey] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    let cancelled = false
    void fetchPlazaGroupInvite(inviteId)
      .then((result) => {
        if (!cancelled) setInvite(result)
      })
      .catch((requestError: unknown) => {
        if (!cancelled) setError(getErrorMessage(requestError, '邀请加载失败'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [inviteId])

  const runAction = async (key: string, action: () => Promise<PlazaGroupInvite>) => {
    try {
      setActionKey(key)
      setError('')
      setInvite(await action())
      onChanged()
    } catch (requestError) {
      setError(getErrorMessage(requestError, '操作失败，请稍后重试'))
    } finally {
      setActionKey('')
    }
  }

  return (
    <ModalFrame onClose={onClose} title={invite?.title ?? '连接邀请'}>
      <div className="overflow-y-auto px-4 py-4 sm:px-5">
        {loading ? (
          <div className="flex min-h-48 items-center justify-center text-sm text-[#707570]">
            <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />
            正在加载
          </div>
        ) : null}
        {error ? <p className="mb-4 rounded bg-[#f5e7e6] px-3 py-2 text-sm text-[#873f43]">{error}</p> : null}
        {invite ? (
          <>
            <div className="flex flex-wrap items-center gap-2 text-xs text-[#647069]">
              <span className="rounded bg-[#e5eee9] px-2 py-1 font-semibold text-[#3e6657]">
                {inviteStatusLabel[invite.status]}
              </span>
              <span>{invite.memberCount}/{invite.maxMembers} 人</span>
              <span>{invite.applicationCount} 个申请</span>
              <span className="inline-flex items-center gap-1">
                <Sparkles className="h-3.5 w-3.5" />
                {invite.inferenceSource === 'model' ? 'AI' : '规则'} {Math.round(invite.inferenceConfidence * 100)}%
              </span>
            </div>
            <p className="mt-3 whitespace-pre-wrap text-sm leading-6 text-[#444945]">{invite.summary}</p>
            <div className="mt-3 flex flex-wrap gap-2 text-xs text-[#626a64]">
              {invite.extractedTime ? (
                <span className="inline-flex items-center gap-1 rounded bg-[#eeefec] px-2 py-1">
                  <Clock3 className="h-3.5 w-3.5" />{invite.extractedTime}
                </span>
              ) : null}
              {invite.extractedLocation ? (
                <span className="inline-flex items-center gap-1 rounded bg-[#eeefec] px-2 py-1">
                  <MapPin className="h-3.5 w-3.5" />{invite.extractedLocation}
                </span>
              ) : null}
              <span className="inline-flex items-center gap-1 rounded bg-[#eeefec] px-2 py-1">
                <Clock3 className="h-3.5 w-3.5" />截止 {formatDateTime(invite.deadline)}
              </span>
            </div>

            <section className="mt-5 border-t border-[#d9ddd9] pt-4">
              <div className="flex items-center justify-between gap-3">
                <h3 className="text-sm font-semibold text-[#303531]">成员与申请</h3>
                <span className="text-xs text-[#777d78]">
                  {invite.revealStrategy === 'on_form' ? '成组后揭示身份' : '接受后向洞主揭示'}
                </span>
              </div>
              <div className="mt-3 flex items-center gap-3 rounded bg-[#ecefeb] px-3 py-2.5">
                <span className="inline-flex h-8 w-8 items-center justify-center rounded bg-[#303633] text-xs font-semibold text-white">主</span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-[#343936]">
                    {invite.owner.revealed ? invite.owner.nickname : invite.owner.alias}
                  </p>
                  {invite.owner.revealed ? (
                    <p className="truncate text-xs text-[#747a75]">{[invite.owner.department, invite.owner.grade].filter(Boolean).join(' · ')}</p>
                  ) : null}
                </div>
                <span className="text-xs text-[#737973]">洞主</span>
              </div>
              <div className="divide-y divide-[#e0e2df]">
                {invite.applications.map((application) => (
                  <div className="flex flex-col gap-2 py-3 sm:flex-row sm:items-center" key={application.id}>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-[#343936]">
                        {application.identity.revealed ? application.identity.nickname : application.identity.alias}
                      </p>
                      <p className="mt-0.5 truncate text-xs text-[#777d78]">
                        {application.identity.revealed
                          ? [application.identity.department, application.identity.grade].filter(Boolean).join(' · ')
                          : statusLabel[application.status]}
                      </p>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                      <span className="text-xs text-[#6d746e]">{statusLabel[application.status]}</span>
                      {invite.isOwner && application.status === 'applied' ? (
                        <>
                          <button
                            aria-label="拒绝申请"
                            className="inline-flex h-8 w-8 items-center justify-center rounded border border-[#d2d4d1] bg-white text-[#7b5555] hover:bg-[#f5eaea]"
                            disabled={Boolean(actionKey)}
                            onClick={() => void runAction(
                              `decline-${application.id}`,
                              () => respondToPlazaGroupApplication(invite.id, application.id, 'declined'),
                            )}
                            title="拒绝"
                            type="button"
                          >
                            <X className="h-4 w-4" />
                          </button>
                          <button
                            aria-label="接受申请"
                            className="inline-flex h-8 w-8 items-center justify-center rounded bg-[#315c4d] text-white hover:bg-[#264b3f]"
                            disabled={Boolean(actionKey)}
                            onClick={() => void runAction(
                              `accept-${application.id}`,
                              () => respondToPlazaGroupApplication(invite.id, application.id, 'accepted'),
                            )}
                            title="接受"
                            type="button"
                          >
                            <Check className="h-4 w-4" />
                          </button>
                        </>
                      ) : null}
                    </div>
                  </div>
                ))}
                {invite.applications.length === 0 ? (
                  <p className="py-5 text-center text-sm text-[#858a85]">暂时没有申请</p>
                ) : null}
              </div>
            </section>

            {invite.status === 'formed' || invite.status === 'kept' ? (
              <section className="mt-4 border-t border-[#d9ddd9] pt-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
                  <div className="min-w-0 flex-1">
                    <h3 className="text-sm font-semibold text-[#303531]">
                      {invite.status === 'kept' ? '群聊已由全员保留' : '临时群续留'}
                    </h3>
                    <p className="mt-1 text-xs text-[#747a75]">
                      {invite.keepVotes}/{invite.requiredKeepVotes} 位成员选择保留 · 到期 {formatDateTime(invite.groupExpiresAt)}
                    </p>
                  </div>
                  {invite.canKeep ? (
                    <Button
                      disabled={Boolean(actionKey)}
                      onClick={() => void runAction('keep', () => keepPlazaGroupInvite(invite.id))}
                      size="sm"
                      variant="secondary"
                    >
                      <Check className="mr-2 h-4 w-4" />保留群聊
                    </Button>
                  ) : null}
                </div>
              </section>
            ) : null}
          </>
        ) : null}
      </div>
      {invite ? (
        <footer className="flex shrink-0 flex-col-reverse gap-2 border-t border-[#deded9] px-4 py-3 sm:flex-row sm:justify-end sm:px-5">
          <Button onClick={onClose} variant="secondary">关闭</Button>
          {invite.canForm ? (
            <Button
              disabled={Boolean(actionKey)}
              onClick={() => void runAction('form', () => formPlazaGroupInvite(invite.id))}
            >
              <Users className="mr-2 h-4 w-4" />确认成组
            </Button>
          ) : null}
          {invite.chatGroupId ? (
            <Link className={buttonVariants()} to={`/chat/group/${invite.chatGroupId}`}>
              <MessageCircleMore className="mr-2 h-4 w-4" />进入群聊
            </Link>
          ) : null}
        </footer>
      ) : null}
    </ModalFrame>
  )
}
