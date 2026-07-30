import type {
  QuestionnaireAnswer,
  QuestionnaireCurrentResult,
  QuestionnaireQuestion,
  QuestionnaireSection,
} from '@/lib/questionnaire-api'
import { surveySections as surveySectionTemplates } from '@/mock/survey-data'
import type {
  SurveyDraft,
  SurveyQuestion,
  SurveyRangeValue,
  SurveySection,
} from '@/types/survey'

function normalizeSectionId(section: string, index: number) {
  const normalized = section
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-+|-+$/g, '')

  return normalized || `section-${index + 1}`
}

function toSurveyQuestion(question: QuestionnaireQuestion): SurveyQuestion {
  const base = {
    id: question.id,
    number: question.order,
    title: question.title,
    required: question.required,
    importantEligible: question.supportsImportance ?? false,
    placeholder: question.placeholder,
  }

  switch (question.type) {
    case 'single_choice':
      if (question.widget === 'select') {
        return {
          ...base,
          kind: 'select',
          options: question.options ?? [],
        }
      }

      return {
        ...base,
        kind: 'single',
        options: question.options ?? [],
      }
    case 'multiple_choice':
      return {
        ...base,
        kind: 'multi',
        minSelections: question.minSelections ?? 1,
        maxSelections: question.maxSelections ?? Math.max(question.options?.length ?? 1, 1),
        groups: [
          {
            label: '可选项',
            options: question.options ?? [],
          },
        ],
      }
    case 'number':
      return {
        ...base,
        kind: 'number',
        min: question.min,
        max: question.max,
      }
    case 'range':
      return {
        ...base,
        kind: 'range',
        min: question.min ?? 0,
        max: question.max ?? 100,
        step: 1,
        lowerLabel: question.leftLabel || '最小值',
        upperLabel: question.rightLabel || '最大值',
        allowUnlimited: question.allowUnlimited ?? false,
        defaultValue: {
          min: question.min ?? 0,
          max: question.max ?? 100,
        } satisfies SurveyRangeValue,
      }
    case 'scale':
      return {
        ...base,
        kind: 'scale',
        min: question.min ?? 1,
        max: question.max ?? 7,
        leftLabel: question.leftLabel || '左',
        rightLabel: question.rightLabel || '右',
      }
    case 'text':
    default:
      return {
        ...base,
        kind: 'text',
        maxLength: question.maxLength,
        placeholder: question.placeholder || '请输入内容',
      }
  }
}

function toAnswerMap(answers: QuestionnaireAnswer[]) {
  return answers.reduce<Record<string, SurveyDraft['answers'][string]>>((acc, answer) => {
    acc[answer.questionId] = answer.value as SurveyDraft['answers'][string]
    return acc
  }, {})
}

function latestUpdatedAt(answers: QuestionnaireAnswer[]) {
  return answers.reduce<string | null>((latest, answer) => {
    if (!latest || answer.updatedAt > latest) {
      return answer.updatedAt
    }
    return latest
  }, null)
}

function resolveSurveySectionMeta(
  section: QuestionnaireSection | undefined,
  fallbackTitle: string,
  index: number,
) {
  const template = surveySectionTemplates[index]

  return {
    id: template?.id || section?.id || normalizeSectionId(fallbackTitle, index),
    title: template?.title || section?.title || fallbackTitle,
    subtitle: template?.subtitle ?? section?.subtitle ?? '',
    description: template?.description ?? section?.description ?? '',
    tip: template?.tip ?? section?.tip ?? '',
  }
}

export function toSurveySections(current: QuestionnaireCurrentResult): SurveySection[] {
  const groups = new Map<string, SurveyQuestion[]>()

  current.questions.forEach((question) => {
    const section = question.section || current.sections[0]?.title || '问卷'
    const questions = groups.get(section) ?? []
    questions.push(toSurveyQuestion(question))
    groups.set(section, questions)
  })

  const sections = current.sections
    .map((section, index) => {
      const questions = groups.get(section.title)

      if (!questions || questions.length === 0) {
        return null
      }

      groups.delete(section.title)

      return {
        ...resolveSurveySectionMeta(section, section.title, index),
        questions,
      }
    })
    .filter((section): section is SurveySection => section !== null)

  const nextIndex = sections.length

  Array.from(groups.entries()).forEach(([sectionTitle, questions], offset) => {
    sections.push({
      ...resolveSurveySectionMeta(undefined, sectionTitle, nextIndex + offset),
      questions,
    })
  })

  return sections
}

export function toSurveyDraft(
  current: QuestionnaireCurrentResult,
  preferredSectionId?: string,
): SurveyDraft {
  const sections = toSurveySections(current)
  const firstSectionId = sections[0]?.id ?? 'questionnaire'
  const sectionIds = new Set(sections.map((section) => section.id))

  return {
    currentSectionId:
      preferredSectionId && sectionIds.has(preferredSectionId)
        ? preferredSectionId
        : firstSectionId,
    answers: toAnswerMap(current.answers),
    importantQuestionIds: current.answers
      .filter((answer) => Boolean(answer.importance))
      .map((answer) => answer.questionId),
    updatedAt: latestUpdatedAt(current.answers),
    status: current.hasCompletedQuestionnaire ? 'completed' : 'draft',
  }
}
