import type { ConversationItem } from '@/lib/chat-api'
import { cn } from '@/lib/utils'

interface Props {
  conversations: ConversationItem[]
  activeId?: number
  onSelect: (id: number) => void
}

export function ConversationList({ conversations, activeId, onSelect }: Props) {
  if (conversations.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-[#7e6d61]">
        暂无会话
      </p>
    )
  }

  return (
    <div className="space-y-1">
      {conversations.map((conv) => (
        <button
          className={cn(
            'flex w-full items-center gap-3 rounded-2xl px-3 py-3 text-left transition-colors',
            activeId === conv.id
              ? 'bg-[#2f221b]/8'
              : 'hover:bg-[#2f221b]/4',
          )}
          key={conv.id}
          onClick={() => onSelect(conv.id)}
          type="button"
        >
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#eee4db] text-sm font-medium text-[#5f4d42]">
            {conv.peerNickname.charAt(0)}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center justify-between">
              <p className="truncate text-sm font-medium text-[#2f221b]">
                {conv.peerNickname}
              </p>
              {conv.lastMessageAt ? (
                <span className="shrink-0 text-[10px] text-[#7e6d61]/60">
                  {new Date(conv.lastMessageAt).toLocaleDateString('zh-CN', {
                    month: 'numeric',
                    day: 'numeric',
                  })}
                </span>
              ) : null}
            </div>
            <div className="flex items-center justify-between gap-2">
              <p className="truncate text-xs text-[#7e6d61]">
                {conv.statusMessage || conv.lastMessage || '暂无消息'}
              </p>
              {conv.unreadCount > 0 ? (
                <span className="ml-2 inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-[#e74c3c] px-1.5 text-[10px] font-medium text-white">
                  {conv.unreadCount > 99 ? '99+' : conv.unreadCount}
                </span>
              ) : null}
            </div>
          </div>
        </button>
      ))}
    </div>
  )
}
