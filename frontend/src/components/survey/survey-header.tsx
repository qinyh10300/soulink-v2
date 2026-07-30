import { FilePenLine } from 'lucide-react'
import { Link } from 'react-router-dom'

interface SurveyHeaderProps {
  answeredCount: number
  totalCount: number
  completionPercent: number
  draftLabel: string
  invalidQuestionNumber?: number | null
  onSave: () => void
  onSaveAndExit: () => void
  saveDisabled?: boolean
  backTo?: string
}

export function SurveyHeader({
  answeredCount,
  totalCount,
  completionPercent,
  draftLabel,
  invalidQuestionNumber = null,
  onSave,
  onSaveAndExit,
  saveDisabled = false,
  backTo = '/dashboard',
}: SurveyHeaderProps) {
  return (
    <div className="sticky top-0 z-40 border-b border-[#eadfd6] bg-[rgba(249,244,239,0.94)] backdrop-blur-xl">
      <div className="mx-auto w-full max-w-[90rem] px-4 py-3 sm:px-6 sm:py-4 lg:px-10">
        <div className="flex flex-wrap items-center gap-3 lg:flex-nowrap lg:gap-5">
          <Link
            className="min-w-fit shrink-0 font-serif text-[1.65rem] leading-none text-[#8f6548] sm:text-4xl"
            to={backTo}
          >
            Soulink
          </Link>

          <div className="flex min-w-[18rem] flex-1 items-center gap-3 sm:min-w-[24rem] sm:gap-4">
            <div className="h-1.5 min-w-0 flex-1 overflow-hidden rounded-full bg-[#eadfd5]">
              <div
                className="h-full rounded-full bg-[linear-gradient(90deg,#8f6548,#b88767)]"
                style={{ width: `${completionPercent}%` }}
              />
            </div>
            <p className="shrink-0 whitespace-nowrap text-[11px] text-[#8b776b] sm:text-lg">
              {answeredCount} / {totalCount} 已完成
            </p>
          </div>

          <div className="ml-auto flex flex-wrap items-center justify-end gap-2 sm:gap-3">
            <div className="inline-flex min-w-0 max-w-[12rem] items-center gap-1.5 rounded-full border border-[#ecdccf] bg-[#fff7f0] px-3 py-1.5 text-[11px] text-[#8f6a54] sm:max-w-none sm:gap-2 sm:px-4 sm:py-2 sm:text-sm">
              <FilePenLine className="shrink-0 sm:h-4 sm:w-4" size={14} />
              <span className="truncate whitespace-nowrap">{draftLabel}</span>
            </div>

            <button
              className="shrink-0 rounded-full border border-[#dccbbb] px-3 py-1.5 text-[12px] font-medium text-[#6a5241] transition-colors hover:border-[#c7b19e] hover:text-[#2f221b] disabled:cursor-not-allowed disabled:opacity-50 sm:px-4 sm:py-2 sm:text-base"
              disabled={saveDisabled}
              onClick={onSave}
              type="button"
            >
              保存
            </button>

            <button
              className="shrink-0 rounded-full px-2 py-1.5 text-[12px] font-medium text-[#5e4636] transition-colors hover:text-[#2f221b] sm:px-4 sm:py-2 sm:text-lg"
              onClick={onSaveAndExit}
              type="button"
            >
              保存并退出
            </button>
          </div>
        </div>

        {invalidQuestionNumber !== null ? (
          <p className="mt-2 text-center text-xs text-[#a35b43] sm:text-sm">
            保存失败，注意题目：{invalidQuestionNumber}
          </p>
        ) : null}
      </div>
    </div>
  )
}
