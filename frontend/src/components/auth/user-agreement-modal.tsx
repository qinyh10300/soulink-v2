import { useEffect } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { agreementMeta, agreementSections } from '@/mock/agreement'
import { cn } from '@/lib/utils'

interface UserAgreementModalProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
}

export function UserAgreementModal({
  open,
  onClose,
  onConfirm,
}: UserAgreementModalProps) {
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

  if (!open || typeof document === 'undefined') {
    return null
  }

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
      <button
        aria-label="关闭用户协议"
        className="absolute inset-0 bg-[rgba(46,33,27,0.34)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[min(90vh,52rem)] w-full max-w-3xl flex-col overflow-hidden rounded-[32px] border border-[#eadfd5] bg-[#fcf8f3] shadow-[0_28px_90px_rgba(66,46,35,0.18)]">
        <div className="flex items-start justify-between gap-6 border-b border-[#eee2d8] px-6 py-5 sm:px-8">
          <div className="space-y-2">
            <p className="text-sm font-semibold tracking-[0.18em] text-[#9c6b48] uppercase">
              {agreementMeta.englishTitle}
            </p>
            <h2 className="font-serif text-3xl text-[#271f1b]">{agreementMeta.title}</h2>
          </div>

          <button
            aria-label="关闭"
            className="inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#e6d9cf] bg-white text-[#755947] transition-colors hover:bg-[#f7f0ea]"
            onClick={onClose}
            type="button"
          >
            <X size={18} />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-6 sm:px-8">
          <div className="space-y-6">
            <section className="space-y-3">
              {agreementMeta.intro.map((paragraph) => (
                <p className="text-[15px] leading-8 text-[#67584d]" key={paragraph}>
                  {paragraph}
                </p>
              ))}
            </section>

            {agreementSections.map((section) => (
              <section className="space-y-3" key={section.title}>
                <h3 className="text-lg font-semibold text-[#2d2521]">{section.title}</h3>

                {'paragraphs' in section && section.paragraphs ? (
                  <div className="space-y-3">
                    {section.paragraphs.map((paragraph) => (
                      <p className="text-[15px] leading-8 text-[#67584d]" key={paragraph}>
                        {paragraph}
                      </p>
                    ))}
                  </div>
                ) : null}

                {'blocks' in section && section.blocks ? (
                  <div className="space-y-4">
                    {section.blocks.map((block) => (
                      <div className="space-y-3" key={block.title}>
                        <h4 className="text-base font-medium text-[#3b3029]">{block.title}</h4>
                        {block.paragraphs.map((paragraph) => (
                          <p className="text-[15px] leading-8 text-[#67584d]" key={paragraph}>
                            {paragraph}
                          </p>
                        ))}
                        {block.list ? (
                          <ul className="space-y-2 pl-5 text-[15px] leading-8 text-[#67584d]">
                            {block.list.map((item) => (
                              <li className="list-disc" key={item}>
                                {item}
                              </li>
                            ))}
                          </ul>
                        ) : null}
                        {block.closing ? (
                          <p className="text-[15px] leading-8 text-[#67584d]">{block.closing}</p>
                        ) : null}
                      </div>
                    ))}
                  </div>
                ) : null}

                {'list' in section && section.list ? (
                  <ul className="space-y-2 pl-5 text-[15px] leading-8 text-[#67584d]">
                    {section.list.map((item) => (
                      <li className="list-disc" key={item}>
                        {item}
                      </li>
                    ))}
                  </ul>
                ) : null}

                {'closing' in section && section.closing ? (
                  <p className="text-[15px] leading-8 text-[#67584d]">{section.closing}</p>
                ) : null}
              </section>
            ))}
          </div>
        </div>

        <div className="border-t border-[#eee2d8] px-6 py-5 sm:px-8">
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            onClick={onConfirm}
            type="button"
          >
            已阅读，同意协议
          </button>
        </div>
      </div>
    </div>,
    document.body,
  )
}
