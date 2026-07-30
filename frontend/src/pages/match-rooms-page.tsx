import {
  AlertCircle,
  ArrowRight,
  CalendarCheck2,
  Clock3,
  HeartHandshake,
  LoaderCircle,
  MapPin,
  ShieldCheck,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { IntentCategoryIcon } from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import { fetchMatchRooms, type MatchRoom } from '@/lib/connection-api'
import { cn } from '@/lib/utils'

function formatMeeting(value?: string) {
  if (!value) {
    return '时间待确认'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    weekday: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

export function MatchRoomsPage() {
  const [rooms, setRooms] = useState<MatchRoom[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      try {
        const result = await fetchMatchRooms()
        if (!cancelled) {
          setRooms(result)
        }
      } catch (error) {
        if (!cancelled) {
          setNotice(getErrorMessage(error, '匹配房间暂时无法加载'))
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="mx-auto w-full max-w-[980px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="border-b border-[#dfdcd6] pb-5">
        <p className="flex items-center gap-1.5 text-xs font-medium text-[#6c4a80]">
          <HeartHandshake size={13} />
          MUTUAL MATCHES
        </p>
        <h1 className="mt-1 text-2xl font-semibold text-[#292724]">匹配房间</h1>
        <p className="mt-1 text-sm text-[#77736c]">只有双方都愿意的连接，才会来到这里。</p>
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <section className="py-6">
        {loading ? (
          <div className="flex min-h-64 items-center justify-center text-sm text-[#77736c]">
            <LoaderCircle className="mr-2 animate-spin" size={17} />
            正在打开双方同意的连接...
          </div>
        ) : rooms.length > 0 ? (
          <div className="space-y-3">
            {rooms.map((room) => (
              <Link
                className="group block rounded-lg border border-[#dfdcd6] bg-white p-4 transition-colors hover:border-[#b9a8c4] sm:p-5"
                key={room.id}
                to={`/matches/${room.id}`}
              >
                <div className="flex items-start gap-3">
                  <IntentCategoryIcon category={room.intentCategory} size={18} />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <h2 className="truncate text-base font-semibold text-[#302e2b]">
                            和 {room.other.nickname}
                          </h2>
                          <span
                            className={cn(
                              'inline-flex items-center gap-1 rounded-md px-2 py-1 text-[11px] font-semibold',
                              room.status === 'open'
                                ? 'bg-[#e8f3ed] text-[#2d6a53]'
                                : 'bg-[#edeae5] text-[#6d6963]',
                            )}
                          >
                            <ShieldCheck size={11} />
                            {room.status === 'open' ? '双方同意' : '交流已结束'}
                          </span>
                        </div>
                        <p className="mt-1 truncate text-sm text-[#706c65]">
                          {room.intentTitle}
                        </p>
                      </div>
                      <ArrowRight
                        className="shrink-0 text-[#8b8780] transition-transform group-hover:translate-x-0.5 group-hover:text-[#5b3477]"
                        size={18}
                      />
                    </div>

                    <p className="mt-3 line-clamp-2 text-xs leading-5 text-[#77736c]">
                      {room.fitReason}
                    </p>

                    <div
                      className={cn(
                        'mt-4 grid gap-2 border-t border-[#ebe8e3] pt-3 text-xs text-[#69655f] sm:grid-cols-2',
                        room.status !== 'open' && 'opacity-60',
                      )}
                    >
                      <span className="inline-flex items-center gap-1.5">
                        <Clock3 size={14} />
                        {formatMeeting(room.scheduledAt)}
                      </span>
                      <span className="inline-flex items-center gap-1.5">
                        <MapPin size={14} />
                        {room.selectedPlace || '地点待确认'}
                      </span>
                    </div>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        ) : (
          <div className="flex min-h-64 flex-col items-center justify-center rounded-lg border border-dashed border-[#d7d2cb] px-6 text-center">
            <CalendarCheck2 className="text-[#8d828f]" size={28} />
            <p className="mt-3 text-sm font-medium text-[#55514c]">还没有匹配房间</p>
            <p className="mt-1 max-w-sm text-xs leading-5 text-[#817d76]">
              当你回应一个意图，而对方也愿意认识你时，系统会自动创建房间。
            </p>
            <Link
              className="mt-4 inline-flex h-9 items-center gap-2 rounded-md bg-[#5b3477] px-3 text-xs font-semibold text-white"
              to="/discover"
            >
              去看看意图流
              <ArrowRight size={14} />
            </Link>
          </div>
        )}
      </section>
    </div>
  )
}
