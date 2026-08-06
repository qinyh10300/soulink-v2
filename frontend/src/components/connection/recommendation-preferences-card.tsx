import {
  AlertCircle,
  BrainCircuit,
  LoaderCircle,
  RotateCcw,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { intentCategoryMeta } from '@/components/connection/intent-category'
import { getErrorMessage } from '@/lib/api'
import {
  fetchRecommendationPreferences,
  resetRecommendationPreferences,
  updateRecommendationPreferences,
  type RecommendationPreferences,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'

const signalLabels: Record<string, string> = {
  saved: '收藏过',
  interested: '表达过兴趣',
  withdrawn: '撤回过回应',
  dismissed: '忽略过',
  accepted_candidate: '接受过连接',
  declined_candidate: '谢绝过连接',
  meet_again: '愿意再次见面',
  do_not_repeat: '不希望再次推荐',
  positive_meetup: '见面体验积极',
  no_show: '发生过爽约',
}

export function RecommendationPreferencesCard() {
  const [preferences, setPreferences] =
    useState<RecommendationPreferences | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [resetConfirming, setResetConfirming] = useState(false)
  const [notice, setNotice] = useState('')

  const load = async () => {
    setLoading(true)
    setNotice('')
    try {
      setPreferences(await fetchRecommendationPreferences())
    } catch (error) {
      setNotice(getErrorMessage(error, '推荐偏好暂时无法加载'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  const toggleLearning = async () => {
    if (!preferences || saving) {
      return
    }
    setSaving(true)
    setNotice('')
    try {
      setPreferences(
        await updateRecommendationPreferences(!preferences.learningEnabled),
      )
    } catch (error) {
      setNotice(getErrorMessage(error, '推荐学习开关暂时无法更新'))
    } finally {
      setSaving(false)
    }
  }

  const resetLearning = async () => {
    if (saving) {
      return
    }
    setSaving(true)
    setNotice('')
    try {
      setPreferences(await resetRecommendationPreferences())
      setResetConfirming(false)
    } catch (error) {
      setNotice(getErrorMessage(error, '推荐偏好暂时无法重置'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-2.5">
          <BrainCircuit className="mt-0.5 shrink-0 text-[#4f7064]" size={17} />
          <div>
            <h2 className="text-sm font-semibold text-[#37332f]">推荐学习</h2>
            <p className="mt-1 text-[11px] leading-5 text-[#77736c]">
              只使用你的兴趣、忽略和私密见后反馈，可随时关闭或清空。
            </p>
          </div>
        </div>
        {preferences ? (
          <button
            aria-label={preferences.learningEnabled ? '关闭推荐学习' : '开启推荐学习'}
            aria-pressed={preferences.learningEnabled}
            className={cn(
              'relative h-6 w-11 shrink-0 rounded-full transition-colors disabled:opacity-50',
              preferences.learningEnabled ? 'bg-[#497363]' : 'bg-[#c9c5bf]',
            )}
            disabled={saving}
            onClick={() => void toggleLearning()}
            title={preferences.learningEnabled ? '关闭推荐学习' : '开启推荐学习'}
            type="button"
          >
            <span
              className={cn(
                'absolute top-1 h-4 w-4 rounded-full bg-white transition-transform',
                preferences.learningEnabled ? 'translate-x-6' : 'translate-x-1',
              )}
            />
          </button>
        ) : null}
      </div>

      {loading ? (
        <div className="mt-5 flex items-center gap-2 text-xs text-[#817d76]">
          <LoaderCircle className="animate-spin" size={14} />
          正在读取学习记录...
        </div>
      ) : notice && !preferences ? (
        <button
          className="mt-4 flex w-full items-start gap-2 text-left text-xs leading-5 text-[#8c463d]"
          onClick={() => void load()}
          type="button"
        >
          <AlertCircle className="mt-0.5 shrink-0" size={14} />
          {notice}，点击重试
        </button>
      ) : preferences ? (
        <>
          <div className="mt-5 border-t border-[#e7e4df] pt-4">
            <div className="flex items-center justify-between gap-3">
              <p className="text-[11px] font-semibold text-[#5f5b55]">目前学到的偏好</p>
              <span className="text-[10px] text-[#8a867f]">
                {preferences.recentSignals.length} 条有效信号
              </span>
            </div>
            {preferences.affinities.length > 0 ? (
              <div className="mt-3 space-y-2.5">
                {preferences.affinities.slice(0, 4).map((affinity) => (
                  <div
                    className="flex items-center justify-between gap-3 text-[11px]"
                    key={affinity.category}
                  >
                    <span className="font-medium text-[#55514c]">
                      {intentCategoryMeta[affinity.category].label}
                    </span>
                    <span
                      className={cn(
                        'text-right',
                        affinity.weight > 0
                          ? 'text-[#3f705e]'
                          : affinity.weight < 0
                            ? 'text-[#946052]'
                            : 'text-[#817d76]',
                      )}
                    >
                      {affinity.summary} · {affinity.signals}
                    </span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="mt-3 text-xs leading-5 text-[#817d76]">
                暂无学习记录，推荐只使用当前意图的明确条件。
              </p>
            )}
          </div>

          {preferences.recentSignals.length > 0 ? (
            <div className="mt-4 border-t border-[#e7e4df] pt-4">
              <p className="text-[11px] font-semibold text-[#5f5b55]">最近依据</p>
              <div className="mt-2 space-y-1.5">
                {preferences.recentSignals.slice(0, 3).map((signal, index) => (
                  <p
                    className="flex items-center justify-between gap-3 text-[10px] leading-4 text-[#77736c]"
                    key={`${signal.type}-${signal.occurredAt}-${index}`}
                  >
                    <span className="truncate">
                      {signalLabels[signal.type] || signal.type} ·{' '}
                      {intentCategoryMeta[signal.category].label}
                    </span>
                    <span className={signal.impact >= 0 ? 'text-[#477062]' : 'text-[#946052]'}>
                      {signal.impact >= 0 ? '+' : ''}
                      {signal.impact}
                    </span>
                  </p>
                ))}
              </div>
            </div>
          ) : null}

          {notice ? (
            <p className="mt-3 flex items-start gap-1.5 text-[11px] leading-5 text-[#8c463d]">
              <AlertCircle className="mt-0.5 shrink-0" size={12} />
              {notice}
            </p>
          ) : null}

          <div className="mt-4 border-t border-[#e7e4df] pt-4">
            {resetConfirming ? (
              <div className="space-y-3">
                <p className="text-[11px] leading-5 text-[#75584e]">
                  清空后无法恢复；收藏和聊天本身不会被删除。
                </p>
                <div className="flex gap-2">
                  <button
                    className="h-8 rounded-md px-3 text-[11px] font-medium text-[#6f6b65]"
                    onClick={() => setResetConfirming(false)}
                    type="button"
                  >
                    取消
                  </button>
                  <button
                    className="inline-flex h-8 items-center gap-1.5 rounded-md bg-[#8b5548] px-3 text-[11px] font-semibold text-white disabled:opacity-50"
                    disabled={saving}
                    onClick={() => void resetLearning()}
                    type="button"
                  >
                    {saving ? <LoaderCircle className="animate-spin" size={12} /> : null}
                    确认清空
                  </button>
                </div>
              </div>
            ) : (
              <button
                className="inline-flex h-8 items-center gap-2 text-[11px] font-medium text-[#78594f] hover:text-[#8b463a]"
                disabled={saving || preferences.recentSignals.length === 0}
                onClick={() => setResetConfirming(true)}
                type="button"
              >
                <RotateCcw size={13} />
                重置推荐学习
              </button>
            )}
          </div>
        </>
      ) : null}
    </section>
  )
}
