import { motion, useReducedMotion } from 'framer-motion'
import { ArrowLeft, ArrowRight } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'

import soulinkWordmark from '@/assets/soulink-interactive-white.svg'
import { buttonVariants } from '@/components/ui/button'
import {
  fetchPkuRegisteredUserCount,
  fetchRegisteredUserCount,
  fetchSchoolRegisteredUserCount,
} from '@/lib/meta-api'
import type { SchoolKey } from '@/lib/school-keys'
import { heroRegistrationCount, heroStats } from '@/mock/landing'
import { cn } from '@/lib/utils'

const heroEase = [0.22, 1, 0.36, 1] as const
const fallbackPkuRegistrationCount = 67

function getFallbackRegistrationCount(schoolKey: SchoolKey, countOverride?: number) {
  if (typeof countOverride === 'number') {
    return countOverride
  }

  switch (schoolKey) {
    case 'thu':
      return heroRegistrationCount
    case 'pku':
      return fallbackPkuRegistrationCount
    default:
      return 0
  }
}

function AnimatedCount({
  value,
  shouldReduceMotion,
}: {
  value: number
  shouldReduceMotion: boolean
}) {
  const [displayValue, setDisplayValue] = useState(shouldReduceMotion ? value : 0)

  useEffect(() => {
    if (shouldReduceMotion) {
      setDisplayValue(value)
      return
    }

    let frameId = 0
    const duration = 1600
    const startTime = performance.now()

    const tick = (now: number) => {
      const progress = Math.min((now - startTime) / duration, 1)
      const easedProgress = 1 - Math.pow(1 - progress, 4)
      setDisplayValue(Math.round(value * easedProgress))

      if (progress < 1) {
        frameId = requestAnimationFrame(tick)
      }
    }

    setDisplayValue(0)
    frameId = requestAnimationFrame(tick)

    return () => cancelAnimationFrame(frameId)
  }, [shouldReduceMotion, value])

  return (
    <span
      aria-label={value.toLocaleString('en-US')}
      className="inline-block min-w-[4.3ch] text-center tabular-nums leading-none"
      role="text"
    >
      {displayValue.toLocaleString('en-US')}
    </span>
  )
}

function RegistrationCountBubble({
  animationDelay,
  shouldReduceMotion,
  value,
}: {
  animationDelay: number
  shouldReduceMotion: boolean
  value: number
}) {
  return (
    <motion.span
      animate={
        shouldReduceMotion
          ? undefined
          : {
              backgroundColor: [
                'rgba(246,234,210,0.14)',
                'rgba(246,234,210,0.20)',
                'rgba(246,234,210,0.14)',
              ],
              borderColor: [
                'rgba(210,170,94,0.88)',
                'rgba(242,206,132,1)',
                'rgba(210,170,94,0.88)',
              ],
              boxShadow: [
                '0 0 0 1px rgba(170,128,62,0.22), 0 0 16px rgba(243,207,125,0.12)',
                '0 0 0 1px rgba(193,149,74,0.38), 0 0 22px rgba(243,207,125,0.22)',
                '0 0 0 1px rgba(170,128,62,0.22), 0 0 16px rgba(243,207,125,0.12)',
              ],
            }
      }
      className="inline-flex shrink-0 items-center justify-center rounded-full border-[2.8px] border-[#d8ad64]/90 bg-[rgba(246,234,210,0.15)] px-[0.8em] py-[0.16em] text-[1.08em] font-bold leading-none text-[#fbedd0] shadow-[0_0_0_1px_rgba(170,128,62,0.12)]"
      transition={
        shouldReduceMotion
          ? undefined
          : {
              delay: animationDelay,
              duration: 4.8,
              ease: 'easeInOut',
              repeat: Infinity,
            }
      }
    >
      <AnimatedCount shouldReduceMotion={shouldReduceMotion} value={value} />
    </motion.span>
  )
}

function SingleSchoolRegistrationBanner({
  count,
  shouldReduceMotion,
  title,
}: {
  count: number
  shouldReduceMotion: boolean
  title: string
}) {
  return (
    <>
      <div className="flex w-full flex-col items-center gap-2.5 sm:hidden">
        <div className="inline-flex items-center gap-[0.55em] font-serif text-[1.02rem] italic leading-none tracking-[0.1em] text-[#f3cfa3]">
          <RegistrationCountBubble
            animationDelay={2}
            shouldReduceMotion={shouldReduceMotion}
            value={count}
          />
          <span className="leading-none">{title}</span>
        </div>
      </div>

      <div className="hidden items-center justify-center gap-5 sm:flex">
        <span className="h-px w-12 bg-[#d7a56e]/60 sm:w-20" />
        <span className="inline-flex items-center gap-[0.45em] font-serif text-[1.2rem] italic leading-none tracking-[0.12em] text-[#f3cfa3] sm:text-[1.4rem]">
          <span className="leading-none">For</span>
          <RegistrationCountBubble
            animationDelay={2}
            shouldReduceMotion={shouldReduceMotion}
            value={count}
          />
          <span className="leading-none">{title}</span>
        </span>
        <span className="h-px w-12 bg-[#d7a56e]/60 sm:w-20" />
      </div>
    </>
  )
}

function DualSchoolRegistrationBanner({
  pkuRegistrationCount,
  shouldReduceMotion,
  thuRegistrationCount,
}: {
  pkuRegistrationCount: number
  shouldReduceMotion: boolean
  thuRegistrationCount: number
}) {
  return (
    <>
      <div className="flex w-full flex-col items-center gap-2.5 sm:hidden">
        <div className="inline-flex items-center gap-[0.55em] font-serif text-[1.02rem] italic leading-none tracking-[0.1em] text-[#f3cfa3]">
          <RegistrationCountBubble
            animationDelay={2}
            shouldReduceMotion={shouldReduceMotion}
            value={thuRegistrationCount}
          />
          <span className="leading-none">THUers</span>
        </div>
        <div className="inline-flex items-center gap-[0.55em] font-serif text-[1.02rem] italic leading-none tracking-[0.1em] text-[#f3cfa3]">
          <RegistrationCountBubble
            animationDelay={2.2}
            shouldReduceMotion={shouldReduceMotion}
            value={pkuRegistrationCount}
          />
          <span className="leading-none">PKUers</span>
        </div>
      </div>

      <div className="hidden items-center justify-center gap-5 sm:flex">
        <span className="h-px w-12 bg-[#d7a56e]/60 sm:w-20" />
        <span className="inline-flex items-center gap-[0.45em] font-serif text-[1.2rem] italic leading-none tracking-[0.12em] text-[#f3cfa3] sm:text-[1.4rem]">
          <span className="leading-none">For</span>
          <RegistrationCountBubble
            animationDelay={2}
            shouldReduceMotion={shouldReduceMotion}
            value={thuRegistrationCount}
          />
          <span className="leading-none">THUers</span>
          <span className="leading-none text-[#d8b48b]">&</span>
          <RegistrationCountBubble
            animationDelay={2.2}
            shouldReduceMotion={shouldReduceMotion}
            value={pkuRegistrationCount}
          />
          <span className="leading-none">PKUers</span>
        </span>
        <span className="h-px w-12 bg-[#d7a56e]/60 sm:w-20" />
      </div>
    </>
  )
}

type LandingHeroProps = {
  countOverride?: number
  currentLocationLabel?: string
  currentSchoolKey?: SchoolKey
  currentTagline?: string
  currentTitle?: string
  onNextSlide?: () => void
  onPreviousSlide?: () => void
  showNextArrow?: boolean
  showPreviousArrow?: boolean
  showRegistrationCount?: boolean
}

export function LandingHero({
  countOverride,
  currentLocationLabel = '',
  currentSchoolKey,
  currentTagline = '',
  currentTitle = 'THUers',
  onNextSlide,
  onPreviousSlide,
  showNextArrow = true,
  showPreviousArrow = true,
  showRegistrationCount = true,
}: LandingHeroProps) {
  const shouldReduceMotion = useReducedMotion() ?? false
  const [thuRegistrationCount, setThuRegistrationCount] = useState(heroRegistrationCount)
  const [pkuRegistrationCount, setPkuRegistrationCount] = useState(fallbackPkuRegistrationCount)
  const [schoolRegistrationCount, setSchoolRegistrationCount] = useState(() =>
    currentSchoolKey ? getFallbackRegistrationCount(currentSchoolKey, countOverride) : 0,
  )

  useEffect(() => {
    if (!currentSchoolKey) {
      return
    }

    setSchoolRegistrationCount(getFallbackRegistrationCount(currentSchoolKey, countOverride))
  }, [countOverride, currentSchoolKey])

  useEffect(() => {
    let cancelled = false

    const loadRegisteredUserCount = async () => {
      try {
        if (currentSchoolKey) {
          const result = await fetchSchoolRegisteredUserCount(currentSchoolKey)

          if (!cancelled) {
            setSchoolRegistrationCount(result.registeredUserCount)
          }

          return
        }

        const [thuResult, pkuResult] = await Promise.all([
          fetchRegisteredUserCount(),
          fetchPkuRegisteredUserCount(),
        ])

        if (cancelled) {
          return
        }

        setThuRegistrationCount(thuResult.registeredUserCount)
        setPkuRegistrationCount(pkuResult.registeredUserCount)
      } catch {
        // Keep the existing fallback values when the API is unavailable.
      }
    }

    void loadRegisteredUserCount()

    return () => {
      cancelled = true
    }
  }, [currentSchoolKey])

  return (
    <section className="relative overflow-hidden">
      <div className="relative min-h-[100svh] w-full overflow-hidden bg-transparent md:bg-[rgba(231,221,211,0.04)]">
        <div className="absolute inset-0 hidden md:block md:bg-[linear-gradient(180deg,rgba(83,56,47,0.08)_0%,rgba(136,101,81,0.03)_34%,rgba(231,221,211,0)_100%)]" />
        <div className="absolute inset-0 hidden md:block md:bg-[radial-gradient(circle_at_top_left,rgba(210,177,157,0.1),transparent_30%),radial-gradient(circle_at_top_right,rgba(194,154,127,0.08),transparent_30%),radial-gradient(circle_at_center,rgba(255,248,241,0.03),transparent_46%)]" />
        <div className="absolute -left-10 top-0 hidden h-[32rem] w-[32rem] rounded-full bg-[#dcb9aa]/10 blur-[120px] md:block" />
        <div className="absolute right-[-8rem] top-16 hidden h-[30rem] w-[30rem] rounded-full bg-[#d5a988]/8 blur-[120px] md:block" />
        <div className="absolute bottom-[-4rem] left-1/2 h-[16rem] w-[92%] -translate-x-1/2 rounded-[100%] bg-[#f8f0e8]/18 blur-[62px]" />
        <div className="absolute bottom-0 left-0 right-0 h-[26rem] bg-[linear-gradient(180deg,rgba(231,221,211,0)_0%,rgba(239,232,225,0.08)_28%,rgba(231,221,211,0.26)_64%,#e7ddd3_100%)]" />

        <div className="absolute left-[-6rem] top-[6rem] hidden h-[18rem] w-[40rem] rotate-[18deg] border-t border-white/14 opacity-35 md:block" />
        <div className="absolute right-[-8rem] top-[4rem] hidden h-[20rem] w-[44rem] -rotate-[16deg] border-t border-white/14 opacity-35 md:block" />
        <div className="absolute left-[8%] top-[9rem] hidden h-4 w-4 rounded-full bg-[#efd1c4]/55 blur-[1px] md:block" />
        <div className="absolute left-[12%] top-[13rem] hidden h-3 w-3 rounded-full bg-[#efd1c4]/55 blur-[1px] md:block" />
        <div className="absolute right-[14%] top-[10rem] hidden h-4 w-4 rounded-full bg-[#f0d0bb]/55 blur-[1px] md:block" />
        <div className="absolute right-[18%] top-[14rem] hidden h-3 w-3 rounded-full bg-[#f0d0bb]/55 blur-[1px] md:block" />

        <div className="pointer-events-none absolute inset-y-0 left-0 right-0 z-20 hidden items-center justify-between px-4 sm:flex lg:px-6">
          <div className="w-12">
            {showPreviousArrow ? (
              <button
                aria-label="Previous slide"
                className="pointer-events-auto inline-flex h-12 w-12 items-center justify-center rounded-full border border-white/18 bg-white/10 text-white backdrop-blur-md transition hover:bg-white/16"
                onClick={onPreviousSlide}
                type="button"
              >
                <ArrowLeft size={18} />
              </button>
            ) : null}
          </div>
          <div className="w-12">
            {showNextArrow ? (
              <button
                aria-label="Next slide"
                className="pointer-events-auto inline-flex h-12 w-12 items-center justify-center rounded-full border border-white/18 bg-white/10 text-white backdrop-blur-md transition hover:bg-white/16"
                onClick={onNextSlide}
                type="button"
              >
                <ArrowRight size={18} />
              </button>
            ) : null}
          </div>
        </div>

        <motion.div
          animate="show"
          className="relative z-10 flex min-h-[96svh] flex-col items-center justify-center px-6 pb-20 pt-28 text-center sm:px-10 sm:pt-32"
          initial="hidden"
          variants={{
            hidden: {},
            show: {
              transition: {
                staggerChildren: 0.1,
              },
            },
          }}
        >
          <motion.div
            className="mt-8 max-w-4xl"
            transition={{ duration: 0.6, ease: heroEase }}
            variants={{
              hidden: { opacity: 0, y: 18 },
              show: { opacity: 1, y: 0 },
            }}
          >
            <h1 className="flex justify-center">
              <motion.div
                animate={
                  shouldReduceMotion
                    ? undefined
                    : {
                        scale: [1, 1.03, 1],
                        y: [0, -10, 0],
                      }
                }
                className="relative w-[min(72vw,39rem)] drop-shadow-[0_20px_36px_rgba(28,15,24,0.16)] lg:w-[min(64vw,43rem)]"
                style={{ top: '-1cm' }}
                transition={
                  shouldReduceMotion
                    ? undefined
                    : {
                        duration: 5.6,
                        ease: 'easeInOut',
                      }
                }
              >
                <img
                  alt="Soulink"
                  className="pointer-events-none block aspect-[1899/926] h-auto w-full select-none"
                  decoding="async"
                  fetchPriority="high"
                  loading="eager"
                  src={soulinkWordmark}
                />
              </motion.div>
            </h1>

            <div className="relative z-20 mx-auto mt-2 inline-flex items-center gap-2 rounded-full border border-white/22 bg-white/22 px-4 py-2 text-sm text-white shadow-[0_12px_30px_rgba(34,19,28,0.18)] backdrop-blur-md">
              <span className="relative flex h-2.5 w-2.5">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-[#8CF5A5] opacity-75" />
                <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-[#7CF29A]" />
              </span>
              <span
                className="text-[1.02rem] tracking-[0.08em] text-white/92"
                style={{
                  fontFamily:
                    "'KaiTi', 'STKaiti', 'Kaiti SC', 'Source Han Serif SC', 'Noto Serif SC', serif",
                }}
              >
                {currentLocationLabel}
              </span>
            </div>

            {showRegistrationCount ? (
              <div className="mx-auto mt-4 flex w-full max-w-[22rem] flex-col items-center gap-3 text-[#f0c38d] sm:max-w-none sm:text-base">
                <div className="flex w-full items-center justify-center gap-3 sm:hidden">
                  <span className="h-px flex-1 bg-[#d7a56e]/45" />
                  <span className="text-[0.78rem] uppercase tracking-[0.3em] text-[#f3cfa3]/80">
                    For
                  </span>
                  <span className="h-px flex-1 bg-[#d7a56e]/45" />
                </div>

                {currentSchoolKey ? (
                  <SingleSchoolRegistrationBanner
                    count={schoolRegistrationCount}
                    shouldReduceMotion={shouldReduceMotion}
                    title={currentTitle}
                  />
                ) : (
                  <DualSchoolRegistrationBanner
                    pkuRegistrationCount={pkuRegistrationCount}
                    shouldReduceMotion={shouldReduceMotion}
                    thuRegistrationCount={thuRegistrationCount}
                  />
                )}
              </div>
            ) : null}

            <p
              className="mx-auto mt-5 max-w-3xl text-balance text-[1.28rem] tracking-[0.15em] text-[#fff1dd] [text-shadow:0_10px_24px_rgba(46,23,15,0.18)] sm:text-[1.82rem]"
              style={{
                fontFamily:
                  "'KaiTi', 'STKaiti', 'Kaiti SC', 'Source Han Serif SC', 'Noto Serif SC', serif",
              }}
            >
              {currentTagline}
            </p>
          </motion.div>

          <motion.div
            className="mt-6 flex flex-col items-center gap-4"
            transition={{ duration: 0.62, ease: heroEase }}
            variants={{
              hidden: { opacity: 0, y: 18 },
              show: { opacity: 1, y: 0 },
            }}
          >
            <Link
              className={cn(
                buttonVariants({ variant: 'secondary', size: 'lg' }),
                'h-14 min-w-44 rounded-full border-none bg-[#f8f3eb] px-8 text-base font-semibold text-[#2b2530] shadow-[0_16px_36px_rgba(32,18,23,0.2)] hover:bg-white',
              )}
              to="/register"
            >
              立即注册
              <ArrowRight className="ml-2" size={18} />
            </Link>

            <div className="flex flex-wrap items-center justify-center gap-3 text-sm text-white/72">
              {heroStats.map((stat) => (
                <div
                  key={stat.label}
                  className="rounded-full border border-white/14 bg-white/8 px-4 py-2 backdrop-blur-md"
                >
                  <span className="font-medium text-white">{stat.value}</span>
                  <span className="ml-2">{stat.label}</span>
                </div>
              ))}
            </div>
          </motion.div>
        </motion.div>
      </div>
    </section>
  )
}
