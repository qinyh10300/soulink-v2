import { CalendarClock, LoaderCircle, RefreshCw } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'

import { MeetupCard } from '@/components/meetup/meetup-card'
import { MeetupComposerDialog } from '@/components/meetup/meetup-composer-dialog'
import { getErrorMessage } from '@/lib/api'
import { fetchMeetups, type Meetup } from '@/lib/meetup-api'

interface Props {
  from: Date
  to: Date
}

export function MeetupAgenda({ from, to }: Props) {
  const [meetups, setMeetups] = useState<Meetup[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [rescheduling, setRescheduling] = useState<Meetup | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setNotice('')
    try {
      setMeetups(await fetchMeetups(undefined, from, to))
    } catch (error) {
      setNotice(getErrorMessage(error, '约见动态加载失败'))
    } finally {
      setLoading(false)
    }
  }, [from, to])

  useEffect(() => {
    const timeout = window.setTimeout(() => void load(), 0)
    return () => window.clearTimeout(timeout)
  }, [load])

  const handleUpdated = (updated: Meetup) => {
    setMeetups((current) => current.map((meetup) => (meetup.id === updated.id ? updated : meetup)))
  }

  if (!loading && !notice && meetups.length === 0) return null

  return (
    <section className="mt-4 border-y border-[#d9dcd8] py-3">
      <div className="flex items-center gap-2">
        <CalendarClock className="h-4 w-4 text-[#5b3477]" />
        <h2 className="text-xs font-semibold text-[#4b4f4a]">约见动态</h2>
        <button
          aria-label="刷新约见动态"
          className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded text-[#676c66] hover:bg-[#e9ece8] disabled:opacity-50"
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
      </div>
      {notice ? <p className="mt-2 text-xs text-[#8c463d]">{notice}</p> : null}
      {meetups.length > 0 ? (
        <div className="mt-3 grid gap-3 lg:grid-cols-2">
          {meetups.map((meetup) => (
            <MeetupCard
              compact
              key={meetup.id}
              meetup={meetup}
              onReschedule={setRescheduling}
              onUpdated={handleUpdated}
            />
          ))}
        </div>
      ) : null}
      {rescheduling ? (
        <MeetupComposerDialog
          context={rescheduling.context}
          defaultCategory={rescheduling.category}
          defaultTitle={rescheduling.title}
          meetup={rescheduling}
          onClose={() => setRescheduling(null)}
          onSaved={handleUpdated}
        />
      ) : null}
    </section>
  )
}
