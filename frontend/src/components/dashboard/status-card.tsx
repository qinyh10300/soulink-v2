import {
  CalendarClock,
  CheckCheck,
  CircleDot,
  ClipboardList,
  type LucideIcon,
} from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { DashboardStatusItem } from '@/types/dashboard'

const toneStyles: Record<
  DashboardStatusItem['tone'],
  { accent: string; iconWrap: string; icon: LucideIcon }
> = {
  success: {
    accent: 'before:bg-[#7dad7f]',
    iconWrap: 'bg-[#eef6ee] text-[#6d9b6f]',
    icon: CheckCheck,
  },
  active: {
    accent: 'before:bg-[#7ca96f]',
    iconWrap: 'bg-[#f0f6eb] text-[#6f985f]',
    icon: CircleDot,
  },
  neutral: {
    accent: 'before:bg-[#c7b8ac]',
    iconWrap: 'bg-[#f7f0ea] text-[#907a6a]',
    icon: CalendarClock,
  },
}

interface StatusCardProps {
  item: DashboardStatusItem
  onAction?: () => void
  onCardClick?: () => void
}

export function StatusCard({ item, onAction, onCardClick }: StatusCardProps) {
  const style =
    item.id === 'survey' && item.progress !== undefined
      ? {
          accent: 'before:bg-[#ba8f68]',
          iconWrap: 'bg-[#f7efe7] text-[#9a7457]',
          icon: ClipboardList,
        }
      : toneStyles[item.tone]
  const Icon = style.icon
  const isClickable = Boolean(onCardClick)

  return (
    <Card
      className={cn(
        'relative overflow-hidden rounded-[30px] border-[#eee4db] bg-[linear-gradient(180deg,rgba(255,251,247,0.96)_0%,rgba(255,248,243,0.93)_100%)] shadow-[0_16px_42px_rgba(72,50,33,0.06)] before:absolute before:inset-y-6 before:left-0 before:w-1 before:rounded-full',
        style.accent,
        isClickable ? 'cursor-pointer transition-transform duration-200 hover:-translate-y-0.5 hover:shadow-[0_20px_48px_rgba(72,50,33,0.08)]' : '',
      )}
      onClick={onCardClick}
    >
      <CardContent className="flex items-start gap-4 px-5 py-6 sm:px-6 sm:py-6">
        <div
          className={cn(
            'flex h-12 w-12 shrink-0 items-center justify-center rounded-full shadow-[0_10px_22px_rgba(85,60,44,0.04)]',
            style.iconWrap,
          )}
        >
          <Icon size={21} />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-[0.92rem] tracking-[0.08em] text-[#9a887b]">{item.title}</p>
          {item.value || item.actionLabel ? (
            <div className="mt-2 flex items-center justify-between gap-3">
              {item.value ? (
                <p className="min-w-0 whitespace-nowrap font-serif text-[1.45rem] leading-none tracking-[-0.03em] text-[#2f221b] sm:text-[1.65rem]">
                  {item.value}
                </p>
              ) : null}
              {item.actionLabel ? (
                <button
                  className="shrink-0 whitespace-nowrap text-[0.82rem] text-[#a07d65] transition-colors hover:text-[#7d5c47] sm:text-[0.88rem]"
                  onClick={(event) => {
                    event.stopPropagation()
                    onAction?.()
                  }}
                  type="button"
                >
                  {item.actionLabel}
                </button>
              ) : null}
            </div>
          ) : null}
          {item.description ? (
            <p className="mt-2 text-[0.98rem] leading-7 text-[#8f7b6f]">{item.description}</p>
          ) : null}
          {typeof item.progress === 'number' ? (
            <div className="mt-5 space-y-2.5">
              <div className="h-2 overflow-hidden rounded-full bg-[#efe5db]">
                <div
                  className="h-full rounded-full bg-[linear-gradient(90deg,#ba8f68,#9d7151)]"
                  style={{ width: `${item.progress}%` }}
                />
              </div>
              <p className="text-[0.72rem] tracking-[0.24em] text-[#a08f83] uppercase">
                已完成 {item.progress}%
              </p>
            </div>
          ) : null}
        </div>
      </CardContent>
    </Card>
  )
}
