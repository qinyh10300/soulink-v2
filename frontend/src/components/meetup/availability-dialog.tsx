import { CalendarPlus, Check, Clock3, LoaderCircle, Plus, Repeat2, Trash2, X } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  fetchAvailability,
  replaceAvailability,
  type AvailabilityWindow,
  type AvailabilityWindowInput,
} from '@/lib/meetup-api'
import { cn } from '@/lib/utils'

interface Props {
  onClose: () => void
  onSaved?: () => void
}

interface DraftWindow extends AvailabilityWindowInput {
  key: string
}

const weekdays = ['一', '二', '三', '四', '五', '六', '日']

function toDateTimeLocal(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  const offset = date.getTimezoneOffset() * 60_000
  return new Date(date.getTime() - offset).toISOString().slice(0, 16)
}

function fromDateTimeLocal(value?: string) {
  return value ? new Date(value).toISOString() : undefined
}

function minutesToTime(value = 0) {
  return `${String(Math.floor(value / 60)).padStart(2, '0')}:${String(value % 60).padStart(2, '0')}`
}

function timeToMinutes(value: string) {
  const [hours, minutes] = value.split(':').map(Number)
  return hours * 60 + minutes
}

function nextHour() {
  const value = new Date()
  value.setMinutes(0, 0, 0)
  value.setHours(value.getHours() + 1)
  return value
}

function mapWindow(window: AvailabilityWindow): DraftWindow {
  return {
    key: `saved-${window.id}`,
    windowType: window.windowType,
    timezone: window.timezone,
    startsAt: window.startsAt,
    endsAt: window.endsAt,
    weekday: window.weekday,
    startMinute: window.startMinute,
    endMinute: window.endMinute,
    validFrom: window.validFrom,
    validUntil: window.validUntil,
  }
}

function newWindow(type: 'weekly' | 'one_time'): DraftWindow {
  if (type === 'weekly') {
    return {
      key: crypto.randomUUID(),
      windowType: 'weekly',
      timezone: 'Asia/Shanghai',
      weekday: Math.max(1, new Date().getDay() || 7),
      startMinute: 18 * 60,
      endMinute: 21 * 60,
    }
  }
  const start = nextHour()
  return {
    key: crypto.randomUUID(),
    windowType: 'one_time',
    timezone: 'Asia/Shanghai',
    startsAt: start.toISOString(),
    endsAt: new Date(start.getTime() + 2 * 60 * 60 * 1000).toISOString(),
  }
}

export function AvailabilityDialog({ onClose, onSaved }: Props) {
  const [windows, setWindows] = useState<DraftWindow[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      try {
        const result = await fetchAvailability()
        if (!cancelled) setWindows(result.map(mapWindow))
      } catch (error) {
        if (!cancelled) setNotice(getErrorMessage(error, '可约时间加载失败'))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const weeklyCount = useMemo(
    () => windows.filter((window) => window.windowType === 'weekly').length,
    [windows],
  )

  const updateWindow = (key: string, patch: Partial<DraftWindow>) => {
    setWindows((current) =>
      current.map((window) => (window.key === key ? { ...window, ...patch } : window)),
    )
  }

  const save = async () => {
    if (saving) return
    setSaving(true)
    setNotice('')
    try {
      await replaceAvailability(
        windows.map((window) => ({
          windowType: window.windowType,
          timezone: window.timezone,
          startsAt: window.startsAt,
          endsAt: window.endsAt,
          weekday: window.weekday,
          startMinute: window.startMinute,
          endMinute: window.endMinute,
          validFrom: window.validFrom,
          validUntil: window.validUntil,
        })),
      )
      onSaved?.()
      onClose()
    } catch (error) {
      setNotice(getErrorMessage(error, '可约时间保存失败'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      aria-labelledby="availability-dialog-title"
      aria-modal="true"
      className="fixed inset-0 z-[80] flex items-end justify-center bg-black/35 sm:items-center sm:p-5"
      role="dialog"
    >
      <div className="flex max-h-[92dvh] w-full max-w-2xl flex-col overflow-hidden rounded-t-lg bg-[#f7f8f6] shadow-2xl sm:rounded-lg">
        <header className="flex shrink-0 items-center gap-3 border-b border-[#dfe1de] bg-white px-4 py-3 sm:px-5">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded bg-[#e5efe9] text-[#315d4f]">
            <Clock3 className="h-4.5 w-4.5" />
          </span>
          <div className="min-w-0">
            <h2 className="text-sm font-semibold text-[#292b28]" id="availability-dialog-title">
              我的可约时间
            </h2>
            <p className="mt-0.5 text-xs text-[#7b807a]">当前 {weeklyCount} 个周期时间</p>
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

          {loading ? (
            <div className="flex h-40 items-center justify-center text-sm text-[#747973]">
              <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />
              正在加载...
            </div>
          ) : (
            <div className="space-y-3">
              {windows.map((window) => (
                <section
                  className="rounded-lg border border-[#d9dcd8] bg-white p-3 sm:p-4"
                  key={window.key}
                >
                  <div className="flex items-center gap-2">
                    {window.windowType === 'weekly' ? (
                      <Repeat2 className="h-4 w-4 text-[#47705d]" />
                    ) : (
                      <CalendarPlus className="h-4 w-4 text-[#6c4a80]" />
                    )}
                    <span className="text-xs font-semibold text-[#454944]">
                      {window.windowType === 'weekly' ? '每周重复' : '仅这一次'}
                    </span>
                    <button
                      aria-label="删除时间段"
                      className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded text-[#8b5862] hover:bg-[#fff0f3]"
                      onClick={() =>
                        setWindows((current) => current.filter((item) => item.key !== window.key))
                      }
                      title="删除"
                      type="button"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>

                  {window.windowType === 'weekly' ? (
                    <div className="mt-3 space-y-3">
                      <div className="grid grid-cols-7 overflow-hidden rounded border border-[#d7dad6]">
                        {weekdays.map((label, index) => (
                          <button
                            className={cn(
                              'h-9 border-r border-[#d7dad6] text-xs last:border-r-0',
                              window.weekday === index + 1
                                ? 'bg-[#376f5b] font-semibold text-white'
                                : 'bg-white text-[#656a64] hover:bg-[#eef3f0]',
                            )}
                            key={label}
                            onClick={() => updateWindow(window.key, { weekday: index + 1 })}
                            type="button"
                          >
                            {label}
                          </button>
                        ))}
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <label className="text-xs text-[#656a64]">
                          开始
                          <input
                            className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 text-sm outline-none focus:border-[#5d7f6d]"
                            onChange={(event) =>
                              updateWindow(window.key, {
                                startMinute: timeToMinutes(event.target.value),
                              })
                            }
                            type="time"
                            value={minutesToTime(window.startMinute)}
                          />
                        </label>
                        <label className="text-xs text-[#656a64]">
                          结束
                          <input
                            className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 text-sm outline-none focus:border-[#5d7f6d]"
                            onChange={(event) =>
                              updateWindow(window.key, {
                                endMinute: timeToMinutes(event.target.value),
                              })
                            }
                            type="time"
                            value={minutesToTime(window.endMinute)}
                          />
                        </label>
                      </div>
                      <label className="block text-xs text-[#656a64]">
                        有效至
                        <input
                          className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 text-sm outline-none focus:border-[#5d7f6d]"
                          min={new Date().toISOString().slice(0, 10)}
                          onChange={(event) =>
                            updateWindow(window.key, {
                              validUntil: event.target.value || undefined,
                            })
                          }
                          type="date"
                          value={window.validUntil ?? ''}
                        />
                      </label>
                    </div>
                  ) : (
                    <div className="mt-3 grid gap-3 sm:grid-cols-2">
                      <label className="text-xs text-[#656a64]">
                        开始
                        <input
                          className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 text-sm outline-none focus:border-[#5d7f6d]"
                          onChange={(event) =>
                            updateWindow(window.key, {
                              startsAt: fromDateTimeLocal(event.target.value),
                            })
                          }
                          type="datetime-local"
                          value={toDateTimeLocal(window.startsAt)}
                        />
                      </label>
                      <label className="text-xs text-[#656a64]">
                        结束
                        <input
                          className="mt-1.5 h-10 w-full rounded border border-[#d7dad6] bg-[#fbfcfa] px-3 text-sm outline-none focus:border-[#5d7f6d]"
                          onChange={(event) =>
                            updateWindow(window.key, {
                              endsAt: fromDateTimeLocal(event.target.value),
                            })
                          }
                          type="datetime-local"
                          value={toDateTimeLocal(window.endsAt)}
                        />
                      </label>
                    </div>
                  )}
                </section>
              ))}

              {windows.length === 0 ? (
                <div className="border-y border-dashed border-[#cfd3ce] py-12 text-center text-sm text-[#7a7f79]">
                  还没有可约时间
                </div>
              ) : null}
            </div>
          )}
        </div>

        <footer className="flex shrink-0 flex-wrap items-center gap-2 border-t border-[#dfe1de] bg-white px-4 py-3 sm:px-5">
          <button
            className="inline-flex h-9 items-center gap-1.5 rounded border border-[#cfd3ce] px-3 text-xs font-medium text-[#535853] hover:bg-[#eef0ed]"
            onClick={() => setWindows((current) => [...current, newWindow('weekly')])}
            type="button"
          >
            <Repeat2 className="h-4 w-4" />
            每周时间
          </button>
          <button
            className="inline-flex h-9 items-center gap-1.5 rounded border border-[#cfd3ce] px-3 text-xs font-medium text-[#535853] hover:bg-[#eef0ed]"
            onClick={() => setWindows((current) => [...current, newWindow('one_time')])}
            type="button"
          >
            <Plus className="h-4 w-4" />
            一次时间
          </button>
          <button
            className="ml-auto inline-flex h-9 items-center gap-1.5 rounded bg-[#337a49] px-4 text-xs font-semibold text-white hover:bg-[#28663c] disabled:opacity-50"
            disabled={loading || saving}
            onClick={() => void save()}
            type="button"
          >
            {saving ? (
              <LoaderCircle className="h-4 w-4 animate-spin" />
            ) : (
              <Check className="h-4 w-4" />
            )}
            保存
          </button>
        </footer>
      </div>
    </div>
  )
}
