import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { SurveyFooterNav } from '@/components/survey/survey-footer-nav'
import { SurveyHeader } from '@/components/survey/survey-header'
import { SurveyQuestionCard } from '@/components/survey/survey-question-card'
import { SurveyRenderer } from '@/components/survey/survey-renderer'
import { SurveySectionTabs } from '@/components/survey/survey-section-tabs'
import { getErrorMessage } from '@/lib/api'
import {
  toSurveyDraft,
  toSurveySections,
} from '@/lib/questionnaire-adapter'
import {
  fetchCurrentQuestionnaire,
  saveQuestionnaireAnswers,
} from '@/lib/questionnaire-api'
import {
  isQuestionnaireEditable,
  resolveQuestionnaireDeadline,
  type QuestionnaireCollectionWindow,
} from '@/lib/schedule'
import {
  getSurveySummary,
} from '@/lib/survey-storage'
import type {
  SurveyAnswerValue,
  SurveyDraft,
  SurveyQuestion,
  SurveySection,
  SurveySectionId,
} from '@/types/survey'

type NoticeTone = 'default' | 'error'
type SaveState = 'idle' | 'saving' | 'saved' | 'error'
const CLOSED_NOTICE = '当前问卷暂时不可编辑，你仍然可以查看已保存的答案'

function isRangeAnswerValue(value: SurveyAnswerValue): value is Extract<
  SurveyAnswerValue,
  { min?: number; max?: number; unlimited?: boolean }
> {
  return (
    typeof value === 'object' &&
    value !== null &&
    !Array.isArray(value) &&
    ('unlimited' in value || ('min' in value && 'max' in value))
  )
}

function formatDraftLabel(_draft: SurveyDraft, saveState: SaveState) {
  return saveState === 'saved' ? '已保存' : '尚未保存'
}

function buildSavePayload(sections: SurveySection[], draft: SurveyDraft) {
  const questions = sections.flatMap((section) => section.questions)

  return questions
    .map((question) => {
      const value = draft.answers[question.id]

      if (value === null || value === undefined) {
        return null
      }

      if (typeof value === 'string' && value.trim() === '') {
        return null
      }

      if (Array.isArray(value) && value.length === 0) {
        return null
      }

      if (question.kind === 'range' && isRangeAnswerValue(value)) {
        if (value.unlimited && !question.allowUnlimited) {
          return null
        }
        if (!value.unlimited && (value.min === undefined || value.max === undefined)) {
          return null
        }
      }

      return {
        questionId: question.id,
        value,
        importance: draft.importantQuestionIds.includes(question.id) || undefined,
      }
    })
    .filter((item): item is NonNullable<typeof item> => item !== null)
}

function getInvalidRequiredQuestionNumbers(
  sections: SurveySection[],
  draft: SurveyDraft,
) {
  return sections
    .flatMap((section) => section.questions)
    .filter(
      (question) => {
        if (question.kind !== 'multi') {
          return false
        }

        const answer = draft.answers[question.id]

        if (!Array.isArray(answer) || answer.length === 0) {
          return false
        }

        return (
          answer.length < question.minSelections ||
          answer.length > question.maxSelections
        )
      },
    )
    .map((question) => question.number)
    .sort((left, right) => left - right)
}

export function SurveyPage() {
  const navigate = useNavigate()
  const [sections, setSections] = useState<SurveySection[]>([])
  const [draft, setDraft] = useState<SurveyDraft | null>(null)
  const [canEdit, setCanEdit] = useState(true)
  const [collectionWindow, setCollectionWindow] = useState<QuestionnaireCollectionWindow | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [hasPendingChanges, setHasPendingChanges] = useState(false)
  const [saveState, setSaveState] = useState<SaveState>('idle')
  const [notice, setNotice] = useState('')
  const [toastNotice, setToastNotice] = useState('')
  const [noticeTone, setNoticeTone] = useState<NoticeTone>('default')
  const [invalidQuestionNumbers, setInvalidQuestionNumbers] = useState<number[]>([])
  const wasEditableRef = useRef(true)

  const summary = useMemo(() => {
    if (!draft) {
      return {
        answeredCount: 0,
        totalCount: 0,
        completionPercent: 0,
        isComplete: false,
        completedSectionIds: [],
        sections: [],
      }
    }

    return getSurveySummary(sections, draft)
  }, [draft, sections])

  const currentSectionIndex = useMemo(() => {
    if (!draft) {
      return 0
    }

    return Math.max(
      sections.findIndex((section) => section.id === draft.currentSectionId),
      0,
    )
  }, [draft, sections])

  const currentSection = sections[currentSectionIndex]

  useEffect(() => {
    let cancelled = false

    const loadQuestionnaire = async () => {
      setIsLoading(true)

      try {
        const current = await fetchCurrentQuestionnaire()
        const mappedSections = toSurveySections(current)
        const nextDraft = toSurveyDraft(current)

        if (cancelled) {
          return
        }

        const nextCollectionWindow = {
          acceptingResponses: current.acceptingResponses,
          collectionStatus: current.collectionStatus,
          collectionEndsAt: current.collectionEndsAt,
        } satisfies QuestionnaireCollectionWindow

        const nextCanEdit = isQuestionnaireEditable(nextCollectionWindow)

        setCollectionWindow(nextCollectionWindow)
        setCanEdit(nextCanEdit)
        setSections(mappedSections)
        setDraft(nextDraft)
        setHasPendingChanges(false)
        setSaveState('idle')
        setNotice('')
        setNoticeTone('default')
        setInvalidQuestionNumbers([])
      } catch (error) {
        if (cancelled) {
          return
        }

        setNotice(getErrorMessage(error, '问卷加载失败，请稍后重试'))
        setNoticeTone('error')
      } finally {
        if (!cancelled) {
          setIsLoading(false)
        }
      }
    }

    void loadQuestionnaire()

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!collectionWindow) {
      return
    }

    const syncEditableState = () => {
      setCanEdit(isQuestionnaireEditable(collectionWindow))
    }

    syncEditableState()

    const deadline = resolveQuestionnaireDeadline(collectionWindow)

    if (!deadline) {
      return
    }

    const timeoutMs = deadline.getTime() - Date.now()

    if (timeoutMs <= 0) {
      return
    }

    const timerId = window.setTimeout(syncEditableState, timeoutMs + 200)

    return () => {
      window.clearTimeout(timerId)
    }
  }, [collectionWindow])

  useEffect(() => {
    if (noticeTone === 'error') {
      return
    }

    setNotice('')
    setNoticeTone('default')

    if (!canEdit && wasEditableRef.current) {
      setToastNotice(CLOSED_NOTICE)
    }
    wasEditableRef.current = canEdit
  }, [canEdit, noticeTone])

  useEffect(() => {
    if (!toastNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setToastNotice('')
    }, 3000)

    return () => window.clearTimeout(timer)
  }, [toastNotice])

  const updateDraft = (updater: (current: SurveyDraft) => SurveyDraft) => {
    setDraft((current) => {
      if (!current) {
        return current
      }

      const next = updater(current)

      return {
        ...next,
        updatedAt: new Date().toISOString(),
      }
    })

    setHasPendingChanges(true)
    setSaveState('idle')
    setInvalidQuestionNumbers([])

    if (noticeTone === 'error') {
      setNotice('')
      setNoticeTone('default')
    }
  }

  const handleAnswerChange = (questionId: string, value: SurveyAnswerValue) => {
    if (!canEdit) {
      return
    }

    updateDraft((current) => ({
      ...current,
      answers: {
        ...current.answers,
        [questionId]: value,
      },
    }))
  }

  const handleImportantToggle = (questionId: string) => {
    if (!canEdit) {
      return
    }

    updateDraft((current) => {
      const exists = current.importantQuestionIds.includes(questionId)

      return {
        ...current,
        importantQuestionIds: exists
          ? current.importantQuestionIds.filter((item) => item !== questionId)
          : [...current.importantQuestionIds, questionId],
      }
    })
  }

  const handleSelectSection = (sectionId: SurveySectionId) => {
    setDraft((current) =>
      current
        ? {
            ...current,
            currentSectionId: sectionId,
          }
        : current,
    )
  }

  const handleSectionShift = (direction: -1 | 1) => {
    const nextIndex = currentSectionIndex + direction

    if (nextIndex < 0 || nextIndex >= sections.length) {
      return
    }

    handleSelectSection(sections[nextIndex].id)
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const persistDraft = async (action: 'save' | 'save_exit') => {
    if (!draft || !canEdit) {
      return false
    }

    const nextInvalidQuestionNumbers = getInvalidRequiredQuestionNumbers(sections, draft)

    if (nextInvalidQuestionNumbers.length > 0) {
      setInvalidQuestionNumbers(nextInvalidQuestionNumbers)
      setNotice('\u95ee\u5377\u4fdd\u5b58\u5931\u8d25\uff0c\u8bf7\u5148\u5b8c\u6210\u4e0d\u7b26\u5408\u8981\u6c42\u7684\u9898\u76ee')
      setNoticeTone('error')
      setSaveState('error')
      return false
    }

    if (!hasPendingChanges) {
      setInvalidQuestionNumbers([])
      setNotice(action === 'save' ? '\u5f53\u524d\u6ca1\u6709\u9700\u8981\u4fdd\u5b58\u7684\u4fee\u6539' : '')
      setNoticeTone('default')
      return true
    }

    const answers = buildSavePayload(sections, draft)
    setSaveState('saving')

    try {
      const result = await saveQuestionnaireAnswers({
        answers,
        action,
      })

      setDraft((current) =>
        current
          ? {
              ...current,
              updatedAt: new Date().toISOString(),
              status: result.hasCompletedQuestionnaire ? 'completed' : 'draft',
            }
          : current,
      )
      setHasPendingChanges(false)
      setSaveState('saved')
      setInvalidQuestionNumbers([])
      setNotice(action === 'save' ? '\u95ee\u5377\u5df2\u4fdd\u5b58' : '')
      setNoticeTone('default')
      return true
    } catch (error) {
      setInvalidQuestionNumbers(getInvalidRequiredQuestionNumbers(sections, draft))
      setNotice(getErrorMessage(error, '\u95ee\u5377\u4fdd\u5b58\u5931\u8d25\uff0c\u8bf7\u7a0d\u540e\u91cd\u8bd5'))
      setNoticeTone('error')
      setSaveState('error')
      return false
    }
  }
  const handleSave = async () => {
    await persistDraft('save')
  }

  const handleSaveAndExit = async () => {
    if (!draft || !canEdit) {
      navigate('/dashboard')
      return
    }

    if (!hasPendingChanges) {
      navigate('/dashboard')
      return
    }

    const saved = await persistDraft('save_exit')
    if (saved) {
      navigate('/dashboard')
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#f7f2ec] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-10 text-center text-lg text-[#7e6d61] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          正在加载问卷...
        </div>
      </div>
    )
  }

  if (!draft || sections.length === 0 || !currentSection) {
    return (
      <div className="min-h-screen bg-[#f7f2ec] px-6 py-20 lg:px-8">
        <div className="mx-auto max-w-3xl rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] p-10 text-center shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
          <p className="text-lg text-[#7e6d61]">{notice || '当前没有可展示的问卷内容'}</p>
        </div>
      </div>
    )
  }

  const noticeClass =
    noticeTone === 'error' ? 'text-base text-[#a35b43]' : 'text-base text-[#7e6d61]'
  const firstInvalidQuestionNumber =
    saveState === 'error' && invalidQuestionNumbers.length > 0
      ? invalidQuestionNumbers[0]
      : null

  return (
    <div className="min-h-screen bg-[#f7f2ec]">
      {toastNotice ? (
        <div className="pointer-events-none fixed inset-x-0 top-6 z-[70] flex justify-center px-4">
          <div className="rounded-full border border-[#e4d5c9] bg-[#fff8f2] px-5 py-2.5 text-sm text-[#7b5f4f] shadow-[0_10px_24px_rgba(71,49,35,0.16)]">
            {toastNotice}
          </div>
        </div>
      ) : null}

      <SurveyHeader
        answeredCount={summary.answeredCount}
        backTo="/dashboard"
        completionPercent={summary.completionPercent}
        draftLabel={formatDraftLabel(draft, saveState)}
        invalidQuestionNumber={firstInvalidQuestionNumber}
        onSave={handleSave}
        onSaveAndExit={handleSaveAndExit}
        saveDisabled={!canEdit || saveState === 'saving'}
        totalCount={summary.totalCount}
      />

      <SurveySectionTabs
        completedSectionIds={summary.completedSectionIds}
        currentSectionId={currentSection.id}
        onSelect={handleSelectSection}
        sections={sections}
      />

      <main className="px-4 pb-10 pt-5 sm:px-6 sm:pb-14 sm:pt-8 lg:px-8">
        <div className="mx-auto w-full max-w-[54rem] space-y-5">
          {notice ? <p className={noticeClass}>{notice}</p> : null}

          <section className="mx-auto max-w-[48rem] space-y-4">
            <div className="space-y-3">
              <h1 className="font-serif text-[2.35rem] leading-none text-[#2f221b] sm:text-5xl">
                {currentSection.title}
              </h1>
              {currentSection.subtitle ? (
                <p
                  className="whitespace-pre-line text-[1.1rem] leading-8 text-[#6f5f55] sm:text-[1.65rem] sm:leading-[2.45rem]"
                  style={{ fontFamily: '"STKaiti", "KaiTi", "Kaiti SC", serif' }}
                >
                  {currentSection.subtitle}
                </p>
              ) : null}
            </div>
          </section>

          <section className="space-y-3">
            {currentSection.questions.map((question: SurveyQuestion) => (
              <SurveyQuestionCard
                disabled={!canEdit}
                important={draft.importantQuestionIds.includes(question.id)}
                key={question.id}
                onImportantToggle={() => handleImportantToggle(question.id)}
                question={question}
              >
                <SurveyRenderer
                  answer={draft.answers[question.id] ?? null}
                  disabled={!canEdit}
                  onAnswerChange={(value) => handleAnswerChange(question.id, value)}
                  question={question}
                />
              </SurveyQuestionCard>
            ))}
          </section>

          <SurveyFooterNav
            currentIndex={currentSectionIndex}
            onNext={() => handleSectionShift(1)}
            onPrevious={() => handleSectionShift(-1)}
            onSaveAndExit={handleSaveAndExit}
            total={sections.length}
          />
        </div>
      </main>
    </div>
  )
}
