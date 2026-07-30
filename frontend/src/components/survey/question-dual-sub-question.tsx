import type { SurveyDualSubQuestion } from '@/types/survey'
import { QuestionScale } from '@/components/survey/question-scale'
import { cn } from '@/lib/utils'

interface QuestionDualSubQuestionProps {
  question: SurveyDualSubQuestion
  value: number | string[] | null
  onChange: (value: number | string[]) => void
  disabled?: boolean
}

export function QuestionDualSubQuestion({
  question,
  value,
  onChange,
  disabled = false,
}: QuestionDualSubQuestionProps) {
  return (
    <div className="space-y-2.5 rounded-[18px] border border-dashed border-[#ede3da] bg-[#fdfaf7] px-3 py-3">
      <h4 className="text-[14px] font-medium leading-[1.35rem] text-[#332821] sm:text-[15px] sm:leading-6">{question.title}</h4>

      {question.kind === 'scale' ? (
        <QuestionScale
          disabled={disabled}
          leftLabel={question.leftLabel}
          max={question.max}
          min={question.min}
          onChange={(nextValue) => onChange(nextValue)}
          rightLabel={question.rightLabel}
          value={typeof value === 'number' ? value : null}
        />
      ) : (
        <div className="space-y-2.5">
          <p className="text-[12px] text-[#8a786c] sm:text-[13px]">
            已选择 {Array.isArray(value) ? value.length : 0} 项，最多 {question.maxSelections}{' '}
            项
          </p>
          {question.groups.map((group) => {
            const selectedValues = Array.isArray(value) ? value : []

            return (
              <div className="space-y-2" key={group.label}>
                <h5 className="text-[14px] font-medium text-[#6c5c51]">{group.label}</h5>
                <div className="flex flex-wrap gap-1.5">
                  {group.options.map((option) => {
                    const active = selectedValues.includes(option.value)

                    return (
                      <button
                        className={cn(
                          'rounded-full border px-3 py-1.5 text-[12px] transition-all',
                          active
                            ? 'border-[#9e6f57] bg-[#fff7f1] font-medium text-[#82553d] shadow-[0_10px_22px_rgba(123,86,59,0.12)]'
                            : 'border-[#e6ddd4] bg-white text-[#6c5d53] hover:border-[#d5c5b8] hover:bg-[#fdfaf7]',
                          disabled ? 'cursor-not-allowed opacity-60 hover:border-[#e6ddd4] hover:bg-white' : '',
                        )}
                        disabled={disabled}
                        key={option.value}
                        onClick={() => {
                          if (active) {
                            onChange(
                              selectedValues.filter((item) => item !== option.value),
                            )
                            return
                          }

                          if (selectedValues.length >= question.maxSelections) {
                            return
                          }

                          onChange([...selectedValues, option.value])
                        }}
                        type="button"
                      >
                        {option.label}
                      </button>
                    )
                  })}
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
