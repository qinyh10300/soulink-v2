import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { DashboardFooter } from '@/components/dashboard/dashboard-footer'
import {
  MatchParticipationModal,
  type WeeklyParticipationAction,
} from '@/components/dashboard/match-participation-modal'
import { QuestionnaireRequiredModal } from '@/components/dashboard/questionnaire-required-modal'
import { SoulPortraitModal } from '@/components/dashboard/soul-portrait-modal'
import { CountdownCard } from '@/components/dashboard/countdown-card'
import { PlatformAnnouncementModal } from '@/components/dashboard/platform-announcement-modal'
import { StatusCard } from '@/components/dashboard/status-card'
import { WelcomeSection } from '@/components/dashboard/welcome-section'
import { getErrorMessage } from '@/lib/api'
import {
  fetchMatchCycleStatus,
  type MatchCycleStatus,
} from '@/lib/match-api'
import {
  fetchQuestionnaireProgress,
  updateQuestionnaireParticipation,
} from '@/lib/questionnaire-api'
import {
  hasSeenDashboardAnnouncement,
  markDashboardAnnouncementSeen,
} from '@/lib/survey-storage'
import { resolveDashboardUser } from '@/lib/user-profile-storage'
import { getNextMatchRevealStart, MATCH_REVEAL_WINDOW_TEXT } from '@/lib/schedule'
import { dashboardContent } from '@/mock/dashboard'

interface SurveyProgressView {
  answeredCount: number
  totalCount: number
  progressPercent: number
  hasCompletedQuestionnaire: boolean
  participatingInMatch: boolean
  resultRevealAt?: string
}

function formatDateTime(value?: string) {
  if (!value) {
    return '待公布'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function hasReachedTime(value?: string) {
  if (!value) {
    return false
  }

  return new Date(value).getTime() <= Date.now()
}

function buildSelectableWeeklyMatchStatus(
  cycleStatus: MatchCycleStatus | null,
  isParticipating: boolean,
  hasCompletedQuestionnaire: boolean,
) {
  if (cycleStatus?.hasCurrentCycle && cycleStatus.revealAt && !hasReachedTime(cycleStatus.revealAt)) {
    const joinedCurrentCycle = cycleStatus.hasUserResult
    return {
      value: joinedCurrentCycle ? '已参加本轮' : '未参加本轮',
      tone: joinedCurrentCycle ? ('active' as const) : ('neutral' as const),
      description: joinedCurrentCycle
        ? `已参加本轮`
        : `未参加本轮`,
      actionLabel: undefined,
    }
  }

  if (!hasCompletedQuestionnaire) {
    return {
      value: '尚未参加',
      tone: 'neutral' as const,
      description: '完成问卷后，才可以选择是否参加下一轮匹配。',
      actionLabel: '参加',
    }
  }

  if (!cycleStatus?.hasCurrentCycle || !cycleStatus.revealAt) {
    return {
      value: isParticipating ? '已选择参加' : '尚未参加',
      tone: isParticipating ? ('active' as const) : ('neutral' as const),
      description: isParticipating
        ? '你已选择加入下一轮匹配。'
        : '当前还没有新的链接周期，你可以先选择是否参加下一轮。',
      actionLabel: isParticipating ? '退出' : '参加',
    }
  }

  return {
    value: isParticipating ? '已选择参加' : '尚未参加',
    tone: isParticipating ? ('active' as const) : ('neutral' as const),
    description: isParticipating
      ? '你已选择加入下一轮匹配，开放后系统会自动纳入。'
      : '你还没有加入下一轮匹配，现在可以先做选择。',
    actionLabel: isParticipating ? '退出' : '参加',
  }
}

export function DashboardPage() {
  const navigate = useNavigate()
  const [announcementOpen, setAnnouncementOpen] = useState(false)
  const [surveyProgress, setSurveyProgress] = useState<SurveyProgressView | null>(null)
  const [cycleStatus, setCycleStatus] = useState<MatchCycleStatus | null>(null)
  const [isParticipationSubmitting, setIsParticipationSubmitting] = useState(false)
  const [matchesNotice, setMatchesNotice] = useState('')
  const [toastNotice, setToastNotice] = useState('')
  const [participationAction, setParticipationAction] = useState<WeeklyParticipationAction | null>(null)
  const [questionnairePromptContext, setQuestionnairePromptContext] = useState<'match' | 'portrait'>('match')
  const [questionnairePromptOpen, setQuestionnairePromptOpen] = useState(false)
  const [portraitOpen, setPortraitOpen] = useState(false)

  const user = useMemo(
    () => resolveDashboardUser(dashboardContent.user),
    [],
  )
  const hasCompletedQuestionnaire = surveyProgress?.hasCompletedQuestionnaire ?? false
  const isParticipating = surveyProgress?.participatingInMatch ?? false
  const heroTargetAt = getNextMatchRevealStart()

  useEffect(() => {
    if (!hasSeenDashboardAnnouncement()) {
      setAnnouncementOpen(true)
    }
  }, [])

  useEffect(() => {
    let cancelled = false

    const loadDashboard = async () => {
      setMatchesNotice('')

      const [progressResult, cycleResult] = await Promise.all([
        fetchQuestionnaireProgress().catch(() => null),
        fetchMatchCycleStatus().catch((error) => error),
      ])

      if (cancelled) {
        return
      }

      if (progressResult) {
        setSurveyProgress({
          answeredCount: progressResult.answeredCount,
          totalCount: progressResult.totalCount,
          progressPercent: progressResult.progressPercent,
          hasCompletedQuestionnaire: progressResult.hasCompletedQuestionnaire,
          participatingInMatch: progressResult.participatingInMatch,
          resultRevealAt: progressResult.resultRevealAt,
        })
      } else {
        setSurveyProgress(null)
      }

      if (cycleResult instanceof Error) {
        setMatchesNotice(getErrorMessage(cycleResult, '配对状态加载失败，请稍后重试'))
      } else {
        setCycleStatus(cycleResult)
      }
    }

    void loadDashboard()

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!toastNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setToastNotice('')
    }, 3000)

    return () => window.clearTimeout(timer)
  }, [toastNotice])

  const weeklyMatchStatus = buildSelectableWeeklyMatchStatus(
    cycleStatus,
    isParticipating,
    hasCompletedQuestionnaire,
  )
  const canChooseWeeklyMatch = Boolean(weeklyMatchStatus.actionLabel)

  const statusItems = dashboardContent.statuses.map((item) => {
    if (item.id === 'survey') {
      if (!surveyProgress) {
        return {
          ...item,
          value: '加载中...',
          tone: 'neutral' as const,
        }
      }

      if (surveyProgress.hasCompletedQuestionnaire) {
        return {
          ...item,
          value: '已完成',
          tone: 'success' as const,
          description: '问卷已全部完成，可以随时查看或修改答案。',
          actionLabel: '查看 / 修改',
          progress: undefined,
        }
      }

      return {
        ...item,
        value: `${surveyProgress.answeredCount} / ${surveyProgress.totalCount}`,
        tone: 'neutral' as const,
        description: '继续填写剩余题目后，才能参与本轮配对。',
        actionLabel: '继续填写',
        progress:
          surveyProgress.progressPercent,
        to: '/survey',
      }
    }

    if (item.id === 'weekly-match') {
      return {
        ...item,
        value: weeklyMatchStatus.value,
        tone: weeklyMatchStatus.tone,
        description: weeklyMatchStatus.description,
        actionLabel: weeklyMatchStatus.actionLabel,
      }
    }

    if (item.id === 'soul-portrait') {
      return item
    }

    return {
      ...item,
      value: cycleStatus?.revealAt ? formatDateTime(cycleStatus.revealAt) : item.value,
      description: cycleStatus?.cycleTitle
        ? `当前周期：${cycleStatus.cycleTitle}`
        : item.description,
    }
  })

  const closeAnnouncement = () => {
    markDashboardAnnouncementSeen()
    setAnnouncementOpen(false)
  }

  const openWeeklyMatchAction = () => {
    if (!hasCompletedQuestionnaire) {
      setQuestionnairePromptContext('match')
      setQuestionnairePromptOpen(true)
      return
    }

    setParticipationAction(isParticipating ? 'leave' : 'join')
  }

  const openSoulPortrait = () => {
    if (!hasCompletedQuestionnaire) {
      setQuestionnairePromptContext('portrait')
      setQuestionnairePromptOpen(true)
      return
    }

    setPortraitOpen(true)
  }

  const closeParticipationModal = () => {
    if (isParticipationSubmitting) {
      return
    }

    setParticipationAction(null)
  }

  const confirmWeeklyMatchAction = async () => {
    if (!participationAction) {
      return
    }

    const nextParticipation = participationAction === 'join'

    setIsParticipationSubmitting(true)

    try {
      const result = await updateQuestionnaireParticipation(nextParticipation)

      setSurveyProgress((current) =>
        current
          ? {
              ...current,
              hasCompletedQuestionnaire: result.hasCompletedQuestionnaire,
              participatingInMatch: result.participatingInMatch,
              resultRevealAt: result.resultRevealAt,
            }
          : {
              answeredCount: 0,
              totalCount: 0,
              progressPercent: 0,
              hasCompletedQuestionnaire: result.hasCompletedQuestionnaire,
              participatingInMatch: result.participatingInMatch,
              resultRevealAt: result.resultRevealAt,
            },
      )
      setMatchesNotice('')
      setParticipationAction(null)
    } catch (error) {
      setToastNotice(getErrorMessage(error, '参与状态更新失败，请稍后重试'))
    } finally {
      setIsParticipationSubmitting(false)
    }
  }

  const getStatusCardClick = (itemId: string) => {
    if (itemId === 'survey') {
      return () => navigate('/survey')
    }

    if (itemId === 'weekly-match' && canChooseWeeklyMatch) {
      return openWeeklyMatchAction
    }

    if (itemId === 'soul-portrait') {
      return openSoulPortrait
    }

    return undefined
  }

  return (
    <div className="min-h-screen bg-[#f6efe8]">
      {toastNotice ? (
        <div className="pointer-events-none fixed inset-x-0 top-6 z-[70] flex justify-center px-4">
          <div className="rounded-full border border-[#e4d5c9] bg-[#fff8f2] px-5 py-2.5 text-sm text-[#7b5f4f] shadow-[0_10px_24px_rgba(71,49,35,0.16)]">
            {toastNotice}
          </div>
        </div>
      ) : null}

      <main className="px-6 pb-0 pt-10 lg:px-8 lg:pt-14">
        <div className="mx-auto w-full max-w-4xl space-y-16">
          <WelcomeSection user={user} />

          <CountdownCard
            description={MATCH_REVEAL_WINDOW_TEXT}
            label={dashboardContent.countdown.label}
            targetAt={heroTargetAt}
          />

          <section className="grid gap-5 lg:grid-cols-3">
            {statusItems.map((item) => (
              <StatusCard
                item={item}
                key={item.id}
                onAction={getStatusCardClick(item.id)}
                onCardClick={getStatusCardClick(item.id)}
              />
            ))}
          </section>

          {matchesNotice ? (
            <p className="text-sm text-[#a35b43]">{matchesNotice}</p>
          ) : null}
        </div>
      </main>

      <DashboardFooter />

      <PlatformAnnouncementModal
        onClose={closeAnnouncement}
        open={announcementOpen}
      />
      <MatchParticipationModal
        action={participationAction}
        isSubmitting={isParticipationSubmitting}
        onClose={closeParticipationModal}
        onConfirm={confirmWeeklyMatchAction}
        open={participationAction !== null}
      />
      <QuestionnaireRequiredModal
        eyebrow={questionnairePromptContext === 'portrait' ? '灵魂画像' : undefined}
        message={
          questionnairePromptContext === 'portrait'
            ? '完成问卷后，才可以生成灵魂画像。点击确认后，将直接前往问卷页面。'
            : undefined
        }
        onClose={() => setQuestionnairePromptOpen(false)}
        onConfirm={() => {
          setQuestionnairePromptOpen(false)
          navigate('/survey')
        }}
        open={questionnairePromptOpen}
        title={questionnairePromptContext === 'portrait' ? '请先完成问卷' : undefined}
      />
      <SoulPortraitModal
        onClose={() => setPortraitOpen(false)}
        open={portraitOpen}
        userName={user.name}
      />
    </div>
  )
}
