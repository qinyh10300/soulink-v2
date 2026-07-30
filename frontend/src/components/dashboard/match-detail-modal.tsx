import { useEffect } from 'react'
import { X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { getFallbackMatchReasonText, getStructuredMatchReason } from '@/lib/match-reason'
import { cn } from '@/lib/utils'
import type { MatchResult } from '@/lib/match-api'

interface MatchDetailModalProps {
  isLoading?: boolean
  match: MatchResult | null
  onClose: () => void
  open: boolean
}

export function MatchDetailModal({
  isLoading = false,
  match,
  onClose,
  open,
}: MatchDetailModalProps) {
  useEffect(() => {
    if (!open) {
      return
    }

    const previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      document.body.style.overflow = previousBodyOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [onClose, open])

  if (!open) {
    return null
  }

  const structuredReason = match ? getStructuredMatchReason(match) : null
  const fallbackReasonText = match ? getFallbackMatchReasonText(match) : ''

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[calc(100vh-2rem)] w-full max-w-2xl flex-col overflow-hidden rounded-[30px] border border-[#eadfd5] bg-[#fcf8f3] p-8 shadow-[0_26px_72px_rgba(61,40,29,0.18)]">
        <button
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#e8ddd3] bg-white text-[#7d6556]"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="min-h-0 flex-1 overflow-y-auto pr-1">
          {isLoading ? (
            <div className="py-12 text-center text-[#7d6556]">正在加载详情...</div>
          ) : !match ? (
            <div className="py-12 text-center text-[#7d6556]">暂无可查看的配对详情。</div>
          ) : (
            <div className="space-y-6">
              <div className="space-y-2">
                <p className="text-sm tracking-[0.2em] text-[#9a6d4d] uppercase">
                  {match.cycleTitle}
                </p>
                <h3 className="font-serif text-4xl text-[#2f221b]">
                  {match.resultStatus === 'matched' ? '本轮配对结果' : '本轮状态'}
                </h3>
                <p className="text-base leading-7 text-[#6f6056]">{match.message}</p>
              </div>

              {match.resultStatus === 'matched' && typeof match.score === 'number' ? (
                <div className="space-y-4 rounded-[24px] border border-[#ebe1d8] bg-[#faf5ef] p-6">
                  <div className="flex items-end justify-between gap-4">
                    <div>
                      <p className="text-sm tracking-[0.16em] text-[#9a6d4d] uppercase">Match Score</p>
                      <p className="mt-2 text-sm text-[#8b7a6f]">匹配度</p>
                    </div>
                    <p className="font-serif text-[2.4rem] leading-none tracking-[-0.05em] text-[#9d463f] sm:text-[2.8rem]">
                      {Math.round(match.score)} / 100
                    </p>
                  </div>
                  <div className="h-2 overflow-hidden rounded-full bg-[#efe8e2]">
                    <div
                      className="h-full rounded-full bg-[linear-gradient(90deg,#8f3f3c,#a25248)]"
                      style={{ width: `${Math.max(0, Math.min(100, match.score))}%` }}
                    />
                  </div>
                </div>
              ) : null}

              {structuredReason || fallbackReasonText ? (
                <div className="space-y-2">
                  <h4 className="text-lg font-medium text-[#2f221b]">匹配说明</h4>
                  <div className="space-y-4 rounded-[24px] border border-[#ebe1d8] bg-[#faf5ef] px-6 py-5 text-base leading-8 text-[#6f6056]">
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

              {match.partner ? (
                <div className="space-y-4">
                  <div className="space-y-1">
                    <h4 className="text-lg font-medium text-[#2f221b]">对方信息</h4>
                    <p className="text-base text-[#6f6056]">
                      {match.partner.nickname || '匿名同学'}
                      {match.partner.schoolName ? ` · ${match.partner.schoolName}` : ''}
                      {match.partner.department ? ` · ${match.partner.department}` : ''}
                      {match.partner.grade ? ` · ${match.partner.grade}` : ''}
                    </p>
                  </div>

                  <div className="grid gap-3">
                    {match.partner.contacts.length > 0 ? (
                      match.partner.contacts.map((contact) => (
                        <div
                          className="rounded-[22px] border border-[#ebe1d8] bg-[#fffaf6] px-5 py-4"
                          key={`${contact.type}-${contact.value}`}
                        >
                          <p className="text-sm text-[#8b7a6f]">{contact.label}</p>
                          <p className="mt-1 text-base text-[#2f221b]">{contact.value}</p>
                        </div>
                      ))
                    ) : (
                      <div className="rounded-[22px] border border-[#ebe1d8] bg-[#fffaf6] px-5 py-4 text-[#6f6056]">
                        对方暂未公开联系方式。
                      </div>
                    )}
                  </div>
                </div>
              ) : null}
            </div>
          )}
        </div>

        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'mt-8 w-full rounded-full bg-[#8f6548] text-white hover:bg-[#7f5a40]',
          )}
          onClick={onClose}
          type="button"
        >
          我知道了
        </button>
      </div>
    </div>
  )
}
