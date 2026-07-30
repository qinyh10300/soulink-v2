import {
  AlertCircle,
  ArrowLeft,
  ArrowRight,
  Check,
  Clock3,
  Eye,
  HeartHandshake,
  LoaderCircle,
  MapPin,
  MessageSquarePlus,
  ShieldCheck,
  Sparkles,
  UserCheck,
  X,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { IntentCategoryIcon, intentCategoryMeta } from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  expressIntentInterest,
  fetchIntent,
  fetchIntentInterests,
  fetchIntentRecommendations,
  respondToIntentInterest,
  updateIntent,
  withdrawIntentInterest,
  type ConnectionIntent,
  type IntentInterest,
  type IntentRecommendation,
} from '@/lib/connection-api'

function formatDateTime(value?: string | null) {
  if (!value) {
    return '未设置'
  }
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function statusText(status: ConnectionIntent['status']) {
  const labels = {
    draft: '草稿',
    active: '匹配中',
    paused: '已暂停',
    fulfilled: '已完成',
    expired: '已结束',
  }
  return labels[status]
}

export function IntentDetailPage() {
  const navigate = useNavigate()
  const params = useParams()
  const intentId = Number(params.intentId)
  const [intent, setIntent] = useState<ConnectionIntent | null>(null)
  const [recommendations, setRecommendations] = useState<IntentRecommendation[]>([])
  const [interests, setInterests] = useState<IntentInterest[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [interestMessage, setInterestMessage] = useState('')
  const [submittingInterest, setSubmittingInterest] = useState(false)
  const [respondingUserId, setRespondingUserId] = useState<number | null>(null)
  const [publishing, setPublishing] = useState(false)

  const load = async () => {
    if (!Number.isInteger(intentId) || intentId <= 0) {
      setNotice('无效的意图编号')
      setLoading(false)
      return
    }
    setLoading(true)
    setNotice('')
    try {
      const result = await fetchIntent(intentId)
      setIntent(result)
      if (result.isOwner) {
        const [recommendationResult, interestResult] = await Promise.allSettled([
          fetchIntentRecommendations(intentId),
          fetchIntentInterests(intentId),
        ])
        setRecommendations(
          recommendationResult.status === 'fulfilled'
            ? recommendationResult.value
            : [],
        )
        setInterests(
          interestResult.status === 'fulfilled' ? interestResult.value : [],
        )
      }
    } catch (error) {
      setNotice(getErrorMessage(error, '这个意图暂时无法打开'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // The route id is the only load key.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [intentId])

  const handleInterest = async () => {
    if (!intent || submittingInterest) {
      return
    }
    setSubmittingInterest(true)
    setNotice('')
    try {
      setIntent(await expressIntentInterest(intent.id, interestMessage))
      setInterestMessage('')
    } catch (error) {
      setNotice(getErrorMessage(error, '回应没有发出去'))
    } finally {
      setSubmittingInterest(false)
    }
  }

  const handleWithdraw = async () => {
    if (!intent || submittingInterest) {
      return
    }
    setSubmittingInterest(true)
    try {
      setIntent(await withdrawIntentInterest(intent.id))
    } catch (error) {
      setNotice(getErrorMessage(error, '暂时无法撤回回应'))
    } finally {
      setSubmittingInterest(false)
    }
  }

  const handleResponse = async (interest: IntentInterest, accepted: boolean) => {
    if (!intent || respondingUserId) {
      return
    }
    setRespondingUserId(interest.userId)
    setNotice('')
    try {
      const result = await respondToIntentInterest(
        intent.id,
        interest.userId,
        accepted,
      )
      setInterests((current) =>
        current.map((item) =>
          item.userId === interest.userId
            ? { ...item, status: accepted ? 'accepted' : 'declined' }
            : item,
        ),
      )
      if (result.room) {
        navigate(`/matches/${result.room.id}`)
      }
    } catch (error) {
      setNotice(getErrorMessage(error, '这个回应暂时无法处理'))
    } finally {
      setRespondingUserId(null)
    }
  }

  const publishDraft = async () => {
    if (!intent || publishing) {
      return
    }
    setPublishing(true)
    setNotice('')
    try {
      setIntent(
        await updateIntent(intent.id, {
          category: intent.category,
          title: intent.title,
          description: intent.description,
          goal: intent.goal,
          timeWindow: intent.timeWindow,
          location: intent.location,
          visibility: intent.visibility,
          status: 'active',
          constraints: intent.constraints,
          dealBreakers: intent.dealBreakers,
          desiredNextStep: intent.desiredNextStep,
          expiresAt: intent.expiresAt,
        }),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '草稿暂时无法发布'))
    } finally {
      setPublishing(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center text-sm text-[#77736c]">
        <LoaderCircle className="mr-2 animate-spin" size={18} />
        正在打开意图...
      </div>
    )
  }

  if (!intent) {
    return (
      <div className="mx-auto max-w-2xl px-4 py-20 text-center">
        <AlertCircle className="mx-auto text-[#9a6155]" size={28} />
        <p className="mt-3 text-sm text-[#6f6b65]">{notice || '找不到这个意图'}</p>
        <Link
          className="mt-4 inline-flex h-9 items-center gap-2 rounded-md border border-[#d8d3cc] px-3 text-xs font-semibold text-[#5b3477]"
          to="/home"
        >
          返回首页
        </Link>
      </div>
    )
  }

  const meta = intentCategoryMeta[intent.category]

  return (
    <div className="mx-auto w-full max-w-[1100px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
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
            <p className="text-xs font-medium text-[#6c4a80]">
              {intent.isOwner ? 'MY INTENT' : 'CAMPUS INTENT'}
            </p>
            <h1 className="truncate text-xl font-semibold text-[#292724] sm:text-2xl">
              {intent.title}
            </h1>
          </div>
        </div>
        <span className="shrink-0 rounded-md bg-[#ede9e4] px-2.5 py-1.5 text-xs font-semibold text-[#69655f]">
          {statusText(intent.status)}
        </span>
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <div className="grid gap-7 py-7 lg:grid-cols-[minmax(0,1.25fr)_minmax(300px,0.75fr)]">
        <main className="min-w-0 space-y-7">
          <section className="rounded-lg border border-[#dfdcd6] bg-white p-5 sm:p-6">
            <div className="flex items-start gap-3">
              <IntentCategoryIcon category={intent.category} size={20} />
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2 text-xs text-[#77736c]">
                  <span className="font-semibold text-[#5b3477]">{meta.label}</span>
                  <span className="inline-flex items-center gap-1">
                    <ShieldCheck size={12} />
                    清华身份已验证
                  </span>
                </div>
                <p className="mt-3 whitespace-pre-wrap text-[15px] leading-7 text-[#494641]">
                  {intent.description}
                </p>
              </div>
            </div>

            <div className="mt-5 grid gap-3 border-t border-[#e8e5e0] pt-4 sm:grid-cols-3">
              <div>
                <p className="text-[11px] font-medium text-[#8a867f]">时间</p>
                <p className="mt-1 flex items-center gap-1.5 text-sm text-[#4f4b46]">
                  <Clock3 size={14} />
                  {intent.timeWindow || '待确认'}
                </p>
              </div>
              <div>
                <p className="text-[11px] font-medium text-[#8a867f]">地点</p>
                <p className="mt-1 flex items-center gap-1.5 text-sm text-[#4f4b46]">
                  <MapPin size={14} />
                  {intent.location || '校内待定'}
                </p>
              </div>
              <div>
                <p className="text-[11px] font-medium text-[#8a867f]">可见范围</p>
                <p className="mt-1 flex items-center gap-1.5 text-sm text-[#4f4b46]">
                  <Eye size={14} />
                  {intent.visibility === 'feed'
                    ? '意图流'
                    : intent.visibility === 'candidates'
                      ? '推荐候选'
                      : '仅 AI'}
                </p>
              </div>
            </div>

            {intent.constraints.length > 0 ? (
              <div className="mt-4 flex flex-wrap gap-2">
                {intent.constraints.map((constraint) => (
                  <span
                    className="rounded-md bg-[#f1efeb] px-2.5 py-1.5 text-xs text-[#65615b]"
                    key={constraint}
                  >
                    {constraint}
                  </span>
                ))}
              </div>
            ) : null}

            <div className="mt-5 rounded-md bg-[#f5f3ef] px-4 py-3">
              <p className="text-[11px] font-medium text-[#85817a]">期待的下一步</p>
              <p className="mt-1 text-sm text-[#56524d]">
                {intent.desiredNextStep || '双方确认后再决定'}
              </p>
            </div>
          </section>

          {intent.isOwner ? (
            <>
              <section>
                <div className="mb-4 flex items-end justify-between gap-3">
                  <div>
                    <h2 className="text-lg font-semibold text-[#302e2b]">三条推荐</h2>
                    <p className="mt-0.5 text-xs text-[#7b7770]">
                      少而具体，按当前意图实时排序
                    </p>
                  </div>
                  <span className="text-xs font-medium text-[#6c4a80]">
                    {recommendations.length} / 3
                  </span>
                </div>
                {recommendations.length > 0 ? (
                  <div className="space-y-3">
                    {recommendations.map((recommendation, index) => (
                      <article
                        className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-4 sm:p-5"
                        key={recommendation.intent.id}
                      >
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex min-w-0 items-start gap-3">
                            <span className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-[#eae3ee] text-xs font-bold text-[#5b3477]">
                              {index + 1}
                            </span>
                            <div className="min-w-0">
                              <Link
                                className="font-semibold text-[#302e2b] hover:text-[#5b3477]"
                                to={`/intents/${recommendation.intent.id}`}
                              >
                                {recommendation.intent.title}
                              </Link>
                              <p className="mt-2 text-sm leading-6 text-[#625e58]">
                                {recommendation.fitReason}
                              </p>
                            </div>
                          </div>
                          <span className="shrink-0 text-sm font-bold text-[#5b3477]">
                            {recommendation.score}%
                          </span>
                        </div>
                        <div className="mt-4 border-t border-[#e6e2dc] pt-3">
                          <p className="text-xs leading-5 text-[#81756f]">
                            <span className="font-semibold">仍需确认：</span>
                            {recommendation.uncertainty}
                          </p>
                          <div className="mt-3 flex flex-wrap items-center justify-between gap-3">
                            <div className="flex flex-wrap gap-1.5">
                              {recommendation.sharedContext.map((item) => (
                                <span
                                  className="rounded-md bg-[#edeae5] px-2 py-1 text-[11px] text-[#69655f]"
                                  key={item}
                                >
                                  {item}
                                </span>
                              ))}
                            </div>
                            <Link
                              className="inline-flex items-center gap-1 text-xs font-semibold text-[#5b3477]"
                              to={`/intents/${recommendation.intent.id}`}
                            >
                              查看
                              <ArrowRight size={13} />
                            </Link>
                          </div>
                        </div>
                      </article>
                    ))}
                  </div>
                ) : (
                  <div className="rounded-lg border border-dashed border-[#d7d2cb] px-5 py-8 text-center text-sm text-[#77736c]">
                    暂时没有达到质量门槛的推荐。
                  </div>
                )}
              </section>

              <section>
                <div className="mb-4">
                  <h2 className="text-lg font-semibold text-[#302e2b]">收到的回应</h2>
                  <p className="mt-0.5 text-xs text-[#7b7770]">
                    接受后才会创建双方可见的匹配房间
                  </p>
                </div>
                {interests.length > 0 ? (
                  <div className="space-y-3">
                    {interests.map((interest) => (
                      <article
                        className="rounded-lg border border-[#dfdcd6] bg-white p-4"
                        key={interest.userId}
                      >
                        <div className="flex items-start gap-3">
                          <span className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#e9f1ed] text-sm font-semibold text-[#2d6a53]">
                            {interest.nickname.slice(0, 1)}
                          </span>
                          <div className="min-w-0 flex-1">
                            <div className="flex flex-wrap items-center gap-2">
                              <h3 className="text-sm font-semibold text-[#37332f]">
                                {interest.nickname}
                              </h3>
                              <span className="text-xs text-[#817d76]">
                                {interest.department} · {interest.grade}
                              </span>
                            </div>
                            {interest.message ? (
                              <p className="mt-2 rounded-md bg-[#f5f3ef] px-3 py-2 text-sm leading-6 text-[#5f5b55]">
                                {interest.message}
                              </p>
                            ) : null}
                            {interest.skills.length > 0 ? (
                              <div className="mt-2 flex flex-wrap gap-1.5">
                                {interest.skills.slice(0, 5).map((skill) => (
                                  <span
                                    className="rounded-md bg-[#efedf1] px-2 py-1 text-[11px] text-[#67566f]"
                                    key={skill}
                                  >
                                    {skill}
                                  </span>
                                ))}
                              </div>
                            ) : null}
                            <div className="mt-3 flex flex-wrap gap-2">
                              {interest.status === 'interested' ? (
                                <>
                                  <button
                                    className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#376f5b] px-3 text-xs font-semibold text-white disabled:opacity-50"
                                    disabled={respondingUserId === interest.userId}
                                    onClick={() => void handleResponse(interest, true)}
                                    type="button"
                                  >
                                    {respondingUserId === interest.userId ? (
                                      <LoaderCircle className="animate-spin" size={14} />
                                    ) : (
                                      <UserCheck size={14} />
                                    )}
                                    愿意认识
                                  </button>
                                  <button
                                    className="inline-flex h-8 items-center gap-1.5 rounded-md px-3 text-xs font-medium text-[#746f69] hover:bg-[#efede9]"
                                    disabled={respondingUserId === interest.userId}
                                    onClick={() => void handleResponse(interest, false)}
                                    type="button"
                                  >
                                    <X size={14} />
                                    这次不合适
                                  </button>
                                </>
                              ) : (
                                <span className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#efede9] px-3 text-xs font-medium text-[#6d6963]">
                                  {interest.status === 'accepted' ? (
                                    <Check size={14} />
                                  ) : (
                                    <X size={14} />
                                  )}
                                  {interest.status === 'accepted' ? '已接受' : '已婉拒'}
                                </span>
                              )}
                            </div>
                          </div>
                        </div>
                      </article>
                    ))}
                  </div>
                ) : (
                  <div className="rounded-lg border border-dashed border-[#d7d2cb] px-5 py-8 text-center text-sm text-[#77736c]">
                    还没有收到回应。
                  </div>
                )}
              </section>
            </>
          ) : null}
        </main>

        <aside className="min-w-0 space-y-5">
          <section className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-5">
            <p className="text-xs font-medium text-[#817d76]">发布者</p>
            <div className="mt-3 flex items-center gap-3">
              <span className="inline-flex h-10 w-10 items-center justify-center rounded-md bg-[#e9f1ed] text-sm font-semibold text-[#2d6a53]">
                {intent.author.nickname.slice(0, 1)}
              </span>
              <span className="min-w-0">
                <span className="block truncate text-sm font-semibold text-[#37332f]">
                  {intent.author.nickname}
                </span>
                <span className="mt-0.5 block truncate text-xs text-[#77736c]">
                  {intent.author.department} · {intent.author.grade}
                </span>
              </span>
            </div>
          </section>

          <section className="rounded-lg border border-[#d9d0df] bg-[#f1ecf4] p-5">
            <div className="flex items-center gap-2 text-sm font-semibold text-[#4d2d66]">
              <Sparkles size={16} />
              SouLink 的判断边界
            </div>
            <p className="mt-3 text-xs leading-6 text-[#6c5b74]">
              系统只依据这条意图中的时间、地点、偏好和显式资料给出线索，不会替你判断一个人是否“注定合适”。
            </p>
          </section>

          {intent.isOwner && intent.status === 'draft' ? (
            <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
              <h2 className="text-sm font-semibold text-[#37332f]">草稿尚未参与匹配</h2>
              <p className="mt-2 text-xs leading-5 text-[#77736c]">
                发布后才会生成推荐，并按你选择的范围展示。
              </p>
              <button
                className="mt-4 inline-flex h-9 w-full items-center justify-center gap-2 rounded-md bg-[#5b3477] text-xs font-semibold text-white disabled:opacity-50"
                disabled={publishing}
                onClick={() => void publishDraft()}
                type="button"
              >
                {publishing ? (
                  <LoaderCircle className="animate-spin" size={15} />
                ) : (
                  <ArrowRight size={15} />
                )}
                发布这个意图
              </button>
            </section>
          ) : null}

          {!intent.isOwner && intent.status === 'active' ? (
            <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
              {intent.interestState === 'interested' ||
              intent.interestState === 'accepted' ? (
                <>
                  <div className="flex items-center gap-2 text-sm font-semibold text-[#2d6a53]">
                    <HeartHandshake size={17} />
                    {intent.interestState === 'accepted'
                      ? '双方都愿意认识'
                      : '已向对方表达兴趣'}
                  </div>
                  <p className="mt-2 text-xs leading-5 text-[#77736c]">
                    {intent.interestState === 'accepted'
                      ? '匹配房间已经创建，可以去确认见面安排。'
                      : '对方接受前，不会看到你的联系方式。'}
                  </p>
                  {intent.interestState === 'interested' ? (
                    <button
                      className="mt-3 inline-flex h-8 items-center gap-1.5 text-xs font-medium text-[#80665c]"
                      disabled={submittingInterest}
                      onClick={() => void handleWithdraw()}
                      type="button"
                    >
                      撤回兴趣
                    </button>
                  ) : (
                    <Link
                      className="mt-3 inline-flex items-center gap-1 text-xs font-semibold text-[#5b3477]"
                      to="/matches"
                    >
                      打开匹配房间
                      <ArrowRight size={13} />
                    </Link>
                  )}
                </>
              ) : (
                <>
                  <div className="flex items-center gap-2 text-sm font-semibold text-[#37332f]">
                    <MessageSquarePlus size={17} />
                    我对这个意图感兴趣
                  </div>
                  <textarea
                    className="mt-3 min-h-24 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fcfbf9] p-3 text-sm leading-5 outline-none focus:border-[#7a5a8e]"
                    maxLength={500}
                    onChange={(event) => setInterestMessage(event.target.value)}
                    placeholder="一句具体的附言，会比“你好”更容易开始。"
                    value={interestMessage}
                  />
                  <button
                    className="mt-3 inline-flex h-9 w-full items-center justify-center gap-2 rounded-md bg-[#5b3477] text-xs font-semibold text-white disabled:opacity-50"
                    disabled={submittingInterest}
                    onClick={() => void handleInterest()}
                    type="button"
                  >
                    {submittingInterest ? (
                      <LoaderCircle className="animate-spin" size={15} />
                    ) : (
                      <MessageSquarePlus size={15} />
                    )}
                    发送兴趣
                  </button>
                </>
              )}
            </section>
          ) : null}

          <section className="rounded-lg border border-[#dfdcd6] bg-[#f5f3ef] p-4 text-xs leading-5 text-[#77736c]">
            <p>
              创建于 {formatDateTime(intent.createdAt)}
              <br />
              到期于 {formatDateTime(intent.expiresAt)}
            </p>
          </section>
        </aside>
      </div>
    </div>
  )
}
