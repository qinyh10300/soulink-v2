import type {
  SurveyAnswerValue,
  SurveyDraft,
  SurveyDualQuestion,
  SurveyMultiSelectQuestion,
  SurveyQuestion,
  SurveySection,
  SurveySectionId,
  SurveySummary,
} from '@/types/survey'
import { getSurveyDisplayProgress } from '@/lib/survey-metadata'

const SURVEY_DRAFT_KEY = 'soulink:survey-draft'
const DASHBOARD_ANNOUNCEMENT_KEY = 'soulink:dashboard-announcement'

function hasWindow() {
  return typeof window !== 'undefined'
}

export function createEmptySurveyDraft(
  firstSectionId: SurveySectionId,
): SurveyDraft {
  return {
    currentSectionId: firstSectionId,
    answers: {},
    importantQuestionIds: [],
    updatedAt: null,
    status: 'draft',
  }
}

export function getSurveyDraft(
  firstSectionId: SurveySectionId,
): SurveyDraft {
  if (!hasWindow()) {
    return createEmptySurveyDraft(firstSectionId)
  }

  const raw = window.localStorage.getItem(SURVEY_DRAFT_KEY)

  if (!raw) {
    return createEmptySurveyDraft(firstSectionId)
  }

  try {
    const parsed = JSON.parse(raw) as SurveyDraft

    return {
      currentSectionId: parsed.currentSectionId ?? firstSectionId,
      answers: parsed.answers ?? {},
      importantQuestionIds: parsed.importantQuestionIds ?? [],
      updatedAt: parsed.updatedAt ?? null,
      status: parsed.status === 'completed' ? 'completed' : 'draft',
    }
  } catch {
    return createEmptySurveyDraft(firstSectionId)
  }
}

export function saveSurveyDraft(draft: SurveyDraft) {
  if (!hasWindow()) {
    return
  }

  window.localStorage.setItem(SURVEY_DRAFT_KEY, JSON.stringify(draft))
}

export function clearSurveyDraft() {
  if (!hasWindow()) {
    return
  }

  window.localStorage.removeItem(SURVEY_DRAFT_KEY)
}

function isMultiAnswerValid(
  question: SurveyMultiSelectQuestion,
  answer: SurveyAnswerValue,
) {
  if (!Array.isArray(answer)) {
    return false
  }

  return (
    answer.length >= question.minSelections &&
    answer.length <= question.maxSelections
  )
}

function isDualAnswerValid(question: SurveyDualQuestion, answer: SurveyAnswerValue) {
  if (!answer || typeof answer !== 'object' || Array.isArray(answer)) {
    return false
  }

  if (
    question.subQuestions[0].kind === 'scale' &&
    question.subQuestions[1].kind === 'scale'
  ) {
    const value = answer as { a?: number | null; b?: number | null }

    return typeof value.a === 'number' && typeof value.b === 'number'
  }

  const value = answer as { a?: string[]; b?: string[] }

  return question.subQuestions.every((subQuestion) => {
    if (subQuestion.kind !== 'multi') {
      return false
    }

    const selected = value[subQuestion.key]

    return (
      Array.isArray(selected) &&
      selected.length >= subQuestion.minSelections &&
      selected.length <= subQuestion.maxSelections
    )
  })
}

function isRangeAnswerValue(answer: SurveyAnswerValue): answer is Extract<
  SurveyAnswerValue,
  { min?: number; max?: number; unlimited?: boolean }
> {
  return (
    typeof answer === 'object' &&
    answer !== null &&
    !Array.isArray(answer) &&
    ('unlimited' in answer || ('min' in answer && 'max' in answer))
  )
}

export function isQuestionAnswered(
  question: SurveyQuestion,
  answer: SurveyAnswerValue,
) {
  if (answer === null || answer === undefined) {
    return false
  }

  switch (question.kind) {
    case 'binary':
    case 'single':
    case 'select':
      return typeof answer === 'string' && answer.trim().length > 0
    case 'scale':
      return typeof answer === 'number'
    case 'number':
      return (
        typeof answer === 'number' &&
        (typeof question.min !== 'number' || answer >= question.min) &&
        (typeof question.max !== 'number' || answer <= question.max)
      )
    case 'text':
      return typeof answer === 'string' && answer.trim().length > 0
    case 'range':
      return (
        isRangeAnswerValue(answer) &&
        ((question.allowUnlimited === true && answer.unlimited === true) ||
          (typeof answer.min === 'number' &&
            typeof answer.max === 'number' &&
            answer.min <= answer.max))
      )
    case 'multi':
      return isMultiAnswerValid(question, answer)
    case 'dual':
      return isDualAnswerValid(question, answer)
    default:
      return false
  }
}

export function getSurveySummary(
  sections: SurveySection[],
  draft: SurveyDraft,
): SurveySummary {
  let answeredCount = 0
  let requiredAnsweredCount = 0
  let requiredTotalCount = 0

  const sectionProgress = sections.map((section) => {
    const allQuestions = section.questions
    const requiredQuestions = section.questions.filter((question) => question.required)
    const answeredAll = allQuestions.filter((question) =>
      isQuestionAnswered(question, draft.answers[question.id]),
    ).length
    const answered = requiredQuestions.filter((question) =>
      isQuestionAnswered(question, draft.answers[question.id]),
    ).length

    answeredCount += answeredAll
    requiredTotalCount += requiredQuestions.length
    requiredAnsweredCount += answered

    return {
      sectionId: section.id,
      answered,
      total: requiredQuestions.length,
      completed: answered === requiredQuestions.length,
    }
  })

  const completedSectionIds = sectionProgress
    .filter((section) => section.completed)
    .map((section) => section.sectionId)
  const isComplete =
    requiredAnsweredCount === requiredTotalCount && requiredTotalCount > 0
  const displayProgress = getSurveyDisplayProgress(answeredCount, isComplete)

  return {
    answeredCount,
    totalCount: displayProgress.totalCount,
    completionPercent: displayProgress.progressPercent,
    isComplete,
    completedSectionIds,
    sections: sectionProgress,
  }
}

export function hasSeenDashboardAnnouncement() {
  if (!hasWindow()) {
    return true
  }

  return window.sessionStorage.getItem(DASHBOARD_ANNOUNCEMENT_KEY) === 'seen'
}

export function markDashboardAnnouncementSeen() {
  if (!hasWindow()) {
    return
  }

  window.sessionStorage.setItem(DASHBOARD_ANNOUNCEMENT_KEY, 'seen')
}

export function resetDashboardAnnouncementSeen() {
  if (!hasWindow()) {
    return
  }

  window.sessionStorage.removeItem(DASHBOARD_ANNOUNCEMENT_KEY)
}
