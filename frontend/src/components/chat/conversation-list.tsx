import { Users } from 'lucide-react'

import { cn } from '@/lib/utils'

export interface ChatThreadItem {
  id: number
  kind: 'direct' | 'group'
  title: string
  subtitle: string
  lastMessageAt?: string
  unreadCount: number
}

interface Props {
  threads: ChatThreadItem[]
  activeKey?: string
  onSelect: (thread: ChatThreadItem) => void
}

function formatThreadTime(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  const today = new Date()
  if (date.toDateString() === today.toDateString()) {
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }
  return date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' })
}

export function ConversationList({ threads, activeKey, onSelect }: Props) {
  if (threads.length === 0) {
    return <p className="py-10 text-center text-sm text-[#858984]">暂无会话</p>
  }

  return (
    <div>
      {threads.map((thread) => {
        const key = `${thread.kind}:${thread.id}`
        return (
          <button
            className={cn(
              'flex w-full items-center gap-3 border-b border-[#e4e5e2] px-3 py-3 text-left transition-colors',
              activeKey === key ? 'bg-[#d9dcda]' : 'hover:bg-[#e7e9e7]',
            )}
            key={key}
            onClick={() => onSelect(thread)}
            type="button"
          >
            <span
              className={cn(
                'flex h-11 w-11 shrink-0 items-center justify-center rounded bg-[#cad9d1] text-sm font-semibold text-[#315d4f]',
                thread.kind === 'group' && 'bg-[#d9d5e5] text-[#5d5072]',
              )}
            >
              {thread.kind === 'group' ? <Users className="h-5 w-5" /> : thread.title.slice(0, 1)}
            </span>
            <span className="min-w-0 flex-1">
              <span className="flex items-center justify-between gap-2">
                <span className="truncate text-sm font-medium text-[#292b28]">{thread.title}</span>
                <time className="shrink-0 text-[10px] text-[#90938f]">
                  {formatThreadTime(thread.lastMessageAt)}
                </time>
              </span>
              <span className="mt-1 flex items-center justify-between gap-2">
                <span className="truncate text-xs text-[#7b7f7a]">{thread.subtitle || '暂无消息'}</span>
                {thread.unreadCount > 0 ? (
                  <span className="inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-[#e14f4f] px-1 text-[10px] font-semibold text-white">
                    {thread.unreadCount > 99 ? '99+' : thread.unreadCount}
                  </span>
                ) : null}
              </span>
            </span>
          </button>
        )
      })}
    </div>
  )
}
