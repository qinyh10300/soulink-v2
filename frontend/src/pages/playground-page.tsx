import {
  CalendarPlus,
  ChevronDown,
  Heart,
  ImagePlus,
  MessageCircle,
  RefreshCw,
  Reply,
  Search,
  Send,
  ShieldCheck,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState, type FormEvent } from 'react'

import { Button } from '@/components/ui/button'
import {
  CreatePlazaGroupInviteDialog,
  PlazaGroupInviteDialog,
  PlazaGroupInviteStrip,
} from '@/components/playground/plaza-group-invite'
import { getErrorMessage } from '@/lib/api'
import {
  applyToPlazaGroupInvite,
  createPlaygroundComment,
  createPlaygroundPost,
  fetchPlaygroundPosts,
  likePlaygroundComment,
  likePlaygroundPost,
  type PlaygroundComment,
  type PlaygroundPost,
  type PlaygroundSort,
  type PlaygroundTopic,
  unlikePlaygroundComment,
  unlikePlaygroundPost,
} from '@/lib/playground-api'
import { cn } from '@/lib/utils'

const topicOptions: Array<{
  value: PlaygroundTopic
  label: string
  accent: string
}> = [
  { value: 'campus', label: '校园', accent: 'bg-[#dce8e1] text-[#315d4f]' },
  { value: 'partner', label: '搭子', accent: 'bg-[#e7e1f0] text-[#5e4d75]' },
  { value: 'study', label: '学习', accent: 'bg-[#dfe8f2] text-[#365a77]' },
  { value: 'emotion', label: '情感', accent: 'bg-[#f1dfdf] text-[#834c4c]' },
  { value: 'secondhand', label: '闲置', accent: 'bg-[#eee5d3] text-[#705a35]' },
]

const topicByValue = Object.fromEntries(
  topicOptions.map((topic) => [topic.value, topic]),
) as Record<PlaygroundTopic, (typeof topicOptions)[number]>

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function parseImageUrls(value: string) {
  return value
    .split(/[\n,，]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function AnonymousBadge({ label, isMe }: { label: string; isMe: boolean }) {
  return (
    <span
      className={cn(
        'inline-flex h-6 shrink-0 items-center rounded px-2 text-xs font-semibold',
        label === '洞主'
          ? 'bg-[#303633] text-white'
          : 'bg-[#eceae6] text-[#5f5b55]',
      )}
    >
      {label}
      {isMe ? ' · 我' : ''}
    </span>
  )
}

interface CommentItemProps {
  comment: PlaygroundComment
  onChangeReply: (value: string) => void
  onLike: (comment: PlaygroundComment) => void
  onReply: (commentId: number) => void
  onSubmitReply: (comment: PlaygroundComment) => void
  replyTargetId: number | null
  replyText: string
}

function CommentItem({
  comment,
  onChangeReply,
  onLike,
  onReply,
  onSubmitReply,
  replyTargetId,
  replyText,
}: CommentItemProps) {
  const isReplying = replyTargetId === comment.id

  return (
    <div className="border-t border-[#e5e3df] py-3 first:border-t-0">
      <div className="flex min-w-0 items-start gap-2">
        <span className="mt-1 w-8 shrink-0 text-right font-mono text-[11px] text-[#9a958d]">
          {comment.floor}F
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <AnonymousBadge label={comment.author.label} isMe={comment.author.isMe} />
            <time className="text-[11px] text-[#99958e]">{formatDate(comment.createdAt)}</time>
          </div>
          <p className="mt-2 whitespace-pre-wrap break-words text-sm leading-6 text-[#34322f]">
            {comment.content}
          </p>
          <div className="mt-2 flex items-center gap-1 text-xs text-[#77736c]">
            <button
              aria-label={comment.likedByMe ? '取消赞同' : '赞同回复'}
              className={cn(
                'inline-flex h-7 items-center gap-1 rounded px-2 hover:bg-[#eceae6]',
                comment.likedByMe && 'text-[#a33f4a]',
              )}
              onClick={() => onLike(comment)}
              title={comment.likedByMe ? '取消赞同' : '赞同'}
              type="button"
            >
              <Heart className={cn('h-3.5 w-3.5', comment.likedByMe && 'fill-current')} />
              {comment.likeCount || ''}
            </button>
            <button
              className="inline-flex h-7 items-center gap-1 rounded px-2 hover:bg-[#eceae6]"
              onClick={() => onReply(comment.id)}
              type="button"
            >
              <Reply className="h-3.5 w-3.5" />
              回复
            </button>
          </div>

          {isReplying ? (
            <div className="mt-2 flex gap-2">
              <input
                aria-label={`回复 ${comment.author.label}`}
                className="h-9 min-w-0 flex-1 rounded border border-[#cbc7c0] bg-white px-3 text-sm outline-none focus:border-[#527465]"
                maxLength={1000}
                onChange={(event) => onChangeReply(event.target.value)}
                placeholder={`回复 ${comment.author.label}`}
                value={replyText}
              />
              <button
                aria-label="发送回复"
                className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded bg-[#2f4f43] text-white hover:bg-[#203b32]"
                onClick={() => onSubmitReply(comment)}
                title="发送回复"
                type="button"
              >
                <Send className="h-4 w-4" />
              </button>
            </div>
          ) : null}

          {comment.replies.length > 0 ? (
            <div className="mt-3 border-l-2 border-[#d7ddd9] pl-3">
              {comment.replies.map((reply) => (
                <CommentItem
                  comment={reply}
                  key={reply.id}
                  onChangeReply={onChangeReply}
                  onLike={onLike}
                  onReply={onReply}
                  onSubmitReply={onSubmitReply}
                  replyTargetId={replyTargetId}
                  replyText={replyText}
                />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function TopicTag({ topic }: { topic: PlaygroundTopic }) {
  const meta = topicByValue[topic]
  return (
    <span className={cn('inline-flex h-6 items-center rounded px-2 text-xs font-medium', meta.accent)}>
      {meta.label}
    </span>
  )
}

export function PlaygroundPage() {
  const [posts, setPosts] = useState<PlaygroundPost[]>([])
  const [nextCursor, setNextCursor] = useState<string | undefined>()
  const [content, setContent] = useState('')
  const [composerTopic, setComposerTopic] = useState<PlaygroundTopic>('campus')
  const [imageInput, setImageInput] = useState('')
  const [showImageInput, setShowImageInput] = useState(false)
  const [selectedTopic, setSelectedTopic] = useState<PlaygroundTopic | ''>('')
  const [sort, setSort] = useState<PlaygroundSort>('latest')
  const [searchDraft, setSearchDraft] = useState('')
  const [query, setQuery] = useState('')
  const [commentDrafts, setCommentDrafts] = useState<Record<number, string>>({})
  const [expandedPosts, setExpandedPosts] = useState<Set<number>>(new Set())
  const [replyTargetId, setReplyTargetId] = useState<number | null>(null)
  const [replyText, setReplyText] = useState('')
  const [notice, setNotice] = useState('正在加载树洞...')
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [createInvitePostId, setCreateInvitePostId] = useState<number | null>(null)
  const [activeInviteId, setActiveInviteId] = useState<number | null>(null)
  const [inviteActionId, setInviteActionId] = useState<number | null>(null)

  const loadPosts = useCallback(async () => {
    try {
      setLoading(true)
      const page = await fetchPlaygroundPosts({
        topic: selectedTopic || undefined,
        sort,
        query,
        limit: 12,
      })
      setPosts(page.items)
      setNextCursor(page.nextCursor)
      setNotice(page.items.length > 0 ? '' : '这里暂时还没有内容。')
    } catch (error) {
      setNotice(getErrorMessage(error, '树洞加载失败，请稍后重试'))
    } finally {
      setLoading(false)
    }
  }, [query, selectedTopic, sort])

  const handleLoadMore = async () => {
    if (!nextCursor || loadingMore) return
    try {
      setLoadingMore(true)
      const page = await fetchPlaygroundPosts({
        topic: selectedTopic || undefined,
        sort,
        query,
        cursor: nextCursor,
        limit: 12,
      })
      setPosts((current) => {
        const known = new Set(current.map((post) => post.id))
        return [...current, ...page.items.filter((post) => !known.has(post.id))]
      })
      setNextCursor(page.nextCursor)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '更多树洞加载失败，请稍后重试'))
    } finally {
      setLoadingMore(false)
    }
  }

  useEffect(() => {
    void loadPosts()
  }, [loadPosts])

  const topicCounts = useMemo(() => {
    const counts = new Map<PlaygroundTopic, number>()
    for (const post of posts) {
      counts.set(post.topic, (counts.get(post.topic) ?? 0) + 1)
    }
    return counts
  }, [posts])

  const handleSearch = (event: FormEvent) => {
    event.preventDefault()
    setQuery(searchDraft.trim())
  }

  const handleCreatePost = async () => {
    const imageUrls = parseImageUrls(imageInput)
    if (!content.trim() && imageUrls.length === 0) {
      setNotice('请先输入内容')
      return
    }

    try {
      setSubmitting(true)
      await createPlaygroundPost({
        topic: composerTopic,
        content: content.trim(),
        imageUrls,
      })
      setContent('')
      setImageInput('')
      setShowImageInput(false)
      setSelectedTopic('')
      setSort('latest')
      setQuery('')
      setSearchDraft('')
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '发布失败，请检查内容后重试'))
    } finally {
      setSubmitting(false)
    }
  }

  const handleTogglePostLike = async (post: PlaygroundPost) => {
    try {
      if (post.likedByMe) {
        await unlikePlaygroundPost(post.id)
      } else {
        await likePlaygroundPost(post.id)
      }
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '操作失败，请稍后重试'))
    }
  }

  const handleToggleCommentLike = async (comment: PlaygroundComment) => {
    try {
      if (comment.likedByMe) {
        await unlikePlaygroundComment(comment.id)
      } else {
        await likePlaygroundComment(comment.id)
      }
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '操作失败，请稍后重试'))
    }
  }

  const handleSubmitComment = async (postId: number) => {
    const draft = commentDrafts[postId]?.trim() ?? ''
    if (!draft) return
    try {
      await createPlaygroundComment({ postId, content: draft })
      setCommentDrafts((current) => ({ ...current, [postId]: '' }))
      setExpandedPosts((current) => new Set(current).add(postId))
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '回复失败，请稍后重试'))
    }
  }

  const handleSubmitReply = async (comment: PlaygroundComment) => {
    const draft = replyText.trim()
    if (!draft) return
    try {
      await createPlaygroundComment({
        postId: comment.postId,
        content: draft,
        parentCommentId: comment.id,
      })
      setReplyText('')
      setReplyTargetId(null)
      setExpandedPosts((current) => new Set(current).add(comment.postId))
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '回复失败，请稍后重试'))
    }
  }

  const handleApplyToInvite = async (inviteId: number) => {
    try {
      setInviteActionId(inviteId)
      setNotice('')
      await applyToPlazaGroupInvite(inviteId)
      await loadPosts()
      setActiveInviteId(inviteId)
    } catch (error) {
      setNotice(getErrorMessage(error, '申请失败，请稍后重试'))
    } finally {
      setInviteActionId(null)
    }
  }

  return (
    <div className="min-h-screen bg-[#eceeed] text-[#292825]">
      <header className="sticky top-15 z-20 border-b border-[#d5d7d4] bg-[rgba(249,250,249,0.94)] backdrop-blur-lg lg:top-0">
        <div className="mx-auto flex max-w-[1180px] items-center gap-3 px-4 py-3 sm:px-6">
          <div className="min-w-0 shrink-0">
            <h1 className="text-lg font-semibold text-[#242825]">清华树洞</h1>
            <p className="text-[11px] text-[#777b77]">THU · SOULINK</p>
          </div>
          <form className="relative ml-auto min-w-0 flex-1 sm:max-w-sm" onSubmit={handleSearch}>
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#8a8d89]" />
            <input
              aria-label="搜索树洞"
              className="h-9 w-full rounded border border-[#cfd2ce] bg-white pl-9 pr-9 text-sm outline-none focus:border-[#527465]"
              maxLength={80}
              onChange={(event) => setSearchDraft(event.target.value)}
              placeholder="搜索洞号或内容"
              value={searchDraft}
            />
            {searchDraft ? (
              <button
                aria-label="清空搜索"
                className="absolute right-1 top-1/2 inline-flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded text-[#777b77] hover:bg-[#eceeed]"
                onClick={() => {
                  setSearchDraft('')
                  setQuery('')
                }}
                title="清空搜索"
                type="button"
              >
                <X className="h-4 w-4" />
              </button>
            ) : null}
          </form>
          <button
            aria-label="刷新树洞"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded border border-[#cfd2ce] bg-white text-[#555a56] hover:bg-[#f2f3f2] disabled:opacity-50"
            disabled={loading}
            onClick={() => void loadPosts()}
            title="刷新"
            type="button"
          >
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      <main className="mx-auto grid max-w-[1180px] grid-cols-1 gap-6 px-4 py-5 sm:px-6 lg:grid-cols-[minmax(0,760px)_280px]">
        <div className="min-w-0 space-y-4">
          <section className="rounded-lg border border-[#d6d6d2] bg-[#fafaf8] shadow-[0_2px_8px_rgba(35,38,35,0.06)]">
            <div className="flex flex-wrap items-center gap-2 border-b border-[#e0dfdb] px-4 py-3">
              <AnonymousBadge label="匿名发布" isMe={false} />
              <select
                aria-label="树洞主题"
                className="h-8 rounded border border-[#cbc9c3] bg-white px-2 text-xs text-[#4e4b46] outline-none focus:border-[#527465]"
                onChange={(event) => setComposerTopic(event.target.value as PlaygroundTopic)}
                value={composerTopic}
              >
                {topicOptions.map((topic) => (
                  <option key={topic.value} value={topic.value}>
                    {topic.label}
                  </option>
                ))}
              </select>
              <span className="ml-auto text-xs text-[#96928b]">{content.length}/2000</span>
            </div>
            <textarea
              aria-label="树洞内容"
              className="min-h-24 w-full resize-y bg-transparent px-4 py-3 text-[15px] leading-7 outline-none placeholder:text-[#9b9891]"
              maxLength={2000}
              onChange={(event) => setContent(event.target.value)}
              placeholder="此刻想说什么？"
              value={content}
            />
            {showImageInput ? (
              <div className="mx-4 mb-3 flex items-center gap-2 border-t border-[#e5e3df] pt-3">
                <ImagePlus className="h-4 w-4 shrink-0 text-[#777b77]" />
                <input
                  aria-label="图片链接"
                  className="h-9 min-w-0 flex-1 rounded border border-[#cbc9c3] bg-white px-3 text-sm outline-none focus:border-[#527465]"
                  onChange={(event) => setImageInput(event.target.value)}
                  placeholder="图片链接，多个链接用逗号分隔"
                  value={imageInput}
                />
              </div>
            ) : null}
            <div className="flex items-center justify-between border-t border-[#e0dfdb] px-3 py-2">
              <button
                aria-label="添加图片链接"
                className={cn(
                  'inline-flex h-9 w-9 items-center justify-center rounded text-[#626660] hover:bg-[#eceeeb]',
                  showImageInput && 'bg-[#e4ebe7] text-[#315d4f]',
                )}
                onClick={() => setShowImageInput((current) => !current)}
                title="添加图片"
                type="button"
              >
                <ImagePlus className="h-4 w-4" />
              </button>
              <Button disabled={submitting} onClick={() => void handleCreatePost()} size="sm">
                <Send className="mr-2 h-4 w-4" />
                {submitting ? '发布中' : '发布'}
              </Button>
            </div>
          </section>

          <div className="flex flex-col gap-3 border-y border-[#d7d9d6] py-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex gap-1 overflow-x-auto pb-1 sm:pb-0">
              <button
                className={cn(
                  'h-8 shrink-0 rounded px-3 text-xs font-medium',
                  selectedTopic === '' ? 'bg-[#303633] text-white' : 'text-[#656862] hover:bg-[#dfe1de]',
                )}
                onClick={() => setSelectedTopic('')}
                type="button"
              >
                全部
              </button>
              {topicOptions.map((topic) => (
                <button
                  className={cn(
                    'h-8 shrink-0 rounded px-3 text-xs font-medium',
                    selectedTopic === topic.value
                      ? 'bg-[#303633] text-white'
                      : 'text-[#656862] hover:bg-[#dfe1de]',
                  )}
                  key={topic.value}
                  onClick={() => setSelectedTopic(topic.value)}
                  type="button"
                >
                  {topic.label}
                </button>
              ))}
            </div>
            <div className="inline-flex h-8 w-fit rounded border border-[#cfd1ce] bg-white p-0.5">
              {(['latest', 'hot'] as PlaygroundSort[]).map((value) => (
                <button
                  className={cn(
                    'rounded px-3 text-xs',
                    sort === value ? 'bg-[#e3e9e5] font-semibold text-[#315d4f]' : 'text-[#77736c]',
                  )}
                  key={value}
                  onClick={() => setSort(value)}
                  type="button"
                >
                  {value === 'latest' ? '最新' : '热议'}
                </button>
              ))}
            </div>
          </div>

          {notice ? (
            <div className="rounded-lg border border-[#d6d6d2] bg-[#fafaf8] px-4 py-3 text-sm text-[#676a65]">
              {notice}
            </div>
          ) : null}

          <section className="space-y-3">
            {posts.map((post) => {
              const expanded = expandedPosts.has(post.id)
              const visibleComments = expanded ? post.comments : post.comments.slice(0, 2)
              return (
                <article
                  className="rounded-lg border border-[#d6d6d2] bg-[#fafaf8] shadow-[0_2px_8px_rgba(35,38,35,0.06)]"
                  id={`hole-${post.id}`}
                  key={post.id}
                >
                  <header className="flex flex-wrap items-center gap-2 border-b border-[#e0dfdb] px-4 py-3">
                    <a
                      className="font-mono text-sm font-semibold text-[#4f665d] hover:underline"
                      href={`#hole-${post.id}`}
                    >
                      #{post.id}
                    </a>
                    <TopicTag topic={post.topic} />
                    <AnonymousBadge label={post.author.label} isMe={post.author.isMe} />
                    <time className="ml-auto text-xs text-[#94908a]">{formatDate(post.createdAt)}</time>
                  </header>

                  <div className="px-4 py-4">
                    {post.content ? (
                      <p className="whitespace-pre-wrap break-words text-[15px] leading-7 text-[#2f2e2b]">
                        {post.content}
                      </p>
                    ) : null}
                    {post.imageUrls.length > 0 ? (
                      <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
                        {post.imageUrls.map((imageUrl) => (
                          <img
                            alt="树洞配图"
                            className="max-h-80 w-full rounded-md border border-[#deddd9] object-cover"
                            key={imageUrl}
                            loading="lazy"
                            src={imageUrl}
                          />
                        ))}
                      </div>
                    ) : null}
                  </div>

                  {post.groupInvite ? (
                    <PlazaGroupInviteStrip
                      busy={inviteActionId === post.groupInvite.id}
                      invite={post.groupInvite}
                      onApply={() => void handleApplyToInvite(post.groupInvite!.id)}
                      onOpen={() => setActiveInviteId(post.groupInvite!.id)}
                    />
                  ) : null}

                  <div className="flex items-center gap-1 border-t border-[#e0dfdb] px-3 py-2 text-sm text-[#666963]">
                    <button
                      aria-label={post.likedByMe ? '取消赞同' : '赞同树洞'}
                      className={cn(
                        'inline-flex h-8 items-center gap-1.5 rounded px-2 hover:bg-[#eceeeb]',
                        post.likedByMe && 'text-[#a33f4a]',
                      )}
                      onClick={() => void handleTogglePostLike(post)}
                      title={post.likedByMe ? '取消赞同' : '赞同'}
                      type="button"
                    >
                      <Heart className={cn('h-4 w-4', post.likedByMe && 'fill-current')} />
                      {post.likeCount}
                    </button>
                    <span className="inline-flex h-8 items-center gap-1.5 px-2">
                      <MessageCircle className="h-4 w-4" />
                      {post.commentCount}
                    </span>
                    {post.author.isMe && !post.groupInvite ? (
                      <button
                        className="ml-auto inline-flex h-8 items-center gap-1.5 rounded px-2 text-xs font-medium text-[#42675a] hover:bg-[#e5ebe7]"
                        onClick={() => setCreateInvitePostId(post.id)}
                        type="button"
                      >
                        <CalendarPlus className="h-4 w-4" />
                        转为连接
                      </button>
                    ) : null}
                  </div>

                  <div className="border-t border-[#e0dfdb] bg-[#f5f5f2] px-4 py-3">
                    {visibleComments.length > 0 ? (
                      <div className="mb-3">
                        {visibleComments.map((comment) => (
                          <CommentItem
                            comment={comment}
                            key={comment.id}
                            onChangeReply={setReplyText}
                            onLike={(target) => void handleToggleCommentLike(target)}
                            onReply={(commentId) => {
                              setReplyTargetId((current) => (current === commentId ? null : commentId))
                              setReplyText('')
                            }}
                            onSubmitReply={(target) => void handleSubmitReply(target)}
                            replyTargetId={replyTargetId}
                            replyText={replyText}
                          />
                        ))}
                      </div>
                    ) : null}
                    {!expanded && post.comments.length > 2 ? (
                      <button
                        className="mb-3 inline-flex items-center gap-1 text-xs font-medium text-[#45685b] hover:underline"
                        onClick={() => setExpandedPosts((current) => new Set(current).add(post.id))}
                        type="button"
                      >
                        查看全部 {post.commentCount} 条回复
                        <ChevronDown className="h-3.5 w-3.5" />
                      </button>
                    ) : null}
                    <div className="flex gap-2">
                      <input
                        aria-label={`回复树洞 ${post.id}`}
                        className="h-9 min-w-0 flex-1 rounded border border-[#cbc9c3] bg-white px-3 text-sm outline-none focus:border-[#527465]"
                        maxLength={1000}
                        onChange={(event) =>
                          setCommentDrafts((current) => ({
                            ...current,
                            [post.id]: event.target.value,
                          }))
                        }
                        onKeyDown={(event) => {
                          if (event.key === 'Enter' && !event.shiftKey) {
                            event.preventDefault()
                            void handleSubmitComment(post.id)
                          }
                        }}
                        placeholder="匿名回复"
                        value={commentDrafts[post.id] ?? ''}
                      />
                      <button
                        aria-label="发送回复"
                        className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded bg-[#2f4f43] text-white hover:bg-[#203b32]"
                        onClick={() => void handleSubmitComment(post.id)}
                        title="发送回复"
                        type="button"
                      >
                        <Send className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </article>
              )
            })}
          </section>
          {nextCursor ? (
            <button
              className="flex h-11 w-full items-center justify-center gap-2 rounded-md border border-[#cfd2ce] bg-[#fafaf8] text-sm font-medium text-[#4e625a] hover:bg-white disabled:opacity-50"
              disabled={loadingMore}
              onClick={() => void handleLoadMore()}
              type="button"
            >
              <RefreshCw className={cn('h-4 w-4', loadingMore && 'animate-spin')} />
              {loadingMore ? '加载中' : '加载更多'}
            </button>
          ) : null}
        </div>

        <aside className="hidden lg:block">
          <div className="sticky top-20 space-y-6">
            <section className="border-t-2 border-[#39443f] pt-3">
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 text-[#416557]" />
                <h2 className="text-sm font-semibold">树洞公约</h2>
              </div>
              <ol className="mt-3 space-y-2 text-xs leading-5 text-[#696c67]">
                <li>1. 只回应内容，不猜测身份</li>
                <li>2. 不发布联系方式和他人隐私</li>
                <li>3. 线下见面先确认边界与地点</li>
              </ol>
            </section>
            <section className="border-t border-[#cfd2ce] pt-3">
              <h2 className="text-sm font-semibold">此刻在聊</h2>
              <div className="mt-3 space-y-2">
                {topicOptions
                  .slice()
                  .sort((left, right) => (topicCounts.get(right.value) ?? 0) - (topicCounts.get(left.value) ?? 0))
                  .map((topic) => (
                    <button
                      className="flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-xs text-[#5e625d] hover:bg-[#dfe1de]"
                      key={topic.value}
                      onClick={() => setSelectedTopic(topic.value)}
                      type="button"
                    >
                      <span>#{topic.label}</span>
                      <span className="font-mono text-[#999c97]">{topicCounts.get(topic.value) ?? 0}</span>
                    </button>
                  ))}
              </div>
            </section>
          </div>
        </aside>
      </main>
      {createInvitePostId ? (
        <CreatePlazaGroupInviteDialog
          onClose={() => setCreateInvitePostId(null)}
          onCreated={(invite) => {
            setCreateInvitePostId(null)
            setActiveInviteId(invite.id)
            void loadPosts()
          }}
          postId={createInvitePostId}
        />
      ) : null}
      {activeInviteId ? (
        <PlazaGroupInviteDialog
          inviteId={activeInviteId}
          key={activeInviteId}
          onChanged={() => void loadPosts()}
          onClose={() => setActiveInviteId(null)}
        />
      ) : null}
    </div>
  )
}
