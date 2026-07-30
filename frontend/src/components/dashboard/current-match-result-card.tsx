import { Mail, MessageCircle, Sparkles } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import type { MatchResult } from '@/lib/match-api'
import { getFallbackMatchReasonText, getStructuredMatchReason } from '@/lib/match-reason'

interface CurrentMatchResultCardProps {
  match: MatchResult | null
  stateTitle?: string
  stateMessage: string
  isContacting?: boolean
  onContact?: (matchId: number) => void
}

function SectionTitle({ children }: { children: string }) {
  return <p className="text-[0.98rem] font-semibold text-[#4a3429]">{children}</p>
}

export function CurrentMatchResultCard({
  match,
  isContacting = false,
  onContact,
  stateTitle = '结果同步中',
  stateMessage,
}: CurrentMatchResultCardProps) {
  if (!match) {
    return (
      <Card className="mx-auto max-w-[920px] overflow-hidden rounded-[32px] border-[#e9ddd3] bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.95),rgba(252,247,242,0.98)_52%,rgba(249,241,233,0.98)_100%)] shadow-[0_18px_46px_rgba(84,57,40,0.08)]">
        <CardContent className="space-y-5 px-5 py-7 sm:px-8 sm:py-8">
          <span className="inline-flex rounded-full border border-[#dbc6bc] bg-[#fff7f1] px-3.5 py-1.5 text-sm font-semibold text-[#9e4f40]">
            本轮结果
          </span>
          <div className="rounded-[28px] border border-[#eadfd6] bg-[#fffaf6] px-6 py-8 text-center">
            <p className="font-serif text-[2rem] tracking-[-0.04em] text-[#2f221b] sm:text-[2.3rem]">
              {stateTitle}
            </p>
            <p className="mt-3 text-[1rem] leading-8 text-[#6f6056]">{stateMessage}</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const matched = match.resultStatus === 'matched'
  const structuredReason = getStructuredMatchReason(match)
  const fallbackReasonText = getFallbackMatchReasonText(match)

  return (
    <Card className="mx-auto max-w-[920px] overflow-hidden rounded-[32px] border-[#e9ddd3] bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.95),rgba(252,247,242,0.98)_52%,rgba(249,241,233,0.98)_100%)] shadow-[0_18px_46px_rgba(84,57,40,0.08)]">
      <CardContent className="space-y-6 px-5 py-6 sm:px-8 sm:py-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <span className="inline-flex rounded-full border border-[#dbc6bc] bg-[#fff7f1] px-3.5 py-1.5 text-sm font-semibold text-[#9e4f40]">
            {match.cycleTitle}
          </span>
          <span
            className={`inline-flex rounded-full px-3.5 py-1.5 text-sm font-semibold ${
              matched ? 'bg-[#edf6ee] text-[#5d8b67]' : 'bg-[#f5ece4] text-[#8d725d]'
            }`}
          >
            {matched ? '已配对' : '未匹配'}
          </span>
        </div>

        <div className="rounded-[28px] border border-[#eadfd6] bg-[#fffaf6] px-6 py-6">
          <p className="text-sm tracking-[0.16em] text-[#9a6d4d] uppercase">
            {matched ? 'Match Result' : 'Status'}
          </p>
          <h3 className="mt-3 font-serif text-[2.1rem] leading-none tracking-[-0.04em] text-[#261814] sm:text-[2.5rem]">
            {matched ? match.partner?.nickname || '匿名同学' : '本轮未匹配成功'}
          </h3>
          <p className="mt-3 text-base leading-8 text-[#6e5d52]">{match.message}</p>
          {matched && typeof match.score === 'number' ? (
            <div className="mt-5 space-y-3">
              <p className="text-sm tracking-[0.16em] text-[#9a6d4d] uppercase">Match Score</p>
              <div className="h-2 overflow-hidden rounded-full bg-[#efe8e2]">
                <div
                  className="h-full rounded-full bg-[linear-gradient(90deg,#8f3f3c,#a25248)]"
                  style={{ width: `${Math.max(0, Math.min(100, match.score))}%` }}
                />
              </div>
            </div>
          ) : null}
          <div className="mt-4 flex flex-wrap gap-3 text-sm text-[#8d7b70]">
            {matched && match.partner ? (
              <span>
                {[match.partner.schoolName, match.partner.department, match.partner.grade].filter(Boolean).join(' · ') || '资料待补充'}
              </span>
            ) : null}
          </div>
        </div>

        {matched && match.partner?.userId && onContact ? (
          <button
            className="inline-flex h-11 items-center justify-center gap-2 rounded-full bg-[#8f3f3c] px-5 text-sm font-semibold text-white transition-colors hover:bg-[#7c3532] disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isContacting}
            onClick={() => onContact(match.matchId)}
            type="button"
          >
            <MessageCircle size={17} />
            {isContacting ? '处理中...' : '进入聊天'}
          </button>
        ) : null}

        {matched && (structuredReason || fallbackReasonText) ? (
          <div className="space-y-3">
            <SectionTitle>匹配说明</SectionTitle>
            <div className="space-y-4 rounded-[24px] border border-[#ebe1d8] bg-[#faf5ef] px-5 py-5 text-[0.98rem] leading-8 text-[#68574d]">
              {structuredReason ? (
                <>
                  {structuredReason.match_reasons?.length ? (
                    <div className="space-y-2">
                      <p className="text-sm font-semibold text-[#4a3429]">匹配原因</p>
                      <ul className="space-y-2">
                        {structuredReason.match_reasons.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                  ) : null}

                  {structuredReason.date_topics?.length ? (
                    <div className="space-y-2">
                      <p className="text-sm font-semibold text-[#4a3429]">推荐话题</p>
                      <ul className="space-y-2">
                        {structuredReason.date_topics.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                  ) : null}

                  {structuredReason.cautions?.length ? (
                    <div className="space-y-2">
                      <p className="text-sm font-semibold text-[#4a3429]">需要留意</p>
                      <ul className="space-y-2">
                        {structuredReason.cautions.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                  ) : null}
                </>
              ) : (
                <p>{fallbackReasonText}</p>
              )}
            </div>
          </div>
        ) : null}

        {matched && match.partner?.contacts.length ? (
          <div className="space-y-3">
            <SectionTitle>联系方式</SectionTitle>
            <div className="grid gap-3">
              {match.partner.contacts.map((contact) => (
                <div
                  className="flex items-center gap-3 rounded-[22px] border border-[#dfc8bf] bg-[#9a403d] px-5 py-4 text-white"
                  key={`${contact.type}-${contact.value}`}
                >
                  <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/14">
                    <Mail size={18} />
                  </span>
                  <div className="min-w-0">
                    <p className="text-xs tracking-[0.14em] text-white/78">{contact.label}</p>
                    <p className="truncate text-base font-semibold">{contact.value}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : null}

        {!matched ? (
          <div className="flex items-center gap-3 rounded-[24px] border border-[#ead9d1] bg-[#fff8f5] px-5 py-4 text-[#7a6257]">
            <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-[#f4e2d7] text-[#b35d50]">
              <Sparkles size={18} />
            </span>
            <p className="text-[0.98rem] leading-7">
              这并不代表连接的结束。继续完善资料并参与下一轮，也许下一次就会遇见合适的人。
            </p>
          </div>
        ) : null}
      </CardContent>
    </Card>
  )
}
