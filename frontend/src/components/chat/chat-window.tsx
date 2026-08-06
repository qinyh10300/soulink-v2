import { Users } from 'lucide-react'
import { Fragment, useCallback, useEffect, useRef, useState } from 'react'

import { ChatInput } from '@/components/chat/chat-input'
import { MessageBubble } from '@/components/chat/message-bubble'
import { TimeSeparator } from '@/components/chat/time-separator'
import { useChat } from '@/contexts/chat-context'
import {
  fetchGroupMessages,
  fetchMessages,
  markConversationRead,
  markGroupRead,
  sendConversationMessage,
  sendGroupMessage,
} from '@/lib/chat-api'
import { getAuthSession } from '@/lib/user-profile-storage'

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
  const { ws, lastMessage, lastGroupMessage, refreshUnread } = useChat()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [sendNotice, setSendNotice] = useState('')
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

  const handleSend = async (content: string) => {
    if (!currentUserId || !thread.canSend) return
    setSendNotice('')
    if (ws?.isConnected) {
      if (thread.kind === 'group') ws.sendGroupMessage(thread.id, content)
      else ws.sendMessage(thread.id, content)
      return
    }
    try {
      const message = thread.kind === 'group'
        ? await sendGroupMessage(thread.id, content)
        : await sendConversationMessage(thread.id, content)
      setMessages((current) => [...current, message])
      onRead?.()
    } catch {
      setSendNotice('消息发送失败，请稍后重试')
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col bg-[#f2f3f1]">
      <header className="flex h-15 shrink-0 items-center border-b border-[#dfe1de] bg-[#f7f8f6] px-4">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-semibold text-[#292b28]">{thread.title}</h2>
          <p className="truncate text-xs text-[#81857f]">{thread.subtitle}</p>
        </div>
        {thread.kind === 'group' ? (
          <button
            aria-label="群聊信息"
            className="ml-auto inline-flex h-9 w-9 items-center justify-center rounded text-[#626761] hover:bg-[#e4e6e3]"
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
      <ChatInput disabled={!thread.canSend} onSend={(content) => void handleSend(content)} />
    </div>
  )
}
