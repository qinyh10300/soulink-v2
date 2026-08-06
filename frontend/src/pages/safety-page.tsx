import {
  AlertCircle,
  ArrowLeft,
  Ban,
  Clock3,
  EyeOff,
  FileWarning,
  Heart,
  LoaderCircle,
  LockKeyhole,
  Send,
  ShieldCheck,
  Trash2,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import {
  appealSafetyReport,
  fetchMySafetyCases,
  type SafetyCase,
  type SafetyReportCategory,
} from '@/lib/safety-api'
import { getAuthSession } from '@/lib/user-profile-storage'

const principles = [
  {
    icon: LockKeyhole,
    title: '双方同意后再揭示',
    copy: '联系方式默认隐藏。进入匹配房间后，每个人仍然单独决定是否向对方显示。',
  },
  {
    icon: EyeOff,
    title: '认真关系单独保护',
    copy: '恋爱意图不进入公开流，公开预览不展示可识别身份，且必须由用户主动开启。',
  },
  {
    icon: Ban,
    title: '随时拒绝与退出',
    copy: '可以跳过推荐、撤回兴趣、暂停匹配、隐藏联系方式，并对用户或内容进行举报。',
  },
  {
    icon: Trash2,
    title: '资料可以编辑和删除',
    copy: 'AI 画像始终可编辑、拒绝或留空；账户和意图可以主动结束或删除。',
  },
]

const caseCategoryLabels: Record<SafetyReportCategory, string> = {
  harassment: '骚扰或冒犯',
  privacy: '隐私泄露',
  impersonation: '冒充或虚假身份',
  unsafe: '线下安全风险',
  spam: '广告或垃圾内容',
  other: '其他风险',
}

function formatCaseDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

function MySafetyCases() {
  const [cases, setCases] = useState<SafetyCase[]>([])
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [appealDrafts, setAppealDrafts] = useState<Record<number, string>>({})
  const [busyID, setBusyID] = useState<number | null>(null)
  const sessionUserID = getAuthSession()?.userId

  useEffect(() => {
    if (!sessionUserID) return
    fetchMySafetyCases()
      .then(setCases)
      .catch((error) => setNotice(getErrorMessage(error, '安全案件暂时无法加载')))
      .finally(() => setLoading(false))
  }, [sessionUserID])

  if (!sessionUserID) return null

  const submitAppeal = async (item: SafetyCase) => {
    const reason = appealDrafts[item.id]?.trim() ?? ''
    if (reason.length < 8 || busyID) {
      setNotice('请用至少 8 个字说明申诉理由。')
      return
    }
    setBusyID(item.id)
    setNotice('')
    try {
      const appeal = await appealSafetyReport(item.id, reason)
      setCases((current) => current.map((entry) => (
        entry.id === item.id ? { ...entry, appeal } : entry
      )))
    } catch (error) {
      setNotice(getErrorMessage(error, '申诉暂时无法提交'))
    } finally {
      setBusyID(null)
    }
  }

  return (
    <section className="mt-12 border-t border-[#dfdcd6] pt-10">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="flex items-center gap-2 text-xs font-semibold text-[#7e3d55]">
            <FileWarning size={14} />
            MY SAFETY CASES
          </p>
          <h2 className="mt-2 text-xl font-semibold">我的安全案件</h2>
          <p className="mt-2 text-sm leading-6 text-[#77736c]">
            这里只展示涉及你的已结案状态，不显示举报人身份、举报详情或聊天证据。
          </p>
        </div>
      </div>

      {notice ? (
        <div className="mt-4 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-3 py-3 text-xs text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={14} />
          {notice}
        </div>
      ) : null}

      {loading ? (
        <div className="flex h-28 items-center justify-center text-sm text-[#77736c]">
          <LoaderCircle className="mr-2 animate-spin" size={16} />
          正在读取案件状态...
        </div>
      ) : cases.length === 0 ? (
        <p className="mt-5 rounded-md border border-[#dfdcd6] bg-white px-4 py-5 text-sm text-[#77736c]">
          当前没有涉及你的已结案安全案件。
        </p>
      ) : (
        <div className="mt-5 divide-y divide-[#e6e3de] overflow-hidden rounded-lg border border-[#dfdcd6] bg-white">
          {cases.map((item) => (
            <article className="p-4 sm:p-5" key={item.id}>
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-xs font-bold text-[#7e3d55]">{item.ticketId}</span>
                <span className="rounded bg-[#f1efeb] px-2 py-1 text-[10px] font-medium text-[#69655f]">
                  {caseCategoryLabels[item.category]}
                </span>
                <span className={`rounded px-2 py-1 text-[10px] font-semibold ${
                  item.status === 'dismissed'
                    ? 'bg-[#e9f2ed] text-[#376f5b]'
                    : 'bg-[#fff1e9] text-[#8e5435]'
                }`}>
                  {item.status === 'dismissed' ? '已撤销处理' : '已处理'}
                </span>
              </div>
              <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1">
                <h3 className="text-sm font-semibold text-[#4f4b46]">
                  {item.contextTitle || '安全场景'}
                </h3>
                <span className="inline-flex items-center gap-1 text-[11px] text-[#817d76]">
                  <Clock3 size={12} />
                  {formatCaseDate(item.updatedAt)}
                </span>
              </div>

              {item.appeal ? (
                <div className="mt-4 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-3 py-3 text-xs leading-5 text-[#67625c]">
                  <p className="font-semibold">
                    申诉状态：{
                      item.appeal.status === 'pending'
                        ? '等待复核'
                        : item.appeal.status === 'accepted'
                          ? '申诉通过'
                          : '申诉未通过'
                    }
                  </p>
                  <p className="mt-1 whitespace-pre-wrap">{item.appeal.reason}</p>
                  {item.appeal.reviewNotes ? (
                    <p className="mt-2 border-t border-[#e6e3de] pt-2 text-[#817d76]">
                      复核说明：{item.appeal.reviewNotes}
                    </p>
                  ) : null}
                </div>
              ) : item.status === 'resolved' ? (
                <div className="mt-4">
                  <textarea
                    aria-label={`${item.ticketId} 申诉理由`}
                    className="min-h-24 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fbfaf8] px-3 py-2 text-sm leading-6 outline-none focus:border-[#8d465d]"
                    maxLength={1000}
                    onChange={(event) => setAppealDrafts((current) => ({
                      ...current,
                      [item.id]: event.target.value,
                    }))}
                    placeholder="如认为处理有误，请说明可以帮助复核的事实。"
                    value={appealDrafts[item.id] ?? ''}
                  />
                  <button
                    className="mt-2 inline-flex h-9 items-center gap-2 rounded-md bg-[#8d465d] px-3 text-xs font-semibold text-white disabled:opacity-50"
                    disabled={busyID === item.id || (appealDrafts[item.id]?.trim().length ?? 0) < 8}
                    onClick={() => void submitAppeal(item)}
                    type="button"
                  >
                    {busyID === item.id ? <LoaderCircle className="animate-spin" size={14} /> : <Send size={14} />}
                    提交申诉
                  </button>
                </div>
              ) : null}
            </article>
          ))}
        </div>
      )}
    </section>
  )
}

export function SafetyPage() {
  return (
    <div className="min-h-screen bg-[#f7f6f3] text-[#292724]">
      <header className="border-b border-[#dfdcd6] bg-[#fbfaf8]">
        <div className="mx-auto flex h-16 max-w-[1040px] items-center justify-between px-4 sm:px-6">
          <Link
            className="inline-flex items-center gap-2 text-sm font-semibold text-[#4f4b46]"
            to="/"
          >
            <ArrowLeft size={16} />
            SouLink
          </Link>
          <Link
            className="inline-flex h-9 items-center rounded-md bg-[#5b3477] px-4 text-xs font-semibold text-white"
            to="/register"
          >
            清华邮箱注册
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-[1040px] px-4 py-12 sm:px-6 sm:py-16">
        <div className="max-w-2xl">
          <p className="flex items-center gap-2 text-xs font-semibold text-[#6c4a80]">
            <ShieldCheck size={14} />
            PRIVACY & SAFETY
          </p>
          <h1 className="mt-3 text-4xl font-semibold">连接值得发生，也值得被保护。</h1>
          <p className="mt-4 text-base leading-8 text-[#706c65]">
            SouLink 把身份验证、意图可见范围、双方同意和联系方式揭示当作产品状态，而不是一段藏在页尾的承诺。
          </p>
        </div>

        <div className="mt-10 grid gap-3 sm:grid-cols-2">
          {principles.map((principle) => {
            const Icon = principle.icon
            return (
              <section
                className="rounded-lg border border-[#dfdcd6] bg-white p-5"
                key={principle.title}
              >
                <Icon size={19} className="text-[#5b3477]" />
                <h2 className="mt-4 text-base font-semibold">{principle.title}</h2>
                <p className="mt-2 text-sm leading-7 text-[#706c65]">{principle.copy}</p>
              </section>
            )
          })}
        </div>

        <MySafetyCases />

        <section className="mt-12 rounded-lg border border-[#dfc7d0] bg-[#fff5f8] p-6">
          <div className="flex items-center gap-2 text-sm font-semibold text-[#7e3d55]">
            <Heart size={17} />
            认真关系匹配的额外边界
          </div>
          <ul className="mt-4 space-y-2 text-sm leading-6 text-[#745663]">
            <li>不在公开意图流展示可识别的恋爱需求。</li>
            <li>不使用恋爱偏好影响项目、学习或普通搭子推荐。</li>
            <li>不生成“命中注定”“完美匹配”等确定性判断。</li>
            <li>暂停或关闭后，不再进入新的认真关系推荐。</li>
          </ul>
        </section>

        <p className="mt-10 text-xs leading-6 text-[#817d76]">
          更完整的数据处理说明见仓库中的隐私文档。遇到骚扰、冒充、隐私泄露或安全风险时，请通过产品内反馈入口提交；紧急情况请优先联系学校或当地公共安全机构。
        </p>
      </main>
    </div>
  )
}
