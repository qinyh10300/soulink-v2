import {
  Activity,
  CalendarCheck2,
  LoaderCircle,
  RefreshCw,
  Repeat2,
  ShieldAlert,
  TrendingUp,
  UsersRound,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import { CampusThemeAdminPanel } from '@/components/admin/campus-theme-admin-panel'
import { fetchProductMetrics, type ProductMetricsDashboard } from '@/lib/product-metrics-api'

type MetricsPeriod = 7 | 28 | 90

const periodOptions: Array<{ value: MetricsPeriod; label: string }> = [
  { value: 7, label: '近 7 天' },
  { value: 28, label: '近 28 天' },
  { value: 90, label: '近 90 天' },
]

function formatPercent(value: number) {
  return `${new Intl.NumberFormat('zh-CN', {
    maximumFractionDigits: 1,
  }).format(value)}%`
}

function formatCompact(value: number) {
  return new Intl.NumberFormat('zh-CN', { notation: 'compact' }).format(value)
}

function formatShortDate(value: string) {
  const date = new Date(`${value}T00:00:00Z`)
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'numeric',
    day: 'numeric',
    timeZone: 'UTC',
  }).format(date)
}

export function ProductMetricsPanel() {
  const [period, setPeriod] = useState<MetricsPeriod>(28)
  const [dashboard, setDashboard] = useState<ProductMetricsDashboard | null>(null)
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [refreshVersion, setRefreshVersion] = useState(0)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setNotice('')
    void fetchProductMetrics(period)
      .then((result) => {
        if (!cancelled) setDashboard(result)
      })
      .catch((error) => {
        if (!cancelled) {
          setNotice(getErrorMessage(error, '产品指标暂时无法加载'))
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [period, refreshVersion])

  const chartMax = useMemo(() => {
    if (!dashboard) return 1
    return Math.max(
      1,
      ...dashboard.daily.map((item) =>
        Math.max(item.activeUsers, item.intentsCreated, item.matchesCreated),
      ),
    )
  }, [dashboard])

  if (loading && !dashboard) {
    return (
      <div className="flex min-h-72 items-center justify-center text-sm text-[#77736c]">
        <LoaderCircle className="mr-2 animate-spin" size={17} />
        正在汇总产品指标...
      </div>
    )
  }

  return (
    <section className="py-6">
      <div className="flex flex-col gap-3 border-b border-[#e3e0db] pb-5 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h2 className="text-sm font-semibold text-[#35322f]">连接质量概览</h2>
          <p className="mt-1 text-xs text-[#817d76]">
            事件由成功的服务端动作生成，不采集聊天正文和联系方式。
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            aria-label="选择统计周期"
            className="h-9 rounded-md border border-[#d8d4cd] bg-white px-3 text-xs text-[#5f5b55] outline-none focus:border-[#527b6d]"
            onChange={(event) => setPeriod(Number(event.target.value) as MetricsPeriod)}
            value={period}
          >
            {periodOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <button
            aria-label="刷新产品指标"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-[#d8d4cd] bg-white text-[#6f6b65] hover:bg-[#f0eeea]"
            onClick={() => setRefreshVersion((current) => current + 1)}
            title="刷新"
            type="button"
          >
            <RefreshCw className={loading ? 'animate-spin' : ''} size={15} />
          </button>
        </div>
      </div>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <ShieldAlert className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      {dashboard ? (
        <>
          <div className="grid gap-3 py-5 sm:grid-cols-2 xl:grid-cols-4">
            {[
              {
                label: '完成见面',
                value: dashboard.northStar.meetupsCompleted,
                detail: `${dashboard.northStar.meetupsConfirmed} 次已确认`,
                icon: CalendarCheck2,
                tone: 'text-[#315c4d] bg-[#e7f0ec]',
              },
              {
                label: '匹配到见面',
                value: formatPercent(dashboard.northStar.matchToMeetupRate),
                detail: '以成功匹配为分母',
                icon: TrendingUp,
                tone: 'text-[#6b4a7f] bg-[#f0eaf3]',
              },
              {
                label: '周期活跃用户',
                value: dashboard.activity.activeUsers,
                detail: `${dashboard.activity.registeredUsers} 名有效用户`,
                icon: UsersRound,
                tone: 'text-[#3d6173] bg-[#e7eff3]',
              },
              {
                label: '重复发起意图',
                value: formatPercent(dashboard.northStar.repeatIntentRate),
                detail: `${dashboard.activity.repeatIntentUsers} 名重复用户`,
                icon: Repeat2,
                tone: 'text-[#775b35] bg-[#f3eee4]',
              },
            ].map((item) => {
              const Icon = item.icon
              return (
                <article
                  className="min-w-0 rounded-lg border border-[#dfdcd6] bg-white p-4"
                  key={item.label}
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="text-xs font-medium text-[#77736c]">{item.label}</span>
                    <span
                      className={`inline-flex h-8 w-8 items-center justify-center rounded-md ${item.tone}`}
                    >
                      <Icon size={15} />
                    </span>
                  </div>
                  <p className="mt-4 text-2xl font-semibold text-[#292724]">{item.value}</p>
                  <p className="mt-1 truncate text-[11px] text-[#918d86]">{item.detail}</p>
                </article>
              )
            })}
          </div>

          <div className="grid gap-6 border-t border-[#e3e0db] py-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
            <section className="min-w-0">
              <div className="flex items-center gap-2">
                <Activity className="text-[#527b6d]" size={16} />
                <h2 className="text-sm font-semibold text-[#35322f]">每日连接活动</h2>
              </div>
              <div className="mt-5 flex h-44 items-end gap-1 overflow-hidden border-b border-[#dcd8d2] px-1 sm:gap-1.5">
                {dashboard.daily.map((item, index) => (
                  <div
                    className="group relative flex h-full min-w-0 flex-1 items-end justify-center"
                    key={item.date}
                    title={`${formatShortDate(item.date)}：${item.activeUsers} 活跃，${item.intentsCreated} 意图，${item.matchesCreated} 匹配`}
                  >
                    <div
                      className="w-full max-w-5 rounded-t-sm bg-[#9db9ae] transition-colors group-hover:bg-[#527b6d]"
                      style={{
                        height: `${Math.max(3, (item.activeUsers / chartMax) * 100)}%`,
                      }}
                    />
                    {dashboard.daily.length <= 14 ||
                    index % Math.ceil(dashboard.daily.length / 7) === 0 ? (
                      <span className="absolute -bottom-5 whitespace-nowrap text-[9px] text-[#918d86]">
                        {formatShortDate(item.date)}
                      </span>
                    ) : null}
                  </div>
                ))}
              </div>
              <div className="mt-7 flex flex-wrap gap-x-5 gap-y-2 text-[11px] text-[#817d76]">
                <span>活跃用户按服务端有效动作去重</span>
                <span>当前周期 {dashboard.period.days} 天</span>
              </div>
            </section>

            <section className="min-w-0 xl:border-l xl:border-[#e3e0db] xl:pl-6">
              <div className="flex items-center justify-between gap-3">
                <h2 className="text-sm font-semibold text-[#35322f]">连接漏斗</h2>
                <span className="text-[11px] text-[#918d86]">按独立用户</span>
              </div>
              <div className="mt-4 space-y-3">
                {dashboard.funnel.map((step, index) => {
                  const firstCount = dashboard.funnel[0]?.count || 1
                  const width = Math.max(3, Math.min(100, (step.count / firstCount) * 100))
                  return (
                    <div key={step.key}>
                      <div className="flex items-center justify-between gap-3 text-xs">
                        <span className="font-medium text-[#5d5953]">{step.label}</span>
                        <span className="text-[#817d76]">
                          {formatCompact(step.count)}
                          {index > 0 ? ` · ${formatPercent(step.conversionFromPrevious)}` : ''}
                        </span>
                      </div>
                      <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-[#ece9e4]">
                        <div
                          className="h-full rounded-full bg-[#607f73]"
                          style={{ width: `${width}%` }}
                        />
                      </div>
                    </div>
                  )
                })}
              </div>
            </section>
          </div>

          <div className="flex flex-col gap-3 border-t border-[#e3e0db] py-5 text-xs text-[#6f6b65] sm:flex-row sm:items-center sm:justify-between">
            <span className="inline-flex items-center gap-2">
              <ShieldAlert className="text-[#8b5b4f]" size={15} />
              {dashboard.safety.openReports} 条待处理安全工单，周期举报率{' '}
              {formatPercent(dashboard.safety.reportRate)}
            </span>
            <span className="inline-flex items-center gap-2 text-[#918d86]">
              <Activity size={14} />
              {dashboard.activity.intentCreators} 名用户发起过意图
            </span>
          </div>
        </>
      ) : null}
      <CampusThemeAdminPanel period={period} />
    </section>
  )
}
