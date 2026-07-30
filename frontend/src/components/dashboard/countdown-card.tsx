import { useEffect, useMemo, useState } from 'react'

import { Card, CardContent } from '@/components/ui/card'
import { getNextMatchRevealStart, MATCH_SIGNUP_DEADLINE_TEXT } from '@/lib/schedule'

interface CountdownCardProps {
  description: string
  label: string
  targetAt: string
}

interface CountdownParts {
  days: string
  hours: string
  minutes: string
  seconds: string
  isComplete: boolean
}

function getCountdownParts(targetAt: string): CountdownParts {
  const diff = new Date(targetAt).getTime() - Date.now()

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

export function CountdownCard({
  description,
  label,
  targetAt,
}: CountdownCardProps) {
  const [currentTargetAt, setCurrentTargetAt] = useState(targetAt)
  const [parts, setParts] = useState(() => getCountdownParts(targetAt))

  useEffect(() => {
    setCurrentTargetAt(targetAt)
    setParts(getCountdownParts(targetAt))
  }, [targetAt])

  useEffect(() => {
    const timer = window.setInterval(() => {
      const nextParts = getCountdownParts(currentTargetAt)

      if (nextParts.isComplete) {
        const nextTargetAt = getNextMatchRevealStart()
        setCurrentTargetAt(nextTargetAt)
        setParts(getCountdownParts(nextTargetAt))
        return
      }

      setParts(nextParts)
    }, 1000)

    return () => window.clearInterval(timer)
  }, [currentTargetAt])

  const units = useMemo(
    () => [
      { key: 'days', value: parts.days, label: '天' },
      { key: 'hours', value: parts.hours, label: '时' },
      { key: 'minutes', value: parts.minutes, label: '分' },
      { key: 'seconds', value: parts.seconds, label: '秒' },
    ],
    [parts.days, parts.hours, parts.minutes, parts.seconds],
  )

  return (
    <Card className="overflow-hidden border-[#eee4db] bg-[linear-gradient(180deg,rgba(255,251,247,0.96)_0%,rgba(255,248,243,0.94)_100%)] shadow-[0_20px_64px_rgba(72,50,33,0.08)]">
      <CardContent className="flex flex-col items-center gap-9 px-6 py-10 text-center sm:px-12 sm:py-14">
        <div className="inline-flex items-center gap-3 rounded-full border border-[#efe2d8] bg-white/76 px-5 py-2 text-[0.95rem] tracking-[0.04em] text-[#8f7c70]">
          <span className="h-2.5 w-2.5 rounded-full bg-[#c38a79]" />
          {label}
        </div>

        <div className="grid grid-cols-2 gap-x-8 gap-y-7 sm:grid-cols-4 sm:gap-x-10 sm:gap-y-8">
          {units.map((unit) => (
            <div className="min-w-[96px] space-y-2.5" key={unit.key}>
              <div className="font-serif text-[3.45rem] leading-none tracking-[-0.04em] text-[#97634d] sm:text-6xl">
                {unit.value}
              </div>
              <div className="text-[0.92rem] tracking-[0.32em] text-[#ad9889]">
                {unit.label}
              </div>
            </div>
          ))}
        </div>

        <div className="space-y-2.5 text-center">
          <p className="max-w-2xl text-[1.02rem] leading-8 text-[#8b776a] sm:text-[1.08rem]">
            {description}
          </p>
          <p className="max-w-2xl text-[1.02rem] leading-8 text-[#8b776a] sm:text-[1.08rem]">
            {MATCH_SIGNUP_DEADLINE_TEXT}
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
