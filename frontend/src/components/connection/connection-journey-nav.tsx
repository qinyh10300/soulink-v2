import {
  ClipboardList,
  Compass,
  HeartHandshake,
  Sparkles,
  type LucideIcon,
} from 'lucide-react'
import { NavLink } from 'react-router-dom'

import { connectionPaths } from '@/lib/connection-routes'
import { cn } from '@/lib/utils'

interface JourneyItem {
  label: string
  compactLabel: string
  to: string
  icon: LucideIcon
  end?: boolean
}

const journeyItems: JourneyItem[] = [
  {
    label: '今日连接',
    compactLabel: '今日',
    to: connectionPaths.root,
    icon: Sparkles,
    end: true,
  },
  {
    label: '校园邀请',
    compactLabel: '邀请',
    to: connectionPaths.discover,
    icon: Compass,
  },
  {
    label: '我的寻找',
    compactLabel: '寻找',
    to: connectionPaths.intents,
    icon: ClipboardList,
  },
  {
    label: '连接房间',
    compactLabel: '房间',
    to: connectionPaths.rooms,
    icon: HeartHandshake,
  },
]

export function ConnectionJourneyNav({
  variant = 'mobile',
}: {
  variant?: 'mobile' | 'sidebar'
}) {
  if (variant === 'sidebar') {
    return (
      <nav aria-label="连接旅程" className="mt-2 space-y-0.5">
        {journeyItems.map((item) => {
          const Icon = item.icon
          return (
            <NavLink
              className={({ isActive }) =>
                cn(
                  'flex h-9 items-center gap-2.5 rounded-md px-2 text-xs font-medium transition-colors',
                  isActive
                    ? 'bg-[#e8f1ed] text-[#315c4d]'
                    : 'text-[#706c65] hover:bg-[#f0eeea] hover:text-[#302e2b]',
                )
              }
              end={item.end}
              key={item.to}
              to={item.to}
            >
              <Icon size={15} strokeWidth={1.8} />
              {item.label}
            </NavLink>
          )
        })}
      </nav>
    )
  }

  return (
    <nav
      aria-label="连接旅程"
      className="grid h-11 shrink-0 grid-cols-4 border-b border-[#dfdcd6] bg-[rgba(251,250,248,0.96)] px-2 backdrop-blur-lg lg:hidden"
    >
      {journeyItems.map((item) => {
        const Icon = item.icon
        return (
          <NavLink
            className={({ isActive }) =>
              cn(
                'relative flex min-w-0 items-center justify-center gap-1.5 text-[11px] font-medium',
                isActive ? 'text-[#315c4d]' : 'text-[#77736c]',
              )
            }
            end={item.end}
            key={item.to}
            to={item.to}
          >
            {({ isActive }) => (
              <>
                <Icon size={14} strokeWidth={1.8} />
                <span>{item.compactLabel}</span>
                {isActive ? (
                  <span className="absolute inset-x-3 bottom-0 h-0.5 bg-[#527b6d]" />
                ) : null}
              </>
            )}
          </NavLink>
        )
      })}
    </nav>
  )
}
