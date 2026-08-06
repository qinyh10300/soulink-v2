import {
  CalendarCheck2,
  LoaderCircle,
  Save,
  Settings2,
  UsersRound,
  X,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import { EngagementToggle } from '@/components/engagement/engagement-toggle'
import { getErrorMessage } from '@/lib/api'
import {
  fetchAdminCampusThemes,
  fetchCampusThemeMetrics,
  updateCampusTheme,
  type CampusTheme,
  type CampusThemeMetric,
  type UpdateCampusThemeInput,
} from '@/lib/engagement-api'

const iconOptions = [
  ['utensils', '餐饮'],
  ['book-open', '学习'],
  ['dumbbell', '运动'],
  ['users-round', '团队'],
  ['sparkles', '精选'],
  ['heart', '关系'],
  ['calendar', '日程'],
]

function formatPercent(value: number) {
  return `${new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 }).format(value)}%`
}

function toDraft(theme: CampusTheme): UpdateCampusThemeInput {
  return {
    title: theme.title,
    description: theme.description,
    defaultPrompt: theme.defaultPrompt,
    scheduleHint: theme.scheduleHint,
    iconName: theme.iconName,
    isActive: theme.isActive,
    displayOrder: theme.displayOrder,
  }
}

export function CampusThemeAdminPanel({ period }: { period: 7 | 28 | 90 }) {
  const [themes, setThemes] = useState<CampusTheme[]>([])
  const [metrics, setMetrics] = useState<CampusThemeMetric[]>([])
  const [loading, setLoading] = useState(true)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [draft, setDraft] = useState<UpdateCampusThemeInput | null>(null)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setNotice('')
    void Promise.all([fetchAdminCampusThemes(), fetchCampusThemeMetrics(period)])
      .then(([nextThemes, nextMetrics]) => {
        if (cancelled) return
        setThemes(nextThemes)
        setMetrics(nextMetrics.themes)
      })
      .catch((error) => {
        if (!cancelled) setNotice(getErrorMessage(error, '校园主题指标暂时无法加载'))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [period])

  const metricByKey = useMemo(
    () => new Map(metrics.map((metric) => [metric.key, metric])),
    [metrics],
  )

  const startEditing = (theme: CampusTheme) => {
    setEditingKey(theme.key)
    setDraft(toDraft(theme))
    setNotice('')
  }

  const save = async () => {
    if (!editingKey || !draft || saving) return
    setSaving(true)
    setNotice('')
    try {
      const updated = await updateCampusTheme(editingKey, draft)
      setThemes((current) =>
        current.map((theme) => (theme.key === updated.key ? updated : theme)),
      )
      setMetrics((current) =>
        current.map((metric) =>
          metric.key === updated.key ? { ...metric, title: updated.title } : metric,
        ),
      )
      setEditingKey(null)
      setDraft(null)
    } catch (error) {
      setNotice(getErrorMessage(error, '校园主题暂时无法保存'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="border-t border-[#e3e0db] py-6">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold text-[#35322f]">校园主题运营</h2>
          <p className="mt-1 text-xs text-[#817d76]">
            供给按活跃意图用户，需求按关注与表达兴趣用户去重。
          </p>
        </div>
        <span className="text-[11px] text-[#918d86]">近 {period} 天</span>
      </div>

      {notice ? (
        <p className="mt-4 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-xs text-[#8c463d]">
          {notice}
        </p>
      ) : null}

      {loading ? (
        <div className="flex min-h-32 items-center justify-center text-sm text-[#77736c]">
          <LoaderCircle className="mr-2 animate-spin" size={16} />
          正在汇总校园主题...
        </div>
      ) : (
        <div className="mt-4 overflow-hidden rounded-lg border border-[#dfdcd6] bg-white">
          {themes.map((theme) => {
            const metric = metricByKey.get(theme.key)
            const editing = editingKey === theme.key && draft
            return (
              <article className="border-b border-[#e6e3de] last:border-b-0" key={theme.key}>
                <div className="flex flex-col gap-4 p-4 lg:flex-row lg:items-center">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="text-sm font-semibold text-[#3f3b37]">{theme.title}</h3>
                      <span className="rounded bg-[#efede9] px-2 py-0.5 text-[10px] font-medium text-[#77736c]">
                        {theme.scheduleHint}
                      </span>
                      {!theme.isActive ? (
                        <span className="rounded bg-[#f4e9e7] px-2 py-0.5 text-[10px] font-semibold text-[#8b5149]">
                          已停用
                        </span>
                      ) : null}
                    </div>
                    <p className="mt-1 line-clamp-1 text-xs text-[#817d76]">
                      {theme.description}
                    </p>
                  </div>
                  <div className="grid shrink-0 grid-cols-2 gap-x-5 gap-y-2 text-xs sm:grid-cols-4">
                    <span>
                      <span className="block text-[10px] text-[#918d86]">供给 / 需求</span>
                      <span className="mt-0.5 flex items-center gap-1 font-semibold text-[#4f4b46]">
                        <UsersRound size={12} />
                        {metric?.supplyUsers ?? 0} / {metric?.demandUsers ?? 0}
                      </span>
                    </span>
                    <span>
                      <span className="block text-[10px] text-[#918d86]">成组率</span>
                      <span className="mt-0.5 block font-semibold text-[#4f4b46]">
                        {formatPercent(metric?.groupFormationRate ?? 0)}
                      </span>
                    </span>
                    <span>
                      <span className="block text-[10px] text-[#918d86]">见面完成率</span>
                      <span className="mt-0.5 flex items-center gap-1 font-semibold text-[#4f4b46]">
                        <CalendarCheck2 size={12} />
                        {formatPercent(metric?.meetupCompletionRate ?? 0)}
                      </span>
                    </span>
                    <button
                      aria-label={`配置${theme.title}`}
                      className="inline-flex h-9 w-9 self-center justify-self-end items-center justify-center rounded-md border border-[#d8d4cd] text-[#6f6b65] hover:bg-[#f0eeea]"
                      onClick={() => (editing ? setEditingKey(null) : startEditing(theme))}
                      title="配置主题"
                      type="button"
                    >
                      {editing ? <X size={14} /> : <Settings2 size={14} />}
                    </button>
                  </div>
                </div>

                {editing ? (
                  <div className="border-t border-[#e8e5e0] bg-[#faf9f7] p-4">
                    <div className="grid gap-3 sm:grid-cols-2">
                      <label className="text-xs font-semibold text-[#5f5b55]">
                        主题名称
                        <input
                          className="mt-1.5 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 font-normal outline-none focus:border-[#527b6d]"
                          maxLength={48}
                          onChange={(event) => setDraft({ ...draft, title: event.target.value })}
                          value={draft.title}
                        />
                      </label>
                      <label className="text-xs font-semibold text-[#5f5b55]">
                        时间提示
                        <input
                          className="mt-1.5 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 font-normal outline-none focus:border-[#527b6d]"
                          maxLength={80}
                          onChange={(event) =>
                            setDraft({ ...draft, scheduleHint: event.target.value })
                          }
                          value={draft.scheduleHint}
                        />
                      </label>
                      <label className="text-xs font-semibold text-[#5f5b55] sm:col-span-2">
                        主题说明
                        <input
                          className="mt-1.5 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 font-normal outline-none focus:border-[#527b6d]"
                          maxLength={240}
                          onChange={(event) =>
                            setDraft({ ...draft, description: event.target.value })
                          }
                          value={draft.description}
                        />
                      </label>
                      <label className="text-xs font-semibold text-[#5f5b55] sm:col-span-2">
                        默认意图
                        <textarea
                          className="mt-1.5 min-h-20 w-full resize-y rounded-md border border-[#d8d4cd] bg-white p-3 font-normal leading-5 outline-none focus:border-[#527b6d]"
                          maxLength={500}
                          onChange={(event) =>
                            setDraft({ ...draft, defaultPrompt: event.target.value })
                          }
                          value={draft.defaultPrompt}
                        />
                      </label>
                    </div>
                    <div className="mt-3 grid gap-3 sm:grid-cols-[minmax(0,1fr)_120px_minmax(180px,auto)] sm:items-end">
                      <label className="text-xs font-semibold text-[#5f5b55]">
                        图标
                        <select
                          className="mt-1.5 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 font-normal outline-none focus:border-[#527b6d]"
                          onChange={(event) => setDraft({ ...draft, iconName: event.target.value })}
                          value={draft.iconName}
                        >
                          {iconOptions.map(([value, label]) => (
                            <option key={value} value={value}>
                              {label}
                            </option>
                          ))}
                        </select>
                      </label>
                      <label className="text-xs font-semibold text-[#5f5b55]">
                        排序
                        <input
                          className="mt-1.5 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 font-normal outline-none focus:border-[#527b6d]"
                          max={1000}
                          min={0}
                          onChange={(event) =>
                            setDraft({ ...draft, displayOrder: Number(event.target.value) })
                          }
                          type="number"
                          value={draft.displayOrder}
                        />
                      </label>
                      <EngagementToggle
                        checked={draft.isActive}
                        label="对用户开放"
                        onChange={(checked) => setDraft({ ...draft, isActive: checked })}
                      />
                    </div>
                    <div className="mt-3 flex justify-end">
                      <button
                        className="inline-flex h-10 items-center gap-2 rounded-md bg-[#315c4d] px-4 text-xs font-semibold text-white hover:bg-[#284f42] disabled:opacity-50"
                        disabled={saving}
                        onClick={() => void save()}
                        type="button"
                      >
                        {saving ? (
                          <LoaderCircle className="animate-spin" size={14} />
                        ) : (
                          <Save size={14} />
                        )}
                        {saving ? '保存中...' : '保存主题'}
                      </button>
                    </div>
                  </div>
                ) : null}
              </article>
            )
          })}
          {themes.length === 0 ? (
            <div className="px-5 py-10 text-center text-sm text-[#817d76]">暂无校园主题</div>
          ) : null}
        </div>
      )}
    </section>
  )
}
