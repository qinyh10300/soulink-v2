import type { ReactNode } from 'react'

import { Footer } from '@/components/layout/footer'
import { Navbar } from '@/components/layout/navbar'
import { ScrollManager } from '@/components/shared/scroll-manager'

interface AppLayoutProps {
  children: ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="relative min-h-screen overflow-hidden">
      <div className="pointer-events-none absolute inset-x-0 top-0 z-0 h-[34rem] bg-[radial-gradient(circle_at_top,_rgba(90,120,184,0.14),_transparent_56%)]" />
      <ScrollManager />
      <Navbar />
      <main className="relative z-10">{children}</main>
      <Footer />
    </div>
  )
}
