import {
  Bookmark,
  BookmarkCheck,
  ChevronRight,
  Clock3,
  MapPin,
  ShieldCheck,
  UsersRound,
} from 'lucide-react'
import { Link } from 'react-router-dom'

import { IntentCategoryIcon, intentCategoryMeta } from '@/components/connection/intent-category'
import type { ConnectionIntent } from '@/lib/connection-api'
import { cn } from '@/lib/utils'

interface IntentCardProps {
  intent: ConnectionIntent
  compact?: boolean
  onSave?: (intent: ConnectionIntent) => void
  saving?: boolean
  footer?: React.ReactNode
}

function formatRelativeDate(value?: string) {
  if (!value) {
    return '刚刚更新'
  }
  const diff = Date.now() - new Date(value).getTime()
  const minutes = Math.max(1, Math.floor(diff / 60_000))
  if (minutes < 60) {
    return `${minutes} 分钟前`
  }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) {
    return `${hours} 小时前`
  }
  return `${Math.floor(hours / 24)} 天前`
}

export function IntentCard({
  intent,
  compact = false,
  onSave,
  saving = false,
  footer,
}: IntentCardProps) {
  const meta = intentCategoryMeta[intent.category]

  return (
    <article className="group border-b border-[#e6e2dc] py-5 first:pt-0 last:border-b-0 last:pb-0">
      <div className="flex items-start gap-3">
        <IntentCategoryIcon category={intent.category} size={18} />

        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-start justify-between gap-3">
            <div className="min-w-0">
              <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-[#77736c]">
                <span className="font-medium text-[#3f6659]">{meta.label}</span>
                <span>{formatRelativeDate(intent.publishedAt ?? intent.updatedAt)}</span>
                {intent.author.verified ? (
                  <span className="inline-flex items-center gap-1">
                    <ShieldCheck size={12} />
                    清华身份
                  </span>
                ) : null}
              </div>
              <Link
                className="mt-1.5 block text-[17px] font-semibold leading-6 text-[#23211f] transition-colors hover:text-[#315c4d]"
                to={`/intents/${intent.id}`}
              >
                {intent.title}
              </Link>
            </div>

            {onSave && !intent.isOwner ? (
              <button
                aria-label={intent.savedByMe ? '取消收藏' : '收藏意图'}
                className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#77736c] transition-colors hover:bg-[#f2efeb] hover:text-[#4d3769] disabled:opacity-50"
                disabled={saving}
                onClick={() => onSave(intent)}
                title={intent.savedByMe ? '取消收藏' : '收藏'}
                type="button"
              >
                {intent.savedByMe ? <BookmarkCheck size={18} /> : <Bookmark size={18} />}
              </button>
            ) : (
              <Link
                aria-label="查看意图"
                className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#8a867f] transition-colors hover:bg-[#f2efeb] hover:text-[#4d3769]"
                title="查看详情"
                to={`/intents/${intent.id}`}
              >
                <ChevronRight size={18} />
              </Link>
            )}
          </div>

          {!compact ? (
            <p className="mt-2 line-clamp-2 text-sm leading-6 text-[#65615b]">
              {intent.description}
            </p>
          ) : null}

          <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-[#77736c]">
            {intent.timeWindow ? (
              <span className="inline-flex items-center gap-1.5">
                <Clock3 size={14} />
                {intent.timeWindow}
              </span>
            ) : null}
            {intent.location ? (
              <span className="inline-flex items-center gap-1.5">
                <MapPin size={14} />
                {intent.location}
              </span>
            ) : null}
            <span className="inline-flex items-center gap-1.5">
              <UsersRound size={14} />
              {intent.author.nickname}
            </span>
          </div>

          {intent.constraints.length > 0 ? (
            <div className="mt-3 flex flex-wrap gap-1.5">
              {intent.constraints.slice(0, 4).map((constraint) => (
                <span
                  className="rounded-md bg-[#f2efeb] px-2 py-1 text-[11px] text-[#65615b]"
                  key={constraint}
                >
                  {constraint}
                </span>
              ))}
            </div>
          ) : null}

          {footer ? <div className={cn('mt-4', compact && 'mt-3')}>{footer}</div> : null}
        </div>
      </div>
    </article>
  )
}
