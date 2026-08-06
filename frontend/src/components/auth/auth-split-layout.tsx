import { ArrowLeft, Sparkles } from 'lucide-react'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { AuthHeroPanel } from '@/components/auth/auth-hero-panel'

interface AuthSplitLayoutProps {
  children: ReactNode
  title: string
  subtitle: string
}

export function AuthSplitLayout({ children, title, subtitle }: AuthSplitLayoutProps) {
  return (
    <div className="min-h-screen bg-white lg:grid lg:grid-cols-[minmax(0,1.12fr)_minmax(30rem,0.88fr)]">
      <AuthHeroPanel subtitle={subtitle} title={title} />

      <section className="relative flex min-h-screen items-center justify-center bg-white px-5 pb-12 pt-24 sm:px-8 lg:px-12 lg:py-14 xl:px-16">
        <div className="absolute inset-x-5 top-6 flex items-center justify-between sm:inset-x-8 lg:inset-x-10">
          <Link className="flex items-center gap-2.5 lg:hidden" to="/">
            <span className="flex h-8 w-8 items-center justify-center rounded-md bg-[#1447e6] text-white">
              <Sparkles size={16} />
            </span>
            <span className="text-[16px] font-semibold">SouLink</span>
          </Link>
          <Link
            className="ml-auto inline-flex h-9 items-center gap-2 rounded-md px-2.5 text-xs font-medium text-[#71717a] transition-colors hover:bg-[#f4f4f5] hover:text-[#18181b]"
            to="/"
          >
            <ArrowLeft size={14} />
            返回首页
          </Link>
        </div>

        <div className="w-full max-w-[27rem]">{children}</div>
      </section>
    </div>
  )
}
