import { Fragment, type ReactNode } from 'react'
import { Feather, Heart, LoaderCircle, Sparkles, X } from 'lucide-react'

import brandIcon from '@/assets/soulink-logo-icon.png'
import type { QuestionnairePortraitResult } from '@/lib/questionnaire-api'

export type SoulPortraitViewStatus = 'idle' | 'loading' | 'thinking' | 'ready' | 'error'

interface SoulPortraitCardProps {
  errorMessage?: string
  onClose: () => void
  portrait: QuestionnairePortraitResult | null
  status: SoulPortraitViewStatus
  userName: string
}

export function SoulPortraitCard({
  errorMessage = '',
  onClose,
  portrait,
  status,
  userName,
}: SoulPortraitCardProps) {
  return (
    <div className="relative mx-auto w-full max-w-5xl overflow-hidden rounded-[34px] border border-[#eadfd2] bg-[#fffdfb] text-[#3f342d] shadow-[0_34px_90px_rgba(40,25,18,0.22)]">
      <button
        aria-label="Close"
        className="absolute right-5 top-5 z-20 inline-flex h-9 w-9 items-center justify-center rounded-full text-[#8f857d] transition-colors hover:bg-[rgba(83,62,50,0.08)] hover:text-[#4b3c33]"
        onClick={onClose}
        type="button"
      >
        <X size={20} />
      </button>

      <div className="relative max-h-[88vh] overflow-y-auto px-5 py-5 sm:px-7 sm:py-7 lg:px-9 lg:py-9">
        <p className="sr-only">{userName} soul portrait</p>

        {status === 'loading' || status === 'thinking' ? (
          <div className="flex min-h-[560px] items-center justify-center">
            <SectionCard label="SOUL PORTRAIT">
              <div className="flex flex-col items-center justify-center gap-5 py-16 text-center">
                <div className="flex h-16 w-16 items-center justify-center rounded-full bg-[linear-gradient(135deg,#f2e5da,#ece1d8)] text-[#9f7d67] shadow-[0_16px_34px_rgba(121,92,70,0.14)]">
                  <LoaderCircle className="animate-spin" size={30} />
                </div>
                <div className="space-y-3">
                  <h3 className="font-serif text-4xl text-[#342822] sm:text-5xl">正在生成灵魂画像</h3>
                  <p className="mx-auto max-w-2xl font-serif text-[1.08rem] leading-[1.85] text-[#6c5e54] sm:text-[1.18rem]">
                    {status === 'loading'
                      ? 'Soulink 正在读取你的问卷回答，并将这些线索整理成一段更完整的关系画像。'
                      : '画像即将完成，我们正在继续获取最终结果，请再稍等片刻。'}
                  </p>
                </div>
              </div>
            </SectionCard>
          </div>
        ) : status === 'error' && errorMessage ? (
          <div className="space-y-5">
            <SectionCard label="SOUL PORTRAIT">
              <div className="py-4">
                <h3 className="font-serif text-4xl text-[#7d4f43] sm:text-[2.8rem]">灵魂画像暂时不可用</h3>
                <p className="mt-5 font-serif text-[1.08rem] leading-[1.85] text-[#875c4d] sm:text-[1.18rem]">
                  {errorMessage}
                </p>
              </div>
            </SectionCard>
          </div>
        ) : status === 'ready' && portrait ? (
          <div className="space-y-5 sm:space-y-6">
            <section className="rounded-[30px] border border-[#eadfd4] bg-[linear-gradient(180deg,rgba(255,252,248,0.96)_0%,rgba(255,248,242,0.92)_100%)] px-6 py-6 shadow-[inset_0_1px_0_rgba(255,255,255,0.84)] sm:px-7 sm:py-7 lg:px-8 lg:py-8">
              <div className="grid items-start gap-7 lg:grid-cols-[minmax(0,1.02fr)_236px] lg:gap-12">
                <div className="space-y-6 sm:space-y-7">
                  <div className="flex items-center gap-3 text-[#b2917d]">
                    <RomanticBadge size="small">
                      <Sparkles size={14} strokeWidth={1.8} />
                    </RomanticBadge>
                    <p className="text-[0.76rem] font-medium tracking-normal uppercase sm:text-[0.86rem]">
                      SOUL SIGNATURE
                    </p>
                  </div>

                  <div className="max-w-[38rem] border-l border-[#ddd2c8] pl-5 sm:pl-6">
                    <h2 className="font-serif text-[2.35rem] leading-[1.04] tracking-[-0.04em] text-[#23314d] sm:text-[3.38rem]">
                      {portrait.soulSignature}
                    </h2>
                  </div>
                </div>

                <div className="flex items-start justify-center lg:justify-end">
                  <PolaroidVisual />
                </div>
              </div>
            </section>

            <SectionCard
              icon={<Feather size={14} strokeWidth={1.8} />}
              label="PERSONAL PROFILE"
            >
              <div className="flex items-start gap-5 sm:gap-6">
                <div className="mt-1 flex w-[84px] shrink-0 justify-center sm:w-[92px]">
                  <LayeredHeartsIcon />
                </div>
                <div className="min-w-0 space-y-3.5 font-serif text-[1.02rem] leading-[1.88] text-[#6d6057] sm:text-[1.12rem]">
                  {renderParagraphs(portrait.personalProfile)}
                </div>
              </div>
            </SectionCard>

            <SectionCard
              icon={<Heart size={14} strokeWidth={1.8} />}
              label="IDEAL RESONANCE"
            >
              <div className="flex items-start gap-5 sm:gap-6">
                <div className="mt-1 flex w-[84px] shrink-0 justify-center sm:w-[92px]">
                  <StormCloudsIcon />
                </div>
                <div className="min-w-0 space-y-3.5 font-serif text-[1.02rem] leading-[1.88] text-[#6d6057] sm:text-[1.12rem]">
                  {renderParagraphs(portrait.idealResonance)}
                </div>
              </div>
            </SectionCard>
          </div>
        ) : null}
      </div>
    </div>
  )
}

function SectionCard({
  children,
  icon,
  label,
}: {
  children: ReactNode
  icon?: ReactNode
  label: string
}) {
  return (
    <section className="rounded-[30px] border border-[#eadfd4] bg-[rgba(255,250,245,0.72)] px-6 py-6 shadow-[inset_0_1px_0_rgba(255,255,255,0.84)] sm:px-7 sm:py-7 lg:px-8 lg:py-8">
      <div className="mb-4 flex items-center gap-3 text-[#b2917d]">
        {icon ? <RomanticBadge size="small">{icon}</RomanticBadge> : null}
        <p className="text-[0.76rem] font-medium tracking-normal uppercase sm:text-[0.86rem]">{label}</p>
      </div>
      {children}
    </section>
  )
}

function RomanticBadge({
  children,
  className = '',
  size,
}: {
  children: ReactNode
  className?: string
  size: 'small' | 'large'
}) {
  const sizeClasses =
    size === 'large'
      ? 'h-14 w-14 text-[#ab7a72] shadow-[0_14px_28px_rgba(73,52,41,0.08),inset_0_1px_0_rgba(255,255,255,0.92)]'
      : 'h-8 w-8 text-[#b5877a] shadow-[0_8px_18px_rgba(73,52,41,0.06),inset_0_1px_0_rgba(255,255,255,0.92)]'

  return (
    <span
      className={`inline-flex shrink-0 items-center justify-center rounded-full border border-[#eadfd6] bg-[radial-gradient(circle_at_30%_25%,#fffdfc_0%,#fff8f4_58%,#f8eee7_100%)] ${sizeClasses}${className ? ` ${className}` : ''}`}
    >
      {children}
    </span>
  )
}

function LayeredHeartsIcon() {
  return (
    <svg
      aria-hidden="true"
      className="overflow-visible"
      fill="none"
      height="58"
      viewBox="0 0 84 58"
      width="84"
    >
      <path
        d="M27.2 39.5c-9.2-5.8-18.2-12.1-18.2-22.3 0-7.2 5.1-12.6 11.9-12.6 4.9 0 8.6 2.4 11 6.6 2.5-4.2 6.5-6.6 11.3-6.6 7 0 12 5.5 12 12.7 0 10.1-9 16.4-18.5 22.2l-4.7 2.9-4.8-2.9Z"
        fill="#E5A58F"
      />
      <path
        d="M41.2 44.8c-9.6-6-19-12.7-19-23.4 0-7.6 5.4-13.3 12.7-13.3 5.2 0 9.2 2.5 11.7 7 2.6-4.5 6.8-7 11.9-7 7.4 0 12.8 5.8 12.8 13.4 0 10.6-9.4 17.3-19.2 23.3l-4.9 3-6-3Z"
        fill="#8F3B39"
        opacity="0.92"
      />
      <path
        d="M52.8 42.4c-9.7-6.1-19.2-12.7-19.2-23.7 0-7.7 5.4-13.5 12.8-13.5 5.2 0 9.3 2.5 11.8 7.1 2.6-4.6 6.8-7.1 12-7.1 7.5 0 13 5.9 13 13.7 0 10.8-9.5 17.5-19.4 23.6L59 45.6l-6.2-3.2Z"
        fill="#F13A1F"
      />
    </svg>
  )
}

function StormCloudsIcon() {
  return (
    <svg
      aria-hidden="true"
      className="overflow-visible"
      fill="none"
      height="58"
      viewBox="0 0 88 58"
      width="88"
    >
      <g filter="url(#cloud-shadow)">
        <path
          d="M18 37.7c0-5.4 4.2-9.7 9.5-10 .9-6.4 6.3-11.3 12.9-11.3 4.6 0 8.7 2.3 11.1 5.9 1.6-1 3.4-1.5 5.4-1.5 5.6 0 10.2 4.4 10.4 10 4.4.7 7.7 4.5 7.7 9 0 5.1-4.1 9.2-9.2 9.2H27.1c-5 0-9.1-4.1-9.1-9.3Z"
          fill="#D7D3CE"
        />
        <path
          d="M12 40.1c0-4.8 3.7-8.8 8.5-9.2.8-5.8 5.8-10.3 11.8-10.3 4.2 0 7.9 2.1 10.1 5.4 1.4-.9 3.1-1.4 4.9-1.4 5.2 0 9.4 4.1 9.5 9.2 4 .7 7.1 4.1 7.1 8.2 0 4.6-3.8 8.4-8.4 8.4H20.3c-4.6 0-8.3-3.8-8.3-8.4Z"
          fill="#C4C0BA"
        />
        <path
          d="M50.5 31.3 42.8 41h5l-4.8 10.4 10.8-13.1h-5.3l2-7Z"
          fill="#D5A5FF"
        />
      </g>
      <defs>
        <filter
          colorInterpolationFilters="sRGB"
          filterUnits="userSpaceOnUse"
          height="48"
          id="cloud-shadow"
          width="70"
          x="9"
          y="12"
        >
          <feDropShadow dx="0" dy="4" floodColor="#b9b2aa" floodOpacity="0.35" stdDeviation="2.5" />
        </filter>
      </defs>
    </svg>
  )
}

function PolaroidVisual() {
  return (
    <div className="flex w-[120px] items-start justify-center pt-1 sm:w-[136px]">
      <img
        alt="Soulink"
        className="h-auto w-full rounded-[24px] border border-[#eadfd6] object-contain shadow-[0_10px_24px_rgba(73,52,41,0.08)]"
        src={brandIcon}
      />
    </div>
  )
}

function renderParagraphs(text: string) {
  return text
    .split(/\n+/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean)
    .map((paragraph, index) => (
      <p key={`${paragraph}-${index}`}>{renderHighlightedText(paragraph)}</p>
    ))
}

function renderHighlightedText(text: string) {
  const matches = [...text.matchAll(/\*\*(.+?)\*\*/g)]
  if (matches.length === 0) {
    return text
  }

  const parts: ReactNode[] = []
  let cursor = 0

  matches.forEach((match, index) => {
    const [raw, content] = match
    const start = match.index ?? 0

    if (start > cursor) {
      parts.push(text.slice(cursor, start))
    }

    parts.push(
      <span className="font-semibold text-[#8e63f1]" key={`${content}-${index}`}>
        {content}
      </span>,
    )
    cursor = start + raw.length
  })

  if (cursor < text.length) {
    parts.push(text.slice(cursor))
  }

  return parts.map((part, index) => <Fragment key={index}>{part}</Fragment>)
}
