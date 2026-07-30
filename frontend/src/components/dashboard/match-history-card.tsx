import type { KeyboardEvent, MouseEvent } from 'react'

import { ChevronRight, MessageCircle } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { MatchHistoryEmptyState, MatchHistoryListItem } from '@/types/dashboard'

interface MatchHistoryCardProps {
  emptyState: MatchHistoryEmptyState
  contactingMatchId?: number | null
  isLoading?: boolean
  items?: MatchHistoryListItem[]
  onContact?: (matchId: number) => void
  onSelect?: (matchId: number) => void
}

export function MatchHistoryCard({
  emptyState,
  contactingMatchId = null,
  isLoading = false,
  items = [],
  onContact,
  onSelect,
}: MatchHistoryCardProps) {
  if (isLoading) {
    return (
      <Card className="rounded-[32px] border-[#eee4db] bg-[rgba(255,251,247,0.94)] shadow-[0_14px_44px_rgba(72,50,33,0.06)]">
        <CardContent className="flex min-h-[18rem] items-center justify-center px-6 py-12 text-center text-lg text-[#7c6a5f]">
          正在加载配对记录...
        </CardContent>
      </Card>
    )
  }

  if (items.length === 0) {
    return (
      <Card className="rounded-[32px] border-[#eee4db] bg-[rgba(255,251,247,0.94)] shadow-[0_14px_44px_rgba(72,50,33,0.06)]">
        <CardContent className="flex min-h-[22rem] flex-col items-center justify-center gap-4 px-6 py-12 text-center">
          <div className="text-5xl">{emptyState.icon}</div>
          <h3 className="font-serif text-[1.45rem] tracking-[-0.03em] text-[#2f221b] sm:text-[1.65rem]">
            {emptyState.title}
          </h3>
          <p className="max-w-xl text-[1.02rem] leading-8 text-[#7c6a5f]">
            {emptyState.description}
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className="rounded-[32px] border-[#eee4db] bg-[rgba(255,251,247,0.94)] shadow-[0_14px_44px_rgba(72,50,33,0.06)]">
      <CardContent className="space-y-4 px-5 py-5 sm:px-6">
        {items.map((item) => {
          const interactive = Boolean(onSelect)
          const contactable = Boolean(item.canContact && onContact)
          const isContacting = contactingMatchId === item.id

          const handleSelect = () => {
            onSelect?.(item.id)
          }

          const handleKeyDown = (event: KeyboardEvent<HTMLElement>) => {
            if (!interactive || event.target !== event.currentTarget) {
              return
            }
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              handleSelect()
            }
          }

          const handleContact = (event: MouseEvent<HTMLButtonElement>) => {
            event.stopPropagation()
            onContact?.(item.id)
          }

          return (
            <article
              className={cn(
                'w-full rounded-[28px] border border-[#efe5dc] bg-[#fffaf6] px-5 py-5 text-left transition-colors',
                interactive ? 'hover:border-[#decdbf] hover:bg-white' : 'cursor-default',
              )}
              key={item.id}
              onClick={interactive ? handleSelect : undefined}
              onKeyDown={handleKeyDown}
              role={interactive ? 'button' : undefined}
              tabIndex={interactive ? 0 : undefined}
            >
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="space-y-2.5">
                  <div className="flex flex-wrap items-center gap-3">
                    <h3 className="font-serif text-[1.8rem] tracking-[-0.03em] text-[#2f221b]">
                      {item.title}
                    </h3>
                    <span
                      className={cn(
                        'inline-flex rounded-full px-3 py-1 text-xs font-semibold tracking-[0.04em]',
                        item.statusTone === 'success'
                          ? 'bg-[#edf6ee] text-[#5d8b67]'
                          : 'bg-[#f5ece4] text-[#8d725d]',
                      )}
                    >
                      {item.statusLabel}
                    </span>
                  </div>
                  <p className="text-[0.92rem] text-[#9a877a]">{item.subtitle}</p>
                  <p className="text-[1rem] leading-7 text-[#6f5f55]">{item.message}</p>
                </div>

                <div className="flex items-center gap-3 text-sm text-[#8d7b70]">
                  <div className="space-y-1 text-right">
                    {item.scoreLabel ? <p>{item.scoreLabel}</p> : null}
                  </div>
                  {contactable ? (
                    <button
                      className="inline-flex h-9 shrink-0 items-center justify-center gap-1.5 rounded-full bg-[#8f3f3c] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#7c3532] disabled:cursor-not-allowed disabled:opacity-60"
                      disabled={isContacting}
                      onClick={handleContact}
                      onKeyDown={(event) => event.stopPropagation()}
                      type="button"
                    >
                      <MessageCircle size={15} />
                      {isContacting ? '处理中...' : '进入聊天'}
                    </button>
                  ) : null}
                  {interactive ? <ChevronRight size={18} /> : null}
                </div>
              </div>
            </article>
          )
        })}
      </CardContent>
    </Card>
  )
}
