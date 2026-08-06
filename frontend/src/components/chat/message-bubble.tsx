import { cn } from '@/lib/utils'

interface Props {
  content: string
  isMine: boolean
  senderName?: string
  time: string
}

export function MessageBubble({ content, isMine, senderName, time }: Props) {
  const formattedTime = new Date(time).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div className={cn('group flex items-start gap-2.5', isMine && 'flex-row-reverse')}>
      <span
        className={cn(
          'flex h-9 w-9 shrink-0 items-center justify-center rounded bg-[#d8ddd9] text-xs font-semibold text-[#4c5952]',
          isMine && 'bg-[#d8e9df] text-[#315d4f]',
        )}
      >
        {(senderName || (isMine ? '我' : '同')).slice(0, 1)}
      </span>
      <div className={cn('flex max-w-[72%] flex-col', isMine ? 'items-end' : 'items-start')}>
        {!isMine && senderName ? (
          <span className="mb-1 text-[11px] text-[#858984]">{senderName}</span>
        ) : null}
        <div
          className={cn(
            'rounded-md px-3.5 py-2.5 text-sm leading-6 shadow-[0_1px_2px_rgba(30,35,31,0.08)]',
            isMine
              ? 'bg-[#95d6a4] text-[#173220]'
              : 'border border-[#e1e2df] bg-white text-[#2f312e]',
          )}
        >
          <p className="whitespace-pre-wrap break-words">{content}</p>
        </div>
      </div>
      <span className="mt-7 shrink-0 text-[10px] text-[#777b77]/0 transition-colors group-hover:text-[#777b77]/60">
        {formattedTime}
      </span>
    </div>
  )
}
