import { Activity, Clock3, LoaderCircle, RefreshCw, TriangleAlert } from 'lucide-react'
import { useEffect, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import { fetchOperationalMetrics, type OperationalMetrics } from '@/lib/operations-api'

function formatDuration(value: number) {
  if (value < 1) return '< 1 ms'
  if (value < 1000) return `${Math.round(value)} ms`
  return `${(value / 1000).toFixed(1)} s`
}

function formatUptime(seconds: number) {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days} 天 ${hours} 小时`
  if (hours > 0) return `${hours} 小时 ${minutes} 分钟`
  return `${Math.max(0, minutes)} 分钟`
}

export function OperationalMetricsPanel() {
  const [metrics, setMetrics] = useState<OperationalMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [notice, setNotice] = useState('')
  const [refreshVersion, setRefreshVersion] = useState(0)

  useEffect(() => {
    let cancelled = false
    void fetchOperationalMetrics()
      .then((result) => {
        if (!cancelled) setMetrics(result)
      })
      .catch((error) => {
        if (!cancelled) setNotice(getErrorMessage(error, '运行指标暂时无法加载'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [refreshVersion])

  const summary = metrics
    ? [
        { label: '运行时间', value: formatUptime(metrics.uptimeSeconds) },
        { label: '请求总量', value: metrics.requests.total.toLocaleString('zh-CN') },
        { label: '服务端错误', value: metrics.requests.serverErrors.toLocaleString('zh-CN') },
        { label: '平均响应', value: formatDuration(metrics.requests.averageDurationMs) },
      ]
    : []

  const refresh = () => {
    setLoading(true)
    setNotice('')
    setRefreshVersion((current) => current + 1)
  }

  return (
    <section className="border-t border-[#e3e0db] py-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <Activity className="text-[#527b6d]" size={16} />
            <h2 className="text-sm font-semibold text-[#35322f]">运行健康</h2>
          </div>
          <p className="mt-1 text-xs text-[#817d76]">
            仅聚合路由模板与任务名，不采集正文、邮箱或查询参数。
          </p>
        </div>
        <button
          aria-label="刷新运行指标"
          className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-[#d8d4cd] bg-white text-[#6f6b65] hover:bg-[#f0eeea]"
          disabled={loading}
          onClick={refresh}
          title="刷新运行指标"
          type="button"
        >
          {loading ? <LoaderCircle className="animate-spin" size={15} /> : <RefreshCw size={15} />}
        </button>
      </div>

      {notice ? (
        <div className="mt-4 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <TriangleAlert className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      {metrics ? (
        <>
          <div className="mt-5 grid grid-cols-2 gap-x-5 gap-y-4 border-y border-[#e3e0db] py-4 lg:grid-cols-4">
            {summary.map((item) => (
              <div className="min-w-0" key={item.label}>
                <p className="text-[11px] text-[#918d86]">{item.label}</p>
                <p className="mt-1 truncate text-base font-semibold text-[#35322f]">{item.value}</p>
              </div>
            ))}
          </div>

          <div className="grid gap-6 pt-5 xl:grid-cols-[minmax(0,1.25fr)_minmax(280px,0.75fr)]">
            <div className="min-w-0 overflow-x-auto">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h3 className="text-xs font-semibold text-[#5d5953]">高频路由</h3>
                <span className="text-[11px] text-[#918d86]">按请求数排序</span>
              </div>
              <table className="w-full min-w-[560px] table-fixed text-left text-xs">
                <thead className="text-[11px] text-[#918d86]">
                  <tr className="border-b border-[#e3e0db]">
                    <th className="w-[55%] pb-2 font-medium">路由</th>
                    <th className="pb-2 text-right font-medium">请求</th>
                    <th className="pb-2 text-right font-medium">4xx / 5xx</th>
                    <th className="pb-2 text-right font-medium">平均耗时</th>
                  </tr>
                </thead>
                <tbody>
                  {metrics.routes.slice(0, 8).map((route) => (
                    <tr className="border-b border-[#ece9e4] last:border-0" key={`${route.method}:${route.route}`}>
                      <td className="truncate py-2.5 pr-4 font-mono text-[11px] text-[#4f5f59]" title={`${route.method} ${route.route}`}>
                        <span className="mr-2 font-semibold text-[#527b6d]">{route.method}</span>
                        {route.route}
                      </td>
                      <td className="py-2.5 text-right text-[#5d5953]">{route.count}</td>
                      <td className="py-2.5 text-right text-[#817d76]">
                        {route.clientErrors} / {route.serverErrors}
                      </td>
                      <td className="py-2.5 text-right text-[#817d76]">
                        {formatDuration(route.averageDurationMs)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {metrics.routes.length === 0 ? (
                <p className="py-8 text-center text-xs text-[#918d86]">尚无请求数据</p>
              ) : null}
            </div>

            <div className="min-w-0 xl:border-l xl:border-[#e3e0db] xl:pl-6">
              <div className="mb-3 flex items-center gap-2">
                <Clock3 className="text-[#775b35]" size={15} />
                <h3 className="text-xs font-semibold text-[#5d5953]">后台任务</h3>
              </div>
              <div className="divide-y divide-[#ece9e4]">
                {metrics.tasks.map((task) => (
                  <div className="flex items-center justify-between gap-4 py-3 first:pt-0" key={task.name}>
                    <div className="min-w-0">
                      <p className="truncate font-mono text-[11px] text-[#4f5f59]" title={task.name}>
                        {task.name}
                      </p>
                      <p className="mt-1 text-[10px] text-[#918d86]">
                        {task.count} 次 · {task.failures} 次失败
                      </p>
                    </div>
                    <span className="shrink-0 text-[11px] text-[#817d76]">
                      {formatDuration(task.averageDurationMs)}
                    </span>
                  </div>
                ))}
              </div>
              {metrics.tasks.length === 0 ? (
                <p className="py-8 text-center text-xs text-[#918d86]">尚无任务数据</p>
              ) : null}
            </div>
          </div>
        </>
      ) : null}
    </section>
  )
}
