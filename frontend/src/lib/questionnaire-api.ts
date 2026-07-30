import { apiRequest } from '@/lib/api'

export interface QuestionnaireOption {
  label: string
  value: string
}

export interface QuestionnaireQuestion {
  id: string
  order: number
  section: string
  title: string
  type: 'single_choice' | 'multiple_choice' | 'number' | 'range' | 'scale' | 'text'
  widget?: string
  required: boolean
  options?: QuestionnaireOption[]
  min?: number
  max?: number
  leftLabel?: string
  rightLabel?: string
  minSelections?: number
  maxSelections?: number
  maxLength?: number
  placeholder?: string
  allowUnlimited?: boolean
  supportsImportance?: boolean
}

export interface QuestionnaireSection {
  id: string
  title: string
  subtitle?: string
  description?: string
  tip?: string
}

export interface QuestionnaireAnswer {
  questionId: string
  value: unknown
  importance?: boolean
  updatedAt: string
}

export interface QuestionnaireCurrentResult {
  questionnaireId: number
  questionnaireKey: string
  title: string
  version: string
  collectionStatus: 'always_open' | 'scheduled' | 'open' | 'closed'
  acceptingResponses: boolean
  collectionStartsAt?: string
  collectionEndsAt?: string
  collectionTimezone?: string
  resultRevealAt?: string
  totalCount: number
  answeredCount: number
  progressPercent: number
  hasCompletedQuestionnaire: boolean
  participatingInMatch: boolean
  sections: QuestionnaireSection[]
  questions: QuestionnaireQuestion[]
  answers: QuestionnaireAnswer[]
}

export interface QuestionnaireProgressResult {
  questionnaireId: number
  questionnaireKey: string
  version: string
  collectionStatus: 'always_open' | 'scheduled' | 'open' | 'closed'
  acceptingResponses: boolean
  collectionStartsAt?: string
  collectionEndsAt?: string
  collectionTimezone?: string
  resultRevealAt?: string
  totalCount: number
  answeredCount: number
  progressPercent: number
  hasCompletedQuestionnaire: boolean
  participatingInMatch: boolean
}

export interface SaveQuestionnaireAnswersResult {
  resultRevealAt?: string
  answeredCount: number
  progressPercent: number
  hasCompletedQuestionnaire: boolean
  participatingInMatch: boolean
}

export interface UpdateQuestionnaireParticipationResult {
  resultRevealAt?: string
  hasCompletedQuestionnaire: boolean
  participatingInMatch: boolean
}

export interface QuestionnairePortraitResult {
  soulSignature: string
  personalProfile: string
  idealResonance: string
  traits: Record<string, unknown>
}

export function fetchCurrentQuestionnaire() {
  return apiRequest<QuestionnaireCurrentResult>('/questionnaires/current', {
    auth: true,
  })
}

export function fetchQuestionnaireProgress() {
  return apiRequest<QuestionnaireProgressResult>('/questionnaires/current/progress', {
    auth: true,
  })
}

export function saveQuestionnaireAnswers(input: {
  answers: Array<{
    questionId: string
    value: unknown
    importance?: boolean
  }>
  action?: 'save' | 'save_exit' | 'submit'
}) {
  return apiRequest<SaveQuestionnaireAnswersResult>('/questionnaires/current/answers', {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function fetchQuestionnairePortrait() {
  return apiRequest<QuestionnairePortraitResult>('/questionnaires/current/portrait', {
    auth: true,
  })
}

export function updateQuestionnaireParticipation(participatingInMatch: boolean) {
  return apiRequest<UpdateQuestionnaireParticipationResult>('/questionnaires/current/participation', {
    method: 'PUT',
    auth: true,
    body: {
      participatingInMatch,
    },
  })
}
