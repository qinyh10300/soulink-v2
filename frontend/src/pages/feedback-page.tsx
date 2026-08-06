import {
  AlertCircle,
  AlertTriangle,
  ArrowLeft,
  Bug,
  Check,
  HeartHandshake,
  Lightbulb,
  LoaderCircle,
  MessageSquareText,
  Send,
  ShieldAlert,
  Star,
} from 'lucide-react'
import { type FormEvent, useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import { submitFeedback } from '@/lib/feedback-api'
import {
  submitSafetyReport,
  type SafetyContextType,
  type SafetyReportCategory,
} from '@/lib/safety-api'
import { getStoredUserProfile } from '@/lib/user-profile-storage'
import {
  connectionIntentPath,
  connectionPaths,
  connectionRoomPath,
} from '@/lib/connection-routes'
import { cn } from '@/lib/utils'

type FeedbackCategory = 'match' | 'bug' | 'feature' | 'other'

const feedbackCategories: Array<{
  id: FeedbackCategory
  icon: typeof HeartHandshake
  label: string
}> = [
  { id: 'match', icon: HeartHandshake, label: '匹配体验' },
  { id: 'bug', icon: Bug, label: '产品问题' },
  { id: 'feature', icon: Lightbulb, label: '功能建议' },
  { id: 'other', icon: MessageSquareText, label: '其他反馈' },
]

const reportReasons: Array<{
  id: SafetyReportCategory
  label: string
}> = [
  { id: 'harassment', label: '骚扰或冒犯' },
  { id: 'privacy', label: '隐私泄露' },
  { id: 'impersonation', label: '冒充或虚假身份' },
  { id: 'unsafe', label: '线下安全风险' },
  { id: 'spam', label: '广告或垃圾内容' },
  { id: 'other', label: '其他风险' },
]

function parseContext(
  params: URLSearchParams,
): { type: SafetyContextType; id: number } | null {
  const roomID = Number(params.get('roomId'))
  if (Number.isInteger(roomID) && roomID > 0) {
    return { type: 'match_room', id: roomID }
  }
  const intentID = Number(params.get('intentId'))
  if (Number.isInteger(intentID) && intentID > 0) {
    return { type: 'intent', id: intentID }
  }
  return null
}

export function FeedbackPage() {
  const [searchParams] = useSearchParams()
  const profile = useMemo(() => getStoredUserProfile(), [])
  const context = useMemo(() => parseContext(searchParams), [searchParams])
  const isSafetyReport = searchParams.get('category') === 'report'
  const subject = searchParams.get('subject')?.trim() ?? ''
  const [selectedCategory, setSelectedCategory] =
    useState<FeedbackCategory>('match')
  const [reportReason, setReportReason] =
    useState<SafetyReportCategory>('harassment')
  const [rating, setRating] = useState(0)
  const [details, setDetails] = useState('')
  const [contactEmail, setContactEmail] = useState(profile.email)
  const [blockUser, setBlockUser] = useState(false)
  const [notice, setNotice] = useState('')
  const [noticeTone, setNoticeTone] = useState<'error' | 'success'>('success')
  const [ticketID, setTicketID] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const returnTo =
    context?.type === 'match_room'
      ? connectionRoomPath(context.id)
      : context?.type === 'intent'
        ? connectionIntentPath(context.id)
        : connectionPaths.root

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (submitting) {
      return
    }

    if (isSafetyReport) {
      if (!context) {
        setNoticeTone('error')
        setNotice('没有找到需要举报的意图或匹配房间，请从对应页面重新进入。')
        return
      }
      if (details.trim().length < 8) {
        setNoticeTone('error')
        setNotice('请用至少 8 个字说明发生了什么。')
        return
      }
    } else if (rating === 0) {
      setNoticeTone('error')
      setNotice('请先给这次体验打分。')
      return
    }

    setSubmitting(true)
    setNotice('')
    try {
      if (isSafetyReport && context) {
        const result = await submitSafetyReport({
          contextType: context.type,
          contextId: context.id,
          category: reportReason,
          details: details.trim(),
          blockUser,
        })
        setTicketID(result.ticketId)
        setNoticeTone('success')
        setNotice(
          result.blocked
            ? '举报已提交，对方已被拉黑，相关房间的交流和联系方式已经关闭。'
            : '举报已提交，安全管理员会在后台处理。',
        )
      } else {
        const contextNote = context
          ? `[${context.type === 'match_room' ? '匹配房间' : '意图'} #${context.id}${subject ? ` · ${subject}` : ''}]\n`
          : ''
        const result = await submitFeedback({
          category: selectedCategory,
          rating,
          details: `${contextNote}${details.trim()}`.trim(),
          contactEmail: contactEmail.trim() || undefined,
        })
        setTicketID(result.ticketId)
        setNoticeTone('success')
        setNotice('反馈已收到，感谢你把真实体验说清楚。')
      }
    } catch (error) {
      setNoticeTone('error')
      setNotice(getErrorMessage(error, isSafetyReport ? '举报暂时无法提交' : '反馈暂时无法提交'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1020px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex items-center gap-3 border-b border-[#dfdcd6] pb-5">
        <Link
          aria-label="返回"
          className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#6f6b65] hover:bg-[#eeece8]"
          title="返回"
          to={returnTo}
        >
          <ArrowLeft size={19} />
        </Link>
        <div>
          <p className="flex items-center gap-1.5 text-xs font-medium text-[#6c4a80]">
            {isSafetyReport ? <ShieldAlert size={13} /> : <MessageSquareText size={13} />}
            {isSafetyReport ? 'SAFETY REPORT' : 'PRODUCT FEEDBACK'}
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">
            {isSafetyReport ? '举报与安全处置' : '告诉我们这次体验如何'}
          </h1>
          <p className="mt-1 text-sm text-[#77736c]">
            {isSafetyReport
              ? '举报会关联当前场景；如感到不安全，可以同时停止双方继续联系。'
              : '具体的反馈会帮助我们判断匹配是否真的有用。'}
          </p>
        </div>
      </header>

      {ticketID ? (
        <section className="mx-auto mt-10 max-w-xl rounded-lg border border-[#bfd4ca] bg-[#eef6f2] p-6 text-center">
          <span className="mx-auto inline-flex h-10 w-10 items-center justify-center rounded-md bg-white text-[#2f6c55]">
            <Check size={20} />
          </span>
          <h2 className="mt-4 text-lg font-semibold text-[#30473e]">{notice}</h2>
          <p className="mt-2 text-xs text-[#65766f]">工单编号：{ticketID}</p>
          <Link
            className="mt-5 inline-flex h-9 items-center rounded-md bg-[#376f5b] px-4 text-xs font-semibold text-white"
            to={returnTo}
          >
            返回原页面
          </Link>
        </section>
      ) : (
        <form
          className="grid gap-7 py-7 lg:grid-cols-[minmax(0,1.2fr)_minmax(260px,0.8fr)]"
          onSubmit={handleSubmit}
        >
          <main className="min-w-0 space-y-6">
            {subject ? (
              <section className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] px-4 py-3">
                <p className="text-[11px] font-medium text-[#817d76]">当前场景</p>
                <p className="mt-1 text-sm font-semibold text-[#4f4b46]">{subject}</p>
              </section>
            ) : null}

            {isSafetyReport ? (
              <section>
                <h2 className="text-sm font-semibold text-[#37332f]">发生了什么</h2>
                <div className="mt-3 grid gap-2 sm:grid-cols-2">
                  {reportReasons.map((reason) => (
                    <button
                      className={cn(
                        'h-10 rounded-md border px-3 text-left text-xs font-medium transition-colors',
                        reportReason === reason.id
                          ? 'border-[#9b5268] bg-[#fff1f5] text-[#7e3d55]'
                          : 'border-[#dedad4] bg-white text-[#65615b] hover:border-[#c8aeb8]',
                      )}
                      key={reason.id}
                      onClick={() => setReportReason(reason.id)}
                      type="button"
                    >
                      {reason.label}
                    </button>
                  ))}
                </div>
              </section>
            ) : (
              <>
                <section>
                  <h2 className="text-sm font-semibold text-[#37332f]">反馈类型</h2>
                  <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
                    {feedbackCategories.map((category) => {
                      const Icon = category.icon
                      return (
                        <button
                          className={cn(
                            'flex h-16 flex-col items-center justify-center gap-1.5 rounded-md border text-xs font-medium transition-colors',
                            selectedCategory === category.id
                              ? 'border-[#7c5a90] bg-[#f0eaf3] text-[#4d2d66]'
                              : 'border-[#dedad4] bg-white text-[#65615b] hover:border-[#bbaac5]',
                          )}
                          key={category.id}
                          onClick={() => setSelectedCategory(category.id)}
                          type="button"
                        >
                          <Icon size={17} />
                          {category.label}
                        </button>
                      )
                    })}
                  </div>
                </section>

                <section>
                  <h2 className="text-sm font-semibold text-[#37332f]">这次体验怎么样</h2>
                  <div className="mt-3 flex items-center gap-1">
                    {Array.from({ length: 5 }, (_, index) => {
                      const value = index + 1
                      return (
                        <button
                          aria-label={`评分 ${value} 星`}
                          className="inline-flex h-10 w-10 items-center justify-center rounded-md hover:bg-[#f1efeb]"
                          key={value}
                          onClick={() => setRating(value)}
                          type="button"
                        >
                          <Star
                            className={
                              value <= rating ? 'text-[#a16d2c]' : 'text-[#c8c3bb]'
                            }
                            fill={value <= rating ? 'currentColor' : 'none'}
                            size={22}
                          />
                        </button>
                      )
                    })}
                    <span className="ml-2 text-xs text-[#817d76]">
                      {rating ? `${rating} / 5` : '尚未评分'}
                    </span>
                  </div>
                </section>
              </>
            )}

            <label className="block text-sm font-semibold text-[#37332f]">
              {isSafetyReport ? '请描述具体经过' : '哪些地方有用，哪些地方不对'}
              <textarea
                className="mt-3 min-h-48 w-full resize-y rounded-lg border border-[#d9d5cf] bg-white p-4 text-sm leading-7 text-[#4f4b46] outline-none placeholder:text-[#9a958d] focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10"
                maxLength={2000}
                onChange={(event) => setDetails(event.target.value)}
                placeholder={
                  isSafetyReport
                    ? '说明发生的时间、行为和你担心的风险。请不要在这里填写无关的敏感信息。'
                    : '例如：推荐理由很具体，但我希望见面地点能显示开放时间。'
                }
                value={details}
              />
              <span className="mt-1 block text-right text-[11px] font-normal text-[#8a867f]">
                {details.length} / 2000
              </span>
            </label>
          </main>

          <aside className="min-w-0 space-y-5">
            {isSafetyReport ? (
              <>
                <section className="rounded-lg border border-[#e0c8d0] bg-[#fff6f8] p-4">
                  <div className="flex items-start gap-2">
                    <AlertTriangle className="mt-0.5 shrink-0 text-[#8d465d]" size={16} />
                    <div>
                      <h2 className="text-xs font-semibold text-[#74394c]">同时拉黑对方</h2>
                      <p className="mt-1 text-[11px] leading-5 text-[#80606a]">
                        会立即关闭双方已有匹配房间、隐藏联系方式，并停止彼此出现在意图流和推荐中。
                      </p>
                    </div>
                  </div>
                  <button
                    className={cn(
                      'mt-4 flex w-full items-center justify-between rounded-md border px-3 py-3 text-left',
                      blockUser
                        ? 'border-[#b86b82] bg-white'
                        : 'border-[#dfcbd2] bg-[#fffafb]',
                    )}
                    onClick={() => setBlockUser((current) => !current)}
                    type="button"
                  >
                    <span className="text-xs font-semibold text-[#693747]">
                      提交后拉黑
                    </span>
                    <span
                      aria-hidden="true"
                      className={cn(
                        'relative h-6 w-11 rounded-full transition-colors',
                        blockUser ? 'bg-[#9b5268]' : 'bg-[#c9c5bf]',
                      )}
                    >
                      <span
                        className={cn(
                          'absolute top-1 h-4 w-4 rounded-full bg-white transition-transform',
                          blockUser ? 'translate-x-6' : 'translate-x-1',
                        )}
                      />
                    </span>
                  </button>
                </section>

                <p className="text-[11px] leading-5 text-[#817d76]">
                  举报材料仅用于安全审核。紧急人身风险请优先联系学校保卫部门或公共安全机构。
                </p>
              </>
            ) : (
              <label className="block rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-4 text-xs font-semibold text-[#5f5b55]">
                联系邮箱（选填）
                <input
                  className="mt-3 h-10 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-sm font-normal outline-none focus:border-[#7a5a8e]"
                  onChange={(event) => setContactEmail(event.target.value)}
                  type="email"
                  value={contactEmail}
                />
                <span className="mt-2 block text-[10px] font-normal leading-4 text-[#817d76]">
                  只有需要进一步确认时才会联系。
                </span>
              </label>
            )}

            {notice && noticeTone === 'error' ? (
              <div className="flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-3 py-3 text-xs leading-5 text-[#8c463d]">
                <AlertCircle className="mt-0.5 shrink-0" size={14} />
                {notice}
              </div>
            ) : null}

            <button
              className={cn(
                'inline-flex h-10 w-full items-center justify-center gap-2 rounded-md text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50',
                isSafetyReport
                  ? 'bg-[#8d465d] hover:bg-[#79394e]'
                  : 'bg-[#5b3477] hover:bg-[#4c2b65]',
              )}
              disabled={submitting}
              type="submit"
            >
              {submitting ? (
                <LoaderCircle className="animate-spin" size={16} />
              ) : (
                <Send size={16} />
              )}
              {submitting
                ? '正在提交...'
                : isSafetyReport
                  ? '提交安全举报'
                  : '提交反馈'}
            </button>
          </aside>
        </form>
      )}
    </div>
  )
}
