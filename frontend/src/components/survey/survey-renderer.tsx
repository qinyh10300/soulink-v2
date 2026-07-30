import { Select } from '@/components/ui/select'
import { QuestionDualSubQuestion } from '@/components/survey/question-dual-sub-question'
import { QuestionMultiSelect } from '@/components/survey/question-multi-select'
import { QuestionNumber } from '@/components/survey/question-number'
import { QuestionRange } from '@/components/survey/question-range'
import { QuestionScale } from '@/components/survey/question-scale'
import { QuestionSingleChoice } from '@/components/survey/question-single-choice'
import type {
  SurveyAnswerValue,
  SurveyDualQuestion,
  SurveyDualMultiValue,
  SurveyDualScaleValue,
  SurveyQuestion,
  SurveyRangeValue,
} from '@/types/survey'

interface SurveyRendererProps {
  question: SurveyQuestion
  answer: SurveyAnswerValue
  onAnswerChange: (value: SurveyAnswerValue) => void
  disabled?: boolean
}

function getDualDefaultValue(question: Extract<SurveyQuestion, { kind: 'dual' }>) {
  if (question.subQuestions[0].kind === 'scale') {
    return { a: null, b: null } satisfies SurveyDualScaleValue
  }

  return { a: [], b: [] } satisfies SurveyDualMultiValue
}

function isRangeValue(answer: SurveyAnswerValue): answer is SurveyRangeValue {
  return (
    typeof answer === 'object' &&
    answer !== null &&
    !Array.isArray(answer) &&
    ('unlimited' in answer || ('min' in answer && 'max' in answer))
  )
}

function getResolvedDualValue(
  question: SurveyDualQuestion,
  answer: SurveyAnswerValue,
): SurveyDualScaleValue | SurveyDualMultiValue {
  if (typeof answer === 'object' && answer !== null && !Array.isArray(answer)) {
    return answer as SurveyDualScaleValue | SurveyDualMultiValue
  }

  return getDualDefaultValue(question)
}

export function SurveyRenderer({
  question,
  answer,
  onAnswerChange,
  disabled = false,
}: SurveyRendererProps) {
  switch (question.kind) {
    case 'binary':
    case 'single':
      return (
        <QuestionSingleChoice
          disabled={disabled}
          onChange={(value) => onAnswerChange(value)}
          options={question.options}
          value={typeof answer === 'string' ? answer : null}
        />
      )
    case 'select':
      return (
        <div className="max-w-[22rem]">
          <Select
            disabled={disabled}
            onChange={(event) => onAnswerChange(event.target.value)}
            value={typeof answer === 'string' ? answer : ''}
          >
            <option disabled value="">
              {question.placeholder ?? 'Please select'}
            </option>
            {question.options.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        </div>
      )
    case 'range':
      return (
        <QuestionRange
          disabled={disabled}
          onChange={(value) => onAnswerChange(value)}
          question={question}
          value={isRangeValue(answer) ? answer : question.defaultValue}
        />
      )
    case 'scale':
      return (
        <QuestionScale
          disabled={disabled}
          leftLabel={question.leftLabel}
          max={question.max}
          min={question.min}
          onChange={(value) => onAnswerChange(value)}
          rightLabel={question.rightLabel}
          value={typeof answer === 'number' ? answer : null}
        />
      )
    case 'number':
      return (
        <QuestionNumber
          disabled={disabled}
          onChange={(value) => onAnswerChange(value)}
          question={question}
          value={typeof answer === 'number' ? answer : null}
        />
      )
    case 'text':
      return (
        <textarea
          className="min-h-[4.5rem] w-full rounded-[18px] border border-[#ded3ca] bg-white/88 px-3 py-2.5 text-[13px] text-[#2d2623] shadow-[0_6px_16px_rgba(95,70,55,0.04)] outline-none transition-all placeholder:text-[#b4a59a] focus:border-[#b98b68] focus:ring-4 focus:ring-[#b98b68]/12"
          disabled={disabled}
          maxLength={question.maxLength}
          onChange={(event) => onAnswerChange(event.target.value)}
          placeholder={question.placeholder ?? 'Enter text'}
          value={typeof answer === 'string' ? answer : ''}
        />
      )
    case 'multi':
      return (
        <QuestionMultiSelect
          disabled={disabled}
          onChange={(value) => onAnswerChange(value)}
          question={question}
          value={Array.isArray(answer) ? answer : []}
        />
      )
    case 'dual': {
      const value = getResolvedDualValue(question, answer)

      return (
        <div className="space-y-2.5">
          {question.subQuestions.map((subQuestion) => (
            <QuestionDualSubQuestion
              disabled={disabled}
              key={subQuestion.key}
              onChange={(nextValue) =>
                onAnswerChange({
                  ...value,
                  [subQuestion.key]: nextValue,
                })
              }
              question={subQuestion}
              value={value[subQuestion.key] ?? null}
            />
          ))}
        </div>
      )
    }
    default:
      return null
  }
}
