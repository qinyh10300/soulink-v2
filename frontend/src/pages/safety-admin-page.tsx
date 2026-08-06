import {
  Activity,
  AlertCircle,
  CheckCircle2,
  Clock3,
  LoaderCircle,
  RefreshCw,
  Save,
  ShieldCheck,
  UserCheck,
  UserRound,
} from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import { ProductMetricsPanel } from '@/components/admin/product-metrics-panel'
import { getErrorMessage } from '@/lib/api'
import {
  fetchSafetyReports,
  updateSafetyReport,
  type SafetyReport,
  type SafetyReportStatus,
} from '@/lib/safety-api'
import { cn } from '@/lib/utils'

const statusOptions: Array<{
  value: SafetyReportStatus | ''
  label: string
}> = [
  { value: '', label: '全部状态' },
  { value: 'open', label: '待处理' },
  { value: 'reviewing', label: '核查中' },
  { value: 'resolved', label: '已处理' },
  { value: 'dismissed', label: '已驳回' },
]

const statusLabels: Record<SafetyReportStatus, string> = {
  open: '待处理',
  reviewing: '核查中',
  resolved: '已处理',
  dismissed: '已驳回',
}

const categoryLabels: Record<SafetyReport['category'], string> = {
  harassment: '骚扰或冒犯',
  privacy: '隐私泄露',
  impersonation: '冒充或虚假身份',
  unsafe: '线下安全风险',
  spam: '广告或垃圾内容',
  other: '其他风险',
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(new Date(value))
}

export function SafetyAdminPage() {
  const [view, setView] = useState<'metrics' | 'reports'>('metrics')
  const [status, setStatus] = useState<SafetyReportStatus | ''>('open')
  const [reports, setReports] = useState<SafetyReport[]>([])
  const [loading, setLoading] = useState(true)
  const [updatingID, setUpdatingID] = useState<number | null>(null)
  const [draftNotes, setDraftNotes] = useState<Record<number, string>>({})
  const [notice, setNotice] = useState('')

  const load = async (nextStatus = status) => {
    setLoading(true)
    setNotice('')
    try {
      const nextReports = await fetchSafetyReports(nextStatus)
      setReports(nextReports)
      setDraftNotes(
        Object.fromEntries(nextReports.map((report) => [report.id, report.reviewNotes])),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '安全工单暂时无法加载'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (view === 'reports') void load(status)
    // The status filter is the only query key.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [status, view])

  const persistReport = async (
    report: SafetyReport,
    input: { status?: SafetyReportStatus; assignToMe?: boolean } = {},
  ) => {
    if (updatingID) {
      return
    }
    setUpdatingID(report.id)
    setNotice('')
    try {
      const updated = await updateSafetyReport(report.id, {
        status: input.status ?? report.status,
        reviewNotes: draftNotes[report.id] ?? report.reviewNotes,
        assignToMe: input.assignToMe ?? false,
      })
      setReports((current) =>
        status && updated.status !== status
          ? current.filter((item) => item.id !== updated.id)
          : current.map((item) => (item.id === updated.id ? updated : item)),
      )
      setDraftNotes((current) => ({
        ...current,
        [updated.id]: updated.reviewNotes,
      }))
    } catch (error) {
      setNotice(getErrorMessage(error, '工单暂时无法更新'))
    } finally {
      setUpdatingID(null)
    }
  }

  return (
    <div className="mx-auto w-full max-w-[1120px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex flex-col gap-4 border-b border-[#dfdcd6] pb-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="flex items-center gap-1.5 text-xs font-medium text-[#526f64]">
            <Activity size={13} />
            PRODUCT OPERATIONS
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">运营与安全</h1>
          <p className="mt-1 text-sm text-[#77736c]">
            关注真实连接结果，同时保留人工审核和用户安全控制。
          </p>
        </div>
        {view === 'reports' ? (
          <div className="flex items-center gap-2">
            <select
              aria-label="筛选处理状态"
              className="h-9 rounded-md border border-[#d8d4cd] bg-white px-3 text-xs text-[#5f5b55] outline-none focus:border-[#7a5a8e]"
              onChange={(event) => setStatus(event.target.value as SafetyReportStatus | '')}
              value={status}
            >
              {statusOptions.map((option) => (
                <option key={option.value || 'all'} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <button
              aria-label="刷新安全工单"
              className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-[#d8d4cd] bg-white text-[#6f6b65] hover:bg-[#f0eeea]"
              onClick={() => void load()}
              title="刷新"
              type="button"
            >
              <RefreshCw size={15} />
            </button>
          </div>
        ) : null}
      </header>

      <div className="mt-5 inline-flex h-10 items-center rounded-md border border-[#d8d4cd] bg-[#efede9] p-1">
        <button
          className={`inline-flex h-8 items-center gap-1.5 rounded px-3 text-xs font-medium transition-colors ${
            view === 'metrics'
              ? 'bg-white text-[#315c4d] shadow-sm'
              : 'text-[#77736c] hover:text-[#35322f]'
          }`}
          onClick={() => setView('metrics')}
          type="button"
        >
          <Activity size={14} />
          产品指标
        </button>
        <button
          className={`inline-flex h-8 items-center gap-1.5 rounded px-3 text-xs font-medium transition-colors ${
            view === 'reports'
              ? 'bg-white text-[#6c4a80] shadow-sm'
              : 'text-[#77736c] hover:text-[#35322f]'
          }`}
          onClick={() => setView('reports')}
          type="button"
        >
          <ShieldCheck size={14} />
          安全工单
        </button>
      </div>

      {view === 'metrics' ? <ProductMetricsPanel /> : null}

      {view === 'reports' && notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      {view === 'reports' ? (
        <section className="py-6">
          {loading ? (
            <div className="flex min-h-64 items-center justify-center text-sm text-[#77736c]">
              <LoaderCircle className="mr-2 animate-spin" size={17} />
              正在读取安全工单...
            </div>
          ) : reports.length > 0 ? (
            <div className="overflow-hidden rounded-lg border border-[#dfdcd6] bg-white">
              {reports.map((report) => (
                <article
                  className="border-b border-[#e6e3de] p-4 last:border-b-0 sm:p-5"
                  key={report.id}
                >
                  <div className="flex flex-col gap-4 lg:flex-row lg:items-start">
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="text-xs font-bold text-[#5b3477]">{report.ticketId}</span>
                        <span
                          className={cn(
                            'rounded-md px-2 py-1 text-[10px] font-semibold',
                            report.status === 'open'
                              ? 'bg-[#fff1e9] text-[#8e5435]'
                              : report.status === 'reviewing'
                                ? 'bg-[#eee8f2] text-[#5b3477]'
                                : 'bg-[#e9f2ed] text-[#376f5b]',
                          )}
                        >
                          {statusLabels[report.status]}
                        </span>
                        <span className="rounded-md bg-[#f1efeb] px-2 py-1 text-[10px] font-medium text-[#69655f]">
                          {categoryLabels[report.category]}
                        </span>
                      </div>

                      <h2 className="mt-3 text-sm font-semibold text-[#37332f]">
                        {report.contextTitle || '未命名场景'}
                      </h2>
                      <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-[#5f5b55]">
                        {report.details}
                      </p>
                      <div className="mt-4 flex flex-wrap gap-x-5 gap-y-2 border-t border-[#eeebe7] pt-3 text-[11px] text-[#817d76]">
                        <span className="inline-flex items-center gap-1.5">
                          <UserRound size={12} />
                          {report.reporterNickname} 举报 {report.reportedUserNickname}
                        </span>
                        <span className="inline-flex items-center gap-1.5">
                          <Clock3 size={12} />
                          {formatDateTime(report.createdAt)}
                        </span>
                        <Link
                          className="font-medium text-[#5b3477]"
                          to={
                            report.contextType === 'match_room'
                              ? `/matches/${report.contextId}`
                              : `/intents/${report.contextId}`
                          }
                        >
                          查看关联场景
                        </Link>
                      </div>
                    </div>

                    <div className="w-full shrink-0 border-t border-[#eeebe7] pt-4 lg:w-80 lg:border-l lg:border-t-0 lg:pl-5 lg:pt-0">
                      <div className="flex min-h-9 items-center justify-between gap-3">
                        <span className="text-[11px] font-medium text-[#77736c]">负责人</span>
                        {report.assignedTo ? (
                          <span className="inline-flex min-w-0 items-center gap-1.5 text-xs font-medium text-[#4f4b46]">
                            <UserCheck className="shrink-0 text-[#527164]" size={14} />
                            <span className="truncate">
                              {report.assignedToNickname || `管理员 #${report.assignedTo}`}
                            </span>
                          </span>
                        ) : (
                          <button
                            className="inline-flex h-8 items-center gap-1.5 rounded-md border border-[#d8d4cd] bg-white px-2.5 text-xs font-medium text-[#5b3477] hover:bg-[#f4f0f6] disabled:opacity-60"
                            disabled={updatingID === report.id}
                            onClick={() => void persistReport(report, { assignToMe: true })}
                            type="button"
                          >
                            <UserCheck size={14} />
                            接手工单
                          </button>
                        )}
                      </div>

                      <label className="mt-3 block text-[11px] font-medium text-[#77736c]">
                        处理状态
                        <span className="relative mt-2 block">
                          <select
                            aria-label={`更新 ${report.ticketId} 的状态`}
                            className="h-9 w-full rounded-md border border-[#d8d4cd] bg-[#fbfaf8] px-2 text-xs text-[#4f4b46] outline-none focus:border-[#7a5a8e] disabled:opacity-60"
                            disabled={updatingID === report.id}
                            onChange={(event) =>
                              void persistReport(report, {
                                status: event.target.value as SafetyReportStatus,
                              })
                            }
                            value={report.status}
                          >
                            {statusOptions
                              .filter(
                                (
                                  option,
                                ): option is {
                                  value: SafetyReportStatus
                                  label: string
                                } => Boolean(option.value),
                              )
                              .map((option) => (
                                <option key={option.value} value={option.value}>
                                  {option.label}
                                </option>
                              ))}
                          </select>
                          {updatingID === report.id ? (
                            <LoaderCircle
                              className="absolute right-2.5 top-2.5 animate-spin text-[#6c4a80]"
                              size={14}
                            />
                          ) : null}
                        </span>
                      </label>

                      <label className="mt-3 block text-[11px] font-medium text-[#77736c]">
                        审核备注
                        <textarea
                          aria-label={`编辑 ${report.ticketId} 的审核备注`}
                          className="mt-2 min-h-24 w-full resize-y rounded-md border border-[#d8d4cd] bg-[#fbfaf8] px-3 py-2 text-xs leading-5 text-[#4f4b46] outline-none focus:border-[#7a5a8e] disabled:opacity-60"
                          disabled={updatingID === report.id}
                          maxLength={2000}
                          onChange={(event) =>
                            setDraftNotes((current) => ({
                              ...current,
                              [report.id]: event.target.value,
                            }))
                          }
                          placeholder="记录核查依据、处理动作和后续事项"
                          value={draftNotes[report.id] ?? report.reviewNotes}
                        />
                      </label>
                      <div className="mt-2 flex items-center justify-between gap-3">
                        <span className="text-[10px] text-[#918d86]">
                          {(draftNotes[report.id] ?? report.reviewNotes).length}
                          /2000
                        </span>
                        <button
                          className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#5b3477] px-3 text-xs font-semibold text-white hover:bg-[#4c2b65] disabled:cursor-not-allowed disabled:opacity-45"
                          disabled={
                            updatingID === report.id ||
                            (draftNotes[report.id] ?? report.reviewNotes) === report.reviewNotes
                          }
                          onClick={() => void persistReport(report)}
                          type="button"
                        >
                          {updatingID === report.id ? (
                            <LoaderCircle className="animate-spin" size={14} />
                          ) : (
                            <Save size={14} />
                          )}
                          保存备注
                        </button>
                      </div>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <div className="flex min-h-64 flex-col items-center justify-center rounded-lg border border-dashed border-[#d7d2cb] px-6 text-center">
              <CheckCircle2 className="text-[#527164]" size={28} />
              <p className="mt-3 text-sm font-medium text-[#55514c]">当前筛选下没有安全工单</p>
            </div>
          )}
        </section>
      ) : null}
    </div>
  )
}
