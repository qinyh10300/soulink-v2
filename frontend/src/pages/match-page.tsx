import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { CurrentMatchResultCard } from '@/components/dashboard/current-match-result-card'
import { DashboardFooter } from '@/components/dashboard/dashboard-footer'
import { MatchDetailModal } from '@/components/dashboard/match-detail-modal'
import { MatchHistoryCard } from '@/components/dashboard/match-history-card'
import {
  getGlobalMatchResultRevealAt,
  hasReachedGlobalMatchRevealAt,
} from '@/config/match-reveal'
import { ApiError, getErrorMessage } from '@/lib/api'
import {
  contactMatchedPartner,
  fetchCurrentMatch,
  fetchMatchCycleStatus,
  fetchMatchDetail,
  fetchMatchHistory,
  type MatchCycleStatus,
  type MatchResult,
} from '@/lib/match-api'
import { dashboardContent } from '@/mock/dashboard'
import type { MatchHistoryListItem } from '@/types/dashboard'

function buildHistoryItems(items: MatchResult[]): MatchHistoryListItem[] {
  return items.map((item) => ({
    id: item.matchId,
    title: item.cycleTitle,
    subtitle:
      item.resultStatus === 'matched'
        ? `匹配对象：${item.partner?.nickname || '匿名同学'}`
        : '本轮未匹配成功',
    message: item.message,
    statusLabel: item.resultStatus === 'matched' ? '已配对' : '未匹配',
    statusTone: item.resultStatus === 'matched' ? 'success' : 'neutral',
    scoreLabel:
      typeof item.score === 'number' ? `分数：${Math.round(item.score)} / 100` : undefined,
    canContact: item.resultStatus === 'matched' && Boolean(item.partner?.userId),
  }))
}

export function MatchPage() {
  const navigate = useNavigate()
  const [cycleStatus, setCycleStatus] = useState<MatchCycleStatus | null>(null)
  const [currentMatch, setCurrentMatch] = useState<MatchResult | null>(null)
  const [historyItems, setHistoryItems] = useState<MatchResult[]>([])
  const [isMatchesLoading, setIsMatchesLoading] = useState(true)
  const [matchesNotice, setMatchesNotice] = useState('')
  const [toastNotice, setToastNotice] = useState('')
  const [contactingMatchId, setContactingMatchId] = useState<number | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailMatch, setDetailMatch] = useState<MatchResult | null>(null)
  const [isGlobalRevealOpen, setIsGlobalRevealOpen] = useState(() =>
    hasReachedGlobalMatchRevealAt(),
  )

  const globalRevealAt = getGlobalMatchResultRevealAt(cycleStatus?.revealAt)
  const canRequestLatestMatchResult = isGlobalRevealOpen
  const visibleHistoryItems = historyItems
  const latestVisibleMatch = useMemo(
    () => {
      if (canRequestLatestMatchResult && cycleStatus?.hasUserResult) {
        return currentMatch
      }

      if (visibleHistoryItems.length > 0) {
        return visibleHistoryItems[0]
      }

      return null
    },
    [canRequestLatestMatchResult, currentMatch, cycleStatus?.hasUserResult, visibleHistoryItems],
  )
  const historyViewItems = useMemo(
    () => buildHistoryItems(visibleHistoryItems),
    [visibleHistoryItems],
  )

  useEffect(() => {
    setIsGlobalRevealOpen(hasReachedGlobalMatchRevealAt(cycleStatus?.revealAt))

    if (!globalRevealAt) {
      return
    }

    const delay = new Date(globalRevealAt).getTime() - Date.now()
    if (delay <= 0) {
      setIsGlobalRevealOpen(true)
      return
    }

    const timer = window.setTimeout(() => {
      setIsGlobalRevealOpen(true)
    }, delay + 500)

    return () => window.clearTimeout(timer)
  }, [cycleStatus?.revealAt, globalRevealAt])

  useEffect(() => {
    if (!cycleStatus?.hasUserResult || currentMatch || !globalRevealAt || !canRequestLatestMatchResult) {
      return
    }

    let cancelled = false

    const loadCurrentMatch = async () => {
      try {
        const current = await fetchCurrentMatch()
        if (!cancelled) {
          setCurrentMatch(current)
        }
      } catch (error) {
        if (cancelled) {
          return
        }

        if (error instanceof ApiError && error.status === 404) {
          setCurrentMatch(null)
          return
        }

        setMatchesNotice(getErrorMessage(error, '本轮结果加载失败，请稍后重试'))
      }
    }

    void loadCurrentMatch()

    return () => {
      cancelled = true
    }
  }, [canRequestLatestMatchResult, currentMatch, cycleStatus?.hasUserResult, globalRevealAt])

  useEffect(() => {
    let cancelled = false

    const loadMatchPage = async () => {
      setIsMatchesLoading(true)
      setMatchesNotice('')

      const [cycleResult, historyResult] = await Promise.all([
        fetchMatchCycleStatus().catch((error) => error),
        fetchMatchHistory(100).catch((error) => error),
      ])

      if (cancelled) {
        return
      }

      let nextCycleStatus: MatchCycleStatus | null = null
      let nextNotice = ''

      if (cycleResult instanceof Error) {
        nextNotice = getErrorMessage(cycleResult, '配对状态加载失败，请稍后重试')
      } else {
        nextCycleStatus = cycleResult
        setCycleStatus(cycleResult)
      }

      if (historyResult instanceof Error) {
        nextNotice ||= getErrorMessage(historyResult, '配对记录加载失败，请稍后重试')
        setHistoryItems([])
      } else {
        setHistoryItems(historyResult.items)
      }

      if (nextCycleStatus?.hasUserResult && hasReachedGlobalMatchRevealAt(nextCycleStatus.revealAt)) {
        try {
          const current = await fetchCurrentMatch()

          if (cancelled) {
            return
          }

          setCurrentMatch(current)
        } catch (error) {
          if (cancelled) {
            return
          }

          if (error instanceof ApiError && error.status === 404) {
            setCurrentMatch(null)
          } else {
            nextNotice ||= getErrorMessage(error, '本轮结果加载失败，请稍后重试')
          }
        }
      } else {
        setCurrentMatch(null)
      }

      setMatchesNotice(nextNotice)
      setIsMatchesLoading(false)
    }

    void loadMatchPage()

    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!toastNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setToastNotice('')
    }, 3000)

    return () => window.clearTimeout(timer)
  }, [toastNotice])

  const openMatchDetail = async (matchId: number) => {
    setDetailOpen(true)
    setDetailLoading(true)

    try {
      const detail = await fetchMatchDetail(matchId)
      setDetailMatch(detail)
    } catch (error) {
      setDetailMatch(null)
      setMatchesNotice(getErrorMessage(error, '配对详情加载失败，请稍后重试'))
    } finally {
      setDetailLoading(false)
    }
  }

  const contactMatchPartner = async (matchId: number) => {
    if (contactingMatchId !== null) {
      return
    }

    setContactingMatchId(matchId)
    setMatchesNotice('')
    try {
      const result = await contactMatchedPartner(matchId)
      if (result.action === 'chat_ready' && typeof result.conversationId === 'number') {
        navigate(`/chat/${result.conversationId}`)
        return
      }

      if (result.action === 'request_sent') {
        setToastNotice(result.message || '好友申请已发送，请等待对方通过后进入聊天。')
        return
      }

      setMatchesNotice('暂时无法进入聊天，请稍后再试。')
    } catch (error) {
      setMatchesNotice(getErrorMessage(error, '无法打开聊天，请稍后再试'))
    } finally {
      setContactingMatchId(null)
    }
  }

  return (
    <div className="min-h-screen bg-[#f6efe8]">
      {toastNotice ? (
        <div className="pointer-events-none fixed inset-x-0 top-6 z-[70] flex justify-center px-4">
          <div className="rounded-full border border-[#e4d5c9] bg-[#fff8f2] px-5 py-2.5 text-sm text-[#7b5f4f] shadow-[0_10px_24px_rgba(71,49,35,0.16)]">
            {toastNotice}
          </div>
        </div>
      ) : null}

      <main className="px-6 pb-0 pt-10 lg:px-8 lg:pt-14">
        <div className="mx-auto w-full max-w-4xl space-y-16">
          <section className="space-y-6" id="current-match">
            <div className="space-y-2">
              <h2 className="font-serif text-[2rem] leading-[1.08] tracking-[-0.03em] text-[#2f221b] sm:text-[2.2rem]">本周匹配</h2>
              <p className="text-[1.02rem] leading-8 text-[#7c6a5f]">
                这里会在统一揭晓时间后展示最近一次匹配结果。
              </p>
            </div>

            <CurrentMatchResultCard
              isContacting={Boolean(latestVisibleMatch && contactingMatchId === latestVisibleMatch.matchId)}
              match={latestVisibleMatch}
              onContact={(matchId) => {
                void contactMatchPartner(matchId)
              }}
              stateTitle={
                !canRequestLatestMatchResult
                  ? '等待揭晓'
                  : cycleStatus?.hasUserResult && !latestVisibleMatch
                    ? '结果同步中'
                    : '暂无结果'
              }
              stateMessage={
                !canRequestLatestMatchResult
                  ? '尚未到统一揭晓时间，结果开放后会在这里展示。'
                  : cycleStatus?.hasUserResult && !latestVisibleMatch
                    ? '最近一次结果正在同步中，请稍后刷新查看。'
                    : '你还没有已揭晓的匹配结果，参与下一轮后会在这里展示。'
              }
            />

            {matchesNotice ? (
              <p className="text-sm text-[#a35b43]">{matchesNotice}</p>
            ) : null}
          </section>

          <section className="space-y-6" id="match-history">
            <div className="space-y-2">
              <h2 className="font-serif text-[2rem] leading-[1.08] tracking-[-0.03em] text-[#2f221b] sm:text-[2.2rem]">配对历史</h2>
              <p className="text-[1.02rem] leading-8 text-[#7c6a5f]">
                历史周期会在统一揭晓时间后按时间倒序展示。
              </p>
            </div>

            <MatchHistoryCard
              contactingMatchId={contactingMatchId}
              emptyState={dashboardContent.history}
              isLoading={isMatchesLoading}
              items={historyViewItems}
              onContact={(matchId) => {
                void contactMatchPartner(matchId)
              }}
              onSelect={(matchId) => {
                void openMatchDetail(matchId)
              }}
            />
          </section>
        </div>
      </main>

      <DashboardFooter />

      <MatchDetailModal
        isLoading={detailLoading}
        match={detailMatch}
        onClose={() => {
          setDetailOpen(false)
          setDetailMatch(null)
        }}
        open={detailOpen}
      />
    </div>
  )
}
