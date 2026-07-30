import type { ReactNode } from 'react'
import { ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'

import { AuthHeroPanel } from '@/components/auth/auth-hero-panel'
import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface AuthSplitLayoutProps {
  children: ReactNode
  imageSrc: string
  imageSources?: {
    avif: string
    webp: string
  }
  imageAlt: string
  title: string
  subtitle: string
  showImageOnMobile?: boolean
}

export function AuthSplitLayout({
  children,
  imageSrc,
  imageSources,
  imageAlt,
  title,
  subtitle,
  showImageOnMobile = false,
}: AuthSplitLayoutProps) {
  return (
    <div className="min-h-screen bg-[#ece3db] lg:grid lg:grid-cols-[minmax(0,1.00fr)_minmax(32rem,1.00fr)]">
      <AuthHeroPanel
        imageAlt={imageAlt}
        imageSources={imageSources}
        imageSrc={imageSrc}
        subtitle={subtitle}
        title={title}
      />

      <section className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[linear-gradient(180deg,#f9f4ef_0%,#f6f0e8_100%)] px-6 py-12 sm:px-8 lg:px-12">
        {showImageOnMobile ? (
          <div className="absolute inset-0 lg:hidden">
            <picture className="absolute inset-0 block">
              {imageSources ? (
                <>
                  <source srcSet={imageSources.avif} type="image/avif" />
                  <source srcSet={imageSources.webp} type="image/webp" />
                </>
              ) : null}
              <img
                alt={imageAlt}
                className="h-full w-full object-cover"
                decoding="async"
                fetchPriority="high"
                height={1024}
                loading="eager"
                src={imageSrc}
                width={1536}
              />
            </picture>
            <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(58,41,35,0.18)_0%,rgba(78,55,45,0.26)_42%,rgba(40,25,22,0.5)_100%)]" />
            <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,216,194,0.2),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(146,92,66,0.18),transparent_30%)]" />
          </div>
        ) : null}

        <Link
          className={cn(
            buttonVariants({ variant: 'secondary', size: 'sm' }),
            'absolute left-6 top-6 z-20 gap-2 rounded-full border border-white/55 bg-[rgba(249,244,239,0.88)] px-4 text-[#5f4c40] shadow-[0_14px_32px_rgba(40,25,22,0.12)] backdrop-blur-md hover:bg-[rgba(255,250,246,0.96)] sm:left-8 sm:top-8 lg:hidden',
          )}
          to="/"
        >
          <ArrowLeft size={16} />
          返回首页
        </Link>

        <div className="relative z-10 w-full max-w-[28rem] rounded-[2rem] border border-white/40 bg-[rgba(249,244,239,0.78)] p-6 shadow-[0_18px_44px_rgba(40,25,22,0.16)] backdrop-blur-md sm:p-8 lg:rounded-none lg:border-none lg:bg-transparent lg:p-0 lg:shadow-none lg:backdrop-blur-0">
          {children}
        </div>
      </section>
    </div>
  )
}
