import {
  ArrowLeft,
  CalendarDays,
  Clock3,
  MapPin,
  MessageCircle,
  UserCheck,
  UserPlus,
  UserRound,
} from 'lucide-react'
import {
  addDays,
  addMinutes,
  format,
  isSameDay,
  parseISO,
  startOfWeek,
} from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { CalendarEventDialog } from '@/components/calendar/calendar-event-dialog'
import { getErrorMessage } from '@/lib/api'
import { fetchUserSchedule, type CalendarEvent } from '@/lib/calendar-api'
import { ensureConversation } from '@/lib/chat-api'
import {
  fetchFriends,
  fetchReceivedRequests,
  fetchSentRequests,
  sendFriendRequest,
} from '@/lib/friend-api'
import { fetchPublicProfile, type PublicProfileDetails } from '@/lib/profile-api'

type RelationshipStatus = 'friend' | 'none' | 'received' | 'sent' | 'unknown'

async function fetchRelationshipStatus(userId: number): Promise<RelationshipStatus> {
  try {
    const [friends, sentRequests, receivedRequests] = await Promise.all([
      fetchFriends(),
      fetchSentRequests(),
      fetchReceivedRequests(),
    ])
    if (friends.some((friend) => friend.user.userId === userId)) return 'friend'
    if (receivedRequests.some((request) => request.user.userId === userId)) return 'received'
    if (sentRequests.some((request) => request.user.userId === userId)) return 'sent'
    return 'none'
  } catch {
    return 'unknown'
  }
}

function formatJoinedAt(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(new Date(value))
}

function findSuggestedStart(events: CalendarEvent[]) {
  const now = new Date()
  let cursor = addMinutes(now, 60 - (now.getMinutes() % 60))
  const deadline = addDays(now, 7)
  while (cursor < deadline) {
    const hour = cursor.getHours()
    if (hour < 9) cursor = new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate(), 9)
    if (hour >= 21) cursor = new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate() + 1, 9)
    const candidateEnd = addMinutes(cursor, 60)
    const conflicts = events.some((event) => {
      const start = parseISO(event.startsAt)
      const end = parseISO(event.endsAt)
      return start < candidateEnd && end > cursor
    })
    if (!conflicts) return cursor
    cursor = addMinutes(cursor, 30)
  }
  return addMinutes(now, 60)
}

export function PublicProfilePage() {
  const { userId } = useParams()
  const navigate = useNavigate()
  const numericUserId = Number(userId)
  const [profile, setProfile] = useState<PublicProfileDetails | null>(null)
  const [schedule, setSchedule] = useState<CalendarEvent[]>([])
  const [inviteOpen, setInviteOpen] = useState(false)
  const [relationshipStatus, setRelationshipStatus] = useState<RelationshipStatus>('unknown')
  const [notice, setNotice] = useState('正在加载个人主页...')
  const weekStart = useMemo(() => startOfWeek(new Date(), { weekStartsOn: 1 }), [])
  const days = useMemo(() => Array.from({ length: 7 }, (_, index) => addDays(weekStart, index)), [weekStart])

  const loadProfile = useCallback(async () => {
    if (!Number.isFinite(numericUserId) || numericUserId <= 0) {
      setNotice('用户不存在')
      return
    }
    try {
      const [profileData, events, relationship] = await Promise.all([
        fetchPublicProfile(numericUserId),
        fetchUserSchedule(numericUserId, weekStart, addDays(weekStart, 7)),
        fetchRelationshipStatus(numericUserId),
      ])
      setProfile(profileData)
      setSchedule(events)
      setRelationshipStatus(relationship)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '个人主页加载失败，请稍后重试'))
    }
  }, [numericUserId, weekStart])

  useEffect(() => {
    void loadProfile()
  }, [loadProfile])

  const handleMessage = async () => {
    if (!profile) return
    if (relationshipStatus !== 'friend') {
      setNotice('成为好友后才能发消息')
      return
    }
    try {
      const conversation = await ensureConversation(numericUserId)
      navigate(`/chat/direct/${conversation.id}`)
    } catch (error) {
      setNotice(getErrorMessage(error, '成为好友后才能发消息'))
    }
  }

  const handleFriendRequest = async () => {
    if (!profile || relationshipStatus !== 'none') return
    try {
      await sendFriendRequest(profile.nickname)
      setRelationshipStatus('sent')
      setNotice('好友申请已发送')
    } catch (error) {
      setNotice(getErrorMessage(error, '好友申请发送失败'))
    }
  }

  return (
    <div className="min-h-screen bg-[#f6f7f5] px-4 py-5 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-5xl">
        <button
          className="inline-flex h-9 items-center gap-2 rounded px-2 text-sm text-[#656a65] hover:bg-[#e7e9e6]"
          onClick={() => navigate(-1)}
          type="button"
        >
          <ArrowLeft className="h-4 w-4" />
          返回
        </button>

        {profile ? (
          <section className="mt-4 border-y border-[#d9dcd8] bg-white">
            <div className="h-24 bg-[linear-gradient(120deg,#dce8e1_0%,#e9e4dc_48%,#e4e5ef_100%)]" />
            <div className="flex flex-col gap-5 px-5 pb-5 sm:flex-row sm:items-end sm:px-7">
              <div className="-mt-10 flex h-20 w-20 shrink-0 items-center justify-center rounded-lg border-4 border-white bg-[#cad9d1] text-3xl font-semibold text-[#315d4f] shadow-sm">
                {profile.nickname.trim().slice(0, 1) || <UserRound className="h-7 w-7" />}
              </div>
              <div className="min-w-0 flex-1 sm:pb-1">
                <h1 className="text-2xl font-semibold text-[#292b28]">{profile.nickname}</h1>
                <p className="mt-1 text-sm text-[#777c77]">
                  {[profile.schoolName, profile.department, profile.grade].filter(Boolean).join(' · ')}
                </p>
                <p className="mt-1 text-xs text-[#959993]">加入于 {formatJoinedAt(profile.createdAt)}</p>
              </div>
              <div className="flex flex-wrap gap-2 sm:pb-1">
                {relationshipStatus === 'none' ? (
                  <button
                    className="inline-flex h-9 items-center gap-2 rounded border border-[#cfd2ce] bg-white px-3 text-sm hover:bg-[#f0f2ef]"
                    onClick={() => void handleFriendRequest()}
                    type="button"
                  >
                    <UserPlus className="h-4 w-4" />
                    加好友
                  </button>
                ) : (
                  <button
                    className="inline-flex h-9 items-center gap-2 rounded border border-[#d9dcd8] bg-[#f3f4f2] px-3 text-sm text-[#747974]"
                    disabled
                    title={relationshipStatus === 'received' ? '请在好友申请中处理' : undefined}
                    type="button"
                  >
                    {relationshipStatus === 'friend' ? (
                      <UserCheck className="h-4 w-4" />
                    ) : (
                      <Clock3 className="h-4 w-4" />
                    )}
                    {relationshipStatus === 'friend'
                      ? '已是好友'
                      : relationshipStatus === 'sent'
                        ? '申请已发送'
                        : relationshipStatus === 'received'
                          ? '待处理申请'
                          : '关系加载中'}
                  </button>
                )}
                <button
                  className="inline-flex h-9 items-center gap-2 rounded border border-[#cfd2ce] bg-white px-3 text-sm hover:bg-[#f0f2ef] disabled:cursor-not-allowed disabled:bg-[#f3f4f2] disabled:text-[#a0a49f]"
                  disabled={relationshipStatus !== 'friend'}
                  onClick={() => void handleMessage()}
                  title={relationshipStatus === 'friend' ? undefined : '成为好友后才能发消息'}
                  type="button"
                >
                  <MessageCircle className="h-4 w-4" />
                  发消息
                </button>
                <button
                  className="inline-flex h-9 items-center gap-2 rounded bg-[#337a49] px-3 text-sm font-medium text-white hover:bg-[#28663c]"
                  onClick={() => setInviteOpen(true)}
                  type="button"
                >
                  <CalendarDays className="h-4 w-4" />
                  约个时间
                </button>
              </div>
            </div>
          </section>
        ) : null}

        {notice ? (
          <div className="mt-4 rounded border border-[#d9dcd8] bg-white px-4 py-3 text-sm text-[#6f746f]">
            {notice}
          </div>
        ) : null}

        {profile ? (
          <section className="mt-6">
            <div className="flex items-end justify-between border-b border-[#d9dcd8] pb-3">
              <div>
                <h2 className="text-lg font-semibold text-[#292b28]">本周时间</h2>
                <p className="mt-1 text-xs text-[#858a84]">{format(weekStart, 'M月d日', { locale: zhCN })}起</p>
              </div>
              <button
                className="inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs font-medium text-[#337a49] hover:bg-[#e9f0eb]"
                onClick={() => setInviteOpen(true)}
                type="button"
              >
                <Clock3 className="h-4 w-4" />
                发起邀约
              </button>
            </div>

            <div className="mt-3 grid gap-px overflow-hidden border border-[#d9dcd8] bg-[#d9dcd8] sm:grid-cols-7">
              {days.map((day) => {
                const dayEvents = schedule.filter((event) => isSameDay(parseISO(event.startsAt), day))
                return (
                  <div className="min-h-32 bg-white p-2" key={day.toISOString()}>
                    <div className="flex items-center justify-between sm:block">
                      <p className="text-[11px] text-[#858a84]">{format(day, 'EEE', { locale: zhCN })}</p>
                      <p className="text-sm font-semibold">{format(day, 'd')}</p>
                    </div>
                    <div className="mt-2 space-y-1.5">
                      {dayEvents.map((event) => (
                        <div
                          className="rounded border-l-2 border-[#638775] bg-[#edf3ef] px-2 py-1.5"
                          key={event.id}
                        >
                          <p className="truncate text-[11px] font-medium text-[#365a49]">{event.title}</p>
                          <p className="mt-0.5 text-[10px] text-[#718078]">
                            {format(parseISO(event.startsAt), 'HH:mm')} - {format(parseISO(event.endsAt), 'HH:mm')}
                          </p>
                          {event.location && !event.isBusyOnly ? (
                            <p className="mt-0.5 flex items-center gap-1 truncate text-[10px] text-[#718078]">
                              <MapPin className="h-3 w-3 shrink-0" />
                              {event.location}
                            </p>
                          ) : null}
                        </div>
                      ))}
                      {dayEvents.length === 0 ? <p className="text-[10px] text-[#adb0ac]">暂无安排</p> : null}
                    </div>
                  </div>
                )
              })}
            </div>
          </section>
        ) : null}
      </div>

      {inviteOpen && profile ? (
        <CalendarEventDialog
          initialStart={findSuggestedStart(schedule)}
          lockedInvitee={{ userId: numericUserId, nickname: profile.nickname }}
          onClose={() => setInviteOpen(false)}
          onSaved={() => {
            setInviteOpen(false)
            setNotice('邀约已发送，对方可以在时间表中回复')
            void loadProfile()
          }}
        />
      ) : null}
    </div>
  )
}
