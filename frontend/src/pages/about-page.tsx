import { useEffect, useState } from 'react'
import {
  ArrowLeft,
  ArrowUpRight,
  Mail,
  MessageSquareQuote,
  Users,
  X,
} from 'lucide-react'
import { createPortal } from 'react-dom'
import { Link } from 'react-router-dom'

import { buttonVariants } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import {
  aboutAllianceParagraphs,
  aboutContactItems,
  aboutStoryParagraphs,
  aboutUsParagraphs,
} from '@/mock/about'

interface WechatGroupModalProps {
  onClose: () => void
  open: boolean
}

function WechatGroupModal({ onClose, open }: WechatGroupModalProps) {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    if (!open) {
      return
    }

    const originalOverflow = document.body.style.overflow

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = originalOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [onClose, open])

  if (!open || !mounted) {
    return null
  }

  return createPortal(
    <div className="fixed inset-0 z-[220] flex items-center justify-center p-4 sm:p-6">
      <button
        aria-label="关闭微信群二维码弹窗"
        className="absolute inset-0 bg-[rgba(46,34,29,0.46)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 w-full max-w-xl rounded-[32px] border border-[rgba(255,255,255,0.62)] bg-[rgba(255,250,247,0.96)] p-6 shadow-[0_28px_72px_rgba(48,35,30,0.22)] sm:p-8">
        <button
          aria-label="关闭"
          className="absolute right-4 top-4 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#eaded4] bg-white/90 text-[#7d6556] transition-colors hover:bg-white"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="space-y-6 text-center">
          <div className="space-y-3">
            <p className="text-sm font-semibold tracking-[0.24em] text-[#9b614a] uppercase">
              Wechat Group
            </p>
            <h3 className="font-serif text-3xl text-[#241d1a] sm:text-4xl">
              Soulink在线交流群
            </h3>
            <p className="mx-auto max-w-md text-base leading-7 text-[#6a5a50]">
              扫码加入微信群，和其他用户一起交流使用体验、产品建议与校园里的新故事
            </p>
          </div>

          <div className="mx-auto flex aspect-square w-full max-w-[18rem] items-center justify-center rounded-[28px] border border-[#eaded4] bg-white p-8 text-sm leading-6 text-[#7d6556] shadow-[0_14px_40px_rgba(88,64,51,0.12)]">
            为保护社群隐私，群二维码不随源码发布。请联系项目团队获取最新二维码。
          </div>
        </div>
      </div>
    </div>,
    document.body,
  )
}

export function AboutPage() {
  const [isWechatModalOpen, setIsWechatModalOpen] = useState(false)

  const renderParagraphs = (
    paragraphs: Array<Array<{ text: string; emphasis?: boolean }>>,
  ) => (
    <div className="space-y-4 text-left text-base leading-8 text-[#584b43] sm:text-[1.05rem] sm:leading-8">
      {paragraphs.map((paragraph, paragraphIndex) => (
        <p key={paragraphIndex} className="indent-[2em]">
          {paragraph.map((part, partIndex) => (
            <span
              className={cn(part.emphasis ? 'font-semibold text-black' : undefined)}
              key={`${paragraphIndex}-${partIndex}`}
            >
              {part.text}
            </span>
          ))}
        </p>
      ))}
    </div>
  )

  return (
    <>
      <div className="bg-[#e7ddd3] px-6 pb-24 pt-32 lg:px-8 lg:pt-36">
        <div className="mx-auto max-w-5xl">
          <section className="py-6 sm:py-10" id="about-soulink">
            <Card className="rounded-[36px] border-white/45 bg-white/52 shadow-[0_16px_50px_rgba(98,72,56,0.08)] backdrop-blur-sm">
              <CardContent className="px-8 py-10 sm:px-12 sm:py-12">
                <div className="mx-auto max-w-3xl space-y-8 text-center">
                  <div className="space-y-3">
                    <p className="text-sm font-semibold tracking-[0.24em] text-[#9b614a] uppercase">
                      About Soulink
                    </p>
                    <h1 className="font-serif text-3xl text-[#241d1a] sm:text-4xl">
                      关于Soulink——故事的开始
                    </h1>
                  </div>

                  {renderParagraphs(aboutStoryParagraphs)}
                </div>
              </CardContent>
            </Card>
          </section>

          <section className="py-6 sm:py-10" id="about-alliance">
            <Card className="rounded-[36px] border-white/45 bg-white/52 shadow-[0_16px_50px_rgba(98,72,56,0.08)] backdrop-blur-sm">
              <CardContent className="px-8 py-10 sm:px-12 sm:py-12">
                <div className="mx-auto max-w-3xl space-y-8 text-center">
                  <div className="space-y-3">
                    <p className="text-sm font-semibold tracking-[0.24em] text-[#9b614a] uppercase">
                      About Alliance
                    </p>
                    <h2 className="font-serif text-3xl text-[#241d1a] sm:text-4xl">
                      为什么我们选择高校联盟
                    </h2>
                  </div>

                  {renderParagraphs(aboutAllianceParagraphs)}
                </div>
              </CardContent>
            </Card>
          </section>

          <section className="py-6 sm:py-10" id="about-us">
            <Card className="rounded-[36px] border-white/45 bg-white/52 shadow-[0_16px_50px_rgba(98,72,56,0.08)] backdrop-blur-sm">
              <CardContent className="px-8 py-10 sm:px-12 sm:py-12">
                <div className="mx-auto max-w-3xl space-y-8 text-center">
                  <div className="space-y-3">
                    <p className="text-sm font-semibold tracking-[0.24em] text-[#9b614a] uppercase">
                      About Us
                    </p>
                    <h2 className="font-serif text-3xl text-[#241d1a] sm:text-4xl">
                      关于我们
                    </h2>
                  </div>

                  {renderParagraphs(aboutUsParagraphs)}
                </div>
              </CardContent>
            </Card>
          </section>

          <section className="py-6 sm:py-10">
            <Card className="rounded-[36px] border-white/45 bg-white/52 shadow-[0_16px_50px_rgba(98,72,56,0.08)] backdrop-blur-sm">
              <CardContent className="px-8 py-10 sm:px-12 sm:py-12">
                <div className="space-y-10 text-center">
                  <div className="space-y-3">
                    <p className="text-sm font-semibold tracking-[0.24em] text-[#9b614a] uppercase">
                      Contact
                    </p>
                    <h2 className="font-serif text-4xl text-[#241d1a] sm:text-5xl">
                      联系我们
                    </h2>
                    <p className="mx-auto max-w-2xl text-lg leading-8 text-[#6a5a50]">
                      如果你对产品有建议、发现问题，或者希望交流想法，欢迎联系或加入我们。
                    </p>
                  </div>

                  <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <Link
                      className="group rounded-[28px] border border-white/55 bg-white/58 px-6 py-7 text-left transition-all duration-200 hover:-translate-y-0.5 hover:bg-white/72"
                      to={aboutContactItems[0].href}
                    >
                      <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#f2e5dc] text-[#8c543e]">
                        <Mail size={20} />
                      </div>
                      <p className="text-sm font-medium tracking-[0.16em] text-[#9b614a] uppercase">
                        {aboutContactItems[0].label}
                      </p>
                      <div className="mt-3 flex items-center justify-between gap-4">
                        <p className="text-xl font-semibold text-[#2f2622]">
                          {aboutContactItems[0].value}
                        </p>
                        <ArrowUpRight
                          className="text-[#8c543e] transition-transform duration-200 group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
                          size={18}
                        />
                      </div>
                    </Link>

                    <Link
                      className="group rounded-[28px] border border-white/55 bg-white/58 px-6 py-7 text-left transition-all duration-200 hover:-translate-y-0.5 hover:bg-white/72"
                      to={aboutContactItems[1].href}
                    >
                      <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#f2e5dc] text-[#8c543e]">
                        <MessageSquareQuote size={20} />
                      </div>
                      <p className="text-sm font-medium tracking-[0.16em] text-[#9b614a] uppercase">
                        {aboutContactItems[1].label}
                      </p>
                      <div className="mt-3 flex items-center justify-between gap-4">
                        <p className="text-xl font-semibold text-[#2f2622]">
                          {aboutContactItems[1].value}
                        </p>
                        <ArrowUpRight
                          className="text-[#8c543e] transition-transform duration-200 group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
                          size={18}
                        />
                      </div>
                    </Link>

                    <button
                      className="group rounded-[28px] border border-white/55 bg-white/58 px-6 py-7 text-left transition-all duration-200 hover:-translate-y-0.5 hover:bg-white/72"
                      onClick={() => setIsWechatModalOpen(true)}
                      type="button"
                    >
                      <div className="mb-5 flex h-12 w-12 items-center justify-center rounded-2xl bg-[#f2e5dc] text-[#8c543e]">
                        <Users size={20} />
                      </div>
                      <p className="text-sm font-medium tracking-[0.16em] text-[#9b614a] uppercase">
                        微信社群
                      </p>
                      <div className="mt-3 flex items-center justify-between gap-4">
                        <div className="space-y-1">
                          <p className="text-xl font-semibold text-[#2f2622]">
                            Soulink在线交流群
                          </p>
                          <p className="text-sm leading-6 text-[#6a5a50]">
                            点击查看群二维码
                          </p>
                        </div>
                        <ArrowUpRight
                          className="shrink-0 text-[#8c543e] transition-transform duration-200 group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
                          size={18}
                        />
                      </div>
                    </button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>

          <div className="flex justify-center pt-6 sm:pt-8">
            <Link
              className={cn(
                buttonVariants({ variant: 'secondary', size: 'lg' }),
                'gap-2 rounded-full border-white/50 bg-white/58 px-7 text-[#2f2622] shadow-[0_12px_30px_rgba(98,72,56,0.08)] hover:bg-white/76',
              )}
              to="/"
            >
              <ArrowLeft size={16} />
              返回首页
            </Link>
          </div>
        </div>
      </div>

      <WechatGroupModal onClose={() => setIsWechatModalOpen(false)} open={isWechatModalOpen} />
    </>
  )
}
