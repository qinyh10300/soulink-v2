import { CalendarClock, LoaderCircle, Plus, RefreshCw, Settings2, X } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { AvailabilityDialog } from '@/components/meetup/availability-dialog'
import { MeetupCard } from '@/components/meetup/meetup-card'
import { MeetupComposerDialog } from '@/components/meetup/meetup-composer-dialog'
import { getErrorMessage } from '@/lib/api'
import { fetchMeetups, type Meetup, type MeetupContext } from '@/lib/meetup-api'

interface PanelProps {
  context: MeetupContext
  defaultCategory?: string
  defaultTitle?: string
  compact?: boolean
}

interface DialogProps extends PanelProps {
  onClose: () => void
  title: string
}

export function MeetupContextPanel({
  context,
  defaultCategory = 'other',
  defaultTitle = '一起在校内见面',
  compact = false,
}: PanelProps) {
  const [meetups, setMeetups] = useState<Meetup[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [composerOpen, setComposerOpen] = useState(false)
  const [rescheduling, setRescheduling] = useState<Meetup | null>(null)
  const [availabilityOpen, setAvailabilityOpen] = useState(false)
  const contextID = context.id
  const contextType = context.type

  const load = useCallback(async () => {
    setLoading(true)
    setNotice('')
    try {
      setMeetups(await fetchMeetups({ id: contextID, type: contextType }))
    } catch (error) {
      setNotice(getErrorMessage(error, '约见记录加载失败'))
    } finally {
      setLoading(false)
    }
  }, [contextID, contextType])

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0)
    return () => window.clearTimeout(timeout)
  }, [load])

  const handleUpdated = (updated: Meetup) => {
    setMeetups((current) => {
      const exists = current.some((meetup) => meetup.id === updated.id)
      return exists
        ? current.map((meetup) => (meetup.id === updated.id ? updated : meetup))
        : [updated, ...current]
    })
  }

  return (
    <section>
      <div className="flex items-center gap-2">
        <CalendarClock className="h-4 w-4 text-[#5b3477]" />
        <h2 className="text-sm font-semibold text-[#373b36]">校内约见</h2>
        <button
          aria-label="刷新约见"
          className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded text-[#676c66] hover:bg-[#eef0ed] disabled:opacity-50"
          disabled={loading}
          onClick={() => void load()}
          title="刷新"
          type="button"
        >
          {loading ? (
            <LoaderCircle className="h-4 w-4 animate-spin" />
          ) : (
            <RefreshCw className="h-4 w-4" />
          )}
        </button>
        <button
          className="inline-flex h-8 items-center gap-1.5 rounded bg-[#5b3477] px-3 text-xs font-semibold text-white hover:bg-[#4b2b62]"
          onClick={() => setComposerOpen(true)}
          type="button"
        >
          <Plus className="h-3.5 w-3.5" />
          发起
        </button>
      </div>

      {notice ? (
        <div className="mt-3 rounded border border-[#e4cfc9] bg-[#fff6f3] px-3 py-2 text-xs text-[#8c463d]">
          {notice}
        </div>
      ) : null}

      {!loading && meetups.length === 0 ? (
        <div className="mt-3 border-y border-dashed border-[#cfd3ce] py-8 text-center">
          <CalendarClock className="mx-auto h-6 w-6 text-[#949993]" />
          <p className="mt-2 text-xs text-[#757a74]">还没有约见提议</p>
          <button
            className="mt-2 inline-flex h-8 items-center gap-1 rounded px-3 text-xs font-medium text-[#5b3477] hover:bg-[#f1edf3]"
            onClick={() => setAvailabilityOpen(true)}
            type="button"
          >
            <Settings2 className="h-3.5 w-3.5" />
            设置可约时间
          </button>
        </div>
      ) : (
        <div className="mt-3 space-y-3">
          {meetups.map((meetup) => (
            <MeetupCard
              compact={compact}
              key={meetup.id}
              meetup={meetup}
              onReschedule={setRescheduling}
              onUpdated={handleUpdated}
            />
          ))}
        </div>
      )}

      {composerOpen || rescheduling ? (
        <MeetupComposerDialog
          context={context}
          defaultCategory={defaultCategory}
          defaultTitle={defaultTitle}
          meetup={rescheduling ?? undefined}
          onClose={() => {
            setComposerOpen(false)
            setRescheduling(null)
          }}
          onOpenAvailability={() => {
            setComposerOpen(false)
            setRescheduling(null)
            setAvailabilityOpen(true)
          }}
          onSaved={handleUpdated}
        />
      ) : null}

      {availabilityOpen ? (
        <AvailabilityDialog
          onClose={() => setAvailabilityOpen(false)}
          onSaved={() => void load()}
        />
      ) : null}
    </section>
  )
}

export function MeetupWorkflowDialog({
  context,
  defaultCategory,
  defaultTitle,
  onClose,
  title,
}: DialogProps) {
  return (
    <div
      aria-labelledby="meetup-workflow-title"
      aria-modal="true"
      className="fixed inset-0 z-[75] flex items-end justify-center bg-black/35 sm:items-center sm:p-5"
      role="dialog"
    >
      <div className="flex max-h-[92dvh] w-full max-w-2xl flex-col overflow-hidden rounded-t-lg bg-[#f7f8f6] shadow-2xl sm:rounded-lg">
        <header className="flex shrink-0 items-center gap-3 border-b border-[#dfe1de] bg-white px-4 py-3 sm:px-5">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded bg-[#eee8f2] text-[#5b3477]">
            <CalendarClock className="h-4.5 w-4.5" />
          </span>
          <div className="min-w-0">
            <h2
              className="truncate text-sm font-semibold text-[#292b28]"
              id="meetup-workflow-title"
            >
              {title}
            </h2>
            <p className="mt-0.5 text-xs text-[#7b807a]">约见与共同空闲</p>
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
          <MeetupContextPanel
            context={context}
            defaultCategory={defaultCategory}
            defaultTitle={defaultTitle}
          />
        </div>
      </div>
    </div>
  )
}
