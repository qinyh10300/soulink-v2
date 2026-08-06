import {
  addDays,
  addHours,
  differenceInMinutes,
  format,
  isSameDay,
  isToday,
  max,
  min,
  parseISO,
  startOfDay,
} from 'date-fns'
import { zhCN } from 'date-fns/locale'

import type { CalendarEvent, CalendarEventType } from '@/lib/calendar-api'
import { cn } from '@/lib/utils'

const START_HOUR = 7
const END_HOUR = 24
const HOUR_HEIGHT = 64

const eventColors: Record<CalendarEventType, string> = {
  meetup: 'border-[#4f806c] bg-[#dceae3] text-[#264c3d]',
  study: 'border-[#4f7796] bg-[#dde9f2] text-[#284d68]',
  sports: 'border-[#9a7948] bg-[#eee5d3] text-[#674d28]',
  meal: 'border-[#a1674c] bg-[#f0e1d8] text-[#6f402c]',
  date: 'border-[#a55d69] bg-[#f0dde1] text-[#743844]',
  project: 'border-[#6e6390] bg-[#e5e0ef] text-[#4e426f]',
  other: 'border-[#6f7771] bg-[#e5e8e5] text-[#474d48]',
  busy: 'border-[#8d918c] bg-[#e4e6e3] text-[#626762]',
}

interface Props {
  events: CalendarEvent[]
  onEventClick: (event: CalendarEvent) => void
  onSlotClick: (startsAt: Date) => void
  weekStart: Date
}

export function WeekCalendarGrid({ events, onEventClick, onSlotClick, weekStart }: Props) {
  const days = Array.from({ length: 7 }, (_, index) => addDays(weekStart, index))
  const hours = Array.from({ length: END_HOUR - START_HOUR }, (_, index) => START_HOUR + index)
  const bodyHeight = hours.length * HOUR_HEIGHT

  return (
    <div className="overflow-x-auto border border-[#d9dcd8] bg-white">
      <div className="min-w-[840px]">
        <div className="grid grid-cols-[56px_repeat(7,minmax(0,1fr))] border-b border-[#d9dcd8] bg-[#fafbf9]">
          <div className="border-r border-[#e1e3e0]" />
          {days.map((day) => (
            <div
              className={cn(
                'border-r border-[#e1e3e0] px-2 py-2 text-center last:border-r-0',
                isToday(day) && 'bg-[#eef5f0]',
              )}
              key={day.toISOString()}
            >
              <p className="text-[11px] text-[#7e827d]">{format(day, 'EEE', { locale: zhCN })}</p>
              <p
                className={cn(
                  'mx-auto mt-1 flex h-7 w-7 items-center justify-center rounded-full text-sm font-semibold',
                  isToday(day) && 'bg-[#337a49] text-white',
                )}
              >
                {format(day, 'd')}
              </p>
            </div>
          ))}
        </div>

        <div className="grid grid-cols-[56px_repeat(7,minmax(0,1fr))]">
          <div className="relative border-r border-[#e1e3e0]" style={{ height: bodyHeight }}>
            {hours.map((hour, index) => (
              <span
                className="absolute right-2 -translate-y-1/2 text-[10px] text-[#8d918b]"
                key={hour}
                style={{ top: index * HOUR_HEIGHT }}
              >
                {String(hour).padStart(2, '0')}:00
              </span>
            ))}
          </div>

          {days.map((day) => {
            const dayStart = startOfDay(day)
            const visibleStart = addHours(dayStart, START_HOUR)
            const visibleEnd = addHours(dayStart, END_HOUR)
            const dayEvents = events.filter((event) => {
              const startsAt = parseISO(event.startsAt)
              const endsAt = parseISO(event.endsAt)
              return startsAt < visibleEnd && endsAt > visibleStart && isSameDay(startsAt, day)
            })
            return (
              <div
                className={cn(
                  'relative border-r border-[#e1e3e0] last:border-r-0',
                  isToday(day) && 'bg-[#fbfdfb]',
                )}
                key={day.toISOString()}
                style={{ height: bodyHeight }}
              >
                {hours.map((hour, index) => (
                  <button
                    aria-label={`${format(day, 'M月d日')} ${hour}:00 新建日程`}
                    className="absolute inset-x-0 border-t border-[#eceeeb] hover:bg-[#f2f6f3]"
                    key={hour}
                    onClick={() => onSlotClick(addHours(dayStart, hour))}
                    style={{ height: HOUR_HEIGHT, top: index * HOUR_HEIGHT }}
                    type="button"
                  />
                ))}

                {dayEvents.map((event) => {
                  const start = max([parseISO(event.startsAt), visibleStart])
                  const end = min([parseISO(event.endsAt), visibleEnd])
                  const top = (differenceInMinutes(start, visibleStart) / 60) * HOUR_HEIGHT
                  const height = Math.max(24, (differenceInMinutes(end, start) / 60) * HOUR_HEIGHT)
                  return (
                    <button
                      className={cn(
                        'absolute inset-x-1 z-10 overflow-hidden rounded border-l-[3px] px-1.5 py-1 text-left shadow-[0_1px_2px_rgba(25,30,26,0.08)] hover:brightness-[0.97]',
                        eventColors[event.eventType],
                        event.attendeeStatus === 'pending' && 'border border-dashed opacity-85',
                      )}
                      key={event.id}
                      onClick={() => onEventClick(event)}
                      style={{ height, top }}
                      type="button"
                    >
                      <span className="block truncate text-[11px] font-semibold">{event.title}</span>
                      {height >= 42 ? (
                        <span className="block truncate text-[10px] opacity-80">
                          {format(parseISO(event.startsAt), 'HH:mm')} · {event.location || '未填写地点'}
                        </span>
                      ) : null}
                    </button>
                  )
                })}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
