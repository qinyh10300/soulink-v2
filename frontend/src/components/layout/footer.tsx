import { useState } from 'react'
import { Link } from 'react-router-dom'

import { ContactUsModal } from '@/components/layout/contact-us-modal'
import { PrivacyModal } from '@/components/layout/privacy-modal'
import { cn } from '@/lib/utils'
import { aboutContactItems } from '@/mock/about'

interface FooterProps {
  className?: string
}

export function Footer({ className }: FooterProps) {
  const [contactModalOpen, setContactModalOpen] = useState(false)
  const [privacyModalOpen, setPrivacyModalOpen] = useState(false)
  const contactEmail = aboutContactItems[0]?.value ?? 'thusoulink@163.com'

  return (
    <>
      <footer
        className={cn(
          'relative z-10 border-t border-[#7f6658] bg-[#41231b] text-[#f5e8df]',
          className,
        )}
      >
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-8 px-6 py-14 lg:flex-row lg:items-end lg:justify-between lg:px-8">
          <div className="space-y-3">
            <p className="font-serif text-4xl text-[#fff6ef]">Soulink</p>
            <p className="text-lg text-[#e8cdbf]">在不同校园里，遇见同频的人</p>
          </div>

          <div className="relative -top-1 flex flex-wrap items-center gap-x-8 gap-y-3 text-base text-[#f1ddd1] lg:justify-end">
            <Link className="transition-colors hover:text-white" to="/">
              首页
            </Link>
            <Link className="transition-colors hover:text-white" to="/about#about-us">
              关于我们
            </Link>
            <button
              className="transition-colors hover:text-white"
              onClick={() => setPrivacyModalOpen(true)}
              type="button"
            >
              隐私协议
            </button>
            <button
              className="transition-colors hover:text-white"
              onClick={() => setContactModalOpen(true)}
              type="button"
            >
              联系我们
            </button>
          </div>
        </div>

        <div className="mx-auto w-full max-w-7xl border-t border-[#6b4132] px-6 py-6 text-sm text-[#d9b7a5] lg:px-8">
          @2026 Soulink Alliance
        </div>
      </footer>

      <PrivacyModal
        onClose={() => setPrivacyModalOpen(false)}
        open={privacyModalOpen}
      />
      <ContactUsModal
        contactEmail={contactEmail}
        onClose={() => setContactModalOpen(false)}
        open={contactModalOpen}
      />
    </>
  )
}
