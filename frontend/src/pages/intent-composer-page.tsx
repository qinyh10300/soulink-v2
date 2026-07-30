import {
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  Check,
  Eye,
  EyeOff,
  LoaderCircle,
  LockKeyhole,
  RotateCcw,
  Save,
  Send,
  ShieldCheck,
  Sparkles,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'

import {
  IntentCategoryIcon,
  intentCategoryMeta,
  intentCategoryOrder,
} from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  createIntent,
  fetchMatchingProfile,
  interpretIntent,
  type IntentInput,
  type IntentInterpretation,
  type IntentVisibility,
  type MatchingProfile,
  type ProjectMemo,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'

const visibilityOptions: Array<{
  value: IntentVisibility
  label: string
  description: string
  icon: typeof Eye
}> = [
  {
    value: 'private',
    label: '仅 AI 搜索',
    description: '不进入公开流',
    icon: EyeOff,
  },
  {
    value: 'candidates',
    label: '仅候选可见',
    description: '只向推荐的人展示',
    icon: LockKeyhole,
  },
  {
    value: 'feed',
    label: '进入意图流',
    description: '校内同学可发现',
    icon: Eye,
  },
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

export function IntentComposerPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const initialText = useMemo(() => {
    const state = location.state as { text?: string } | null
    return state?.text?.trim() ?? ''
  }, [location.state])

  const [rawText, setRawText] = useState(initialText)
  const [interpretation, setInterpretation] =
    useState<IntentInterpretation | null>(null)
  const [draft, setDraft] = useState<IntentInput>(blankDraft)
  const [profile, setProfile] = useState<MatchingProfile | null>(null)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [interpreting, setInterpreting] = useState(false)
  const [submitting, setSubmitting] = useState<'draft' | 'active' | null>(null)
  const [notice, setNotice] = useState('')
  const [didAutoInterpret, setDidAutoInterpret] = useState(false)

  async function runInterpretation(value = rawText) {
    const text = value.trim()
    if (!text || interpreting) {
      return
    }
    setInterpreting(true)
    setNotice('')
    try {
      const result = await interpretIntent(text)
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
    if (!initialText || didAutoInterpret) {
      return
    }
    setDidAutoInterpret(true)
    void runInterpretation(initialText)
    // The initial route state is intentionally consumed once.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [didAutoInterpret, initialText])

  const updateDraft = <K extends keyof IntentInput>(
    key: K,
    value: IntentInput[K],
  ) => {
    setDraft((current) => ({ ...current, [key]: value }))
  }

  const answerQuestion = (questionId: string, value: string) => {
    setAnswers((current) => ({ ...current, [questionId]: value }))
    if (questionId === 'timeWindow') {
      updateDraft('timeWindow', value === '时间都可以' ? '时间灵活' : value)
    }
    if (questionId === 'location') {
      updateDraft('location', value === '校内都可以' ? '校内均可' : value.replace('附近', ''))
    }
    if (questionId === 'skills') {
      updateDraft('constraints', Array.from(new Set([...draft.constraints, value])))
      setDraft((current) => ({
        ...current,
        projectMemo: {
          ...(current.projectMemo ?? emptyProjectMemo()),
          requiredSkills: Array.from(
            new Set([
              ...(current.projectMemo?.requiredSkills ?? []),
              value,
            ]),
          ),
        },
      }))
    }
    if (questionId === 'projectStage') {
      setDraft((current) => ({
        ...current,
        projectMemo: {
          ...(current.projectMemo ?? emptyProjectMemo()),
          stage: value,
        },
      }))
    }
    if (questionId === 'category') {
      const category =
        value === '一起吃饭或活动'
          ? 'meal'
          : value === '学习或项目协作'
            ? 'project'
            : 'romantic'
      updateDraft('category', category)
      if (category === 'romantic') {
        updateDraft('visibility', 'candidates')
      }
      updateDraft(
        'projectMemo',
        category === 'project'
          ? draft.projectMemo ?? emptyProjectMemo()
          : null,
      )
    }
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

  const romanticNeedsOptIn =
    draft.category === 'romantic' && profile && !profile.romanticOptIn
  const canSubmit =
    Boolean(draft.title.trim() && draft.description.trim()) &&
    !romanticNeedsOptIn

  return (
    <div className="mx-auto w-full max-w-[1180px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex items-center justify-between gap-4 border-b border-[#dfdcd6] pb-5">
        <div className="flex min-w-0 items-center gap-3">
          <button
            aria-label="返回"
            className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#6f6b65] hover:bg-[#eeece8]"
            onClick={() => navigate(-1)}
            title="返回"
            type="button"
          >
            <ArrowLeft size={19} />
          </button>
          <div className="min-w-0">
            <p className="text-xs font-medium text-[#6c4a80]">NEW INTENT</p>
            <h1 className="truncate text-xl font-semibold text-[#292724] sm:text-2xl">
              把一个念头变成可匹配的意图
            </h1>
          </div>
        </div>
        {interpretation ? (
          <button
            className="inline-flex h-9 items-center gap-2 rounded-md px-3 text-sm font-medium text-[#66625c] hover:bg-[#eeece8]"
            onClick={() => {
              setInterpretation(null)
              setDraft(blankDraft)
              setAnswers({})
            }}
            type="button"
          >
            <RotateCcw size={15} />
            <span className="hidden sm:inline">重新描述</span>
          </button>
        ) : null}
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      {!interpretation ? (
        <section className="mx-auto max-w-3xl py-10 sm:py-14">
          <div className="mb-5 flex items-center gap-2 text-sm font-medium text-[#5b3477]">
            <Sparkles size={17} />
            先说人话就好
          </div>
          <textarea
            aria-label="描述意图"
            autoFocus
            className="min-h-52 w-full resize-y rounded-lg border border-[#d8d4cd] bg-white p-5 text-lg leading-8 text-[#292724] outline-none transition-shadow placeholder:text-[#9a958d] focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10"
            maxLength={1200}
            onChange={(event) => setRawText(event.target.value)}
            placeholder="例如：想找一个能每周投入 10 小时、偏前端和工程实现的 CTO，一起做 AI 社交产品。"
            value={rawText}
          />
          <div className="mt-4 flex items-center justify-between gap-3">
            <span className="text-xs text-[#817d76]">
              {rawText.length} / 1200
            </span>
            <button
              className="inline-flex h-11 items-center gap-2 rounded-md bg-[#5b3477] px-5 text-sm font-semibold text-white hover:bg-[#4c2b65] disabled:cursor-not-allowed disabled:opacity-50"
              disabled={!rawText.trim() || interpreting}
              onClick={() => void runInterpretation()}
              type="button"
            >
              {interpreting ? (
                <LoaderCircle className="animate-spin" size={17} />
              ) : (
                <Sparkles size={17} />
              )}
              {interpreting ? '正在整理...' : '让 SouLink 整理'}
            </button>
          </div>
        </section>
      ) : (
        <div className="grid gap-7 py-7 lg:grid-cols-[minmax(0,0.75fr)_minmax(0,1.25fr)]">
          <aside className="min-w-0 space-y-5">
            <section className="rounded-lg border border-[#dad5ce] bg-[#f1ecf4] p-5">
              <div className="flex items-center justify-between gap-3">
                <span className="inline-flex items-center gap-2 text-sm font-semibold text-[#4d2d66]">
                  <Sparkles size={16} />
                  AI 整理结果
                </span>
                <span className="rounded-md bg-white/75 px-2 py-1 text-[11px] font-medium text-[#694a7d]">
                  {Math.round(interpretation.confidence * 100)}% 置信
                </span>
              </div>
              <p className="mt-4 text-sm leading-6 text-[#5f5267]">
                {interpretation.goal}
              </p>
              <div className="mt-4 border-t border-[#dcd2e1] pt-3">
                {interpretation.inferences.map((inference) => (
                  <p
                    className="flex items-start gap-2 py-1 text-xs leading-5 text-[#736679]"
                    key={inference}
                  >
                    <Check className="mt-0.5 shrink-0" size={13} />
                    {inference}
                  </p>
                ))}
              </div>
            </section>

            {interpretation.questions.length > 0 ? (
              <section>
                <h2 className="text-sm font-semibold text-[#36332f]">还差一点关键信息</h2>
                <div className="mt-3 space-y-4">
                  {interpretation.questions.map((question) => (
                    <div
                      className="border-b border-[#e0ddd7] pb-4 last:border-b-0"
                      key={question.id}
                    >
                      <p className="text-sm leading-6 text-[#514d48]">{question.prompt}</p>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {question.options.map((option) => (
                          <button
                            className={cn(
                              'h-8 rounded-md border px-3 text-xs font-medium transition-colors',
                              answers[question.id] === option
                                ? 'border-[#6a477e] bg-[#eee7f2] text-[#4c2e61]'
                                : 'border-[#d9d5cf] bg-[#fbfaf8] text-[#6d6963] hover:border-[#bbaac5]',
                            )}
                            key={option}
                            onClick={() => answerQuestion(question.id, option)}
                            type="button"
                          >
                            {option}
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </section>
            ) : null}

            <section
              className={cn(
                'flex items-start gap-3 rounded-md border px-4 py-3',
                interpretation.safety.level === 'review'
                  ? 'border-[#e8d6ae] bg-[#fff9e9] text-[#775c24]'
                  : 'border-[#cfe0d7] bg-[#f0f7f3] text-[#426456]',
              )}
            >
              <ShieldCheck className="mt-0.5 shrink-0" size={17} />
              <p className="text-xs leading-5">{interpretation.safety.notice}</p>
            </section>
          </aside>

          <section className="min-w-0 rounded-lg border border-[#ddd9d3] bg-white">
            <div className="border-b border-[#e4e1dc] px-5 py-4 sm:px-6">
              <div className="flex items-center gap-3">
                <IntentCategoryIcon category={draft.category} size={18} />
                <div>
                  <p className="text-xs text-[#7d7972]">发布前确认</p>
                  <h2 className="text-base font-semibold text-[#302e2b]">
                    这会成为别人看到的需求
                  </h2>
                </div>
              </div>
            </div>

            <div className="space-y-6 p-5 sm:p-6">
              <div>
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-title">
                  标题
                </label>
                <input
                  className="mt-2 h-11 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm text-[#302e2b] outline-none focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10"
                  id="intent-title"
                  maxLength={80}
                  onChange={(event) => updateDraft('title', event.target.value)}
                  value={draft.title}
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-description">
                  具体描述
                </label>
                <textarea
                  className="mt-2 min-h-28 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fcfbf9] p-3 text-sm leading-6 text-[#302e2b] outline-none focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10"
                  id="intent-description"
                  maxLength={1200}
                  onChange={(event) => updateDraft('description', event.target.value)}
                  value={draft.description}
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-goal">
                  希望这次连接达成什么
                </label>
                <input
                  className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                  id="intent-goal"
                  maxLength={200}
                  onChange={(event) => updateDraft('goal', event.target.value)}
                  placeholder="例如：先确认双方是否适合长期协作"
                  value={draft.goal}
                />
              </div>

              <div>
                <span className="text-xs font-semibold text-[#5f5b55]">意图类型</span>
                <div className="mt-2 grid grid-cols-3 gap-2 sm:grid-cols-5">
                  {intentCategoryOrder.map((category) => {
                    const meta = intentCategoryMeta[category]
                    const Icon = meta.icon
                    return (
                      <button
                        className={cn(
                          'flex min-h-16 flex-col items-center justify-center gap-1 rounded-md border px-2 py-2 text-center text-[11px] font-medium transition-colors',
                          draft.category === category
                            ? 'border-[#735087] bg-[#f1ebf4] text-[#4d2d66]'
                            : 'border-[#dedad4] bg-[#fcfbf9] text-[#6f6b65] hover:border-[#bbaac5]',
                        )}
                        key={category}
                        onClick={() => {
                          updateDraft('category', category)
                          if (category === 'romantic') {
                            updateDraft('visibility', 'candidates')
                          }
                          updateDraft(
                            'projectMemo',
                            category === 'project' || category === 'cofounder'
                              ? draft.projectMemo ?? emptyProjectMemo()
                              : null,
                          )
                        }}
                        type="button"
                      >
                        <Icon className={meta.iconClassName} size={16} />
                        {meta.shortLabel}
                      </button>
                    )
                  })}
                </div>
              </div>

              {(draft.category === 'project' ||
                draft.category === 'cofounder') &&
              draft.projectMemo ? (
                <section className="border-y border-[#ddd7e1] bg-[#faf8fb] px-1 py-5">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <h3 className="text-sm font-semibold text-[#4d2d66]">
                        项目 memo
                      </h3>
                      <p className="mt-1 text-[11px] text-[#81768a]">
                        这些信息会参与工程伙伴与 CTO 匹配。
                      </p>
                    </div>
                    <span className="rounded-md bg-[#eee8f2] px-2 py-1 text-[10px] font-semibold text-[#5b3477]">
                      PROJECT
                    </span>
                  </div>

                  <div className="mt-5 space-y-4">
                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      正在做什么
                      <textarea
                        className="mt-2 min-h-24 w-full resize-y rounded-md border border-[#d9d5cf] bg-white p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                        maxLength={600}
                        onChange={(event) =>
                          updateProjectMemo('building', event.target.value)
                        }
                        placeholder="产品解决什么问题，已经有了哪些证据。"
                        value={draft.projectMemo.building}
                      />
                    </label>

                    <div className="grid gap-4 sm:grid-cols-2">
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        当前阶段
                        <input
                          className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm outline-none focus:border-[#7a5a8e]"
                          maxLength={240}
                          onChange={(event) =>
                            updateProjectMemo('stage', event.target.value)
                          }
                          placeholder="想法 / 验证需求 / 已有 MVP"
                          value={draft.projectMemo.stage}
                        />
                      </label>
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        每周投入
                        <input
                          className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm outline-none focus:border-[#7a5a8e]"
                          maxLength={240}
                          onChange={(event) =>
                            updateProjectMemo('weeklyTime', event.target.value)
                          }
                          placeholder="例如：每周 8 到 12 小时"
                          value={draft.projectMemo.weeklyTime}
                        />
                      </label>
                    </div>

                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      权益或补偿预期
                      <input
                        className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm outline-none focus:border-[#7a5a8e]"
                        maxLength={240}
                        onChange={(event) =>
                          updateProjectMemo('compensation', event.target.value)
                        }
                        placeholder="可留空，也可以说明股权、薪资或尚待讨论"
                        value={draft.projectMemo.compensation}
                      />
                    </label>

                    <div className="grid gap-4 sm:grid-cols-2">
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        必备技能
                        <input
                          className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm outline-none focus:border-[#7a5a8e]"
                          onChange={(event) =>
                            updateProjectMemo(
                              'requiredSkills',
                              splitTags(event.target.value),
                            )
                          }
                          placeholder="React，工程架构，AI 应用"
                          value={draft.projectMemo.requiredSkills.join('，')}
                        />
                      </label>
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        加分技能
                        <input
                          className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm outline-none focus:border-[#7a5a8e]"
                          onChange={(event) =>
                            updateProjectMemo(
                              'niceToHaveSkills',
                              splitTags(event.target.value),
                            )
                          }
                          placeholder="Go，推荐系统，校园产品"
                          value={draft.projectMemo.niceToHaveSkills.join('，')}
                        />
                      </label>
                    </div>

                    <div className="grid gap-4 sm:grid-cols-2">
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        协作方式
                        <textarea
                          className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#d9d5cf] bg-white p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                          maxLength={240}
                          onChange={(event) =>
                            updateProjectMemo('workingStyle', event.target.value)
                          }
                          placeholder="如何决策、同步和交付。"
                          value={draft.projectMemo.workingStyle}
                        />
                      </label>
                      <label className="block text-xs font-semibold text-[#5f5b55]">
                        风险偏好
                        <textarea
                          className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#d9d5cf] bg-white p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                          maxLength={240}
                          onChange={(event) =>
                            updateProjectMemo('riskAppetite', event.target.value)
                          }
                          placeholder="对早期不确定性、现金流或长期投入的看法。"
                          value={draft.projectMemo.riskAppetite}
                        />
                      </label>
                    </div>
                  </div>
                </section>
              ) : null}

              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-time">
                    时间窗口
                  </label>
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    id="intent-time"
                    onChange={(event) => updateDraft('timeWindow', event.target.value)}
                    placeholder="例如：今晚 / 本周"
                    value={draft.timeWindow}
                  />
                </div>
                <div>
                  <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-location">
                    校内地点
                  </label>
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    id="intent-location"
                    onChange={(event) => updateDraft('location', event.target.value)}
                    placeholder="例如：紫荆 / 图书馆"
                    value={draft.location}
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-constraints">
                  重要偏好
                </label>
                <input
                  className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                  id="intent-constraints"
                  onChange={(event) =>
                    updateDraft('constraints', splitTags(event.target.value))
                  }
                  placeholder="用逗号分隔，例如：低压力，守时，每周可投入 10 小时"
                  value={draft.constraints.join('，')}
                />
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-deal-breakers">
                  明确不合适的情况
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    id="intent-deal-breakers"
                    onChange={(event) =>
                      updateDraft('dealBreakers', splitTags(event.target.value))
                    }
                    placeholder="用逗号分隔，可留空"
                    value={draft.dealBreakers.join('，')}
                  />
                </label>
                <label className="text-xs font-semibold text-[#5f5b55]" htmlFor="intent-next-step">
                  希望的下一步
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    id="intent-next-step"
                    maxLength={200}
                    onChange={(event) =>
                      updateDraft('desiredNextStep', event.target.value)
                    }
                    placeholder="例如：先聊 30 分钟"
                    value={draft.desiredNextStep}
                  />
                </label>
              </div>

              <div>
                <span className="text-xs font-semibold text-[#5f5b55]">谁可以看见</span>
                <div className="mt-2 grid gap-2 sm:grid-cols-3">
                  {visibilityOptions.map((option) => {
                    const Icon = option.icon
                    const disabled =
                      draft.category === 'romantic' && option.value !== 'candidates'
                    return (
                      <button
                        className={cn(
                          'min-h-16 rounded-md border px-3 py-2 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-45',
                          draft.visibility === option.value
                            ? 'border-[#735087] bg-[#f1ebf4]'
                            : 'border-[#dedad4] bg-[#fcfbf9] hover:border-[#bbaac5]',
                        )}
                        disabled={disabled}
                        key={option.value}
                        onClick={() => updateDraft('visibility', option.value)}
                        type="button"
                      >
                        <span className="flex items-center gap-2 text-xs font-semibold text-[#4e4a45]">
                          <Icon size={14} />
                          {option.label}
                        </span>
                        <span className="mt-1 block text-[11px] text-[#817d76]">
                          {option.description}
                        </span>
                      </button>
                    )
                  })}
                </div>
              </div>

              {romanticNeedsOptIn ? (
                <div className="rounded-md border border-[#e5cad3] bg-[#fff5f8] px-4 py-3">
                  <p className="text-sm font-medium text-[#7e3d55]">认真关系匹配尚未开启</p>
                  <p className="mt-1 text-xs leading-5 text-[#8e6473]">
                    这是一个单独的私密选择。开启后，关系偏好仍不会出现在公开意图流。
                  </p>
                  <Link
                    className="mt-2 inline-flex items-center gap-1 text-xs font-semibold text-[#7e3d55]"
                    to="/profile"
                  >
                    前往 AI 档案开启
                    <ArrowRight size={13} />
                  </Link>
                </div>
              ) : null}
            </div>

            <footer className="flex flex-col-reverse gap-2 border-t border-[#e4e1dc] bg-[#fbfaf8] px-5 py-4 sm:flex-row sm:items-center sm:justify-end sm:px-6">
              <button
                className="inline-flex h-10 items-center justify-center gap-2 rounded-md border border-[#d6d1ca] bg-white px-4 text-sm font-semibold text-[#605c56] hover:bg-[#f6f4f1] disabled:opacity-50"
                disabled={!canSubmit || submitting !== null}
                onClick={() => void handleSubmit('draft')}
                type="button"
              >
                {submitting === 'draft' ? (
                  <LoaderCircle className="animate-spin" size={16} />
                ) : (
                  <Save size={16} />
                )}
                保存草稿
              </button>
              <button
                className="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-[#5b3477] px-5 text-sm font-semibold text-white hover:bg-[#4c2b65] disabled:cursor-not-allowed disabled:opacity-50"
                disabled={!canSubmit || submitting !== null}
                onClick={() => void handleSubmit('active')}
                type="button"
              >
                {submitting === 'active' ? (
                  <LoaderCircle className="animate-spin" size={16} />
                ) : (
                  <Send size={16} />
                )}
                发布并开始匹配
              </button>
            </footer>
          </section>
        </div>
      )}
    </div>
  )
}
