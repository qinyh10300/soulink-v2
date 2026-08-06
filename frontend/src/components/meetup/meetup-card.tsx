import {
  AlertTriangle,
  CalendarCheck2,
  CalendarClock,
  Check,
  CheckCircle2,
  Clock3,
  LoaderCircle,
  MapPin,
  MessageSquareText,
  RefreshCw,
  RotateCcw,
  UserRoundCheck,
  UserRoundX,
  X,
} from 'lucide-react'
import { useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  cancelMeetup,
  completeMeetup,
  respondMeetup,
  revokeMeetupFeedback,
  submitMeetupFeedback,
  type FeedbackInput,
  type Meetup,
} from '@/lib/meetup-api'
import { cn } from '@/lib/utils'

interface Props {
  meetup: Meetup
  onReschedule?: (meetup: Meetup) => void
  onUpdated: (meetup: Meetup) => void
  compact?: boolean
}

type ConfirmAction = 'decline' | 'cancel' | 'completed' | 'no_show' | null

const statusLabels: Record<Meetup['status'], string> = {
  proposed: '待确认',
  reschedule_requested: '待确认改期',
  confirmed: '已确认',
  declined: '已拒绝',
  cancelled: '已取消',
  completed: '已完成',
  no_show: '未见面',
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    weekday: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

export function MeetupCard({ meetup, onReschedule, onUpdated, compact = false }: Props) {
  const [busyAction, setBusyAction] = useState('')
  const [confirmAction, setConfirmAction] = useState<ConfirmAction>(null)
  const [notice, setNotice] = useState('')
  const [feedbackOpen, setFeedbackOpen] = useState(false)
  const [feedback, setFeedback] = useState<FeedbackInput>({
    attendance: meetup.status === 'no_show' ? 'no_show' : 'attended',
    rating: null,
    wouldMeetAgain: null,
    safetyConcern: false,
    privateNotes: '',
  })

  const update = async (key: string, action: () => Promise<Meetup>) => {
    if (busyAction) return
    setBusyAction(key)
    setNotice('')
    try {
      const result = await action()
      onUpdated(result)
      setConfirmAction(null)
      setFeedbackOpen(false)
    } catch (error) {
      setNotice(getErrorMessage(error, '约见状态更新失败'))
    } finally {
      setBusyAction('')
    }
  }

  const runConfirmedAction = () => {
    switch (confirmAction) {
      case 'decline':
        return update('decline', () => respondMeetup(meetup.id, 'decline', meetup.version))
      case 'cancel':
        return update('cancel', () => cancelMeetup(meetup.id, meetup.version, '参与者取消'))
      case 'completed':
        return update('completed', () => completeMeetup(meetup.id, meetup.version, 'completed'))
      case 'no_show':
        return update('no_show', () => completeMeetup(meetup.id, meetup.version, 'no_show'))
      default:
        return Promise.resolve()
    }
  }

  const canLeaveFeedback =
    meetup.status === 'completed' ||
    meetup.status === 'no_show' ||
    meetup.status === 'cancelled' ||
    meetup.reminder.stage === 'feedback_due'

  return (
    <article className="rounded-lg border border-[#d9dcd8] bg-white p-3.5 sm:p-4">
      <div className="flex min-w-0 items-start gap-3">
        <span
          className={cn(
            'mt-0.5 inline-flex h-9 w-9 shrink-0 items-center justify-center rounded',
            meetup.status === 'confirmed' || meetup.status === 'completed'
              ? 'bg-[#e5efe9] text-[#315d4f]'
              : meetup.status === 'proposed' || meetup.status === 'reschedule_requested'
                ? 'bg-[#eee8f2] text-[#5b3477]'
                : 'bg-[#eeece8] text-[#6d6963]',
          )}
        >
          {meetup.status === 'confirmed' || meetup.status === 'completed' ? (
            <CalendarCheck2 className="h-4.5 w-4.5" />
          ) : (
            <CalendarClock className="h-4.5 w-4.5" />
          )}
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <h3 className="min-w-0 truncate text-sm font-semibold text-[#30332f]">
              {meetup.title}
            </h3>
            <span className="shrink-0 rounded bg-[#eef0ed] px-2 py-1 text-[10px] font-semibold text-[#626761]">
              {statusLabels[meetup.status]}
            </span>
          </div>
          <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1.5 text-xs text-[#6f746e]">
            <span className="inline-flex items-center gap-1.5">
              <Clock3 className="h-3.5 w-3.5" />
              {formatDateTime(meetup.startsAt)}
            </span>
            {meetup.location ? (
              <span className="inline-flex min-w-0 items-center gap-1.5">
                <MapPin className="h-3.5 w-3.5 shrink-0" />
                <span className="truncate">{meetup.location}</span>
              </span>
            ) : null}
          </div>
        </div>
      </div>

      {!compact ? (
        <div className="mt-3 flex flex-wrap gap-1.5">
          {meetup.participants.map((participant) => (
            <span
              className={cn(
                'inline-flex items-center gap-1 rounded px-2 py-1 text-[10px]',
                participant.status === 'accepted'
                  ? 'bg-[#edf5f0] text-[#37624f]'
                  : participant.status === 'declined'
                    ? 'bg-[#fff0f2] text-[#8b4559]'
                    : 'bg-[#f1f0ed] text-[#77736c]',
              )}
              key={participant.userId}
            >
              {participant.status === 'accepted' ? (
                <UserRoundCheck className="h-3 w-3" />
              ) : participant.status === 'declined' ? (
                <UserRoundX className="h-3 w-3" />
              ) : (
                <Clock3 className="h-3 w-3" />
              )}
              {participant.isMe ? '你' : participant.nickname}
            </span>
          ))}
        </div>
      ) : null}

      {meetup.reminder.text ? (
        <div
          className={cn(
            'mt-3 flex items-start gap-2 border-l-2 px-3 py-2 text-xs leading-5',
            meetup.reminder.stage === 'soon' || meetup.reminder.stage === 'underway'
              ? 'border-[#b37a3e] bg-[#fff8ee] text-[#77552f]'
              : 'border-[#7c9b89] bg-[#f2f6f3] text-[#52665b]',
          )}
        >
          <Clock3 className="mt-0.5 h-3.5 w-3.5 shrink-0" />
          {meetup.reminder.text}
        </div>
      ) : null}

      {notice ? (
        <div className="mt-3 rounded border border-[#e4cfc9] bg-[#fff6f3] px-3 py-2 text-xs text-[#8c463d]">
          {notice}
        </div>
      ) : null}

      {meetup.myFeedback ? (
        <div className="mt-3 border-t border-[#e1e3df] pt-3">
          <div className="flex items-center gap-2 text-xs text-[#52665b]">
            <CheckCircle2 className="h-4 w-4" />
            私密反馈已保存
            <button
              className="ml-auto inline-flex h-8 items-center gap-1 rounded px-2 text-[#6d716c] hover:bg-[#eef0ed] disabled:opacity-50"
              disabled={Boolean(busyAction)}
              onClick={() => void update('revoke-feedback', () => revokeMeetupFeedback(meetup.id))}
              type="button"
            >
              <RotateCcw className="h-3.5 w-3.5" />
              撤回
            </button>
          </div>
        </div>
      ) : null}

      {feedbackOpen && !meetup.myFeedback ? (
        <div className="mt-3 space-y-3 border-t border-[#e1e3df] pt-3">
          <div className="grid grid-cols-3 overflow-hidden rounded border border-[#d7dad6]">
            {(
              [
                ['attended', '顺利见面'],
                ['no_show', '对方未出现'],
                ['cancelled', '临时取消'],
              ] as const
            ).map(([value, label]) => (
              <button
                className={cn(
                  'min-h-9 border-r border-[#d7dad6] px-2 text-[11px] last:border-r-0',
                  feedback.attendance === value
                    ? 'bg-[#376f5b] font-semibold text-white'
                    : 'bg-white text-[#656a64] hover:bg-[#eef3f0]',
                )}
                key={value}
                onClick={() => setFeedback((current) => ({ ...current, attendance: value }))}
                type="button"
              >
                {label}
              </button>
            ))}
          </div>
          {feedback.attendance === 'attended' ? (
            <div className="grid gap-3 sm:grid-cols-2">
              <label className="text-xs text-[#646963]">
                体验评分
                <select
                  className="mt-1.5 h-9 w-full rounded border border-[#d7dad6] bg-white px-2 text-xs"
                  onChange={(event) =>
                    setFeedback((current) => ({
                      ...current,
                      rating: event.target.value ? Number(event.target.value) : null,
                    }))
                  }
                  value={feedback.rating ?? ''}
                >
                  <option value="">暂不评分</option>
                  {[5, 4, 3, 2, 1].map((rating) => (
                    <option key={rating} value={rating}>
                      {rating} 分
                    </option>
                  ))}
                </select>
              </label>
              <div>
                <p className="text-xs text-[#646963]">愿意再次联系</p>
                <div className="mt-1.5 grid grid-cols-2 overflow-hidden rounded border border-[#d7dad6]">
                  {[
                    [true, '愿意'],
                    [false, '不愿意'],
                  ].map(([value, label]) => (
                    <button
                      className={cn(
                        'h-9 border-r border-[#d7dad6] text-xs last:border-r-0',
                        feedback.wouldMeetAgain === value
                          ? 'bg-[#5b3477] font-semibold text-white'
                          : 'bg-white text-[#656a64] hover:bg-[#f1edf3]',
                      )}
                      key={String(value)}
                      onClick={() =>
                        setFeedback((current) => ({
                          ...current,
                          wouldMeetAgain: value as boolean,
                        }))
                      }
                      type="button"
                    >
                      {label as string}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          ) : null}
          <textarea
            className="min-h-20 w-full resize-y rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 py-2 text-xs leading-5 outline-none focus:border-[#5d7f6d]"
            maxLength={1000}
            onChange={(event) =>
              setFeedback((current) => ({ ...current, privateNotes: event.target.value }))
            }
            placeholder="仅你可见的备注"
            value={feedback.privateNotes}
          />
          <label className="flex items-center gap-2 text-xs text-[#73525a]">
            <input
              checked={feedback.safetyConcern}
              className="h-4 w-4 accent-[#8b4559]"
              onChange={(event) =>
                setFeedback((current) => ({
                  ...current,
                  safetyConcern: event.target.checked,
                }))
              }
              type="checkbox"
            />
            这次见面涉及安全问题
          </label>
          <div className="flex justify-end gap-2">
            <button
              className="h-8 rounded px-3 text-xs text-[#666b65] hover:bg-[#eef0ed]"
              onClick={() => setFeedbackOpen(false)}
              type="button"
            >
              取消
            </button>
            <button
              className="inline-flex h-8 items-center gap-1.5 rounded bg-[#376f5b] px-3 text-xs font-semibold text-white disabled:opacity-50"
              disabled={Boolean(busyAction)}
              onClick={() =>
                void update('feedback', () => submitMeetupFeedback(meetup.id, feedback))
              }
              type="button"
            >
              {busyAction === 'feedback' ? (
                <LoaderCircle className="h-3.5 w-3.5 animate-spin" />
              ) : (
                <Check className="h-3.5 w-3.5" />
              )}
              保存反馈
            </button>
          </div>
        </div>
      ) : null}

      {confirmAction ? (
        <div className="mt-3 flex flex-wrap items-center gap-2 border-t border-[#e7d7d2] pt-3">
          <AlertTriangle className="h-4 w-4 text-[#9a6155]" />
          <span className="min-w-0 flex-1 text-xs text-[#75524b]">
            {confirmAction === 'decline'
              ? '确认拒绝这项约见？'
              : confirmAction === 'cancel'
                ? '确认取消已经发出的约见？'
                : confirmAction === 'completed'
                  ? '确认这次见面已经完成？'
                  : '确认这次约见未能见面？'}
          </span>
          <button
            aria-label="返回"
            className="inline-flex h-8 w-8 items-center justify-center rounded text-[#666b65] hover:bg-[#eef0ed]"
            onClick={() => setConfirmAction(null)}
            title="返回"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
          <button
            className="inline-flex h-8 items-center gap-1.5 rounded bg-[#8b4559] px-3 text-xs font-semibold text-white disabled:opacity-50"
            disabled={Boolean(busyAction)}
            onClick={() => void runConfirmedAction()}
            type="button"
          >
            {busyAction ? <LoaderCircle className="h-3.5 w-3.5 animate-spin" /> : null}
            确认
          </button>
        </div>
      ) : (
        <div className="mt-3 flex flex-wrap items-center gap-2 border-t border-[#e1e3df] pt-3">
          {meetup.canRespond ? (
            <>
              <button
                className="inline-flex h-8 items-center gap-1.5 rounded bg-[#337a49] px-3 text-xs font-semibold text-white disabled:opacity-50"
                disabled={Boolean(busyAction)}
                onClick={() =>
                  void update('accept', () => respondMeetup(meetup.id, 'accept', meetup.version))
                }
                type="button"
              >
                {busyAction === 'accept' ? (
                  <LoaderCircle className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Check className="h-3.5 w-3.5" />
                )}
                接受
              </button>
              <button
                className="h-8 rounded px-3 text-xs text-[#8b4559] hover:bg-[#fff0f3]"
                onClick={() => setConfirmAction('decline')}
                type="button"
              >
                拒绝
              </button>
            </>
          ) : null}
          {meetup.canReschedule && onReschedule ? (
            <button
              className="inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs text-[#5e635d] hover:bg-[#eef0ed]"
              onClick={() => onReschedule(meetup)}
              type="button"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              改期
            </button>
          ) : null}
          {meetup.canComplete ? (
            <>
              <button
                className="inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs text-[#37624f] hover:bg-[#edf5f0]"
                onClick={() => setConfirmAction('completed')}
                type="button"
              >
                <CheckCircle2 className="h-3.5 w-3.5" />
                已见面
              </button>
              <button
                className="h-8 rounded px-2 text-xs text-[#7a5c45] hover:bg-[#fff7ed]"
                onClick={() => setConfirmAction('no_show')}
                type="button"
              >
                未见面
              </button>
            </>
          ) : null}
          {canLeaveFeedback && !meetup.myFeedback ? (
            <button
              className="inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs text-[#5b3477] hover:bg-[#f1edf3]"
              onClick={() => setFeedbackOpen(true)}
              type="button"
            >
              <MessageSquareText className="h-3.5 w-3.5" />
              私密反馈
            </button>
          ) : null}
          {meetup.canCancel ? (
            <button
              className="ml-auto h-8 rounded px-2 text-xs text-[#7b7067] hover:bg-[#f2efeb]"
              onClick={() => setConfirmAction('cancel')}
              type="button"
            >
              取消约见
            </button>
          ) : null}
        </div>
      )}
    </article>
  )
}
