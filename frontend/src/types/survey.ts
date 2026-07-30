export type SurveySectionId = string

export interface SurveyOption {
  label: string
  value: string
}

export interface SurveyOptionGroup {
  label: string
  options: SurveyOption[]
}

export interface SurveyRangeValue {
  min?: number
  max?: number
  unlimited?: boolean
}

export interface SurveyDualScaleValue {
  a: number | null
  b: number | null
}

export interface SurveyDualMultiValue {
  a: string[]
  b: string[]
}

export type SurveyAnswerValue =
  | string
  | string[]
  | number
  | SurveyRangeValue
  | SurveyDualScaleValue
  | SurveyDualMultiValue
  | null

interface SurveyQuestionBase {
  id: string
  number: number
  title: string
  required: boolean
  importantEligible?: boolean
  note?: string
  placeholder?: string
}

export interface SurveyBinaryQuestion extends SurveyQuestionBase {
  kind: 'binary'
  options: SurveyOption[]
}

export interface SurveySingleChoiceQuestion extends SurveyQuestionBase {
  kind: 'single'
  options: SurveyOption[]
}

export interface SurveySelectQuestion extends SurveyQuestionBase {
  kind: 'select'
  options: SurveyOption[]
}

export interface SurveyRangeQuestion extends SurveyQuestionBase {
  kind: 'range'
  min: number
  max: number
  step: number
  unit?: string
  lowerLabel: string
  upperLabel: string
  allowUnlimited?: boolean
  defaultValue: SurveyRangeValue
}

export interface SurveyScaleQuestion extends SurveyQuestionBase {
  kind: 'scale'
  min: number
  max: number
  leftLabel: string
  rightLabel: string
}

export interface SurveyNumberQuestion extends SurveyQuestionBase {
  kind: 'number'
  min?: number
  max?: number
  unit?: string
}

export interface SurveyTextQuestion extends SurveyQuestionBase {
  kind: 'text'
  maxLength?: number
}

export interface SurveyMultiSelectQuestion extends SurveyQuestionBase {
  kind: 'multi'
  minSelections: number
  maxSelections: number
  groups: SurveyOptionGroup[]
}

export interface SurveyDualScaleSubQuestion {
  key: 'a' | 'b'
  kind: 'scale'
  title: string
  min: number
  max: number
  leftLabel: string
  rightLabel: string
}

export interface SurveyDualMultiSubQuestion {
  key: 'a' | 'b'
  kind: 'multi'
  title: string
  minSelections: number
  maxSelections: number
  groups: SurveyOptionGroup[]
}

export type SurveyDualSubQuestion =
  | SurveyDualScaleSubQuestion
  | SurveyDualMultiSubQuestion

export interface SurveyDualQuestion extends SurveyQuestionBase {
  kind: 'dual'
  subQuestions: [SurveyDualSubQuestion, SurveyDualSubQuestion]
}

export type SurveyQuestion =
  | SurveyBinaryQuestion
  | SurveySingleChoiceQuestion
  | SurveySelectQuestion
  | SurveyRangeQuestion
  | SurveyScaleQuestion
  | SurveyNumberQuestion
  | SurveyTextQuestion
  | SurveyMultiSelectQuestion
  | SurveyDualQuestion

export interface SurveySection {
  id: SurveySectionId
  title: string
  subtitle: string
  description: string
  tip: string
  questions: SurveyQuestion[]
}

export interface SurveyDraft {
  currentSectionId: SurveySectionId
  answers: Record<string, SurveyAnswerValue>
  importantQuestionIds: string[]
  updatedAt: string | null
  status: 'draft' | 'completed'
}

export interface SurveySectionProgress {
  sectionId: SurveySectionId
  answered: number
  total: number
  completed: boolean
}

export interface SurveySummary {
  answeredCount: number
  totalCount: number
  completionPercent: number
  isComplete: boolean
  completedSectionIds: SurveySectionId[]
  sections: SurveySectionProgress[]
}
