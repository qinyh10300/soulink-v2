import {
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Clock3,
  MapPin,
  Plus,
  Settings2,
} from 'lucide-react'
import {
  addDays,
  addHours,
  addWeeks,
  endOfWeek,
  format,
  isSameDay,
  parseISO,
  setMinutes,
  setSeconds,
  startOfWeek,
} from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useCallback, useEffect, useMemo, useState } from 'react'

import { CalendarEventDialog } from '@/components/calendar/calendar-event-dialog'
import { WeekCalendarGrid } from '@/components/calendar/week-calendar-grid'
import { AvailabilityDialog } from '@/components/meetup/availability-dialog'
import { MeetupAgenda } from '@/components/meetup/meetup-agenda'
import { getErrorMessage } from '@/lib/api'
import { fetchCalendarEvents, type CalendarEvent } from '@/lib/calendar-api'
import { fetchFriends, type FriendItem } from '@/lib/friend-api'

function nextHour() {
  return addHours(setSeconds(setMinutes(new Date(), 0), 0), 1)
}

export function CalendarPage() {
  const [weekStart, setWeekStart] = useState(() => startOfWeek(new Date(), { weekStartsOn: 1 }))
  const [events, setEvents] = useState<CalendarEvent[]>([])
  const [friends, setFriends] = useState<FriendItem[]>([])
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null)
  const [draftStart, setDraftStart] = useState<Date | null>(null)
  const [availabilityOpen, setAvailabilityOpen] = useState(false)
  const [notice, setNotice] = useState('正在加载日程...')

  const weekEnd = useMemo(() => endOfWeek(weekStart, { weekStartsOn: 1 }), [weekStart])

  const load = useCallback(async () => {
    try {
      const [nextEvents, nextFriends] = await Promise.all([
        fetchCalendarEvents(weekStart, addDays(weekStart, 7)),
        fetchFriends(),
      ])
      setEvents(nextEvents)
      setFriends(nextFriends)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '日程加载失败'))
    }
  }, [weekStart])

  useEffect(() => {
    void load()
  }, [load])

  const pendingInvitations = events.filter((event) => event.attendeeStatus === 'pending')
  const days = Array.from({ length: 7 }, (_, index) => addDays(weekStart, index))

  const openSlot = (start: Date) => {
    if (start < new Date()) {
      setNotice('不能在过去的时间创建日程')
      return
    }
    setDraftStart(start)
  }

  const closeDialog = () => {
    setDraftStart(null)
    setSelectedEvent(null)
  }

  const handleSaved = () => {
    closeDialog()
    void load()
  }

  return (
    <div className="min-h-screen bg-[#f6f7f5] px-4 py-5 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-[1320px]">
        <header className="flex flex-wrap items-center gap-3 border-b border-[#d9dcd8] pb-4">
          <div className="flex min-w-0 items-center gap-3">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded bg-[#dce8e1] text-[#315d4f]">
              <CalendarDays className="h-5 w-5" />
            </span>
            <div>
              <h1 className="text-xl font-semibold text-[#292b28]">时间表</h1>
              <p className="text-xs text-[#7e827d]">{format(weekStart, 'yyyy 年 M 月', { locale: zhCN })}</p>
            </div>
          </div>

          <div className="ml-auto flex items-center gap-1">
            <button
              className="h-9 rounded border border-[#cfd2ce] bg-white px-3 text-xs font-medium hover:bg-[#eef0ed]"
              onClick={() => setWeekStart(startOfWeek(new Date(), { weekStartsOn: 1 }))}
              type="button"
            >
              今天
            </button>
            <button
              aria-label="上一周"
              className="inline-flex h-9 w-9 items-center justify-center rounded hover:bg-[#e7e9e6]"
              onClick={() => setWeekStart((current) => addWeeks(current, -1))}
              title="上一周"
              type="button"
            >
              <ChevronLeft className="h-5 w-5" />
            </button>
            <button
              aria-label="下一周"
              className="inline-flex h-9 w-9 items-center justify-center rounded hover:bg-[#e7e9e6]"
              onClick={() => setWeekStart((current) => addWeeks(current, 1))}
              title="下一周"
              type="button"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
            <button
              aria-label="设置可约时间"
              className="ml-2 inline-flex h-9 w-9 items-center justify-center rounded border border-[#cfd2ce] bg-white text-[#59605a] hover:bg-[#eef0ed]"
              onClick={() => setAvailabilityOpen(true)}
              title="可约时间"
              type="button"
            >
              <Settings2 className="h-4 w-4" />
            </button>
            <button
              className="inline-flex h-9 items-center gap-2 rounded bg-[#337a49] px-3 text-sm font-medium text-white hover:bg-[#28663c]"
              onClick={() => setDraftStart(nextHour())}
              type="button"
            >
              <Plus className="h-4 w-4" />
              新建
            </button>
          </div>
        </header>

        {pendingInvitations.length > 0 ? (
          <section className="mt-4 border-y border-[#d9dcd8] bg-[#f0f5f1] px-3 py-3">
            <div className="flex gap-3 overflow-x-auto">
              <span className="inline-flex shrink-0 items-center gap-1.5 text-xs font-semibold text-[#3f6854]">
                <Clock3 className="h-4 w-4" />
                待回复
              </span>
              {pendingInvitations.map((event) => (
                <button
                  className="shrink-0 rounded bg-white px-3 py-1.5 text-left text-xs shadow-sm hover:bg-[#fafbf9]"
                  key={event.id}
                  onClick={() => setSelectedEvent(event)}
                  type="button"
                >
                  <span className="font-medium">{event.title}</span>
                  <span className="ml-2 text-[#858984]">
                    {format(parseISO(event.startsAt), 'M月d日 HH:mm')}
                  </span>
                </button>
              ))}
            </div>
          </section>
        ) : null}

        <MeetupAgenda from={weekStart} to={addDays(weekStart, 7)} />

        {notice ? (
          <div className="mt-4 rounded border border-[#d9dcd8] bg-white px-4 py-3 text-sm text-[#70746f]">
            {notice}
          </div>
        ) : null}

        <section className="mt-4 hidden md:block">
          <WeekCalendarGrid
            events={events}
            onEventClick={setSelectedEvent}
            onSlotClick={openSlot}
            weekStart={weekStart}
          />
        </section>

        <section className="mt-4 divide-y divide-[#dfe1de] border-y border-[#d9dcd8] md:hidden">
          {days.map((day) => {
            const dayEvents = events.filter((event) => isSameDay(parseISO(event.startsAt), day))
            return (
              <div className="grid grid-cols-[68px_1fr] gap-3 py-3" key={day.toISOString()}>
                <div>
                  <p className="text-xs text-[#7e827d]">{format(day, 'EEE', { locale: zhCN })}</p>
                  <p className="mt-1 text-xl font-semibold">{format(day, 'd')}</p>
                </div>
                <div className="space-y-2">
                  {dayEvents.map((event) => (
                    <button
                      className="w-full rounded border border-[#d7dcd8] bg-white p-3 text-left"
                      key={event.id}
                      onClick={() => setSelectedEvent(event)}
                      type="button"
                    >
                      <span className="block text-sm font-medium">{event.title}</span>
                      <span className="mt-1 flex items-center gap-3 text-xs text-[#7d817c]">
                        <span className="inline-flex items-center gap-1">
                          <Clock3 className="h-3.5 w-3.5" />
                          {format(parseISO(event.startsAt), 'HH:mm')} - {format(parseISO(event.endsAt), 'HH:mm')}
                        </span>
                        {event.location ? (
                          <span className="inline-flex min-w-0 items-center gap-1 truncate">
                            <MapPin className="h-3.5 w-3.5 shrink-0" />
                            {event.location}
                          </span>
                        ) : null}
                      </span>
                    </button>
                  ))}
                  {dayEvents.length === 0 ? (
                    <button
                      className="h-10 w-full rounded border border-dashed border-[#cfd2ce] text-xs text-[#858984]"
                      onClick={() => openSlot(addHours(day, 18))}
                      type="button"
                    >
                      添加日程
                    </button>
                  ) : null}
                </div>
              </div>
            )
          })}
        </section>

        <p className="mt-3 text-xs text-[#858984]">
          {format(weekStart, 'M月d日')} - {format(weekEnd, 'M月d日')}
        </p>
      </div>

      {draftStart || selectedEvent ? (
        <CalendarEventDialog
          event={selectedEvent}
          friends={friends}
          initialStart={draftStart ?? undefined}
          onClose={closeDialog}
          onSaved={handleSaved}
        />
      ) : null}
      {availabilityOpen ? (
        <AvailabilityDialog
          onClose={() => setAvailabilityOpen(false)}
          onSaved={() => void load()}
        />
      ) : null}
    </div>
  )
}
