import { apiRequest } from '@/lib/api'

export interface PlaygroundAuthor {
  label: string
  isMe: boolean
  isOwner: boolean
}

export type PlaygroundTopic =
  | 'campus'
  | 'partner'
  | 'study'
  | 'emotion'
  | 'secondhand'

export type PlaygroundSort = 'latest' | 'hot'

export interface PlaygroundComment {
  id: number
  postId: number
  author: PlaygroundAuthor
  parentCommentId?: number
  content: string
  likeCount: number
  replyCount: number
  floor: number
  likedByMe: boolean
  createdAt: string
  replies: PlaygroundComment[]
}

export interface PlaygroundPost {
  id: number
  author: PlaygroundAuthor
  topic: PlaygroundTopic
  content: string
  imageUrls: string[]
  likeCount: number
  commentCount: number
  likedByMe: boolean
  createdAt: string
  comments: PlaygroundComment[]
}

export interface PlaygroundLikeResult {
  liked: boolean
  likeCount: number
}

export function fetchPlaygroundPosts(options: {
  topic?: PlaygroundTopic
  sort?: PlaygroundSort
  query?: string
} = {}) {
  const params = new URLSearchParams()
  if (options.topic) params.set('topic', options.topic)
  if (options.sort) params.set('sort', options.sort)
  if (options.query?.trim()) params.set('q', options.query.trim())
  const query = params.size > 0 ? `?${params.toString()}` : ''
  return apiRequest<PlaygroundPost[]>(`/playground/posts${query}`, { auth: true })
}

export function createPlaygroundPost(input: {
  topic: PlaygroundTopic
  content: string
  imageUrls: string[]
}) {
  return apiRequest<PlaygroundPost>('/playground/posts', {
    method: 'POST',
    auth: true,
    body: input,
  })
}

export function likePlaygroundPost(postId: number) {
  return apiRequest<PlaygroundLikeResult>(`/playground/posts/${postId}/likes`, {
    method: 'POST',
    auth: true,
  })
}

export function unlikePlaygroundPost(postId: number) {
  return apiRequest<PlaygroundLikeResult>(`/playground/posts/${postId}/likes`, {
    method: 'DELETE',
    auth: true,
  })
}

export function createPlaygroundComment(input: {
  postId: number
  content: string
  parentCommentId?: number
}) {
  return apiRequest<PlaygroundComment>(`/playground/posts/${input.postId}/comments`, {
    method: 'POST',
    auth: true,
    body: {
      content: input.content,
      parentCommentId: input.parentCommentId ?? 0,
    },
  })
}

export function likePlaygroundComment(commentId: number) {
  return apiRequest<PlaygroundLikeResult>(`/playground/comments/${commentId}/likes`, {
    method: 'POST',
    auth: true,
  })
}

export function unlikePlaygroundComment(commentId: number) {
  return apiRequest<PlaygroundLikeResult>(`/playground/comments/${commentId}/likes`, {
    method: 'DELETE',
    auth: true,
  })
}
