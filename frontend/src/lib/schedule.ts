const THURSDAY = 4
const QUESTIONNAIRE_DEADLINE_HOUR = 14

export type QuestionnaireCollectionStatus =
  | 'always_open'
  | 'scheduled'
  | 'open'
  | 'closed'

export interface QuestionnaireCollectionWindow {
  acceptingResponses: boolean
  collectionStatus: QuestionnaireCollectionStatus
  collectionEndsAt?: string
}

export const MATCH_REVEAL_WINDOW_TEXT = '每周四晚上 20:00-21:00 开启匹配结果揭晓'
export const MATCH_SIGNUP_DEADLINE_TEXT = '周四14:00截止问卷填写和匹配报名'
export const MATCH_REVEAL_WINDOW_SHORT_TEXT = '周四 20:00-21:00'
export const QUESTIONNAIRE_DEADLINE_TEXT = '每周四下午 14:00 截止'
export const QUESTIONNAIRE_DEADLINE_SHORT_TEXT = '周四 14:00'

function getWeeklyQuestionnaireDeadline(baseDate: Date) {
  const deadline = new Date(baseDate)
  const day = deadline.getDay()
  const daysUntilThursday = (THURSDAY - day + 7) % 7

  deadline.setDate(deadline.getDate() + daysUntilThursday)
  deadline.setHours(QUESTIONNAIRE_DEADLINE_HOUR, 0, 0, 0)

  if (day === 0 || day > THURSDAY) {
    deadline.setDate(deadline.getDate() - 7)
  }

  return deadline
}

export function resolveQuestionnaireDeadline(
  window: QuestionnaireCollectionWindow,
  now = new Date(),
) {
  if (window.collectionEndsAt) {
    const explicitDeadline = new Date(window.collectionEndsAt)

    if (!Number.isNaN(explicitDeadline.getTime())) {
      return explicitDeadline
    }
  }

  if (window.collectionStatus === 'always_open') {
    return null
  }

  return getWeeklyQuestionnaireDeadline(now)
}

export function isQuestionnaireEditable(
  window: QuestionnaireCollectionWindow,
  now = new Date(),
) {
  if (!window.acceptingResponses || window.collectionStatus === 'closed') {
    return false
  }

  const deadline = resolveQuestionnaireDeadline(window, now)

  if (!deadline) {
    return true
  }

  return now.getTime() < deadline.getTime()
}

export function getNextMatchRevealStart() {
  const now = new Date()
  const next = new Date(now)
  const day = next.getDay()
  const daysUntilThursday = (THURSDAY - day + 7) % 7

  next.setDate(next.getDate() + daysUntilThursday)
  next.setHours(20, 0, 0, 0)

  if (daysUntilThursday === 0) {
    const revealWindowEnd = new Date(next)
    revealWindowEnd.setHours(21, 0, 0, 0)

    if (now.getTime() < revealWindowEnd.getTime()) {
      return next.toISOString()
    }
  }

  if (next.getTime() <= now.getTime()) {
    next.setDate(next.getDate() + 7)
  }

  return next.toISOString()
}

export function formatMatchRevealStartLabel(value: Date) {
  const weekdays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

  return `${value.getFullYear()} 年 ${value.getMonth() + 1} 月 ${value.getDate()} 日 ${weekdays[value.getDay()]} 20:00 开始`
}
