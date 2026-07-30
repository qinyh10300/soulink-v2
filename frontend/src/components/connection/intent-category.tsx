import {
  CalendarDays,
  Coffee,
  Dumbbell,
  Heart,
  Lightbulb,
  MessageCircle,
  Rocket,
  Sparkles,
  UsersRound,
  type LucideIcon,
} from 'lucide-react'

import type { IntentCategory } from '@/lib/connection-api'

interface IntentCategoryMeta {
  label: string
  shortLabel: string
  description: string
  icon: LucideIcon
  iconClassName: string
  surfaceClassName: string
}

export const intentCategoryOrder: IntentCategory[] = [
  'meal',
  'study',
  'sport',
  'conversation',
  'event',
  'project',
  'cofounder',
  'romantic',
  'other',
]

export const intentCategoryMeta: Record<IntentCategory, IntentCategoryMeta> = {
  meal: {
    label: '饭搭子',
    shortLabel: '吃饭',
    description: '今天就能发生的一顿饭',
    icon: Coffee,
    iconClassName: 'text-[#9b4d2f]',
    surfaceClassName: 'bg-[#fff1e8]',
  },
  study: {
    label: '学习搭子',
    shortLabel: '学习',
    description: '自习、备考或长期节奏',
    icon: Lightbulb,
    iconClassName: 'text-[#35647a]',
    surfaceClassName: 'bg-[#eaf4f7]',
  },
  sport: {
    label: '运动搭子',
    shortLabel: '运动',
    description: '跑步、球类与日常训练',
    icon: Dumbbell,
    iconClassName: 'text-[#287058]',
    surfaceClassName: 'bg-[#e8f5ef]',
  },
  event: {
    label: '活动同伴',
    shortLabel: '活动',
    description: '讲座、演出与校园现场',
    icon: CalendarDays,
    iconClassName: 'text-[#7e5d1e]',
    surfaceClassName: 'bg-[#fbf3d9]',
  },
  conversation: {
    label: '认真聊聊',
    shortLabel: '聊天',
    description: '低压力但有内容的交流',
    icon: MessageCircle,
    iconClassName: 'text-[#4f648f]',
    surfaceClassName: 'bg-[#edf1fa]',
  },
  romantic: {
    label: '认真关系',
    shortLabel: '关系',
    description: '私密、慢速、双方自愿',
    icon: Heart,
    iconClassName: 'text-[#a33e5f]',
    surfaceClassName: 'bg-[#faebf0]',
  },
  project: {
    label: '项目伙伴',
    shortLabel: '项目',
    description: '技能互补的短中期协作',
    icon: Rocket,
    iconClassName: 'text-[#67449b]',
    surfaceClassName: 'bg-[#f1ecf8]',
  },
  cofounder: {
    label: 'CTO / 联创',
    shortLabel: '联创',
    description: '愿景、能力与投入的长期匹配',
    icon: UsersRound,
    iconClassName: 'text-[#8b4a36]',
    surfaceClassName: 'bg-[#f8ede7]',
  },
  other: {
    label: '开放寻找',
    shortLabel: '其他',
    description: '还没被分类的一次连接',
    icon: Sparkles,
    iconClassName: 'text-[#5c6472]',
    surfaceClassName: 'bg-[#eef0f3]',
  },
}

export function IntentCategoryIcon({
  category,
  size = 18,
}: {
  category: IntentCategory
  size?: number
}) {
  const meta = intentCategoryMeta[category]
  const Icon = meta.icon

  return (
    <span
      className={`inline-flex shrink-0 items-center justify-center rounded-md ${meta.surfaceClassName} ${meta.iconClassName}`}
      style={{ height: size + 16, width: size + 16 }}
    >
      <Icon aria-hidden="true" size={size} strokeWidth={1.8} />
    </span>
  )
}
