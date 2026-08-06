import {
  AlertCircle,
  ArrowLeft,
  Check,
  CheckCircle2,
  CircleX,
  ContactRound,
  Copy,
  Flag,
  LoaderCircle,
  MessageCircle,
  Send,
  ShieldCheck,
  Sparkles,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { IntentCategoryIcon } from '@/components/connection/intent-category'
import { MeetupContextPanel } from '@/components/meetup/meetup-context-panel'
import { getErrorMessage } from '@/lib/api'
import {
  createMatchRoomMessage,
  fetchMatchRoom,
  fetchMatchRoomMessages,
  setMatchRoomContactSharing,
  updateMatchRoomStatus,
  type MatchRoom,
  type MatchRoomMessage,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'

function formatMessageTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

export function MatchRoomPage() {
  const navigate = useNavigate()
  const params = useParams()
  const roomId = Number(params.roomId)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [room, setRoom] = useState<MatchRoom | null>(null)
  const [messages, setMessages] = useState<MatchRoomMessage[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [sharingContact, setSharingContact] = useState(false)
  const [message, setMessage] = useState('')
  const [sendingMessage, setSendingMessage] = useState(false)
  const [copied, setCopied] = useState(false)
  const [updatingStatus, setUpdatingStatus] = useState<'closed' | null>(null)
  const [confirmingClose, setConfirmingClose] = useState(false)

  useEffect(() => {
    if (!Number.isInteger(roomId) || roomId <= 0) {
      setNotice('无效的匹配房间编号')
      setLoading(false)
      return
    }
    let cancelled = false
    const load = async () => {
      try {
        const [roomResult, messageResult] = await Promise.all([
          fetchMatchRoom(roomId),
          fetchMatchRoomMessages(roomId),
        ])
        if (cancelled) {
          return
        }
        setRoom(roomResult)
        setMessages(messageResult)
      } catch (error) {
        if (!cancelled) {
          setNotice(getErrorMessage(error, '这个匹配房间暂时无法打开'))
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
  }, [roomId])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, [messages])

  const toggleContact = async () => {
    if (!room || sharingContact) {
      return
    }
    setSharingContact(true)
    try {
      setRoom(
        await setMatchRoomContactSharing(room.id, !room.myContactShared),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '联系方式状态暂时无法更新'))
    } finally {
      setSharingContact(false)
    }
  }

  const sendMessage = async () => {
    if (!room || !message.trim() || sendingMessage) {
      return
    }
    setSendingMessage(true)
    try {
      const created = await createMatchRoomMessage(room.id, message.trim())
      setMessages((current) => [...current, created])
      setMessage('')
    } catch (error) {
      setNotice(getErrorMessage(error, '消息没有发出去'))
    } finally {
      setSendingMessage(false)
    }
  }

  const copyContact = async () => {
    if (!room?.other.contact?.value) {
      return
    }
    await navigator.clipboard.writeText(room.other.contact.value)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1600)
  }

  const changeRoomStatus = async (status: 'closed') => {
    if (!room || updatingStatus) {
      return
    }
    setUpdatingStatus(status)
    setNotice('')
    try {
      setRoom(await updateMatchRoomStatus(room.id, status))
      setConfirmingClose(false)
    } catch (error) {
      setNotice(getErrorMessage(error, '匹配房间状态暂时无法更新'))
    } finally {
      setUpdatingStatus(null)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center text-sm text-[#77736c]">
        <LoaderCircle className="mr-2 animate-spin" size={18} />
        正在准备匹配房间...
      </div>
    )
  }

  if (!room) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-20 text-center">
        <AlertCircle className="mx-auto text-[#9a6155]" size={28} />
        <p className="mt-3 text-sm text-[#6f6b65]">{notice || '找不到这个房间'}</p>
        <Link
          className="mt-4 inline-flex h-9 items-center rounded-md border border-[#d8d3cc] px-3 text-xs font-semibold text-[#5b3477]"
          to="/matches"
        >
          返回匹配房间
        </Link>
      </div>
    )
  }

  const feedbackSubject = `和 ${room.other.nickname} 的匹配 · ${room.intentTitle}`
  const reportPath = `/feedback?${new URLSearchParams({
    category: 'report',
    roomId: String(room.id),
    subject: feedbackSubject,
  }).toString()}`
  const roomClosed = room.status === 'closed'
  const meetingConfirmed = room.status === 'met'

  return (
    <div className="mx-auto w-full max-w-[1180px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex items-center justify-between gap-4 border-b border-[#dfdcd6] pb-5">
        <div className="flex min-w-0 items-center gap-3">
          <button
            aria-label="返回"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#6f6b65] hover:bg-[#eeece8]"
            onClick={() => navigate(-1)}
            title="返回"
            type="button"
          >
            <ArrowLeft size={19} />
          </button>
          <IntentCategoryIcon category={room.intentCategory} size={18} />
          <div className="min-w-0">
            <p className="truncate text-xs font-medium text-[#6c4a80]">
              {room.intentTitle}
            </p>
            <h1 className="truncate text-xl font-semibold text-[#292724] sm:text-2xl">
              你和 {room.other.nickname}
            </h1>
          </div>
        </div>
        <span
          className={cn(
            'inline-flex shrink-0 items-center gap-1.5 rounded-md px-2.5 py-1.5 text-xs font-semibold',
            roomClosed
              ? 'bg-[#edeae5] text-[#6d6963]'
              : meetingConfirmed
                ? 'bg-[#eee8f2] text-[#5b3477]'
                : 'bg-[#e8f3ed] text-[#2d6a53]',
          )}
        >
          {meetingConfirmed ? <CheckCircle2 size={13} /> : <ShieldCheck size={13} />}
          {roomClosed ? '连接已结束' : meetingConfirmed ? '已确认见面' : '双方同意'}
        </span>
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <div className="grid gap-6 py-7 lg:grid-cols-[minmax(0,0.9fr)_minmax(380px,1.1fr)]">
        <div className="min-w-0 space-y-5">
          <section className="rounded-lg border border-[#d9d0df] bg-[#f1ecf4] p-5">
            <div className="flex items-center gap-2 text-sm font-semibold text-[#4d2d66]">
              <Sparkles size={16} />
              为什么把你们介绍到这里
            </div>
            <p className="mt-3 text-sm leading-7 text-[#5e5066]">{room.fitReason}</p>
            {room.sharedContext.length > 0 ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {room.sharedContext.map((item) => (
                  <span
                    className="rounded-md bg-white/70 px-2.5 py-1.5 text-xs text-[#685672]"
                    key={item}
                  >
                    {item}
                  </span>
                ))}
              </div>
            ) : null}
            <p className="mt-4 border-t border-[#dcd2e1] pt-3 text-xs leading-5 text-[#75687b]">
              {room.uncertainty}
            </p>
          </section>

          <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
            <MeetupContextPanel
              context={{ id: room.id, type: 'match_room' }}
              defaultCategory={room.intentCategory}
              defaultTitle={room.intentTitle}
            />
          </section>

          <section className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-5">
            <h2 className="flex items-center gap-2 text-sm font-semibold text-[#37332f]">
              <ContactRound size={16} />
              联系方式
            </h2>
            <p className="mt-2 text-xs leading-5 text-[#77736c]">
              每个人单独决定是否向对方揭示；关闭后会立即隐藏。
            </p>
            <button
              className={cn(
                'mt-4 flex w-full items-center justify-between rounded-md border px-3 py-3 text-left',
                room.myContactShared
                  ? 'border-[#bcd4c9] bg-[#edf6f1]'
                  : 'border-[#d9d5cf] bg-white',
              )}
              disabled={sharingContact || roomClosed}
              onClick={() => void toggleContact()}
              type="button"
            >
              <span>
                <span className="block text-sm font-medium text-[#47433f]">
                  向 {room.other.nickname} 显示我的联系方式
                </span>
                <span className="mt-0.5 block text-xs text-[#817d76]">
                  {room.myContactShared ? '当前已显示' : '当前未显示'}
                </span>
              </span>
              <span
                aria-hidden="true"
                className={cn(
                  'relative h-6 w-11 rounded-full transition-colors',
                  room.myContactShared ? 'bg-[#376f5b]' : 'bg-[#c9c5bf]',
                )}
              >
                <span
                  className={cn(
                    'absolute top-1 h-4 w-4 rounded-full bg-white transition-transform',
                    room.myContactShared ? 'translate-x-6' : 'translate-x-1',
                  )}
                />
              </span>
            </button>

            {room.other.contact ? (
              <div className="mt-3 flex items-center gap-3 rounded-md border border-[#d5dfda] bg-white px-3 py-3">
                <span className="min-w-0 flex-1">
                  <span className="block text-[11px] text-[#817d76]">
                    {room.other.nickname} 已向你显示
                  </span>
                  <span className="mt-0.5 block truncate text-sm font-semibold text-[#3c4e46]">
                    {room.other.contact.value}
                  </span>
                </span>
                <button
                  aria-label="复制联系方式"
                  className="inline-flex h-8 w-8 items-center justify-center rounded-md text-[#546a61] hover:bg-[#eef3f0]"
                  onClick={() => void copyContact()}
                  title={copied ? '已复制' : '复制'}
                  type="button"
                >
                  {copied ? <Check size={15} /> : <Copy size={15} />}
                </button>
              </div>
            ) : (
              <p className="mt-3 rounded-md bg-[#f2efeb] px-3 py-2 text-xs text-[#77736c]">
                对方还没有选择显示联系方式，站内消息仍然可用。
              </p>
            )}
          </section>

          <section className="flex flex-wrap items-center gap-2 border-t border-[#dfdcd6] pt-4">
            <Link
              className="inline-flex h-9 items-center gap-2 rounded-md px-3 text-xs font-medium text-[#8b4559] hover:bg-[#fff0f4]"
              to={reportPath}
            >
              <Flag size={14} />
              举报与拉黑
            </Link>
            {!roomClosed ? (
              <button
                className="inline-flex h-9 items-center gap-2 rounded-md px-3 text-xs font-medium text-[#77736c] hover:bg-[#eeece8]"
                onClick={() => setConfirmingClose(true)}
                type="button"
              >
                <CircleX size={14} />
                结束连接
              </button>
            ) : null}
            {confirmingClose ? (
              <div className="mt-2 flex w-full flex-col gap-3 rounded-md border border-[#ead4ce] bg-[#fff6f3] p-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-xs leading-5 text-[#7a5149]">
                  结束后将停止站内消息，并立即收回双方联系方式。
                </p>
                <div className="flex shrink-0 gap-2">
                  <button
                    className="h-8 rounded-md px-3 text-xs font-medium text-[#6f6b65]"
                    onClick={() => setConfirmingClose(false)}
                    type="button"
                  >
                    取消
                  </button>
                  <button
                    className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#8b4559] px-3 text-xs font-semibold text-white disabled:opacity-50"
                    disabled={updatingStatus !== null}
                    onClick={() => void changeRoomStatus('closed')}
                    type="button"
                  >
                    {updatingStatus === 'closed' ? (
                      <LoaderCircle className="animate-spin" size={13} />
                    ) : null}
                    确认结束
                  </button>
                </div>
              </div>
            ) : null}
          </section>
        </div>

        <section className="flex min-h-[620px] min-w-0 flex-col rounded-lg border border-[#dfdcd6] bg-white">
          <div className="border-b border-[#e4e1dc] px-4 py-4 sm:px-5">
            <h2 className="flex items-center gap-2 text-sm font-semibold text-[#37332f]">
              <MessageCircle size={16} />
              和 {room.other.nickname} 先聊几句
            </h2>
            <div className="mt-3 flex gap-2 overflow-x-auto pb-1">
              {room.conversationStarters.map((starter) => (
                <button
                  className="max-w-64 shrink-0 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 py-2 text-left text-[11px] leading-4 text-[#69655f] hover:border-[#bbaac5]"
                  key={starter}
                  onClick={() => setMessage(starter)}
                  disabled={roomClosed}
                  type="button"
                >
                  {starter}
                </button>
              ))}
            </div>
          </div>

          <div className="min-h-0 flex-1 space-y-3 overflow-y-auto bg-[#faf9f7] p-4 sm:p-5">
            {messages.length > 0 ? (
              messages.map((item) => {
                const mine = item.senderId === room.me.userId
                return (
                  <div
                    className={cn('flex', mine ? 'justify-end' : 'justify-start')}
                    key={item.id}
                  >
                    <div className="max-w-[82%]">
                      <p
                        className={cn(
                          'mb-1 text-[10px] text-[#8a867f]',
                          mine && 'text-right',
                        )}
                      >
                        {mine ? '你' : item.senderNickname} ·{' '}
                        {formatMessageTime(item.createdAt)}
                      </p>
                      <p
                        className={cn(
                          'rounded-lg px-3.5 py-2.5 text-sm leading-6',
                          mine
                            ? 'bg-[#5b3477] text-white'
                            : 'border border-[#e1ded8] bg-white text-[#4f4b46]',
                        )}
                      >
                        {item.content}
                      </p>
                    </div>
                  </div>
                )
              })
            ) : (
              <div className="flex h-full min-h-64 items-center justify-center text-center">
                <div>
                  <MessageCircle className="mx-auto text-[#aaa49c]" size={25} />
                  <p className="mt-3 text-sm text-[#77736c]">
                    从一个具体问题开始，会比“在吗”更轻松。
                  </p>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          <div className="border-t border-[#e4e1dc] bg-white p-3 sm:p-4">
            <div className="flex items-end gap-2">
              <textarea
                className="min-h-11 max-h-28 flex-1 resize-none rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 py-2.5 text-sm leading-5 outline-none focus:border-[#7a5a8e]"
                maxLength={1000}
                onChange={(event) => setMessage(event.target.value)}
                disabled={roomClosed}
                onKeyDown={(event) => {
                  if (
                    event.key === 'Enter' &&
                    !event.shiftKey &&
                    !event.nativeEvent.isComposing
                  ) {
                    event.preventDefault()
                    void sendMessage()
                  }
                }}
                placeholder={roomClosed ? '这个匹配房间已停止交流' : '写一句具体的话...'}
                rows={1}
                value={message}
              />
              <button
                aria-label="发送消息"
                className="inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-[#5b3477] text-white disabled:opacity-50"
                disabled={roomClosed || !message.trim() || sendingMessage}
                onClick={() => void sendMessage()}
                title="发送"
                type="button"
              >
                {sendingMessage ? (
                  <LoaderCircle className="animate-spin" size={17} />
                ) : (
                  <Send size={17} />
                )}
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
  )
}
