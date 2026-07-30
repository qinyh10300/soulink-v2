interface Props {
  time: string
}

export function TimeSeparator({ time }: Props) {
  return (
    <div className="flex items-center gap-3 py-2">
      <div className="h-px flex-1 bg-[#eee4db]" />
      <span className="shrink-0 text-[11px] text-[#7e6d61]/70">
        {formatSeparatorTime(time)}
      </span>
      <div className="h-px flex-1 bg-[#eee4db]" />
    </div>
  )
}

function formatSeparatorTime(iso: string): string {
  const date = new Date(iso)
  const now = new Date()

  const isToday =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()

  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  const isYesterday =
    date.getFullYear() === yesterday.getFullYear() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getDate() === yesterday.getDate()

  const hhmm = date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
  })

  if (isToday) return hhmm
  if (isYesterday) return `昨天 ${hhmm}`

  const isSameYear = date.getFullYear() === now.getFullYear()
  if (isSameYear) {
    return date.toLocaleDateString('zh-CN', {
      month: 'numeric',
      day: 'numeric',
    }) + ' ' + hhmm
  }

  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
  }) + ' ' + hhmm
}
