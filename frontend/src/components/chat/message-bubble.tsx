import { cn } from '@/lib/utils'

interface Props {
  content: string
  isMine: boolean
  time: string
}

export function MessageBubble({ content, isMine, time }: Props) {
  const formattedTime = new Date(time).toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div
      className={cn(
        'group flex items-end gap-2',
        isMine ? 'flex-row-reverse' : 'flex-row',
      )}
    >
      <div
        className={cn(
          'max-w-[75%] rounded-2xl px-4 py-2.5 text-sm leading-relaxed',
          isMine
            ? 'rounded-br-md bg-[#2f221b] text-white'
            : 'rounded-bl-md border border-[#eee4db] bg-white text-[#2f221b]',
        )}
      >
        <p className="whitespace-pre-wrap break-words">{content}</p>
      </div>
      <span className="mb-1 shrink-0 text-[10px] text-[#7e6d61]/0 transition-colors group-hover:text-[#7e6d61]/60">
        {formattedTime}
      </span>
    </div>
  )
}
