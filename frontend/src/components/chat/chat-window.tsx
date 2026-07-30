import { Fragment, useCallback, useEffect, useRef, useState } from 'react'

import { ChatInput } from '@/components/chat/chat-input'
import { MessageBubble } from '@/components/chat/message-bubble'
import { TimeSeparator } from '@/components/chat/time-separator'
import { useChat } from '@/contexts/chat-context'
import {
  fetchMessages,
  markConversationRead,
  type ConversationItem,
  type MessageItem,
} from '@/lib/chat-api'
import { getAuthSession } from '@/lib/user-profile-storage'

const TIME_GAP_MS = 3 * 60 * 1000

interface Props {
  conversation: ConversationItem
  onRead?: () => void
}

export function ChatWindow({ conversation, onRead }: Props) {
  const { ws, lastMessage, refreshUnread } = useChat()
  const [messages, setMessages] = useState<MessageItem[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const scrollRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const isInitialLoad = useRef(true)
  const currentUserId = getAuthSession()?.userId

  const loadMessages = useCallback(async () => {
    try {
      const result = await fetchMessages(conversation.id, { limit: 50 })
      setMessages(result.items.reverse())
      setHasMore(result.hasMore)
      await markConversationRead(conversation.id)
      refreshUnread()
      onRead?.()
    } catch {
      // ignore
    }
  }, [conversation.id, onRead, refreshUnread])

  useEffect(() => {
    loadMessages()
  }, [loadMessages])

  useEffect(() => {
    if (lastMessage && lastMessage.conversationId === conversation.id) {
      setMessages((prev) => [...prev, lastMessage])
      if (lastMessage.senderId !== currentUserId) {
        markConversationRead(conversation.id)
          .then(() => onRead?.())
          .catch(() => {})
        refreshUnread()
      }
    }
  }, [lastMessage, conversation.id, currentUserId, onRead, refreshUnread])

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
    const prevScrollHeight = container?.scrollHeight ?? 0

    try {
      const oldest = messages[0]
      const result = await fetchMessages(conversation.id, {
        before: oldest.id,
        limit: 50,
      })
      setMessages((prev) => [...result.items.reverse(), ...prev])
      setHasMore(result.hasMore)

      requestAnimationFrame(() => {
        if (container) {
          container.scrollTop = container.scrollHeight - prevScrollHeight
        }
      })
    } catch {
      // ignore
    } finally {
      setLoadingMore(false)
    }
  }, [hasMore, loadingMore, messages, conversation.id])

  useEffect(() => {
    const container = scrollRef.current
    if (!container) return

    function onScroll() {
      if (container && container.scrollTop < 80) {
        handleLoadMore()
      }
    }

    container.addEventListener('scroll', onScroll, { passive: true })
    return () => container.removeEventListener('scroll', onScroll)
  }, [handleLoadMore])

  function handleSend(content: string) {
    if (!ws || !currentUserId || !conversation.canSend) return
    ws.sendMessage(conversation.id, content)
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="border-b border-[#eee4db] px-4 py-3">
        <h2 className="text-sm font-semibold text-[#2f221b]">
          {conversation.peerNickname}
        </h2>
        <p className="text-xs text-[#7e6d61]">
          {conversation.peerDepartment} · {conversation.peerGrade}
        </p>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-4" ref={scrollRef}>
        {loadingMore ? (
          <div className="mb-3 text-center">
            <span className="text-xs text-[#7e6d61]">加载中...</span>
          </div>
        ) : null}

        <div className="space-y-3">
          {messages.map((msg, idx) => {
            const prev = idx > 0 ? messages[idx - 1] : null
            const gap = prev
              ? new Date(msg.createdAt).getTime() -
                new Date(prev.createdAt).getTime()
              : Infinity
            const showSep = gap >= TIME_GAP_MS

            return (
              <Fragment key={msg.id}>
                {showSep ? (
                  <TimeSeparator time={msg.createdAt} />
                ) : null}
                <MessageBubble
                  content={msg.content}
                  isMine={msg.senderId === currentUserId}
                  time={msg.createdAt}
                />
              </Fragment>
            )
          })}
        </div>
        <div ref={bottomRef} />
      </div>

      {conversation.statusMessage ? (
        <div className="border-t border-[#eee4db] bg-[#f4ebe3] px-4 py-3 text-sm text-[#7e6d61]">
          {conversation.statusMessage}
        </div>
      ) : null}

      <ChatInput disabled={!conversation.canSend} onSend={handleSend} />
    </div>
  )
}
