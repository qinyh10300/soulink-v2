import { motion, type Variants } from 'framer-motion'
import { useEffect, useState } from 'react'

import { cn } from '@/lib/utils'
import {
  formatMatchRevealStartLabel,
  getNextMatchRevealStart,
} from '@/lib/schedule'

const sectionEase = [0.22, 1, 0.36, 1] as const

const countdownNumberVariant: Variants = {
  initial: { opacity: 0.55, y: 6, filter: 'blur(1px)' },
  animate: {
    opacity: 1,
    y: 0,
    filter: 'blur(0px)',
    transition: { duration: 0.3, ease: sectionEase },
  },
}

function getCountdownParts(targetAt: string, now = new Date()) {
  const diff = new Date(targetAt).getTime() - now.getTime()

  if (diff <= 0) {
    return {
      days: '00',
      hours: '00',
      minutes: '00',
      seconds: '00',
      isComplete: true,
    }
  }

  const totalSeconds = Math.floor(diff / 1000)

  const days = Math.floor(totalSeconds / 86400)
  const hours = Math.floor((totalSeconds % 86400) / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  return {
    days: String(days).padStart(2, '0'),
    hours: String(hours).padStart(2, '0'),
    minutes: String(minutes).padStart(2, '0'),
    seconds: String(seconds).padStart(2, '0'),
    isComplete: false,
  }
}

type MatchRevealCountdownProps = {
  className?: string
}

function CountdownUnit({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="flex min-w-[4.15rem] flex-col items-center gap-2 sm:min-w-[5.15rem]">
      <motion.span
        animate="animate"
        className="tabular-nums text-[3.2rem] font-normal leading-none tracking-[-0.095em] text-[#2c241f] sm:text-[4.1rem] xl:text-[4.6rem]"
        initial="initial"
        key={value}
        style={{ fontFamily: '"Playfair Display", "Bodoni Moda", "Times New Roman", serif' }}
        variants={countdownNumberVariant}
      >
        {value}
      </motion.span>
      <span
        className="text-[1rem] text-[#a88f7d] sm:text-[1.12rem]"
        style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
      >
        {label}
      </span>
    </div>
  )
}

function CountdownDivider({ delay = 0 }: { delay?: number }) {
  return (
    <motion.span
      animate={{ opacity: [0.38, 0.9, 0.38] }}
      className="mx-[-0.18rem] pb-8 text-[1.72rem] text-[#d1b9a6] sm:mx-[-0.25rem] sm:pb-9 sm:text-[2.3rem]"
      style={{ fontFamily: '"Cormorant Garamond", serif' }}
      transition={{
        delay,
        duration: 1.8,
        ease: 'easeInOut',
        repeat: Number.POSITIVE_INFINITY,
      }}
    >
      :
    </motion.span>
  )
}

export function MatchRevealCountdown({ className }: MatchRevealCountdownProps) {
  const [currentTargetAt, setCurrentTargetAt] = useState(() => getNextMatchRevealStart())
  const [countdown, setCountdown] = useState(() => getCountdownParts(getNextMatchRevealStart()))

  useEffect(() => {
    const timer = window.setInterval(() => {
      const nextCountdown = getCountdownParts(currentTargetAt)

      if (nextCountdown.isComplete) {
        const nextTargetAt = getNextMatchRevealStart()
        setCurrentTargetAt(nextTargetAt)
        setCountdown(getCountdownParts(nextTargetAt))
        return
      }

      setCountdown(nextCountdown)
    }, 1000)

    return () => window.clearInterval(timer)
  }, [currentTargetAt])

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-[2.25rem] border border-[#efe4d9] bg-[linear-gradient(180deg,rgba(255,255,255,0.98)_0%,rgba(252,248,244,0.96)_100%)] shadow-[0_26px_60px_rgba(160,131,104,0.12)]',
        className,
      )}
    >
      <div className="pointer-events-none absolute inset-0">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.96),transparent_40%),radial-gradient(circle_at_bottom,rgba(246,237,229,0.72),transparent_46%)]" />
      </div>

      <div className="relative px-7 py-8 text-center sm:px-10 sm:py-10">
        <div className="space-y-5">
          <p
            className="text-[1.08rem] tracking-[0.18em] text-[#b0927e] sm:text-[1.35rem]"
            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
          >
            距下次配对揭晓
          </p>

          <div className="flex flex-wrap items-end justify-center gap-x-1.5 gap-y-3 text-[#2f2923] sm:flex-nowrap sm:gap-x-2.5">
            <CountdownUnit label="天" value={countdown.days} />
            <CountdownDivider />
            <CountdownUnit label="时" value={countdown.hours} />
            <CountdownDivider delay={0.12} />
            <CountdownUnit label="分" value={countdown.minutes} />
            <CountdownDivider delay={0.24} />
            <CountdownUnit label="秒" value={countdown.seconds} />
          </div>

          <p
            className="text-[1.02rem] tracking-[0.06em] text-[#947f70] sm:text-[1.16rem]"
            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
          >
            {formatMatchRevealStartLabel(new Date(currentTargetAt))}
          </p>
        </div>
      </div>
    </div>
  )
}
