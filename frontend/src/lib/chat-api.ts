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

export interface GroupMemberItem {
  userId: number
  nickname: string
  department: string
  grade: string
  role: 'owner' | 'member'
  isMe: boolean
  joinedAt: string
}

export interface GroupItem {
  id: number
  name: string
  ownerId: number
  isOwner: boolean
  memberCount: number
  members: GroupMemberItem[]
  lastMessage: string
  lastMessageAt?: string
  lastMessageSender?: string
  unreadCount: number
  createdAt: string
}

export interface GroupMessageItem {
  id: number
  groupId: number
  senderId: number
  senderNickname: string
  content: string
  contentType: string
  createdAt: string
}

export interface GroupMessagesResult {
  items: GroupMessageItem[]
  hasMore: boolean
}

export function fetchConversations() {
  return apiRequest<ConversationItem[]>('/conversations', { auth: true })
}

export function ensureConversation(peerUserId: number) {
  return apiRequest<ConversationItem>('/conversations', {
    method: 'POST',
    auth: true,
    body: { peerUserId },
  })
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

export function sendConversationMessage(conversationId: number, content: string) {
  return apiRequest<MessageItem>(`/conversations/${conversationId}/messages`, {
    method: 'POST',
    auth: true,
    body: { content, contentType: 'text' },
  })
}

export function fetchGroups() {
  return apiRequest<GroupItem[]>('/chat-groups', { auth: true })
}

export function fetchGroup(groupId: number) {
  return apiRequest<GroupItem>(`/chat-groups/${groupId}`, { auth: true })
}

export function createGroup(input: { name: string; memberIds: number[] }) {
  return apiRequest<GroupItem>('/chat-groups', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function renameGroup(groupId: number, name: string) {
  return apiRequest<GroupItem>(`/chat-groups/${groupId}`, {
    method: 'PUT',
    auth: true,
    body: { name },
  })
}

export function addGroupMembers(groupId: number, memberIds: number[]) {
  return apiRequest<GroupItem>(`/chat-groups/${groupId}/members`, {
    method: 'POST',
    auth: true,
    body: { memberIds },
  })
}

export function removeGroupMember(groupId: number, userId: number) {
  return apiRequest<null>(`/chat-groups/${groupId}/members/${userId}`, {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchGroupMessages(
  groupId: number,
  options?: { before?: number; limit?: number },
) {
  const params = new URLSearchParams()
  if (options?.before) params.set('before', String(options.before))
  if (options?.limit) params.set('limit', String(options.limit))
  const query = params.size > 0 ? `?${params.toString()}` : ''
  return apiRequest<GroupMessagesResult>(`/chat-groups/${groupId}/messages${query}`, {
    auth: true,
  })
}

export function sendGroupMessage(groupId: number, content: string) {
  return apiRequest<GroupMessageItem>(`/chat-groups/${groupId}/messages`, {
    method: 'POST',
    auth: true,
    body: { content, contentType: 'text' },
  })
}

export function markGroupRead(groupId: number) {
  return apiRequest<null>(`/chat-groups/${groupId}/read`, {
    method: 'POST',
    auth: true,
  })
}

export function fetchUnreadCount() {
  return apiRequest<{ unreadCount: number }>('/conversations/unread-count', {
    auth: true,
  })
}
