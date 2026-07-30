import { Star } from 'lucide-react'
import type { ReactNode } from 'react'

import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import type { SurveyQuestion } from '@/types/survey'

interface SurveyQuestionCardProps {
  question: SurveyQuestion
  important: boolean
  onImportantToggle: () => void
  children: ReactNode
  disabled?: boolean
}

function formatQuestionConstraint(question: SurveyQuestion) {
  const inferUnit = (placeholder?: string) => {
    if (!placeholder) {
      return undefined
    }

    const normalized = placeholder.toLowerCase()
    if (normalized.includes('cm')) {
      return 'cm'
    }

    return undefined
  }

  const formatRangeText = (min?: number, max?: number, unit?: string) => {
    const suffix = unit ? ` ${unit}` : ''

    if (typeof min === 'number' && typeof max === 'number') {
      return `${min}-${max}${suffix}`
    }

    if (typeof min === 'number') {
      return `>= ${min}${suffix}`
    }

    if (typeof max === 'number') {
      return `<= ${max}${suffix}`
    }

    return null
  }

  switch (question.kind) {
    case 'number': {
      const unit = question.unit ?? inferUnit(question.placeholder)
      return unit ? formatRangeText(question.min, question.max, unit) : null
    }
    case 'range':
      return question.unit
        ? formatRangeText(question.min, question.max, question.unit)
        : null
    default:
      return null
  }
}

export function SurveyQuestionCard({
  question,
  important,
  onImportantToggle,
  children,
  disabled = false,
}: SurveyQuestionCardProps) {
  const constraintText = formatQuestionConstraint(question)

  return (
    <Card className="mx-auto w-full max-w-[48rem] border-[#eee4db] bg-[rgba(255,251,247,0.96)] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
      <CardContent className="space-y-2 px-3.5 py-3 sm:px-5 sm:py-3.5">
        <div className="space-y-1.5">
          <div className="flex items-start gap-2">
            <div className="flex h-6.5 w-6.5 shrink-0 items-center justify-center rounded-full bg-[#f4efe9] text-[10px] font-semibold text-[#8f7d70] sm:h-7 sm:w-7 sm:text-[10px]">
              {question.number}
            </div>

            <div className="min-w-0 flex-1 space-y-0.5">
              <div className="flex flex-col items-start gap-1.5 sm:flex-row sm:flex-wrap sm:items-start sm:justify-between">
                <h3 className="min-w-0 w-full text-[0.9rem] font-medium leading-5 text-[#2f241e] sm:min-w-[12rem] sm:flex-1 sm:text-[0.98rem] sm:leading-6">
                  <span>{question.title}</span>
                  {constraintText ? (
                    <span className="ml-1 text-[0.8rem] font-normal text-[#8b776b]">
                      ({constraintText})
                    </span>
                  ) : null}
                </h3>

                {question.importantEligible ? (
                  <button
                    className={cn(
                      'inline-flex shrink-0 items-center gap-1 self-end rounded-full border px-2 py-0.75 text-[10px] transition-colors sm:self-auto sm:px-2.5 sm:text-[10.5px]',
                      important
                        ? 'border-[#e3bf69] bg-[#fff7df] text-[#8d6a19]'
                        : 'border-[#ebe3db] bg-white text-[#8e7d73] hover:border-[#dac9b7] hover:text-[#725b49]',
                      disabled ? 'cursor-not-allowed opacity-60 hover:border-[#ebe3db] hover:text-[#8e7d73]' : '',
                    )}
                    disabled={disabled}
                    onClick={onImportantToggle}
                    type="button"
                  >
                    <Star className={important ? 'fill-current' : ''} size={13} />
                    {important ? '已设为重要' : '设为重要'}
                  </button>
                ) : null}
              </div>

              {question.note ? (
                <p className="max-w-3xl text-[12px] leading-[1.15rem] text-[#8b776b] sm:text-[12.5px] sm:leading-5">
                  {question.note}
                </p>
              ) : null}
            </div>
          </div>
        </div>

        {children}
      </CardContent>
    </Card>
  )
}
