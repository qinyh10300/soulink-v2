import { useEffect } from 'react'
import { ClipboardList, X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface QuestionnaireRequiredModalProps {
  eyebrow?: string
  message?: string
  onClose: () => void
  onConfirm: () => void
  open: boolean
  title?: string
}

export function QuestionnaireRequiredModal({
  eyebrow = '本周配对',
  message = '完成问卷后，才可以参加本周配对。点击确认后，将直接前往问卷页面。',
  onClose,
  onConfirm,
  open,
  title = '请填写问卷',
}: QuestionnaireRequiredModalProps) {
  useEffect(() => {
    if (!open) {
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open])

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        aria-label="关闭填写问卷提示"
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 w-full max-w-xl rounded-[30px] border border-[#eadfd5] bg-[#fcf8f3] p-8 shadow-[0_26px_72px_rgba(61,40,29,0.18)]">
        <button
          aria-label="关闭"
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#e8ddd3] bg-white text-[#7d6556]"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="flex items-center gap-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-[linear-gradient(135deg,#f3e8dc,#ead9c8)] text-[#8f6548] shadow-[0_14px_28px_rgba(111,84,61,0.14)]">
            <ClipboardList size={24} />
          </div>
          <div className="space-y-1">
            <p className="text-sm uppercase tracking-[0.22em] text-[#a4846d]">{eyebrow}</p>
            <h3 className="font-serif text-3xl text-[#2f221b]">{title}</h3>
          </div>
        </div>

        <p className="mt-6 text-base leading-8 text-[#6f6056]">
          {message}
        </p>

        <div className="mt-8 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <button
            className={cn(
              buttonVariants({ size: 'lg', variant: 'secondary' }),
              'rounded-full border-[#e7d8cb] bg-white/92 text-[#7d6556] hover:bg-white',
            )}
            onClick={onClose}
            type="button"
          >
            稍后再说
          </button>
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'rounded-full bg-[#8f6548] text-white hover:bg-[#7f5a40]',
            )}
            onClick={onConfirm}
            type="button"
          >
            去填写问卷
          </button>
        </div>
      </div>
    </div>
  )
}
