import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from 'react'

import {
  fetchUnreadCount,
  type GroupMessageItem,
  type MessageItem,
} from '@/lib/chat-api'
import { fetchReceivedRequests } from '@/lib/friend-api'
import { SoulinkWebSocket, type WSEnvelope } from '@/lib/websocket'
import { useEngagement } from '@/contexts/engagement-context'

interface ChatContextValue {
  ws: SoulinkWebSocket | null
  totalUnread: number
  pendingRequests: number
  refreshUnread: () => void
  refreshPendingRequests: () => void
  lastMessage: MessageItem | null
  lastGroupMessage: GroupMessageItem | null
  conversationVersion: number
}

const ChatContext = createContext<ChatContextValue>({
  ws: null,
  totalUnread: 0,
  pendingRequests: 0,
  refreshUnread: () => {},
  refreshPendingRequests: () => {},
  lastMessage: null,
  lastGroupMessage: null,
  conversationVersion: 0,
})

export function ChatProvider({ children }: { children: ReactNode }) {
  const { refreshNotifications } = useEngagement()
  const [ws, setWs] = useState<SoulinkWebSocket | null>(null)
  const [totalUnread, setTotalUnread] = useState(0)
  const [pendingRequests, setPendingRequests] = useState(0)
  const [lastMessage, setLastMessage] = useState<MessageItem | null>(null)
  const [lastGroupMessage, setLastGroupMessage] = useState<GroupMessageItem | null>(null)
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
        void refreshNotifications()
      }
      if (env.type === 'group_message') {
        const message = env.data as GroupMessageItem
        setLastGroupMessage(message)
        refreshUnread()
        void refreshNotifications()
      }
      if (env.type === 'read') {
        refreshUnread()
      }
      if (env.type === 'friend_request' || env.type === 'friend_accepted') {
        refreshPendingRequests()
      }
      if (env.type === 'conversation_updated' || env.type === 'group_updated') {
        setConversationVersion((current) => current + 1)
      }
    })

    return () => {
      unsubscribe()
      instance.disconnect()
      setWs(null)
    }
  }, [refreshNotifications, refreshUnread, refreshPendingRequests])

  return (
    <ChatContext.Provider
      value={{
        ws,
        totalUnread,
        pendingRequests,
        refreshUnread,
        refreshPendingRequests,
        lastMessage,
        lastGroupMessage,
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
