export const SURVEY_DECLARED_QUESTION_COUNT = 52
export const SURVEY_RENDERED_QUESTION_COUNT = SURVEY_DECLARED_QUESTION_COUNT
export const SURVEY_MISSING_QUESTION_NUMBERS: number[] = []

export function getSurveyDisplayProgress(
  answeredCount: number,
  isComplete = false,
) {
  const totalCount = SURVEY_DECLARED_QUESTION_COUNT

  return {
    totalCount,
    progressPercent: isComplete ? 100 : Math.round((answeredCount / totalCount) * 100),
  }
}
