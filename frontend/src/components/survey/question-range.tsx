import { cn } from '@/lib/utils'
import type { SurveyRangeQuestion, SurveyRangeValue } from '@/types/survey'

interface QuestionRangeProps {
  question: SurveyRangeQuestion
  value: SurveyRangeValue
  onChange: (value: SurveyRangeValue) => void
  disabled?: boolean
}

function formatValue(value: number, unit?: string) {
  return unit ? `${value} ${unit}` : String(value)
}

function resolveRangeValue(question: SurveyRangeQuestion, value: SurveyRangeValue) {
  if (question.allowUnlimited && value.unlimited) {
    return {
      min: question.min,
      max: question.max,
    }
  }

  return {
    min: value.min ?? question.min,
    max: value.max ?? question.max,
  }
}

export function QuestionRange({
  question,
  value,
  onChange,
  disabled = false,
}: QuestionRangeProps) {
  const resolvedValue = resolveRangeValue(question, value)
  const isUnlimited = question.allowUnlimited && value.unlimited

  return (
    <div className="space-y-4">
      {question.allowUnlimited ? (
        <button
          className={cn(
            'inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] transition-colors',
            isUnlimited
              ? 'border-[#8f6548] bg-[#fff7f1] text-[#7a5038]'
              : 'border-[#e5d8cf] bg-white text-[#7e6d61] hover:border-[#d5c5b8]',
            disabled ? 'cursor-not-allowed opacity-60 hover:border-[#e5d8cf]' : '',
          )}
          disabled={disabled}
          onClick={() =>
            onChange(
              isUnlimited
                ? {
                    min: question.min,
                    max: question.max,
                  }
                : { unlimited: true },
            )
          }
          type="button"
        >
          {isUnlimited ? '已选择不限' : '选择不限'}
        </button>
      ) : null}

      <div
        className={cn(
          isUnlimited ? 'pointer-events-none opacity-45' : '',
          disabled ? 'pointer-events-none opacity-60' : '',
        )}
      >
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between gap-4 text-[12px] font-medium text-[#7e6d61] sm:text-[13px]">
              <span>{question.lowerLabel}</span>
              <span className="font-serif text-[1.15rem] text-[#8f6548] sm:text-[1.3rem]">
                {formatValue(resolvedValue.min, question.unit)}
              </span>
            </div>
            <input
              className="h-2 w-full cursor-pointer appearance-none rounded-full bg-[#eee5dc] accent-[#8f6548]"
              disabled={disabled}
              max={question.max}
              min={question.min}
              onChange={(event) => {
                const nextMin = Math.min(Number(event.target.value), resolvedValue.max)
                onChange({
                  min: nextMin,
                  max: resolvedValue.max,
                })
              }}
              step={question.step}
              type="range"
              value={resolvedValue.min}
            />
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between gap-4 text-[12px] font-medium text-[#7e6d61] sm:text-[13px]">
              <span>{question.upperLabel}</span>
              <span className="font-serif text-[1.15rem] text-[#8f6548] sm:text-[1.3rem]">
                {formatValue(resolvedValue.max, question.unit)}
              </span>
            </div>
            <input
              className="h-2 w-full cursor-pointer appearance-none rounded-full bg-[#eee5dc] accent-[#8f6548]"
              disabled={disabled}
              max={question.max}
              min={question.min}
              onChange={(event) => {
                const nextMax = Math.max(Number(event.target.value), resolvedValue.min)
                onChange({
                  min: resolvedValue.min,
                  max: nextMax,
                })
              }}
              step={question.step}
              type="range"
              value={resolvedValue.max}
            />
          </div>
        </div>
      </div>
    </div>
  )
}
