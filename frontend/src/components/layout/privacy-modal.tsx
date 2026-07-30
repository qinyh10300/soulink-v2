import { useEffect } from 'react'
import { ShieldCheck, X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { privacyPolicyMeta, privacyPolicySections } from '@/mock/privacy'

interface PrivacyModalProps {
  open: boolean
  onClose: () => void
}

export function PrivacyModal({ open, onClose }: PrivacyModalProps) {
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

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center p-4 sm:p-6">
      <button
        aria-label="关闭隐私说明弹窗"
        className="absolute inset-0 bg-[rgba(48,34,28,0.4)] backdrop-blur-[5px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[min(88vh,52rem)] w-full max-w-3xl flex-col overflow-hidden rounded-[32px] border border-[#eadfd5] bg-[#fdf9f4] shadow-[0_32px_88px_rgba(59,39,28,0.22)]">
        <button
          aria-label="关闭"
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#eadfd5] bg-white/86 text-[#8b7160] transition-colors hover:bg-white"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="border-b border-[#efe5dc] px-7 pb-6 pt-8 sm:px-10">
          <div className="flex items-center gap-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[linear-gradient(135deg,#f6ddd3,#f0c4bb)] text-[#8f5d47] shadow-[0_12px_24px_rgba(143,93,71,0.14)]">
              <ShieldCheck size={22} />
            </div>
            <div className="min-w-0">
              <h2 className="font-serif text-3xl text-[#2f221b] sm:text-4xl">
                {privacyPolicyMeta.title}
              </h2>
              <p className="mt-1 text-sm text-[#8a7667]">{privacyPolicyMeta.updatedAt}</p>
            </div>
          </div>
          <div className="mt-5 space-y-3 text-[15px] leading-7 text-[#5d5048] sm:text-base">
            {privacyPolicyMeta.intro.map((paragraph) => (
              <p key={paragraph}>{paragraph}</p>
            ))}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-7 py-7 sm:px-10">
          <div className="space-y-7">
            {privacyPolicySections.map((section) => (
              <section
                className="rounded-[24px] border border-[#efe4db] bg-[rgba(255,255,255,0.72)] px-5 py-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)] sm:px-6"
                key={section.title}
              >
                <h3 className="font-serif text-2xl text-[#2f221b]">{section.title}</h3>

                {'blocks' in section && section.blocks ? (
                  <div className="mt-4 space-y-4 text-[15px] leading-7 text-[#5d5048] sm:text-base">
                    {section.blocks.map((block) => (
                      <div key={block.heading}>
                        <p className="font-medium text-[#3e3028]">{block.heading}</p>
                        <p className="mt-1">{block.text}</p>
                      </div>
                    ))}
                  </div>
                ) : null}

                {'paragraphs' in section && section.paragraphs ? (
                  <div className="mt-4 space-y-3 text-[15px] leading-7 text-[#5d5048] sm:text-base">
                    {section.paragraphs.map((paragraph) => (
                      <p key={paragraph}>{paragraph}</p>
                    ))}
                  </div>
                ) : null}

                {'list' in section && section.list ? (
                  <ul className="mt-4 space-y-2 pl-5 text-[15px] leading-7 text-[#5d5048] sm:text-base">
                    {section.list.map((item) => (
                      <li className="list-disc" key={item}>
                        {item}
                      </li>
                    ))}
                  </ul>
                ) : null}

                {'closing' in section && section.closing ? (
                  <p className="mt-4 text-[15px] leading-7 text-[#5d5048] sm:text-base">
                    {section.closing}
                  </p>
                ) : null}

                {section.title === '9. 联系我们' ? (
                  <p className="mt-4 text-[15px] leading-7 text-[#5d5048] sm:text-base">
                    <span className="font-medium text-[#3e3028]">联系邮箱：</span>
                    <a
                      className="ml-2 text-[#8f6548] underline decoration-[#d4b5a0] underline-offset-4 hover:text-[#744d35]"
                      href={`mailto:${privacyPolicyMeta.contactEmail}`}
                    >
                      {privacyPolicyMeta.contactEmail}
                    </a>
                  </p>
                ) : null}
              </section>
            ))}
          </div>
        </div>

        <div className="border-t border-[#efe5dc] px-7 py-5 sm:px-10">
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_30px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            onClick={onClose}
            type="button"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}
