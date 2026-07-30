import type { SurveyOption } from '@/types/survey'
import { cn } from '@/lib/utils'

interface QuestionSingleChoiceProps {
  options: SurveyOption[]
  value: string | null
  onChange: (value: string) => void
  disabled?: boolean
}

export function QuestionSingleChoice({
  options,
  value,
  onChange,
  disabled = false,
}: QuestionSingleChoiceProps) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {options.map((option) => {
        const active = value === option.value

        return (
          <button
            className={cn(
              'rounded-full border px-3 py-1.5 text-[13px] font-medium transition-all',
              active
                ? 'border-[#8f6548] bg-[#fff7f1] text-[#7a5038] shadow-[0_10px_22px_rgba(123,86,59,0.12)]'
                : 'border-[#e7ddd3] bg-white text-[#6d5f55] hover:border-[#d5c5b8] hover:bg-[#fdfaf7]',
              disabled ? 'cursor-not-allowed opacity-60 hover:border-[#e7ddd3] hover:bg-white' : '',
            )}
            disabled={disabled}
            key={option.value}
            onClick={() => onChange(option.value)}
            type="button"
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}
