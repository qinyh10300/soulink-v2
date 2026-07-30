import { Compass } from 'lucide-react'

import type { DashboardUser } from '@/types/dashboard'

interface WelcomeSectionProps {
  user: DashboardUser
}

export function WelcomeSection({ user }: WelcomeSectionProps) {
  return (
    <section className="space-y-4 sm:space-y-5">
      <h1 className="flex flex-wrap items-center gap-3 font-serif text-[2.2rem] leading-[1.04] tracking-[-0.03em] text-[#2d211a] sm:text-[2.8rem] lg:text-[3.2rem]">
        <span className="inline-flex flex-wrap items-center gap-2.5">
          <span className="inline-flex items-center gap-2 px-1.5 py-1 text-[0.82em] text-[#a67552]">
            <span className="h-1.5 w-1.5 rounded-full bg-[#d1ae8f]" />
            <span>{user.name}</span>
            <span className="h-1.5 w-1.5 rounded-full bg-[#d1ae8f]" />
          </span>
          <span className="text-[2rem] leading-[1.08] tracking-[-0.03em] sm:text-[2.2rem]">
            的主页
          </span>
        </span>
        <span className="-mt-1 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#eadccf] bg-[rgba(255,250,245,0.94)] text-[#ad7b58] shadow-[0_14px_28px_rgba(73,49,33,0.08)] sm:h-11 sm:w-11">
          <Compass size={19} strokeWidth={1.7} />
        </span>
      </h1>
      <p className="max-w-2xl text-[1.08rem] leading-8 tracking-[0.01em] text-[#8a7366] sm:text-[1.18rem] sm:leading-9">
        每一次认真表达，都会为你留下更靠近同频的坐标
      </p>
    </section>
  )
}
