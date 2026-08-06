import { CalendarClock, Flag, Sparkles, Users } from 'lucide-react'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'

import { ChatInput } from '@/components/chat/chat-input'
import { ChatAssistantPanel } from '@/components/chat/chat-assistant-panel'
import { MessageBubble } from '@/components/chat/message-bubble'
import { MessageRiskReviewDialog } from '@/components/chat/message-risk-review-dialog'
import { TimeSeparator } from '@/components/chat/time-separator'
import { MeetupWorkflowDialog } from '@/components/meetup/meetup-context-panel'
import { useChat } from '@/contexts/chat-context'
import {
  fetchGroupMessages,
  fetchMessages,
  checkMessageSafety,
  markConversationRead,
  markGroupRead,
  sendConversationMessage,
  sendGroupMessage,
  type AssistantActionCard,
  type MessageSafetyAssessment,
} from '@/lib/chat-api'
import { getAuthSession } from '@/lib/user-profile-storage'
import { cn } from '@/lib/utils'

const TIME_GAP_MS = 3 * 60 * 1000

interface ChatWindowThread {
  id: number
  kind: 'direct' | 'group'
  title: string
  subtitle: string
  canSend: boolean
  statusMessage?: string
}

interface ChatMessage {
  id: number
  senderId: number
  senderNickname?: string
  content: string
  createdAt: string
}

interface Props {
  thread: ChatWindowThread
  onOpenGroupInfo?: () => void
  onRead?: () => void
}

export function ChatWindow({ thread, onOpenGroupInfo, onRead }: Props) {
  const { lastMessage, lastGroupMessage, refreshUnread } = useChat()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [sendNotice, setSendNotice] = useState('')
  const [meetupsOpen, setMeetupsOpen] = useState(false)
  const [assistantOpen, setAssistantOpen] = useState(false)
  const [composerDraft, setComposerDraft] = useState('')
  const [composerRevision, setComposerRevision] = useState(0)
  const [meetupDraftTitle, setMeetupDraftTitle] = useState('')
  const [riskReview, setRiskReview] = useState<{
    content: string
    assessment: MessageSafetyAssessment
  } | null>(null)
  const [riskSending, setRiskSending] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const isInitialLoad = useRef(true)
  const currentUserId = getAuthSession()?.userId

  const markRead = useCallback(async () => {
    if (thread.kind === 'group') {
      await markGroupRead(thread.id)
    } else {
      await markConversationRead(thread.id)
    }
    refreshUnread()
    onRead?.()
  }, [onRead, refreshUnread, thread.id, thread.kind])

  const loadMessages = useCallback(async () => {
    try {
      if (thread.kind === 'group') {
        const result = await fetchGroupMessages(thread.id, { limit: 50 })
        setMessages(result.items.slice().reverse())
        setHasMore(result.hasMore)
      } else {
        const result = await fetchMessages(thread.id, { limit: 50 })
        setMessages(result.items.slice().reverse())
        setHasMore(result.hasMore)
      }
      await markRead()
    } catch {
      setSendNotice('消息加载失败')
    }
  }, [markRead, thread.id, thread.kind])

  useEffect(() => {
    isInitialLoad.current = true
    setMessages([])
    setSendNotice('')
    void loadMessages()
  }, [loadMessages])

  useEffect(() => {
    if (thread.kind !== 'direct' || !lastMessage || lastMessage.conversationId !== thread.id) return
    setMessages((current) =>
      current.some((message) => message.id === lastMessage.id)
        ? current
        : [...current, lastMessage],
    )
    if (lastMessage.senderId !== currentUserId) void markRead()
  }, [currentUserId, lastMessage, markRead, thread.id, thread.kind])

  useEffect(() => {
    if (thread.kind !== 'group' || !lastGroupMessage || lastGroupMessage.groupId !== thread.id) return
    setMessages((current) =>
      current.some((message) => message.id === lastGroupMessage.id)
        ? current
        : [...current, lastGroupMessage],
    )
    if (lastGroupMessage.senderId !== currentUserId) void markRead()
  }, [currentUserId, lastGroupMessage, markRead, thread.id, thread.kind])

  useEffect(() => {
    if (messages.length === 0) return
    if (isInitialLoad.current) {
      isInitialLoad.current = false
      bottomRef.current?.scrollIntoView({ behavior: 'instant' })
    } else {
      bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }
  }, [messages.length])

  const handleLoadMore = useCallback(async () => {
    if (!hasMore || loadingMore || messages.length === 0) return
    setLoadingMore(true)
    const container = scrollRef.current
    const previousHeight = container?.scrollHeight ?? 0
    try {
      const oldestID = messages[0].id
      if (thread.kind === 'group') {
        const result = await fetchGroupMessages(thread.id, { before: oldestID, limit: 50 })
        setMessages((current) => [...result.items.slice().reverse(), ...current])
        setHasMore(result.hasMore)
      } else {
        const result = await fetchMessages(thread.id, { before: oldestID, limit: 50 })
        setMessages((current) => [...result.items.slice().reverse(), ...current])
        setHasMore(result.hasMore)
      }
      requestAnimationFrame(() => {
        if (container) container.scrollTop = container.scrollHeight - previousHeight
      })
    } catch {
      setSendNotice('更早的消息加载失败')
    } finally {
      setLoadingMore(false)
    }
  }, [hasMore, loadingMore, messages, thread.id, thread.kind])

  useEffect(() => {
    const container = scrollRef.current
    if (!container) return
    const onScroll = () => {
      if (container.scrollTop < 80) void handleLoadMore()
    }
    container.addEventListener('scroll', onScroll, { passive: true })
    return () => container.removeEventListener('scroll', onScroll)
  }, [handleLoadMore])

  const sendCheckedMessage = async (content: string, riskAcknowledged: boolean) => {
    if (!currentUserId || !thread.canSend) return false
    setSendNotice('')
    try {
      const message = thread.kind === 'group'
        ? await sendGroupMessage(thread.id, content, riskAcknowledged)
        : await sendConversationMessage(thread.id, content, riskAcknowledged)
      setMessages((current) => current.some((item) => item.id === message.id) ? current : [...current, message])
      onRead?.()
      return true
    } catch {
      setSendNotice('消息发送失败，请稍后重试')
      return false
    }
  }

  const handleSend = async (content: string) => {
    if (!currentUserId || !thread.canSend) return false
    try {
      const assessment = await checkMessageSafety(thread.kind, thread.id, content)
      if (assessment.level !== 'clear') {
        setRiskReview({ content, assessment })
        return false
      }
    } catch {
      setSendNotice('消息安全检查失败，请稍后重试')
      return false
    }
    return sendCheckedMessage(content, false)
  }

  return (
    <div className="relative flex h-full min-h-0 flex-col bg-[#f2f3f1]">
      <header className="flex h-15 shrink-0 items-center border-b border-[#dfe1de] bg-[#f7f8f6] px-4">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-semibold text-[#292b28]">{thread.title}</h2>
          <p className="truncate text-xs text-[#81857f]">{thread.subtitle}</p>
        </div>
        <button
          aria-label="聊天助手"
          className={cn(
            'ml-auto inline-flex h-9 w-9 items-center justify-center rounded text-[#626761] hover:bg-[#e4e6e3]',
            assistantOpen && 'bg-[#e2ebe6] text-[#315c4d]',
          )}
          onClick={() => setAssistantOpen((current) => !current)}
          title="聊天助手"
          type="button"
        >
          <Sparkles className="h-5 w-5" />
        </button>
        <button
          aria-label="约见与共同空闲"
          className="inline-flex h-9 w-9 items-center justify-center rounded text-[#626761] hover:bg-[#e4e6e3]"
          onClick={() => setMeetupsOpen(true)}
          title="约见"
          type="button"
        >
          <CalendarClock className="h-5 w-5" />
        </button>
        {thread.kind === 'direct' ? (
          <Link
            aria-label="举报当前会话"
            className="inline-flex h-9 w-9 items-center justify-center rounded text-[#626761] hover:bg-[#f1e6e8] hover:text-[#8d465d]"
            title="举报当前会话"
            to={`/feedback?category=report&contextType=direct&contextId=${thread.id}&subject=${encodeURIComponent(thread.title)}`}
          >
            <Flag className="h-4.5 w-4.5" />
          </Link>
        ) : null}
        {thread.kind === 'group' ? (
          <button
            aria-label="群聊信息"
            className="inline-flex h-9 w-9 items-center justify-center rounded text-[#626761] hover:bg-[#e4e6e3]"
            onClick={onOpenGroupInfo}
            title="群聊信息"
            type="button"
          >
            <Users className="h-5 w-5" />
          </button>
        ) : null}
      </header>

      <div className="flex-1 overflow-y-auto px-4 py-5 sm:px-7" ref={scrollRef}>
        {loadingMore ? <p className="mb-3 text-center text-xs text-[#81857f]">加载中...</p> : null}
        <div className="space-y-4">
          {messages.map((message, index) => {
            const previous = index > 0 ? messages[index - 1] : null
            const gap = previous
              ? new Date(message.createdAt).getTime() - new Date(previous.createdAt).getTime()
              : Infinity
            return (
              <Fragment key={message.id}>
                {gap >= TIME_GAP_MS ? <TimeSeparator time={message.createdAt} /> : null}
                <MessageBubble
                  content={message.content}
                  isMine={message.senderId === currentUserId}
                  senderName={thread.kind === 'group' ? message.senderNickname : thread.title}
                  time={message.createdAt}
                />
              </Fragment>
            )
          })}
        </div>
        <div ref={bottomRef} />
      </div>

      {thread.statusMessage || sendNotice ? (
        <div className="border-t border-[#dfe1de] bg-[#f8eee4] px-4 py-2 text-center text-xs text-[#80654f]">
          {thread.statusMessage || sendNotice}
        </div>
      ) : null}
      <ChatInput
        disabled={!thread.canSend}
        initialText={composerDraft}
        key={`${thread.kind}:${thread.id}:${composerRevision}`}
        onSend={handleSend}
      />
      {meetupsOpen ? (
        <MeetupWorkflowDialog
          context={{
            id: thread.id,
            type: thread.kind === 'group' ? 'group' : 'direct',
          }}
          defaultTitle={meetupDraftTitle || (thread.kind === 'group' ? thread.title : `和 ${thread.title} 见面`)}
          onClose={() => {
            setMeetupsOpen(false)
            setMeetupDraftTitle('')
          }}
          title={thread.kind === 'group' ? `${thread.title} · 约见` : `和 ${thread.title} 约见`}
        />
      ) : null}
      {assistantOpen ? (
        <ChatAssistantPanel
          contextId={thread.id}
          contextType={thread.kind}
          onClose={() => setAssistantOpen(false)}
          onOpenMeetup={(card?: AssistantActionCard) => {
            setMeetupDraftTitle(card?.detail ?? '')
            setMeetupsOpen(true)
            setAssistantOpen(false)
          }}
          onUseDraft={(draft) => {
            setComposerDraft(draft)
            setComposerRevision((current) => current + 1)
            setAssistantOpen(false)
          }}
        />
      ) : null}
      {riskReview ? (
        <MessageRiskReviewDialog
          assessment={riskReview.assessment}
          content={riskReview.content}
          onCancel={() => setRiskReview(null)}
          onConfirm={() => {
            void (async () => {
              setRiskSending(true)
              const sent = await sendCheckedMessage(riskReview.content, true)
              setRiskSending(false)
              if (!sent) return
              setRiskReview(null)
              setComposerDraft('')
              setComposerRevision((current) => current + 1)
            })()
          }}
          sending={riskSending}
        />
      ) : null}
    </div>
  )
}
