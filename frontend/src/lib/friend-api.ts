import { apiRequest } from '@/lib/api'

export interface FriendUser {
  userId: number
  nickname: string
  email: string
  department: string
  grade: string
}

export interface FriendItem {
  friendshipId: number
  user: FriendUser
  createdAt: string
}

export interface FriendRequest {
  friendshipId: number
  user: FriendUser
  createdAt: string
}

export function sendFriendRequest(nickname: string) {
  return apiRequest<{ friendshipId: number }>('/friends/request', {
    method: 'POST',
    auth: true,
    body: { nickname },
  })
}

export function acceptFriendRequest(friendshipId: number) {
  return apiRequest<{ friendshipId: number; conversationId: number }>(
    '/friends/accept',
    {
      method: 'POST',
      auth: true,
      body: { friendshipId },
    },
  )
}

export function rejectFriendRequest(friendshipId: number) {
  return apiRequest<null>('/friends/reject', {
    method: 'POST',
    auth: true,
    body: { friendshipId },
  })
}

export function fetchFriends() {
  return apiRequest<FriendItem[]>('/friends', { auth: true })
}

export function fetchReceivedRequests() {
  return apiRequest<FriendRequest[]>('/friends/requests/received', {
    auth: true,
  })
}

export function fetchSentRequests() {
  return apiRequest<FriendRequest[]>('/friends/requests/sent', { auth: true })
}

export function removeFriend(friendshipId: number) {
  return apiRequest<null>(`/friends/${friendshipId}`, {
    method: 'DELETE',
    auth: true,
  })
}
