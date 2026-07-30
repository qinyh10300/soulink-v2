import { useEffect, useMemo, useState } from 'react'

import { Input } from '@/components/ui/input'
import type { SurveyNumberQuestion } from '@/types/survey'

interface QuestionNumberProps {
  question: SurveyNumberQuestion
  value: number | null
  onChange: (value: number | null) => void
  disabled?: boolean
}

function clampToBounds(value: number, min?: number, max?: number) {
  if (typeof min === 'number' && value < min) {
    return min
  }

  if (typeof max === 'number' && value > max) {
    return max
  }

  return value
}

function countDigits(value: number) {
  return Math.abs(value).toString().length
}

function canCompleteWithinRange(raw: string, min?: number, max?: number) {
  if (raw.length === 0) {
    return true
  }

  if (!/^\d+$/.test(raw)) {
    return false
  }

  if (typeof min !== 'number' && typeof max !== 'number') {
    return true
  }

  const prefix = Number(raw)

  if (!Number.isFinite(prefix)) {
    return false
  }

  const maxDigits =
    typeof max === 'number'
      ? countDigits(max)
      : Math.max(raw.length, typeof min === 'number' ? countDigits(min) : raw.length)

  for (let totalDigits = raw.length; totalDigits <= maxDigits; totalDigits += 1) {
    const multiplier = 10 ** (totalDigits - raw.length)
    const candidateMin = prefix * multiplier
    const candidateMax = candidateMin + (multiplier - 1)
    const lowerBound = typeof min === 'number' ? min : Number.NEGATIVE_INFINITY
    const upperBound = typeof max === 'number' ? max : Number.POSITIVE_INFINITY

    if (candidateMax >= lowerBound && candidateMin <= upperBound) {
      return true
    }
  }

  return false
}

export function QuestionNumber({
  question,
  value,
  onChange,
  disabled = false,
}: QuestionNumberProps) {
  const resolvedValue = useMemo(() => {
    if (typeof value !== 'number' || Number.isNaN(value)) {
      return null
    }

    return clampToBounds(value, question.min, question.max)
  }, [question.max, question.min, value])

  const [inputValue, setInputValue] = useState(
    resolvedValue === null ? '' : String(resolvedValue),
  )

  useEffect(() => {
    setInputValue(resolvedValue === null ? '' : String(resolvedValue))
  }, [resolvedValue])

  useEffect(() => {
    if (typeof value === 'number' && resolvedValue !== value) {
      onChange(resolvedValue)
    }
  }, [onChange, resolvedValue, value])

  const maxLength = typeof question.max === 'number' ? countDigits(question.max) : undefined

  return (
    <div className="max-w-[16rem]">
      <Input
        disabled={disabled}
        inputMode="numeric"
        maxLength={maxLength}
        onBlur={() => {
          if (inputValue.length === 0) {
            onChange(null)
            return
          }

          const normalized = clampToBounds(Number(inputValue), question.min, question.max)
          setInputValue(String(normalized))
          onChange(normalized)
        }}
        onChange={(event) => {
          const nextValue = event.target.value.trim()

          if (nextValue === '') {
            setInputValue('')
            onChange(null)
            return
          }

          if (!canCompleteWithinRange(nextValue, question.min, question.max)) {
            return
          }

          setInputValue(nextValue)

          const parsedValue = Number(nextValue)
          const isWithinRange =
            !Number.isNaN(parsedValue) &&
            (typeof question.min !== 'number' || parsedValue >= question.min) &&
            (typeof question.max !== 'number' || parsedValue <= question.max)

          onChange(isWithinRange ? parsedValue : null)
        }}
        pattern="[0-9]*"
        placeholder={question.placeholder ?? 'Enter a number'}
        type="text"
        value={inputValue}
      />
    </div>
  )
}
