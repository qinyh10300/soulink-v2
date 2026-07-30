import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { DashboardFeatureItem } from '@/types/dashboard'

interface DashboardFeatureCardProps {
  item: DashboardFeatureItem
  onAction?: () => void
}

export function DashboardFeatureCard({
  item,
  onAction,
}: DashboardFeatureCardProps) {
  const isDark = Boolean(item.emphasized)

  return (
    <Card
      className={cn(
        'h-full rounded-[32px] border-[#eee4db] transition-transform duration-200 hover:-translate-y-1',
        isDark
          ? 'border-transparent bg-[linear-gradient(180deg,#8a5746_0%,#744739_100%)] text-white shadow-[0_20px_48px_rgba(83,48,35,0.24)]'
          : 'bg-[linear-gradient(180deg,rgba(255,251,247,0.96)_0%,rgba(255,248,243,0.93)_100%)] text-[#2f221b] shadow-[0_14px_38px_rgba(72,50,33,0.06)]',
      )}
    >
      <CardContent className="flex h-full flex-col gap-5 p-7">
        <div className="text-[2.6rem]">{item.icon}</div>
        <div className="space-y-3.5">
          <h3
            className={cn(
              'font-serif text-[1.45rem] leading-[1.08] tracking-[-0.03em] sm:text-[1.65rem]',
              isDark ? 'text-white' : '',
            )}
          >
            {item.title}
          </h3>
          <p
            className={cn(
              'text-[1.02rem] leading-8',
              isDark ? 'text-white/82' : 'text-[#78685d]',
            )}
          >
            {item.description}
          </p>
        </div>

        <div className="mt-auto pt-4">
          {isDark ? (
            item.disabled ? (
              <div className="inline-flex rounded-full bg-white/14 px-4 py-2 text-sm tracking-[0.04em] text-white/90">
                {item.meta}
              </div>
            ) : (
              <button
                className="inline-flex rounded-full bg-white/14 px-4 py-2 text-sm tracking-[0.04em] text-white/90 transition-colors hover:bg-white/20"
                onClick={onAction}
                type="button"
              >
                {item.meta}
              </button>
            )
          ) : item.disabled ? (
            <div className="inline-flex rounded-full bg-[#eadacc] px-4 py-2 text-sm tracking-[0.04em] text-[#7a5f50]">
              {item.meta}
            </div>
          ) : (
            <button
              className="inline-flex rounded-full bg-[#eadacc] px-4 py-2 text-sm tracking-[0.04em] text-[#7a5f50] transition-colors hover:bg-[#e1cfbf] hover:text-[#644b3f]"
              onClick={onAction}
              type="button"
            >
              {item.meta}
            </button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
