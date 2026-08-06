import { CalendarClock, Check, MapPin, Trash2, X } from 'lucide-react'
import { addHours, format, parseISO } from 'date-fns'
import { useEffect, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  cancelCalendarEvent,
  createCalendarEvent,
  respondCalendarEvent,
  updateCalendarEvent,
  type CalendarEvent,
  type CalendarEventInput,
  type CalendarEventType,
  type CalendarVisibility,
} from '@/lib/calendar-api'
import type { FriendItem } from '@/lib/friend-api'
import { cn } from '@/lib/utils'

const eventTypes: Array<{
  value: Exclude<CalendarEventType, 'busy'>
  label: string
}> = [
  { value: 'meetup', label: '见面' },
  { value: 'study', label: '学习' },
  { value: 'sports', label: '运动' },
  { value: 'meal', label: '吃饭' },
  { value: 'date', label: '约会' },
  { value: 'project', label: '项目' },
  { value: 'other', label: '其他' },
]

const visibilityOptions: Array<{ value: CalendarVisibility; label: string }> = [
  { value: 'busy_only', label: '仅显示忙碌' },
  { value: 'public', label: '公开详情' },
  { value: 'private', label: '私人日程' },
]

function toLocalInput(date: Date) {
  return format(date, "yyyy-MM-dd'T'HH:mm")
}

interface Props {
  event?: CalendarEvent | null
  friends?: FriendItem[]
  initialStart?: Date
  lockedInvitee?: { userId: number; nickname: string }
  onClose: () => void
  onSaved: () => void
}

export function CalendarEventDialog({
  event,
  friends = [],
  initialStart,
  lockedInvitee,
  onClose,
  onSaved,
}: Props) {
  const defaultStart = initialStart ?? addHours(new Date(), 1)
  const [eventType, setEventType] = useState<Exclude<CalendarEventType, 'busy'>>('meetup')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [location, setLocation] = useState('')
  const [startsAt, setStartsAt] = useState(toLocalInput(defaultStart))
  const [endsAt, setEndsAt] = useState(toLocalInput(addHours(defaultStart, 1)))
  const [visibility, setVisibility] = useState<CalendarVisibility>('busy_only')
  const [invitees, setInvitees] = useState<Set<number>>(
    new Set(lockedInvitee ? [lockedInvitee.userId] : []),
  )
  const [submitting, setSubmitting] = useState(false)
  const [notice, setNotice] = useState('')

  const editable = !event || event.isOrganizer

  useEffect(() => {
    if (event) {
      if (event.eventType !== 'busy') setEventType(event.eventType)
      setTitle(event.title)
      setDescription(event.description)
      setLocation(event.location)
      setStartsAt(toLocalInput(parseISO(event.startsAt)))
      setEndsAt(toLocalInput(parseISO(event.endsAt)))
      setVisibility(event.visibility)
      setInvitees(new Set(event.attendees.map((attendee) => attendee.userId)))
      return
    }
    const start = initialStart ?? addHours(new Date(), 1)
    setEventType('meetup')
    setTitle(lockedInvitee ? `和 ${lockedInvitee.nickname} 见面` : '')
    setDescription('')
    setLocation('')
    setStartsAt(toLocalInput(start))
    setEndsAt(toLocalInput(addHours(start, 1)))
    setVisibility('busy_only')
    setInvitees(new Set(lockedInvitee ? [lockedInvitee.userId] : []))
  }, [event, initialStart, lockedInvitee])

  const handleSave = async () => {
    if (!title.trim()) {
      setNotice('请填写日程标题')
      return
    }
    const start = new Date(startsAt)
    const end = new Date(endsAt)
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || end <= start) {
      setNotice('结束时间需要晚于开始时间')
      return
    }
    const input: CalendarEventInput = {
      eventType,
      title: title.trim(),
      description: description.trim(),
      location: location.trim(),
      startsAt: start.toISOString(),
      endsAt: end.toISOString(),
      visibility,
      inviteeIds: [...invitees],
    }
    try {
      setSubmitting(true)
      if (event) await updateCalendarEvent(event.id, input)
      else await createCalendarEvent(input)
      onSaved()
    } catch (error) {
      setNotice(getErrorMessage(error, '日程保存失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleCancel = async () => {
    if (!event) return
    try {
      setSubmitting(true)
      await cancelCalendarEvent(event.id)
      onSaved()
    } catch (error) {
      setNotice(getErrorMessage(error, '日程取消失败'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleRSVP = async (status: 'accepted' | 'declined') => {
    if (!event) return
    try {
      setSubmitting(true)
      await respondCalendarEvent(event.id, status)
      onSaved()
    } catch (error) {
      setNotice(getErrorMessage(error, '邀请回复失败'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-60 flex items-center justify-center bg-black/45 p-4">
      <section
        aria-modal="true"
        className="flex max-h-[92dvh] w-full max-w-xl flex-col rounded-lg bg-white shadow-2xl"
        role="dialog"
      >
        <header className="flex h-14 shrink-0 items-center border-b border-[#e1e3e0] px-4">
          <CalendarClock className="mr-2 h-5 w-5 text-[#3f7651]" />
          <h2 className="font-semibold">{event ? '日程详情' : '新建日程'}</h2>
          <button
            aria-label="关闭"
            className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded hover:bg-[#eceeeb]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </header>

        <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-4 sm:p-5">
          <div className="flex flex-wrap gap-1.5">
            {eventTypes.map((type) => (
              <button
                className={cn(
                  'h-8 rounded px-3 text-xs font-medium',
                  eventType === type.value
                    ? 'bg-[#315d4f] text-white'
                    : 'bg-[#eef0ed] text-[#676b66] hover:bg-[#e2e5e1]',
                )}
                disabled={!editable}
                key={type.value}
                onClick={() => setEventType(type.value)}
                type="button"
              >
                {type.label}
              </button>
            ))}
          </div>

          <label className="block">
            <span className="mb-1.5 block text-xs font-medium text-[#686c67]">标题</span>
            <input
              className="h-10 w-full rounded border border-[#cfd2ce] px-3 text-sm outline-none focus:border-[#4d8960] disabled:bg-[#f1f2f0]"
              disabled={!editable}
              maxLength={120}
              onChange={(inputEvent) => setTitle(inputEvent.target.value)}
              value={title}
            />
          </label>

          <div className="grid gap-3 sm:grid-cols-2">
            <label>
              <span className="mb-1.5 block text-xs font-medium text-[#686c67]">开始</span>
              <input
                className="h-10 w-full rounded border border-[#cfd2ce] px-3 text-sm outline-none focus:border-[#4d8960] disabled:bg-[#f1f2f0]"
                disabled={!editable}
                onChange={(inputEvent) => setStartsAt(inputEvent.target.value)}
                type="datetime-local"
                value={startsAt}
              />
            </label>
            <label>
              <span className="mb-1.5 block text-xs font-medium text-[#686c67]">结束</span>
              <input
                className="h-10 w-full rounded border border-[#cfd2ce] px-3 text-sm outline-none focus:border-[#4d8960] disabled:bg-[#f1f2f0]"
                disabled={!editable}
                onChange={(inputEvent) => setEndsAt(inputEvent.target.value)}
                type="datetime-local"
                value={endsAt}
              />
            </label>
          </div>

          <label className="block">
            <span className="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-[#686c67]">
              <MapPin className="h-3.5 w-3.5" />
              地点
            </span>
            <input
              className="h-10 w-full rounded border border-[#cfd2ce] px-3 text-sm outline-none focus:border-[#4d8960] disabled:bg-[#f1f2f0]"
              disabled={!editable}
              maxLength={160}
              onChange={(inputEvent) => setLocation(inputEvent.target.value)}
              placeholder="例如：紫荆园一层"
              value={location}
            />
          </label>

          <label className="block">
            <span className="mb-1.5 block text-xs font-medium text-[#686c67]">备注</span>
            <textarea
              className="min-h-20 w-full resize-y rounded border border-[#cfd2ce] px-3 py-2 text-sm leading-6 outline-none focus:border-[#4d8960] disabled:bg-[#f1f2f0]"
              disabled={!editable}
              maxLength={2000}
              onChange={(inputEvent) => setDescription(inputEvent.target.value)}
              value={description}
            />
          </label>

          {editable ? (
            <label className="block">
              <span className="mb-1.5 block text-xs font-medium text-[#686c67]">个人主页可见范围</span>
              <select
                className="h-10 w-full rounded border border-[#cfd2ce] bg-white px-3 text-sm outline-none focus:border-[#4d8960]"
                onChange={(inputEvent) => setVisibility(inputEvent.target.value as CalendarVisibility)}
                value={visibility}
              >
                {visibilityOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          ) : null}

          {lockedInvitee ? (
            <div className="rounded border border-[#d8ddd9] bg-[#f3f6f4] px-3 py-2.5 text-sm text-[#4f5f56]">
              邀请：{lockedInvitee.nickname}
            </div>
          ) : !event && friends.length > 0 ? (
            <section>
              <div className="mb-2 flex items-center justify-between">
                <h3 className="text-xs font-medium text-[#686c67]">邀请好友</h3>
                <span className="text-xs text-[#8a8e88]">{invitees.size} 人</span>
              </div>
              <div className="max-h-44 overflow-y-auto rounded border border-[#e0e2df]">
                {friends.map((friend) => {
                  const checked = invitees.has(friend.user.userId)
                  return (
                    <button
                      className="flex w-full items-center gap-3 border-b border-[#eceeeb] px-3 py-2 text-left last:border-0 hover:bg-[#f5f6f4]"
                      key={friend.friendshipId}
                      onClick={() =>
                        setInvitees((current) => {
                          const next = new Set(current)
                          if (checked) next.delete(friend.user.userId)
                          else next.add(friend.user.userId)
                          return next
                        })
                      }
                      type="button"
                    >
                      <span className="flex h-8 w-8 items-center justify-center rounded bg-[#dce8e1] text-xs font-semibold text-[#315d4f]">
                        {friend.user.nickname.slice(0, 1)}
                      </span>
                      <span className="min-w-0 flex-1 truncate text-sm">{friend.user.nickname}</span>
                      <span
                        className={cn(
                          'inline-flex h-5 w-5 items-center justify-center rounded border',
                          checked ? 'border-[#3f8053] bg-[#3f8053] text-white' : 'border-[#bfc3be]',
                        )}
                      >
                        {checked ? <Check className="h-3.5 w-3.5" /> : null}
                      </span>
                    </button>
                  )
                })}
              </div>
            </section>
          ) : event?.attendees.length ? (
            <section>
              <h3 className="mb-2 text-xs font-medium text-[#686c67]">参与人</h3>
              <div className="divide-y divide-[#eceeeb] border-y border-[#e0e2df]">
                {event.attendees.map((attendee) => (
                  <div className="flex items-center gap-3 py-2" key={attendee.userId}>
                    <span className="min-w-0 flex-1 truncate text-sm">{attendee.nickname}</span>
                    <span className="text-xs text-[#858984]">
                      {attendee.status === 'accepted'
                        ? '已接受'
                        : attendee.status === 'declined'
                          ? '已拒绝'
                          : '待回复'}
                    </span>
                  </div>
                ))}
              </div>
            </section>
          ) : null}

          {event && !event.isOrganizer ? (
            <div className="rounded border border-[#d8ddd9] bg-[#f3f6f4] px-3 py-2.5 text-sm text-[#4f5f56]">
              发起人：{event.organizerNickname}
            </div>
          ) : null}
          {notice ? <p className="text-sm text-[#a04b4b]">{notice}</p> : null}
        </div>

        <footer className="flex shrink-0 flex-wrap items-center gap-2 border-t border-[#e1e3e0] px-4 py-3">
          {event?.isOrganizer ? (
            <button
              className="inline-flex h-9 items-center gap-1.5 rounded px-3 text-sm text-[#a04b4b] hover:bg-[#f5e8e8]"
              disabled={submitting}
              onClick={() => void handleCancel()}
              type="button"
            >
              <Trash2 className="h-4 w-4" />
              取消日程
            </button>
          ) : null}
          {event && !event.isOrganizer && event.attendeeStatus !== 'declined' ? (
            <div className="flex gap-2">
              <button
                className="h-9 rounded px-3 text-sm text-[#a04b4b] hover:bg-[#f5e8e8]"
                disabled={submitting}
                onClick={() => void handleRSVP('declined')}
                type="button"
              >
                拒绝
              </button>
              <button
                className="h-9 rounded bg-[#337a49] px-3 text-sm font-medium text-white hover:bg-[#28663c]"
                disabled={submitting}
                onClick={() => void handleRSVP('accepted')}
                type="button"
              >
                接受邀请
              </button>
            </div>
          ) : null}
          <div className="ml-auto flex gap-2">
            <button className="h-9 rounded px-4 text-sm hover:bg-[#eceeeb]" onClick={onClose} type="button">
              关闭
            </button>
            {editable ? (
              <button
                className="h-9 rounded bg-[#337a49] px-4 text-sm font-medium text-white hover:bg-[#28663c] disabled:bg-[#a9b9ad]"
                disabled={submitting}
                onClick={() => void handleSave()}
                type="button"
              >
                {submitting ? '保存中' : '保存'}
              </button>
            ) : null}
          </div>
        </footer>
      </section>
    </div>
  )
}
