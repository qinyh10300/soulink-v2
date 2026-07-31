import {
  AlertCircle,
  ArrowLeft,
  Check,
  LoaderCircle,
  RotateCcw,
  Send,
  ShieldCheck,
  Sparkles,
  X,
} from 'lucide-react'
import { AnimatePresence, motion } from 'framer-motion'
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from 'react'
import { useLocation, useNavigate } from 'react-router-dom'

import {
  IntentCategoryIcon,
  intentCategoryMeta,
} from '@/components/connection/intent-category'
import { IntentReviewPanel } from '@/components/connection/intent-review-panel'
import { getErrorMessage } from '@/lib/api'
import {
  createIntent,
  fetchMatchingProfile,
  interpretIntent,
  type ClarifyingQuestion,
  type IntentInput,
  type IntentInterpretation,
  type MatchingProfile,
  type ProjectMemo,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'

const starterPrompts = [
  '今晚想在紫荆找个饭搭子',
  '找一个能一起做产品的项目伙伴',
  '想找人每周固定跑两次步',
  '为我的 AI 项目找 CTO',
]

const blankDraft: IntentInput = {
  category: 'other',
  title: '',
  description: '',
  goal: '',
  timeWindow: '',
  location: '',
  visibility: 'candidates',
  status: 'active',
  constraints: [],
  dealBreakers: [],
  desiredNextStep: '',
  projectMemo: null,
  expiresAt: null,
}

function emptyProjectMemo(): ProjectMemo {
  return {
    building: '',
    stage: '',
    weeklyTime: '',
    compensation: '',
    requiredSkills: [],
    niceToHaveSkills: [],
    workingStyle: '',
    riskAppetite: '',
  }
}

function toDraft(interpretation: IntentInterpretation): IntentInput {
  const isProject =
    interpretation.category === 'project' ||
    interpretation.category === 'cofounder'

  return {
    category: interpretation.category,
    title: interpretation.title,
    description: interpretation.description,
    goal: interpretation.goal,
    timeWindow: interpretation.timeWindow,
    location: interpretation.location,
    visibility: interpretation.visibility,
    status: 'active',
    constraints: interpretation.constraints,
    dealBreakers: [],
    desiredNextStep: interpretation.desiredNextStep,
    projectMemo: isProject
      ? interpretation.projectMemo ?? {
          ...emptyProjectMemo(),
          building: interpretation.description,
        }
      : null,
    expiresAt: null,
  }
}

function splitTags(value: string) {
  return value
    .split(/[,，、\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function AssistantMark({ small = false }: { small?: boolean }) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-md bg-[#263b35] text-white',
        small ? 'h-8 w-8' : 'h-10 w-10',
      )}
    >
      <Sparkles size={small ? 15 : 18} strokeWidth={1.8} />
    </span>
  )
}

function AssistantMessage({
  children,
  className,
}: {
  children: React.ReactNode
  className?: string
}) {
  return (
    <motion.div
      animate={{ opacity: 1, y: 0 }}
      className={cn('flex items-start gap-3', className)}
      initial={{ opacity: 0, y: 8 }}
      transition={{ duration: 0.2 }}
    >
      <AssistantMark small />
      <div className="min-w-0 flex-1 pt-1 text-sm leading-6 text-[#45413d]">
        {children}
      </div>
    </motion.div>
  )
}

function UserMessage({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      animate={{ opacity: 1, y: 0 }}
      className="flex justify-end"
      initial={{ opacity: 0, y: 8 }}
      transition={{ duration: 0.2 }}
    >
      <div className="max-w-[88%] rounded-lg bg-[#e9e5df] px-4 py-2.5 text-sm leading-6 text-[#34312e] sm:max-w-[76%]">
        {children}
      </div>
    </motion.div>
  )
}

function categoryFromAnswer(value: string) {
  if (/饭|活动/.test(value)) {
    return 'meal' as const
  }
  if (/学习|项目|协作/.test(value)) {
    return 'project' as const
  }
  if (/关系|对象|恋爱/.test(value)) {
    return 'romantic' as const
  }
  return null
}

export function IntentComposerPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const conversationEndRef = useRef<HTMLDivElement>(null)
  const didAutoInterpretRef = useRef(false)
  const initialText = useMemo(() => {
    const state = location.state as { text?: string } | null
    return state?.text?.trim() ?? ''
  }, [location.state])

  const [composerText, setComposerText] = useState('')
  const [requestText, setRequestText] = useState(initialText)
  const [interpretation, setInterpretation] =
    useState<IntentInterpretation | null>(null)
  const [draft, setDraft] = useState<IntentInput>(blankDraft)
  const [profile, setProfile] = useState<MatchingProfile | null>(null)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [interpreting, setInterpreting] = useState(false)
  const [submitting, setSubmitting] = useState<'draft' | 'active' | null>(null)
  const [notice, setNotice] = useState('')
  const [mobileReviewOpen, setMobileReviewOpen] = useState(false)

  const currentQuestion = useMemo(
    () =>
      interpretation?.questions.find((question) => !answers[question.id]) ??
      null,
    [answers, interpretation],
  )
  const answeredCount = interpretation
    ? interpretation.questions.filter((question) => answers[question.id]).length
    : 0
  const readyToReview = Boolean(interpretation && !currentQuestion)

  const runInterpretation = async (value: string) => {
    const text = value.trim()
    if (!text || interpreting) {
      return
    }

    const nextRequest =
      interpretation && requestText
        ? `${requestText}\n补充：${text}`
        : text
    setRequestText(nextRequest)
    setComposerText('')
    setInterpreting(true)
    setNotice('')

    try {
      const result = await interpretIntent(nextRequest)
      setInterpretation(result)
      setDraft(toDraft(result))
      setAnswers({})
    } catch (error) {
      setNotice(getErrorMessage(error, 'SouLink 暂时没有理解这句话，请换一种说法'))
    } finally {
      setInterpreting(false)
    }
  }

  useEffect(() => {
    void fetchMatchingProfile()
      .then(setProfile)
      .catch(() => setProfile(null))
  }, [])

  useEffect(() => {
    if (!initialText || didAutoInterpretRef.current) {
      return
    }
    didAutoInterpretRef.current = true
    void runInterpretation(initialText)
    // Route state is consumed only once when the conversation opens.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [initialText])

  useEffect(() => {
    conversationEndRef.current?.scrollIntoView({
      behavior: 'smooth',
      block: 'nearest',
    })
  }, [answers, interpretation, interpreting, notice])

  const updateDraft = <K extends keyof IntentInput>(
    key: K,
    value: IntentInput[K],
  ) => {
    setDraft((current) => ({ ...current, [key]: value }))
  }

  const updateProjectMemo = <K extends keyof ProjectMemo>(
    key: K,
    value: ProjectMemo[K],
  ) => {
    setDraft((current) => ({
      ...current,
      projectMemo: {
        ...(current.projectMemo ?? emptyProjectMemo()),
        [key]: value,
      },
    }))
  }

  const answerQuestion = (question: ClarifyingQuestion, value: string) => {
    const answer = value.trim()
    if (!answer) {
      return
    }

    setAnswers((current) => ({ ...current, [question.id]: answer }))
    setComposerText('')
    setDraft((current) => {
      const next = { ...current }

      if (question.id === 'timeWindow') {
        next.timeWindow = answer === '时间都可以' ? '时间灵活' : answer
      }
      if (question.id === 'location') {
        next.location =
          answer === '校内都可以' ? '校内均可' : answer.replace('附近', '')
      }
      if (question.id === 'skills') {
        const skills = splitTags(answer)
        next.constraints = Array.from(
          new Set([...current.constraints, ...skills]),
        )
        next.projectMemo = {
          ...(current.projectMemo ?? emptyProjectMemo()),
          requiredSkills: Array.from(
            new Set([
              ...(current.projectMemo?.requiredSkills ?? []),
              ...skills,
            ]),
          ),
        }
      }
      if (question.id === 'projectStage') {
        next.projectMemo = {
          ...(current.projectMemo ?? emptyProjectMemo()),
          stage: answer,
        }
      }
      if (question.id === 'category') {
        const category = categoryFromAnswer(answer)
        if (category) {
          next.category = category
          next.visibility =
            category === 'romantic' ? 'candidates' : next.visibility
          next.projectMemo =
            category === 'project'
              ? current.projectMemo ?? emptyProjectMemo()
              : null
        }
      }

      return next
    })
  }

  const handleComposerSubmit = () => {
    if (!composerText.trim() || interpreting) {
      return
    }
    if (currentQuestion) {
      answerQuestion(currentQuestion, composerText)
      return
    }
    void runInterpretation(composerText)
  }

  const handleComposerKeyDown = (
    event: KeyboardEvent<HTMLTextAreaElement>,
  ) => {
    if (
      event.key === 'Enter' &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing
    ) {
      event.preventDefault()
      handleComposerSubmit()
    }
  }

  const handleSubmit = async (status: 'draft' | 'active') => {
    if (submitting) {
      return
    }
    setSubmitting(status)
    setNotice('')
    try {
      const created = await createIntent({
        ...draft,
        status,
        visibility:
          draft.category === 'romantic' ? 'candidates' : draft.visibility,
      })
      navigate(`/intents/${created.id}`, { replace: true })
    } catch (error) {
      setNotice(getErrorMessage(error, '意图暂时无法保存'))
    } finally {
      setSubmitting(null)
    }
  }

  const resetConversation = () => {
    setComposerText('')
    setRequestText('')
    setInterpretation(null)
    setDraft(blankDraft)
    setAnswers({})
    setNotice('')
    setMobileReviewOpen(false)
    didAutoInterpretRef.current = true
  }

  const composerPlaceholder = currentQuestion
    ? '也可以直接输入你的答案'
    : interpretation
      ? '再补充一句，SouLink 会重新整理'
      : '描述你想找的人，以及想一起做什么'

  return (
    <div className="h-[calc(100dvh-8.75rem)] overflow-hidden bg-[#f7f6f3] lg:h-screen xl:grid xl:grid-cols-[minmax(0,1fr)_360px]">
      <section className="flex min-h-0 min-w-0 flex-col">
        <header className="flex h-16 shrink-0 items-center justify-between border-b border-[#dfdcd6] px-4 sm:px-6">
          <div className="flex min-w-0 items-center gap-3">
            <button
              aria-label="返回"
              className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#6f6b65] hover:bg-[#eeece8]"
              onClick={() => navigate(-1)}
              title="返回"
              type="button"
            >
              <ArrowLeft size={18} />
            </button>
            <div className="min-w-0">
              <div className="flex items-center gap-2">
                <h1 className="truncate text-sm font-semibold text-[#292724]">
                  SouLink 助手
                </h1>
                <span className="h-1.5 w-1.5 rounded-full bg-[#4f846f]" />
              </div>
              <p className="mt-0.5 truncate text-[11px] text-[#817d76]">
                清华校内连接
              </p>
            </div>
          </div>

          {requestText ? (
            <button
              aria-label="重新开始"
              className="inline-flex h-9 items-center gap-2 rounded-md px-3 text-xs font-semibold text-[#66625c] hover:bg-[#eeece8]"
              onClick={resetConversation}
              title="重新开始"
              type="button"
            >
              <RotateCcw size={15} />
              <span className="hidden sm:inline">重新开始</span>
            </button>
          ) : null}
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {!requestText && !interpreting ? (
            <div className="mx-auto flex min-h-full w-full max-w-2xl flex-col items-center justify-center px-5 py-10 text-center">
              <AssistantMark />
              <h2 className="mt-5 text-2xl font-semibold text-[#262421] sm:text-[30px]">
                现在，想认识谁？
              </h2>
              <p className="mt-2 max-w-md text-sm leading-6 text-[#77736c]">
                从一件想一起做的事开始。
              </p>
              <div className="mt-8 grid w-full gap-2 sm:grid-cols-2">
                {starterPrompts.map((prompt) => (
                  <button
                    className="min-h-12 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 py-2.5 text-left text-xs leading-5 text-[#5f5b55] transition-colors hover:border-[#9eaea8] hover:bg-white hover:text-[#2f5146]"
                    key={prompt}
                    onClick={() => void runInterpretation(prompt)}
                    type="button"
                  >
                    {prompt}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div className="mx-auto w-full max-w-3xl space-y-7 px-4 py-7 sm:px-7 sm:py-9">
              <AssistantMessage>
                告诉我你想和谁建立连接，以及你们准备一起做什么。
              </AssistantMessage>

              {requestText ? <UserMessage>{requestText}</UserMessage> : null}

              <AnimatePresence mode="popLayout">
                {interpreting ? (
                  <AssistantMessage key="thinking">
                    <span className="inline-flex items-center gap-2 text-[#66625c]">
                      <LoaderCircle className="animate-spin" size={15} />
                      正在理解你的目标和边界...
                    </span>
                  </AssistantMessage>
                ) : null}
              </AnimatePresence>

              {interpretation && !interpreting ? (
                <>
                  <AssistantMessage>
                    <div className="flex items-start gap-3">
                      <IntentCategoryIcon
                        category={interpretation.category}
                        size={16}
                      />
                      <div className="min-w-0">
                        <p>
                          我先把它理解为
                          <span className="font-semibold text-[#2f4f45]">
                            「{interpretation.title}」
                          </span>
                          。
                        </p>
                        <p className="mt-1 text-[#706c65]">
                          {interpretation.goal}
                        </p>
                      </div>
                    </div>

                    {interpretation.inferences.length > 0 ? (
                      <div className="mt-4 space-y-1.5 border-l-2 border-[#cddbd5] pl-3 text-xs leading-5 text-[#706c65]">
                        {interpretation.inferences.slice(0, 3).map((inference) => (
                          <p className="flex items-start gap-1.5" key={inference}>
                            <Check
                              className="mt-0.5 shrink-0 text-[#477062]"
                              size={13}
                            />
                            {inference}
                          </p>
                        ))}
                      </div>
                    ) : null}
                  </AssistantMessage>

                  {interpretation.questions.map((question) => {
                    const answer = answers[question.id]
                    const isCurrent = currentQuestion?.id === question.id
                    if (!answer && !isCurrent) {
                      return null
                    }

                    return (
                      <div className="space-y-4" key={question.id}>
                        <AssistantMessage>
                          <p>{question.prompt}</p>
                          {isCurrent ? (
                            <div className="mt-3 flex flex-wrap gap-2">
                              {question.options.map((option) => (
                                <button
                                  className="min-h-9 rounded-md border border-[#ccd5d1] bg-[#f9fbfa] px-3 py-1.5 text-xs font-medium text-[#365b4f] transition-colors hover:border-[#6f9185] hover:bg-[#edf4f1]"
                                  key={option}
                                  onClick={() => answerQuestion(question, option)}
                                  type="button"
                                >
                                  {option}
                                </button>
                              ))}
                            </div>
                          ) : null}
                        </AssistantMessage>
                        {answer ? <UserMessage>{answer}</UserMessage> : null}
                      </div>
                    )
                  })}

                  {readyToReview ? (
                    <AssistantMessage>
                      <p>
                        信息够了。我已经整理成一条
                        {intentCategoryMeta[draft.category].label}
                        意图，你可以确认后开始匹配。
                      </p>
                      <button
                        className="mt-3 inline-flex h-9 items-center gap-2 rounded-md border border-[#aebdb7] px-3 text-xs font-semibold text-[#315c4d] hover:bg-[#edf4f1] xl:hidden"
                        onClick={() => setMobileReviewOpen(true)}
                        type="button"
                      >
                        查看意图草稿
                      </button>
                    </AssistantMessage>
                  ) : null}

                  <div
                    className={cn(
                      'flex items-start gap-2 border-t pt-4 text-xs leading-5',
                      interpretation.safety.level === 'review'
                        ? 'border-[#ead9b2] text-[#7c642f]'
                        : 'border-[#d7e3dd] text-[#65766f]',
                    )}
                  >
                    <ShieldCheck className="mt-0.5 shrink-0" size={14} />
                    {interpretation.safety.notice}
                  </div>
                </>
              ) : null}

              {notice ? (
                <div className="flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
                  <AlertCircle className="mt-0.5 shrink-0" size={16} />
                  {notice}
                </div>
              ) : null}

              <div ref={conversationEndRef} />
            </div>
          )}
        </div>

        <footer className="shrink-0 border-t border-[#dfdcd6] bg-[rgba(247,246,243,0.96)] px-3 py-3 backdrop-blur-lg sm:px-6 sm:py-4">
          <div className="mx-auto max-w-3xl">
            {currentQuestion ? (
              <div className="mb-2 flex items-center justify-between gap-3 px-1 text-[11px] text-[#77736c]">
                <span>
                  还差 {interpretation!.questions.length - answeredCount} 个问题
                </span>
                <button
                  className="font-semibold text-[#477062] hover:text-[#315c4d]"
                  onClick={() => setMobileReviewOpen(true)}
                  type="button"
                >
                  先看草稿
                </button>
              </div>
            ) : null}
            <div className="flex items-end gap-2 rounded-lg border border-[#d6d2cb] bg-white p-2 shadow-[0_8px_24px_rgba(35,33,31,0.06)] focus-within:border-[#6f9185] focus-within:ring-3 focus-within:ring-[#55786c]/10">
              <textarea
                aria-label="给 SouLink 助手发消息"
                autoFocus={!initialText}
                className="max-h-28 min-h-10 min-w-0 flex-1 resize-none border-0 bg-transparent px-2 py-2 text-sm leading-6 text-[#292724] outline-none placeholder:text-[#9a958d]"
                maxLength={1200}
                onChange={(event) => setComposerText(event.target.value)}
                onKeyDown={handleComposerKeyDown}
                placeholder={composerPlaceholder}
                rows={1}
                value={composerText}
              />
              <button
                aria-label="发送"
                className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#263b35] text-white transition-colors hover:bg-[#1d302a] disabled:cursor-not-allowed disabled:bg-[#cbc8c2]"
                disabled={!composerText.trim() || interpreting}
                onClick={handleComposerSubmit}
                title="发送"
                type="button"
              >
                {interpreting ? (
                  <LoaderCircle className="animate-spin" size={17} />
                ) : (
                  <Send size={17} />
                )}
              </button>
            </div>
          </div>
        </footer>
      </section>

      {interpretation ? (
        <aside className="hidden min-h-0 border-l border-[#dfdcd6] xl:block">
          <IntentReviewPanel
            draft={draft}
            onChange={updateDraft}
            onProjectChange={updateProjectMemo}
            onSubmit={(status) => void handleSubmit(status)}
            profile={profile}
            submitting={submitting}
          />
        </aside>
      ) : (
        <aside className="hidden border-l border-[#dfdcd6] bg-[#fbfaf8] px-6 py-8 xl:block">
          <p className="text-[11px] font-semibold text-[#817d76]">意图草稿</p>
          <div className="mt-16 border-y border-[#e2dfd9] py-6 text-sm leading-6 text-[#77736c]">
            对话开始后，关键信息会出现在这里。
          </div>
        </aside>
      )}

      {mobileReviewOpen && interpretation ? (
        <div className="fixed inset-0 z-60 bg-black/30 xl:hidden">
          <button
            aria-label="关闭意图草稿"
            className="absolute inset-0 h-full w-full cursor-default"
            onClick={() => setMobileReviewOpen(false)}
            type="button"
          />
          <div className="absolute inset-x-0 bottom-0 h-[92dvh] overflow-hidden rounded-t-lg border-t border-[#d8d5ce] bg-[#fbfaf8] shadow-[0_-18px_50px_rgba(35,33,31,0.16)]">
            <button
              aria-label="关闭意图草稿"
              className="absolute right-3 top-3 z-10 inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#efede9] text-[#625e58] hover:bg-[#e6e3de]"
              onClick={() => setMobileReviewOpen(false)}
              title="关闭"
              type="button"
            >
              <X size={18} />
            </button>
            <IntentReviewPanel
              draft={draft}
              onChange={updateDraft}
              onProjectChange={updateProjectMemo}
              onSubmit={(status) => void handleSubmit(status)}
              profile={profile}
              submitting={submitting}
            />
          </div>
        </div>
      ) : null}
    </div>
  )
}
