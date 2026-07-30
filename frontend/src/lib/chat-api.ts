import { apiRequest } from '@/lib/api'

export interface ConversationItem {
  id: number
  peerUserId: number
  peerNickname: string
  peerDepartment: string
  peerGrade: string
  lastMessage: string
  lastMessageAt?: string
  unreadCount: number
  canSend: boolean
  statusMessage?: string
  createdAt: string
}

export interface MessageItem {
  id: number
  conversationId: number
  senderId: number
  content: string
  contentType: string
  createdAt: string
}

export interface MessagesResult {
  items: MessageItem[]
  hasMore: boolean
}

export function fetchConversations() {
  return apiRequest<ConversationItem[]>('/conversations', { auth: true })
}

export function fetchMessages(
  conversationId: number,
  options?: { before?: number; limit?: number },
) {
  const params = new URLSearchParams()
  if (options?.before) params.set('before', String(options.before))
  if (options?.limit) params.set('limit', String(options.limit))
  const qs = params.toString()
  return apiRequest<MessagesResult>(
    `/conversations/${conversationId}/messages${qs ? `?${qs}` : ''}`,
    { auth: true },
  )
}

export function markConversationRead(conversationId: number) {
  return apiRequest<null>(`/conversations/${conversationId}/read`, {
    method: 'POST',
    auth: true,
  })
}

export function fetchUnreadCount() {
  return apiRequest<{ unreadCount: number }>('/conversations/unread-count', {
    auth: true,
  })
}
