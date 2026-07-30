import { Menu, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'

import { BrandMark } from '@/components/shared/brand-mark'
import { WechatGroupModal } from '@/components/shared/wechat-group-modal'
import { buttonVariants } from '@/components/ui/button'
import { hasActiveAuthSession } from '@/lib/user-profile-storage'
import { cn } from '@/lib/utils'

export function Navbar() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [isScrolled, setIsScrolled] = useState(false)
  const [isWechatModalOpen, setIsWechatModalOpen] = useState(false)
  const location = useLocation()
  const isAllianceHome = location.pathname === '/'
  const isSchoolPage = location.pathname.startsWith('/schools/')
  const hasActiveSession = hasActiveAuthSession()
  const authActionLabel = hasActiveSession ? '前往主页' : '登录'
  const authActionTarget = hasActiveSession ? '/dashboard' : '/login'

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 18)
    }

    handleScroll()
    window.addEventListener('scroll', handleScroll, { passive: true })

    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const useTransparentHeader = (isAllianceHome || isSchoolPage) && !isScrolled && !mobileMenuOpen
  const useTopDarkText = (isAllianceHome || isSchoolPage) && !isScrolled && !mobileMenuOpen
  const glowingBrownText =
    'text-[#4b2f1f] [text-shadow:0_0_10px_rgba(239,196,92,0.55),0_0_18px_rgba(201,156,54,0.34)]'
  const schoolBrandTitleClass = cn('transition-colors font-bold', glowingBrownText)
  const schoolBrandSubtitleClass = cn('transition-colors font-semibold', glowingBrownText)
  const schoolNavLinkClass = cn(
    'rounded-full bg-transparent font-bold hover:bg-[#4b2f1f]/6 hover:text-[#4b2f1f]',
    glowingBrownText,
  )
  const homeAboutLinkClass = cn(
    'rounded-full bg-transparent font-bold hover:bg-[#4b2f1f]/6 hover:text-[#4b2f1f]',
    glowingBrownText,
  )

  return (
    <>
      <header
        className={cn(
          'fixed inset-x-0 top-0 z-40 transition-all duration-300',
          useTransparentHeader
            ? 'border-b border-transparent bg-transparent'
            : 'border-b border-white/40 bg-white/80 shadow-none backdrop-blur-xl md:shadow-[0_12px_40px_rgba(15,23,42,0.08)]',
        )}
      >
        <div className="mx-auto flex w-full max-w-6xl items-center justify-between px-6 py-2 lg:px-8">
          <BrandMark
            iconClassName={cn(
              'shadow-none transition-colors md:shadow-[0_14px_28px_rgba(39,76,155,0.18)]',
              useTransparentHeader ? 'border-white/20' : 'border-white/70',
            )}
            subtitleClassName={cn(
              isSchoolPage
                ? schoolBrandSubtitleClass
                : 'transition-colors',
              !isSchoolPage &&
                (useTopDarkText
                  ? 'text-[#231c18]/72'
                  : useTransparentHeader
                    ? 'text-white/68'
                    : 'text-muted-foreground'),
            )}
            titleClassName={cn(
              isSchoolPage
                ? schoolBrandTitleClass
                : 'transition-colors',
              !isSchoolPage &&
                (useTopDarkText
                  ? 'text-[#231c18]'
                  : useTransparentHeader
                    ? 'text-white'
                    : 'text-primary/75'),
            )}
          />

          <div className="hidden items-center gap-3 md:flex">
            {isSchoolPage ? (
              <Link
                className={cn(
                  buttonVariants({
                    variant: useTransparentHeader ? 'ghost' : 'secondary',
                    size: 'sm',
                  }),
                  isSchoolPage
                    ? schoolNavLinkClass
                    : useTopDarkText
                      ? 'rounded-full bg-transparent text-[#231c18] hover:bg-black/5 hover:text-[#231c18]'
                      : useTransparentHeader
                        ? 'rounded-full bg-transparent text-white/88 hover:bg-white/10 hover:text-white'
                        : 'shadow-none',
                )}
                to="/"
              >
                首页
              </Link>
            ) : null}
            <Link
                className={cn(
                  buttonVariants({
                    variant: useTransparentHeader ? 'ghost' : 'secondary',
                    size: 'sm',
                  }),
                  isSchoolPage
                    ? schoolNavLinkClass
                    : isAllianceHome
                  ? homeAboutLinkClass
                  : useTopDarkText
                    ? 'rounded-full bg-transparent text-[#231c18] hover:bg-black/5 hover:text-[#231c18]'
                    : useTransparentHeader
                      ? 'rounded-full bg-transparent text-white/88 hover:bg-white/10 hover:text-white'
                      : 'shadow-none',
              )}
              to="/about#about-soulink"
            >
              关于Soulink
            </Link>
            {isAllianceHome ? (
              <button
                className={cn(
                  buttonVariants({
                    variant: useTransparentHeader ? 'ghost' : 'secondary',
                    size: 'sm',
                  }),
                  useTopDarkText
                    ? 'rounded-full bg-transparent text-[#231c18] hover:bg-black/5 hover:text-[#231c18]'
                    : useTransparentHeader
                      ? 'rounded-full bg-transparent text-white/88 hover:bg-white/10 hover:text-white'
                      : 'shadow-none',
                )}
                onClick={() => setIsWechatModalOpen(true)}
                type="button"
              >
                加入社群
              </button>
            ) : null}
            <Link
              className={cn(
                buttonVariants({ variant: 'secondary', size: 'sm' }),
                useTransparentHeader
                  ? 'rounded-full border-white/20 bg-white text-[#2c2630] shadow-[0_10px_26px_rgba(34,19,28,0.12)]'
                  : '',
              )}
              to={authActionTarget}
            >
              {authActionLabel}
            </Link>
          </div>

          <button
            aria-label={mobileMenuOpen ? '关闭菜单' : '打开菜单'}
            className={cn(
              'inline-flex h-11 w-11 items-center justify-center rounded-full border md:hidden',
              useTransparentHeader
                ? 'border-white/20 bg-white/10 text-white backdrop-blur-md'
                : 'border-border bg-white/90 text-foreground',
            )}
            onClick={() => setMobileMenuOpen((open) => !open)}
            type="button"
          >
            {mobileMenuOpen ? <X size={18} /> : <Menu size={18} />}
          </button>
        </div>

        {mobileMenuOpen ? (
          <div className="border-t border-white/15 bg-[rgba(32,24,33,0.88)] px-6 py-5 shadow-soft backdrop-blur-xl md:hidden">
            <div className="mx-auto flex max-w-6xl flex-col gap-4">
              <div className="grid grid-cols-2 gap-3 pt-2">
                {isSchoolPage ? (
                  <Link
                    className={cn(
                      buttonVariants({ variant: 'ghost', size: 'sm' }),
                      isSchoolPage
                        ? schoolNavLinkClass
                        : 'text-white/88 hover:bg-white/10 hover:text-white',
                    )}
                    onClick={() => setMobileMenuOpen(false)}
                    to="/"
                  >
                    首页
                  </Link>
                ) : null}
                <Link
                    className={cn(
                      buttonVariants({ variant: 'ghost', size: 'sm' }),
                      isSchoolPage
                        ? schoolNavLinkClass
                        : isAllianceHome
                      ? homeAboutLinkClass
                      : 'text-white/88 hover:bg-white/10 hover:text-white',
                    )}
                  onClick={() => setMobileMenuOpen(false)}
                  to="/about#about-soulink"
                >
                  关于Soulink
                </Link>
                {isAllianceHome ? (
                  <button
                    className={cn(
                      buttonVariants({ variant: 'ghost', size: 'sm' }),
                      'text-white/88 hover:bg-white/10 hover:text-white',
                    )}
                    onClick={() => {
                      setMobileMenuOpen(false)
                      setIsWechatModalOpen(true)
                    }}
                    type="button"
                  >
                    加入社群
                  </button>
                ) : null}
                <Link
                  className={cn(
                    buttonVariants({ variant: 'secondary', size: 'sm' }),
                    'border-white/20 bg-white text-[#2c2630]',
                  )}
                  onClick={() => setMobileMenuOpen(false)}
                  to={authActionTarget}
                >
                  {authActionLabel}
                </Link>
              </div>
            </div>
          </div>
        ) : null}
      </header>

      <WechatGroupModal
        onClose={() => setIsWechatModalOpen(false)}
        open={isWechatModalOpen}
      />
    </>
  )
}
