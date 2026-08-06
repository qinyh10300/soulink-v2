import {
  ArrowRight,
  BookOpen,
  Calendar,
  Dumbbell,
  Heart,
  Sparkles,
  UtensilsCrossed,
  UsersRound,
  type LucideIcon,
} from 'lucide-react'

import { useEngagement } from '@/contexts/engagement-context'
import type { CampusTheme } from '@/lib/engagement-api'
import { cn } from '@/lib/utils'

const themeIcons: Record<string, LucideIcon> = {
  utensils: UtensilsCrossed,
  'book-open': BookOpen,
  dumbbell: Dumbbell,
  'users-round': UsersRound,
  heart: Heart,
  calendar: Calendar,
  sparkles: Sparkles,
}

const themeTones: Record<string, string> = {
  meal: 'bg-[#edf3e9] text-[#4e6f43]',
  study: 'bg-[#e8f0f4] text-[#3c6478]',
  sport: 'bg-[#f5eee3] text-[#765d38]',
  romantic: 'bg-[#f5e9ed] text-[#8b5063]',
  project: 'bg-[#eee9f2] text-[#644c75]',
  event: 'bg-[#f2ece7] text-[#795f50]',
}

export function CampusThemeStrip({
  onSelect,
}: {
  onSelect: (theme: CampusTheme) => void
}) {
  const { themes } = useEngagement()
  if (themes.length === 0) return null

  const selectedCount = themes.filter((theme) => theme.selected).length

  return (
    <section className="mt-7 border-t border-[#dedad4] pt-6">
      <div className="flex items-end justify-between gap-3">
        <div>
          <h3 className="text-sm font-semibold text-[#34312e]">校园主题</h3>
          <p className="mt-0.5 text-[11px] text-[#817d76]">
            {selectedCount > 0 ? `${selectedCount} 个已关注主题` : '从一个具体时刻开始'}
          </p>
        </div>
        <span className="text-[10px] font-medium text-[#918d86]">清华园 · 本周</span>
      </div>
      <div className="mt-3 grid gap-2 sm:grid-cols-2">
        {themes.map((theme) => {
          const Icon = themeIcons[theme.iconName] ?? Sparkles
          return (
            <button
              className={cn(
                'group flex min-h-18 items-center gap-3 rounded-md border px-3 py-3 text-left transition-colors',
                theme.selected
                  ? 'border-[#c6d5cf] bg-[#f2f6f4] hover:border-[#9eb7ad]'
                  : 'border-[#dedad4] bg-[#fbfaf8] hover:border-[#aebdb7] hover:bg-white',
              )}
              key={theme.key}
              onClick={() => onSelect(theme)}
              type="button"
            >
              <span
                className={cn(
                  'inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md',
                  themeTones[theme.category] ?? themeTones.event,
                )}
              >
                <Icon size={16} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="flex items-center gap-2">
                  <span className="truncate text-xs font-semibold text-[#45413d]">
                    {theme.title}
                  </span>
                  <span className="shrink-0 text-[10px] text-[#817d76]">
                    {theme.scheduleHint}
                  </span>
                </span>
                <span className="mt-1 line-clamp-1 block text-[11px] text-[#817d76]">
                  {theme.description}
                </span>
              </span>
              <ArrowRight
                className="shrink-0 text-[#aaa59d] transition-transform group-hover:translate-x-0.5 group-hover:text-[#527164]"
                size={14}
              />
            </button>
          )
        })}
      </div>
    </section>
  )
}
