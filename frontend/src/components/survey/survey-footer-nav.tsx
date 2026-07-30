import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface SurveyFooterNavProps {
  currentIndex: number
  total: number
  onNext: () => void
  onPrevious: () => void
  onSaveAndExit: () => void
}

export function SurveyFooterNav({
  currentIndex,
  total,
  onNext,
  onPrevious,
  onSaveAndExit,
}: SurveyFooterNavProps) {
  const isLastSection = currentIndex === total - 1

  return (
    <div className="space-y-4 pt-2 text-center sm:space-y-6 sm:pt-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
        <p className="order-1 text-sm text-[#8d7d70] sm:order-2 sm:text-xl">
          {currentIndex + 1} / {total}
        </p>

        <button
          className={cn(
            buttonVariants({ variant: 'secondary', size: 'lg' }),
            'order-2 w-full rounded-full border-[#ece2d8] bg-white/88 px-6 text-sm text-[#7e6f64] sm:order-1 sm:w-auto sm:px-8 sm:text-base',
          )}
          disabled={currentIndex === 0}
          onClick={onPrevious}
          type="button"
        >
          上一部分
        </button>

        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'order-3 w-full rounded-full bg-[#f4e9dd] px-6 text-sm text-[#48372c] shadow-none hover:bg-[#efdfcf] sm:w-auto sm:px-8 sm:text-base',
          )}
          disabled={!isLastSection && currentIndex === total - 1}
          onClick={isLastSection ? onSaveAndExit : onNext}
          type="button"
        >
          {isLastSection ? '保存并退出' : '下一部分'}
        </button>
      </div>

      <p className="px-2 text-sm leading-6 text-[#a08f83] sm:text-base">
        你可以随时保存当前进度，稍后继续填写问卷。
      </p>
    </div>
  )
}
