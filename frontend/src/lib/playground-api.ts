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
  groupInvite?: PlazaGroupInviteSummary
}

export interface PlaygroundPostPage {
  items: PlaygroundPost[]
  nextCursor?: string
}

export type PlazaGroupInviteType = 'event' | 'group'
export type PlazaGroupRevealStrategy = 'on_accept' | 'on_form'
export type PlazaGroupInviteStatus =
  | 'open'
  | 'formed'
  | 'expired'
  | 'archived'
  | 'kept'
  | 'cancelled'
export type PlazaGroupApplicationStatus =
  | 'applied'
  | 'accepted'
  | 'declined'
  | 'waitlisted'
  | 'left'

export interface PlazaGroupIdentity {
  alias: string
  revealed: boolean
  userId?: number
  nickname?: string
  department?: string
  grade?: string
}

export interface PlazaGroupApplication {
  id: number
  status: PlazaGroupApplicationStatus
  identity: PlazaGroupIdentity
  createdAt: string
}

export interface PlazaGroupInviteSummary {
  id: number
  postId: number
  inviteType: PlazaGroupInviteType
  title: string
  summary: string
  extractedTime: string
  extractedLocation: string
  minMembers: number
  maxMembers: number
  memberCount: number
  applicationCount: number
  revealStrategy: PlazaGroupRevealStrategy
  deadline: string
  groupExpiresAt: string
  status: PlazaGroupInviteStatus
  isOwner: boolean
  canApply: boolean
  viewerApplicationStatus?: PlazaGroupApplicationStatus
  inferenceSource: 'model' | 'deterministic'
  inferenceConfidence: number
  promptVersion: string
}

export interface PlazaGroupInvite extends PlazaGroupInviteSummary {
  owner: PlazaGroupIdentity
  applications: PlazaGroupApplication[]
  chatGroupId?: number
  canForm: boolean
  canKeep: boolean
  keepVotes: number
  requiredKeepVotes: number
}

export interface PlaygroundLikeResult {
  liked: boolean
  likeCount: number
}

export function fetchPlaygroundPosts(options: {
  topic?: PlaygroundTopic
  sort?: PlaygroundSort
  query?: string
  cursor?: string
  limit?: number
} = {}) {
  const params = new URLSearchParams()
  if (options.topic) params.set('topic', options.topic)
  if (options.sort) params.set('sort', options.sort)
  if (options.query?.trim()) params.set('q', options.query.trim())
  if (options.cursor) params.set('cursor', options.cursor)
  if (options.limit) params.set('limit', String(options.limit))
  const query = params.size > 0 ? `?${params.toString()}` : ''
  return apiRequest<PlaygroundPostPage>(`/playground/posts/page${query}`, { auth: true })
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

export function createPlazaGroupInvite(input: {
  postId: number
  inviteType: PlazaGroupInviteType
  title?: string
  summary?: string
  minMembers: number
  maxMembers: number
  revealStrategy: PlazaGroupRevealStrategy
  deadline: string
  groupExpiresAt: string
}) {
  return apiRequest<PlazaGroupInvite>(`/playground/posts/${input.postId}/group-invite`, {
    method: 'POST',
    auth: true,
    body: {
      inviteType: input.inviteType,
      title: input.title ?? '',
      summary: input.summary ?? '',
      minMembers: input.minMembers,
      maxMembers: input.maxMembers,
      revealStrategy: input.revealStrategy,
      deadline: input.deadline,
      groupExpiresAt: input.groupExpiresAt,
    },
  })
}

export function fetchMyPlazaGroupInvites() {
  return apiRequest<PlazaGroupInvite[]>('/playground/group-invites/mine', { auth: true })
}

export function fetchPlazaGroupInvite(inviteId: number) {
  return apiRequest<PlazaGroupInvite>(`/playground/group-invites/${inviteId}`, { auth: true })
}

export function applyToPlazaGroupInvite(inviteId: number) {
  return apiRequest<PlazaGroupInvite>(`/playground/group-invites/${inviteId}/applications`, {
    method: 'POST',
    auth: true,
  })
}

export function respondToPlazaGroupApplication(
  inviteId: number,
  applicationId: number,
  status: 'accepted' | 'declined',
) {
  return apiRequest<PlazaGroupInvite>(
    `/playground/group-invites/${inviteId}/applications/${applicationId}`,
    { method: 'PUT', auth: true, body: { status } },
  )
}

export function formPlazaGroupInvite(inviteId: number) {
  return apiRequest<PlazaGroupInvite>(`/playground/group-invites/${inviteId}/form`, {
    method: 'POST',
    auth: true,
  })
}

export function keepPlazaGroupInvite(inviteId: number) {
  return apiRequest<PlazaGroupInvite>(`/playground/group-invites/${inviteId}/keep`, {
    method: 'POST',
    auth: true,
  })
}
