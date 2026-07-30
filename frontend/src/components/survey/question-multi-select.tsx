import type { SurveyMultiSelectQuestion } from '@/types/survey'
import { cn } from '@/lib/utils'

interface QuestionMultiSelectProps {
  question: SurveyMultiSelectQuestion
  value: string[]
  onChange: (value: string[]) => void
  disabled?: boolean
}

export function QuestionMultiSelect({
  question,
  value,
  onChange,
  disabled = false,
}: QuestionMultiSelectProps) {
  const toggleValue = (nextValue: string) => {
    const exists = value.includes(nextValue)

    if (exists) {
      onChange(value.filter((item) => item !== nextValue))
      return
    }

    if (value.length >= question.maxSelections) {
      return
    }

    onChange([...value, nextValue])
  }

  return (
    <div className="space-y-4">
      <p className="text-[12px] text-[#8a786c] sm:text-[13px]">
        已选择 {value.length} 项，需选择 {question.minSelections} - {question.maxSelections}{' '}
        项
      </p>

      {question.groups.map((group) => (
        <div className="space-y-2" key={group.label}>
          <h4 className="text-[14px] font-medium text-[#6c5c51]">{group.label}</h4>
          <div className="flex flex-wrap gap-1.5">
            {group.options.map((option) => {
              const active = value.includes(option.value)

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
                  onClick={() => toggleValue(option.value)}
                  type="button"
                >
                  {option.label}
                </button>
              )
            })}
          </div>
        </div>
      ))}
    </div>
  )
}
