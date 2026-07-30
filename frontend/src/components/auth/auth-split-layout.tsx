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
    <div className="min-h-screen bg-[#f7f6f3] lg:grid lg:grid-cols-[minmax(0,1fr)_minmax(30rem,0.8fr)]">
      <AuthHeroPanel
        imageAlt={imageAlt}
        imageSources={imageSources}
        imageSrc={imageSrc}
        subtitle={subtitle}
        title={title}
      />

      <section className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[#f7f6f3] px-6 py-12 sm:px-8 lg:px-12">
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
            <div className="absolute inset-0 bg-[rgba(24,20,26,0.52)]" />
          </div>
        ) : null}

        <Link
          className={cn(
            buttonVariants({ variant: 'secondary', size: 'sm' }),
            'absolute left-6 top-6 z-20 gap-2 border border-[#d9d5cf] bg-white px-4 text-[#5f5b55] shadow-sm sm:left-8 sm:top-8 lg:hidden',
          )}
          to="/"
        >
          <ArrowLeft size={16} />
          返回首页
        </Link>

        <div className="relative z-10 w-full max-w-[28rem] rounded-lg border border-[#dfdcd6] bg-white p-6 shadow-[0_18px_44px_rgba(35,33,31,0.08)] sm:p-8 lg:border-none lg:bg-transparent lg:p-0 lg:shadow-none">
          {children}
        </div>
      </section>
    </div>
  )
}
