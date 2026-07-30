import { Mail } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import type { WeeklyMatchMockResult } from '@/types/dashboard'

interface WeeklyMatchResultCardProps {
  result: WeeklyMatchMockResult
}

function SectionTitle({ children }: { children: string }) {
  return <p className="text-[0.98rem] font-semibold text-[#4a3429]">{children}</p>
}

export function WeeklyMatchResultCard({ result }: WeeklyMatchResultCardProps) {
  return (
    <Card className="mx-auto max-w-[920px] overflow-hidden rounded-[32px] border-[#e9ddd3] bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.95),rgba(252,247,242,0.98)_52%,rgba(249,241,233,0.98)_100%)] shadow-[0_18px_46px_rgba(84,57,40,0.08)]">
      <CardContent className="space-y-6 px-5 py-6 sm:px-8 sm:py-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <span className="inline-flex rounded-full border border-[#dbc6bc] bg-[#fff7f1] px-3.5 py-1.5 text-sm font-semibold text-[#9e4f40]">
            {result.cycleLabel}
          </span>
        </div>

        <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center gap-4">
            <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-[linear-gradient(180deg,#fff7e8,#fde8d6)] text-[2rem] shadow-[0_10px_24px_rgba(94,66,45,0.08)]">
              {result.avatarEmoji}
            </div>
            <div className="space-y-1.5">
              <h3 className="font-serif text-[1.8rem] leading-none tracking-[-0.04em] text-[#261814] sm:text-[2rem]">
                {result.name}
              </h3>
              <p className="text-base text-[#6e5d52]">
                {result.department} · {result.grade}
              </p>
            </div>
          </div>

          <div className="shrink-0 text-left lg:text-right">
            <p className="font-serif text-[2.6rem] leading-none tracking-[-0.05em] text-[#9d463f] sm:text-[3.2rem]">
              {result.score.toFixed(2)}%
            </p>
            <p className="mt-2 text-sm tracking-[0.16em] text-[#a08a7e]">匹配度</p>
          </div>
        </div>

        <div className="h-2 overflow-hidden rounded-full bg-[#efe8e2]">
          <div
            className="h-full rounded-full bg-[linear-gradient(90deg,#8f3f3c,#a25248)]"
            style={{ width: `${result.score}%` }}
          />
        </div>

        <div className="space-y-3">
          <SectionTitle>匹配原因：</SectionTitle>
          <ul className="space-y-3 text-[0.98rem] leading-8 text-[#68574d]">
            {result.matchReasons.map((reason) => (
              <li className="flex gap-3" key={reason}>
                <span className="pt-1.5 text-[#9d463f]">•</span>
                <span>{reason}</span>
              </li>
            ))}
          </ul>
        </div>

        <div className="space-y-3">
          <SectionTitle>话题推荐：</SectionTitle>
          <div className="flex flex-wrap gap-2.5">
            {result.suggestedTopics.map((topic) => (
              <span
                className="rounded-full border border-[#e4cfc5] bg-[#fff6f1] px-3.5 py-1.5 text-sm text-[#a05447]"
                key={topic}
              >
                {topic}
              </span>
            ))}
          </div>
        </div>

        <div className="space-y-3">
          <SectionTitle>需要留意：</SectionTitle>
          <div className="rounded-[24px] border border-[#ead9d1] bg-[#fff8f5] px-4 py-4">
            <ul className="space-y-3 text-[0.95rem] leading-7 text-[#7a6257]">
              {result.cautions.map((item) => (
                <li className="flex gap-3" key={item}>
                  <span className="pt-1.5 text-[#b35d50]">•</span>
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="flex items-center gap-3 rounded-[24px] border border-[#dfc8bf] bg-[#9a403d] px-5 py-4 text-white">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-white/14">
            <Mail size={18} />
          </span>
          <div className="min-w-0">
            <p className="text-xs tracking-[0.14em] text-white/78">{result.contactLabel}</p>
            <p className="truncate text-base font-semibold">{result.contactValue}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
