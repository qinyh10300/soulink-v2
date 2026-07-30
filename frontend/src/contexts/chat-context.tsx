import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'

import { fetchUnreadCount, type MessageItem } from '@/lib/chat-api'
import { fetchReceivedRequests } from '@/lib/friend-api'
import { SoulinkWebSocket, type WSEnvelope } from '@/lib/websocket'

interface ChatContextValue {
  ws: SoulinkWebSocket | null
  totalUnread: number
  pendingRequests: number
  refreshUnread: () => void
  refreshPendingRequests: () => void
  lastMessage: MessageItem | null
  conversationVersion: number
}

const ChatContext = createContext<ChatContextValue>({
  ws: null,
  totalUnread: 0,
  pendingRequests: 0,
  refreshUnread: () => {},
  refreshPendingRequests: () => {},
  lastMessage: null,
  conversationVersion: 0,
})

export function ChatProvider({ children }: { children: ReactNode }) {
  const [ws, setWs] = useState<SoulinkWebSocket | null>(null)
  const [totalUnread, setTotalUnread] = useState(0)
  const [pendingRequests, setPendingRequests] = useState(0)
  const [lastMessage, setLastMessage] = useState<MessageItem | null>(null)
  const [conversationVersion, setConversationVersion] = useState(0)

  const refreshUnread = useCallback(() => {
    fetchUnreadCount()
      .then((data) => setTotalUnread(data.unreadCount))
      .catch(() => {})
  }, [])

  const refreshPendingRequests = useCallback(() => {
    fetchReceivedRequests()
      .then((reqs) => setPendingRequests(reqs.length))
      .catch(() => {})
  }, [])

  useEffect(() => {
    const instance = new SoulinkWebSocket()
    setWs(instance)
    instance.connect()
    refreshUnread()
    refreshPendingRequests()

    const unsubscribe = instance.onEvent((env: WSEnvelope) => {
      if (env.type === 'message') {
        const msg = env.data as MessageItem
        setLastMessage(msg)
        refreshUnread()
      }
      if (env.type === 'read') {
        refreshUnread()
      }
      if (env.type === 'friend_request' || env.type === 'friend_accepted') {
        refreshPendingRequests()
      }
      if (env.type === 'conversation_updated') {
        setConversationVersion((current) => current + 1)
      }
    })

    return () => {
      unsubscribe()
      instance.disconnect()
      setWs(null)
    }
  }, [refreshUnread, refreshPendingRequests])

  return (
    <ChatContext.Provider
      value={{
        ws,
        totalUnread,
        pendingRequests,
        refreshUnread,
        refreshPendingRequests,
        lastMessage,
        conversationVersion,
      }}
    >
      {children}
    </ChatContext.Provider>
  )
}

export function useChat() {
  return useContext(ChatContext)
}
