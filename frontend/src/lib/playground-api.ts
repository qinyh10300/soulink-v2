import { apiRequest } from '@/lib/api'

export interface PlaygroundAuthor {
  userId: number
  nickname: string
  schoolName: string
  department: string
  grade: string
}

export interface PlaygroundComment {
  id: number
  postId: number
  author: PlaygroundAuthor
  parentCommentId?: number
  content: string
  likeCount: number
  replyCount: number
  likedByMe: boolean
  createdAt: string
  replies: PlaygroundComment[]
}

export interface PlaygroundPost {
  id: number
  author: PlaygroundAuthor
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

export function fetchPlaygroundPosts(authorId?: number) {
  const query = authorId ? `?authorId=${encodeURIComponent(String(authorId))}` : ''
  return apiRequest<PlaygroundPost[]>(`/playground/posts${query}`, { auth: true })
}

export function createPlaygroundPost(input: { content: string; imageUrls: string[] }) {
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
