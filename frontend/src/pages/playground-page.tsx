import {
  ArrowLeft,
  Heart,
  ImagePlus,
  MessageCircle,
  RefreshCw,
  Reply,
  Send,
  UserRound,
} from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { Button } from '@/components/ui/button'
import { getErrorMessage } from '@/lib/api'
import {
  createPlaygroundComment,
  createPlaygroundPost,
  fetchPlaygroundPosts,
  likePlaygroundComment,
  likePlaygroundPost,
  type PlaygroundComment,
  type PlaygroundPost,
  unlikePlaygroundComment,
  unlikePlaygroundPost,
} from '@/lib/playground-api'
import { cn } from '@/lib/utils'

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}

function parseImageUrls(value: string) {
  return value
    .split(/[\n,，]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function Avatar({ name }: { name: string }) {
  return (
    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#f1e2d4] text-sm font-semibold text-[#8f5c42]">
      {name.trim().slice(0, 1) || <UserRound className="h-4 w-4" />}
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
    <div className="space-y-3">
      <div className="flex gap-3">
        <Link to={`/profile/${comment.author.userId}`}>
          <Avatar name={comment.author.nickname} />
        </Link>
        <div className="min-w-0 flex-1 rounded-[24px] bg-white/82 px-4 py-3">
          <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
            <Link
              className="font-semibold text-[#2f221b] hover:text-[#8f5c42]"
              to={`/profile/${comment.author.userId}`}
            >
              {comment.author.nickname}
            </Link>
            <span className="text-xs text-[#a08d81]">{formatDate(comment.createdAt)}</span>
          </div>
          <p className="mt-2 whitespace-pre-wrap break-words text-sm leading-7 text-[#4d4038]">
            {comment.content}
          </p>
          <div className="mt-3 flex flex-wrap items-center gap-2 text-xs text-[#8a7467]">
            <button
              className={cn(
                'inline-flex items-center gap-1 rounded-full px-2 py-1 transition-colors hover:bg-[#f3ece4]',
                comment.likedByMe && 'text-[#b94848]',
              )}
              onClick={() => onLike(comment)}
              type="button"
            >
              <Heart className={cn('h-3.5 w-3.5', comment.likedByMe && 'fill-current')} />
              {comment.likeCount}
            </button>
            <button
              className="inline-flex items-center gap-1 rounded-full px-2 py-1 transition-colors hover:bg-[#f3ece4]"
              onClick={() => onReply(comment.id)}
              type="button"
            >
              <Reply className="h-3.5 w-3.5" />
              回复
            </button>
          </div>
        </div>
      </div>

      {isReplying ? (
        <div className="ml-12 flex gap-2">
          <input
            className="min-h-10 flex-1 rounded-full border border-[#eadfd5] bg-white px-4 text-sm outline-none focus:border-[#c49a78]"
            onChange={(event) => onChangeReply(event.target.value)}
            placeholder={`回复 ${comment.author.nickname}`}
            value={replyText}
          />
          <button
            className="inline-flex h-10 items-center justify-center rounded-full bg-[#2f221b] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#4b372c]"
            onClick={() => onSubmitReply(comment)}
            type="button"
          >
            发送
          </button>
        </div>
      ) : null}

      {comment.replies.length > 0 ? (
        <div className="ml-12 space-y-3 border-l border-[#eadfd5] pl-4">
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
  )
}

export function PlaygroundPage() {
  const [posts, setPosts] = useState<PlaygroundPost[]>([])
  const [content, setContent] = useState('')
  const [imageInput, setImageInput] = useState('')
  const [commentDrafts, setCommentDrafts] = useState<Record<number, string>>({})
  const [replyTargetId, setReplyTargetId] = useState<number | null>(null)
  const [replyText, setReplyText] = useState('')
  const [notice, setNotice] = useState('正在加载广场内容...')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)

  const loadPosts = useCallback(async () => {
    try {
      setLoading(true)
      const nextPosts = await fetchPlaygroundPosts()
      setPosts(nextPosts)
      setNotice(nextPosts.length > 0 ? '' : '还没有帖子，来发布第一条动态吧。')
    } catch (error) {
      setNotice(getErrorMessage(error, '广场加载失败，请稍后重试'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadPosts()
  }, [loadPosts])

  const handleCreatePost = async () => {
    const imageUrls = parseImageUrls(imageInput)
    if (!content.trim() && imageUrls.length === 0) {
      setNotice('请先输入文字或图片链接')
      return
    }

    try {
      setSubmitting(true)
      await createPlaygroundPost({ content: content.trim(), imageUrls })
      setContent('')
      setImageInput('')
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
      setNotice(getErrorMessage(error, '点赞失败，请稍后重试'))
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
      setNotice(getErrorMessage(error, '评论点赞失败，请稍后重试'))
    }
  }

  const handleSubmitComment = async (postId: number) => {
    const draft = commentDrafts[postId]?.trim() ?? ''
    if (!draft) {
      return
    }
    try {
      await createPlaygroundComment({ postId, content: draft })
      setCommentDrafts((current) => ({ ...current, [postId]: '' }))
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '评论失败，请稍后重试'))
    }
  }

  const handleSubmitReply = async (comment: PlaygroundComment) => {
    const draft = replyText.trim()
    if (!draft) {
      return
    }
    try {
      await createPlaygroundComment({
        postId: comment.postId,
        content: draft,
        parentCommentId: comment.id,
      })
      setReplyText('')
      setReplyTargetId(null)
      await loadPosts()
    } catch (error) {
      setNotice(getErrorMessage(error, '回复失败，请稍后重试'))
    }
  }

  return (
    <div className="min-h-screen bg-[#f7f2ec] px-4 pb-12 pt-6 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Link
              className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#eee4db] bg-white/80 text-[#5f4d42] transition-colors hover:bg-white"
              to="/dashboard"
            >
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <div>
              <h1 className="text-2xl font-semibold text-[#2f221b]">广场</h1>
              <p className="text-sm text-[#8a7467]">分享校园生活，给感兴趣的人点赞和评论。</p>
            </div>
          </div>
          <Button disabled={loading} onClick={() => void loadPosts()} size="sm" variant="secondary">
            <RefreshCw className="mr-1 h-4 w-4" />
            刷新
          </Button>
        </div>

        <section className="rounded-[30px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-5 shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          <textarea
            className="min-h-28 w-full resize-none rounded-[24px] border border-[#eadfd5] bg-white/90 p-4 text-sm leading-7 text-[#2f221b] outline-none transition-colors placeholder:text-[#b6a599] focus:border-[#c49a78]"
            maxLength={2000}
            onChange={(event) => setContent(event.target.value)}
            placeholder="有什么想分享的？"
            value={content}
          />
          <div className="mt-3 flex items-center gap-2 rounded-[22px] border border-dashed border-[#e0d1c4] bg-[#fbf4ed] px-4 py-3 text-sm text-[#7b675b]">
            <ImagePlus className="h-4 w-4 shrink-0" />
            <input
              className="min-w-0 flex-1 bg-transparent outline-none placeholder:text-[#b6a599]"
              onChange={(event) => setImageInput(event.target.value)}
              placeholder="图片链接，用逗号分隔，最多 9 张"
              value={imageInput}
            />
          </div>
          <div className="mt-4 flex items-center justify-between gap-3">
            <p className="text-xs text-[#a08d81]">当前版本使用图片 URL；后续可接入对象存储上传。</p>
            <Button disabled={submitting} onClick={() => void handleCreatePost()}>
              <Send className="mr-2 h-4 w-4" />
              {submitting ? '发布中...' : '发布'}
            </Button>
          </div>
        </section>

        {notice ? (
          <div className="rounded-2xl border border-[#eadfd5] bg-white/75 px-4 py-3 text-sm text-[#7b675b]">
            {notice}
          </div>
        ) : null}

        <section className="space-y-5">
          {posts.map((post) => (
            <article
              className="rounded-[30px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-5 shadow-[0_18px_54px_rgba(72,50,33,0.06)]"
              key={post.id}
            >
              <header className="flex items-center gap-3">
                <Link to={`/profile/${post.author.userId}`}>
                  <Avatar name={post.author.nickname} />
                </Link>
                <div className="min-w-0">
                  <Link
                    className="font-semibold text-[#2f221b] hover:text-[#8f5c42]"
                    to={`/profile/${post.author.userId}`}
                  >
                    {post.author.nickname}
                  </Link>
                  <p className="truncate text-xs text-[#9d8a7f]">
                    {[post.author.schoolName, post.author.department, post.author.grade]
                      .filter(Boolean)
                      .join(' · ')}
                    {post.author.schoolName || post.author.department || post.author.grade ? ' · ' : ''}
                    {formatDate(post.createdAt)}
                  </p>
                </div>
              </header>

              {post.content ? (
                <p className="mt-4 whitespace-pre-wrap break-words text-[15px] leading-8 text-[#3f332d]">
                  {post.content}
                </p>
              ) : null}

              {post.imageUrls.length > 0 ? (
                <div className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {post.imageUrls.map((imageUrl) => (
                    <img
                      alt=""
                      className="max-h-80 w-full rounded-[24px] border border-[#eee4db] object-cover"
                      key={imageUrl}
                      loading="lazy"
                      src={imageUrl}
                    />
                  ))}
                </div>
              ) : null}

              <div className="mt-4 flex items-center gap-3 border-t border-[#f0e4da] pt-4 text-sm text-[#7b675b]">
                <button
                  className={cn(
                    'inline-flex items-center gap-2 rounded-full px-3 py-2 transition-colors hover:bg-[#f3ece4]',
                    post.likedByMe && 'text-[#b94848]',
                  )}
                  onClick={() => void handleTogglePostLike(post)}
                  type="button"
                >
                  <Heart className={cn('h-4 w-4', post.likedByMe && 'fill-current')} />
                  {post.likeCount}
                </button>
                <span className="inline-flex items-center gap-2 rounded-full px-3 py-2">
                  <MessageCircle className="h-4 w-4" />
                  {post.commentCount}
                </span>
              </div>

              <div className="mt-4 flex gap-2">
                <input
                  className="min-h-11 flex-1 rounded-full border border-[#eadfd5] bg-white px-4 text-sm outline-none focus:border-[#c49a78]"
                  onChange={(event) =>
                    setCommentDrafts((current) => ({
                      ...current,
                      [post.id]: event.target.value,
                    }))
                  }
                  placeholder="写下评论..."
                  value={commentDrafts[post.id] ?? ''}
                />
                <button
                  className="inline-flex h-11 items-center justify-center rounded-full bg-[#2f221b] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#4b372c]"
                  onClick={() => void handleSubmitComment(post.id)}
                  type="button"
                >
                  评论
                </button>
              </div>

              {post.comments.length > 0 ? (
                <div className="mt-5 space-y-4">
                  {post.comments.map((comment) => (
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
            </article>
          ))}
        </section>
      </div>
    </div>
  )
}
