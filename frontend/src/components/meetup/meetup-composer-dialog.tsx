import {
  CalendarClock,
  Check,
  Clock3,
  LoaderCircle,
  MapPin,
  RefreshCw,
  Users,
  X,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  createMeetup,
  rescheduleMeetup,
  suggestMeetupTimes,
  type CampusPlace,
  type Meetup,
  type MeetupContext,
  type MeetupSuggestions,
} from '@/lib/meetup-api'
import { cn } from '@/lib/utils'

interface Props {
  context: MeetupContext
  defaultCategory?: string
  defaultTitle?: string
  meetup?: Meetup
  onClose: () => void
  onOpenAvailability?: () => void
  onSaved: (meetup: Meetup) => void
}

function toDateTimeLocal(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

function initialStart() {
  const start = new Date()
  start.setMinutes(start.getMinutes() < 30 ? 30 : 60, 0, 0)
  return start.toISOString()
}

export function MeetupComposerDialog({
  context,
  defaultCategory = 'other',
  defaultTitle = '一起在校内见面',
  meetup,
  onClose,
  onOpenAvailability,
  onSaved,
}: Props) {
  const [title, setTitle] = useState(meetup?.title ?? defaultTitle)
  const [startsAt, setStartsAt] = useState(meetup?.startsAt ?? initialStart())
  const [durationMinutes, setDurationMinutes] = useState(() => {
    if (!meetup) return 60
    return Math.max(
      30,
      Math.round(
        (new Date(meetup.endsAt).getTime() - new Date(meetup.startsAt).getTime()) / 60_000,
      ),
    )
  })
  const [location, setLocation] = useState(meetup?.location ?? '')
  const [suggestions, setSuggestions] = useState<MeetupSuggestions | null>(null)
  const [loadingSuggestions, setLoadingSuggestions] = useState(true)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState('')

  const loadSuggestions = async () => {
    setLoadingSuggestions(true)
    setNotice('')
    const from = new Date()
    const to = new Date(from.getTime() + 14 * 24 * 60 * 60 * 1000)
    try {
      const result = await suggestMeetupTimes(context, {
        from: from.toISOString(),
        to: to.toISOString(),
        durationMinutes,
      })
      setSuggestions(result)
      if (!meetup && result.candidates[0]) {
        setStartsAt(result.candidates[0].startsAt)
      }
      if (!location && result.places[0]) {
        setLocation(result.places[0].name)
      }
    } catch (error) {
      setNotice(getErrorMessage(error, '共同空闲暂时无法计算'))
    } finally {
      setLoadingSuggestions(false)
    }
  }

  useEffect(() => {
    const timeout = window.setTimeout(() => void loadSuggestions(), 0)
    // The user explicitly refreshes after changing duration to avoid noisy requests.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    return () => window.clearTimeout(timeout)
  }, [context.id, context.type])

  const endsAt = useMemo(
    () => new Date(new Date(startsAt).getTime() + durationMinutes * 60_000).toISOString(),
    [durationMinutes, startsAt],
  )

  const save = async () => {
    if (saving || !startsAt) return
    setSaving(true)
    setNotice('')
    try {
      const result = meetup
        ? await rescheduleMeetup(meetup.id, {
            startsAt: new Date(startsAt).toISOString(),
            endsAt,
            location,
            expectedVersion: meetup.version,
          })
        : await createMeetup({
            context,
            title,
            category: defaultCategory,
            startsAt: new Date(startsAt).toISOString(),
            endsAt,
            location,
          })
      onSaved(result)
      onClose()
    } catch (error) {
      setNotice(getErrorMessage(error, meetup ? '改期没有保存' : '约见没有发出'))
    } finally {
      setSaving(false)
    }
  }

  const places: CampusPlace[] = suggestions?.places ?? meetup?.suggestedPlaces ?? []

  return (
    <div
      aria-labelledby="meetup-composer-title"
      aria-modal="true"
      className="fixed inset-0 z-[85] flex items-end justify-center bg-black/35 sm:items-center sm:p-5"
      role="dialog"
    >
      <div className="flex max-h-[92dvh] w-full max-w-xl flex-col overflow-hidden rounded-t-lg bg-[#f7f8f6] shadow-2xl sm:rounded-lg">
        <header className="flex shrink-0 items-center gap-3 border-b border-[#dfe1de] bg-white px-4 py-3 sm:px-5">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded bg-[#eee8f2] text-[#5b3477]">
            <CalendarClock className="h-4.5 w-4.5" />
          </span>
          <div>
            <h2 className="text-sm font-semibold text-[#292b28]" id="meetup-composer-title">
              {meetup ? '提出改期' : '发起约见'}
            </h2>
            <p className="mt-0.5 text-xs text-[#7b807a]">清华校内</p>
          </div>
          <button
            aria-label="关闭"
            className="ml-auto inline-flex h-9 w-9 items-center justify-center rounded text-[#686d67] hover:bg-[#eef0ed]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X className="h-5 w-5" />
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto p-4 sm:p-5">
          {notice ? (
            <div className="mb-4 rounded border border-[#e4cfc9] bg-[#fff6f3] px-3 py-2 text-xs text-[#8c463d]">
              {notice}
            </div>
          ) : null}

          {!meetup ? (
            <label className="block text-xs font-medium text-[#5e635d]">
              主题
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-white px-3 text-sm outline-none focus:border-[#6f5081]"
                maxLength={120}
                onChange={(event) => setTitle(event.target.value)}
                value={title}
              />
            </label>
          ) : null}

          <section className={cn('border-b border-[#dfe1de] pb-4', !meetup && 'mt-4')}>
            <div className="flex items-center gap-2">
              <Clock3 className="h-4 w-4 text-[#47705d]" />
              <h3 className="text-xs font-semibold text-[#454944]">共同空闲</h3>
              <button
                aria-label="重新计算共同空闲"
                className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded text-[#626761] hover:bg-[#e9ece8] disabled:opacity-50"
                disabled={loadingSuggestions}
                onClick={() => void loadSuggestions()}
                title="重新计算"
                type="button"
              >
                {loadingSuggestions ? (
                  <LoaderCircle className="h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="h-4 w-4" />
                )}
              </button>
            </div>

            {suggestions?.candidates.length ? (
              <div className="mt-3 grid gap-2">
                {suggestions.candidates.map((candidate) => (
                  <button
                    className={cn(
                      'rounded border px-3 py-2.5 text-left transition-colors',
                      startsAt === candidate.startsAt
                        ? 'border-[#6f5081] bg-[#f1edf3]'
                        : 'border-[#d7dad6] bg-white hover:border-[#a995b4]',
                    )}
                    key={candidate.startsAt}
                    onClick={() => setStartsAt(candidate.startsAt)}
                    type="button"
                  >
                    <span className="block text-xs font-semibold text-[#3e423e]">
                      {candidate.label}
                    </span>
                    <span className="mt-1 block text-[10px] leading-4 text-[#747973]">
                      {candidate.explanation}
                    </span>
                  </button>
                ))}
              </div>
            ) : !loadingSuggestions ? (
              <div className="mt-3 border-y border-dashed border-[#cfd3ce] py-4 text-center">
                <p className="text-xs text-[#727771]">
                  {suggestions?.missingAvailabilityCount
                    ? `${suggestions.missingAvailabilityCount} 位参与者尚未设置可约时间`
                    : '未来两周暂时没有共同空闲'}
                </p>
                {onOpenAvailability ? (
                  <button
                    className="mt-2 h-8 rounded px-3 text-xs font-medium text-[#5b3477] hover:bg-[#f1edf3]"
                    onClick={onOpenAvailability}
                    type="button"
                  >
                    设置我的可约时间
                  </button>
                ) : null}
              </div>
            ) : null}
          </section>

          <section className="mt-4 grid gap-3 sm:grid-cols-[1fr_140px]">
            <label className="text-xs font-medium text-[#5e635d]">
              时间
              <input
                className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-white px-3 text-sm outline-none focus:border-[#6f5081]"
                min={toDateTimeLocal(new Date().toISOString())}
                onChange={(event) => setStartsAt(new Date(event.target.value).toISOString())}
                type="datetime-local"
                value={toDateTimeLocal(startsAt)}
              />
            </label>
            <label className="text-xs font-medium text-[#5e635d]">
              时长
              <select
                className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-white px-3 text-sm outline-none focus:border-[#6f5081]"
                onChange={(event) => setDurationMinutes(Number(event.target.value))}
                value={durationMinutes}
              >
                <option value={30}>30 分钟</option>
                <option value={60}>1 小时</option>
                <option value={90}>1.5 小时</option>
                <option value={120}>2 小时</option>
                <option value={180}>3 小时</option>
              </select>
            </label>
          </section>

          <section className="mt-4">
            <label className="text-xs font-medium text-[#5e635d]">
              地点
              <span className="relative mt-1.5 block">
                <MapPin className="pointer-events-none absolute left-3 top-3 h-4 w-4 text-[#777c76]" />
                <input
                  className="h-10 w-full rounded border border-[#d7dad6] bg-white pl-9 pr-3 text-sm outline-none focus:border-[#6f5081]"
                  maxLength={160}
                  onChange={(event) => setLocation(event.target.value)}
                  placeholder="选择或输入校内地点"
                  value={location}
                />
              </span>
            </label>
            <div className="mt-2 flex gap-2 overflow-x-auto pb-1">
              {places.map((place) => (
                <button
                  className={cn(
                    'shrink-0 rounded border px-2.5 py-1.5 text-[11px]',
                    location === place.name
                      ? 'border-[#6f5081] bg-[#f1edf3] text-[#503461]'
                      : 'border-[#d7dad6] bg-white text-[#666b65] hover:border-[#a995b4]',
                  )}
                  key={place.name}
                  onClick={() => setLocation(place.name)}
                  type="button"
                >
                  {place.name}
                </button>
              ))}
            </div>
          </section>

          {suggestions ? (
            <div className="mt-4 flex items-center gap-2 text-[11px] text-[#6f746e]">
              <Users className="h-3.5 w-3.5" />
              {suggestions.participantCount} 位参与者
            </div>
          ) : null}
        </div>

        <footer className="flex shrink-0 justify-end gap-2 border-t border-[#dfe1de] bg-white px-4 py-3 sm:px-5">
          <button
            className="h-9 rounded px-3 text-xs font-medium text-[#656a64] hover:bg-[#eef0ed]"
            onClick={onClose}
            type="button"
          >
            取消
          </button>
          <button
            className="inline-flex h-9 items-center gap-1.5 rounded bg-[#5b3477] px-4 text-xs font-semibold text-white hover:bg-[#4c2b64] disabled:opacity-50"
            disabled={saving || !startsAt || !title.trim()}
            onClick={() => void save()}
            type="button"
          >
            {saving ? (
              <LoaderCircle className="h-4 w-4 animate-spin" />
            ) : (
              <Check className="h-4 w-4" />
            )}
            {meetup ? '发送改期' : '发出提议'}
          </button>
        </footer>
      </div>
    </div>
  )
}
