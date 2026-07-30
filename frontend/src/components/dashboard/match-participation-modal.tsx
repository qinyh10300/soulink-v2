import { useEffect } from 'react'
import { CircleDot, X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export type WeeklyParticipationAction = 'join' | 'leave'

interface MatchParticipationModalProps {
  action: WeeklyParticipationAction | null
  onClose: () => void
  onConfirm: () => void
  isSubmitting?: boolean
  open: boolean
}

const modalCopy: Record<
  WeeklyParticipationAction,
  {
    title: string
    description: string
    confirmLabel: string
  }
> = {
  join: {
    title: '确认参加这次链接？',
    description: '确认后，你会被标记为参加本轮链接；在结果揭晓前，你仍可以随时退出。',
    confirmLabel: '确认参加',
  },
  leave: {
    title: '确认退出这次链接？',
    description: '确认后，你将不再参与本轮链接；如果改变主意，之后仍可以重新参加。',
    confirmLabel: '确认退出',
  },
}

export function MatchParticipationModal({
  action,
  onClose,
  onConfirm,
  isSubmitting = false,
  open,
}: MatchParticipationModalProps) {
  useEffect(() => {
    if (!open) {
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !isSubmitting) {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isSubmitting, onClose, open])

  if (!open || !action) {
    return null
  }

  const copy = modalCopy[action]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        aria-label="关闭参与确认弹窗"
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
        disabled={isSubmitting}
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 w-full max-w-xl rounded-[30px] border border-[#eadfd5] bg-[#fcf8f3] p-8 shadow-[0_26px_72px_rgba(61,40,29,0.18)]">
        <button
          aria-label="关闭"
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#e8ddd3] bg-white text-[#7d6556]"
          disabled={isSubmitting}
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[linear-gradient(135deg,#f3e8dc,#ead9c8)] text-[#8f6548] shadow-[0_14px_28px_rgba(111,84,61,0.14)]">
            <CircleDot size={24} />
          </div>
          <div className="space-y-1">
            <p className="text-sm uppercase tracking-[0.22em] text-[#a4846d]">本轮链接</p>
            <h3 className="font-serif text-3xl text-[#2f221b]">{copy.title}</h3>
          </div>
        </div>

        <p className="mt-6 text-base leading-8 text-[#6f6056]">{copy.description}</p>

        <div className="mt-8 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <button
            className={cn(
              buttonVariants({ size: 'lg', variant: 'secondary' }),
              'rounded-full border-[#e7d8cb] bg-white/92 text-[#7d6556] hover:bg-white disabled:cursor-not-allowed disabled:opacity-60',
            )}
            disabled={isSubmitting}
            onClick={onClose}
            type="button"
          >
            再想想
          </button>
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'rounded-full bg-[#8f6548] text-white hover:bg-[#7f5a40] disabled:cursor-not-allowed disabled:opacity-60',
            )}
            disabled={isSubmitting}
            onClick={onConfirm}
            type="button"
          >
            {isSubmitting ? '提交中...' : copy.confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
