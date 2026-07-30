import { Link } from 'react-router-dom'

import brandIcon from '@/assets/soulink-logo-icon.png'
import { cn } from '@/lib/utils'

interface BrandMarkProps {
  className?: string
  iconClassName?: string
  titleClassName?: string
  subtitleClassName?: string
  to?: string
}

export function BrandMark({
  className,
  iconClassName,
  subtitleClassName,
  titleClassName,
  to = '/',
}: BrandMarkProps) {
  return (
    <Link className={cn('flex items-center gap-3', className)} to={to}>
      <img
        alt="Soulink"
        className={cn('h-12 w-12 rounded-2xl border object-cover', iconClassName)}
        decoding="async"
        fetchPriority="high"
        height="48"
        loading="eager"
        src={brandIcon}
        width="48"
      />
      <div>
        <p className={cn('text-sm font-semibold tracking-[0.12em] uppercase', titleClassName)}>
          Soulink
        </p>
        <p className={cn('text-sm', subtitleClassName)}>AI 社交匹配平台</p>
      </div>
    </Link>
  )
}
