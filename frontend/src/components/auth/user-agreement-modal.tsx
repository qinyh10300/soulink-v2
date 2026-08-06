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
        className="absolute inset-0 bg-black/45 backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[min(90vh,52rem)] w-full max-w-3xl flex-col overflow-hidden rounded-lg border border-[#e4e4e7] bg-white shadow-[0_28px_90px_rgba(9,9,11,0.22)]">
        <div className="flex items-start justify-between gap-6 border-b border-[#e4e4e7] px-6 py-5 sm:px-8">
          <div className="space-y-2">
            <p className="text-xs font-semibold text-[#1447e6] uppercase">
              {agreementMeta.englishTitle}
            </p>
            <h2 className="text-3xl font-semibold text-[#09090b]">{agreementMeta.title}</h2>
          </div>

          <button
            aria-label="关闭"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-[#e4e4e7] bg-white text-[#71717a] transition-colors hover:bg-[#f4f4f5]"
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
                <p className="text-[15px] leading-8 text-[#52525b]" key={paragraph}>
                  {paragraph}
                </p>
              ))}
            </section>

            {agreementSections.map((section) => (
              <section className="space-y-3" key={section.title}>
                <h3 className="text-lg font-semibold text-[#18181b]">{section.title}</h3>

                {'paragraphs' in section && section.paragraphs ? (
                  <div className="space-y-3">
                    {section.paragraphs.map((paragraph) => (
                      <p className="text-[15px] leading-8 text-[#52525b]" key={paragraph}>
                        {paragraph}
                      </p>
                    ))}
                  </div>
                ) : null}

                {'blocks' in section && section.blocks ? (
                  <div className="space-y-4">
                    {section.blocks.map((block) => (
                      <div className="space-y-3" key={block.title}>
                        <h4 className="text-base font-medium text-[#27272a]">{block.title}</h4>
                        {block.paragraphs.map((paragraph) => (
                          <p className="text-[15px] leading-8 text-[#52525b]" key={paragraph}>
                            {paragraph}
                          </p>
                        ))}
                        {block.list ? (
                          <ul className="space-y-2 pl-5 text-[15px] leading-8 text-[#52525b]">
                            {block.list.map((item) => (
                              <li className="list-disc" key={item}>
                                {item}
                              </li>
                            ))}
                          </ul>
                        ) : null}
                        {block.closing ? (
                          <p className="text-[15px] leading-8 text-[#52525b]">{block.closing}</p>
                        ) : null}
                      </div>
                    ))}
                  </div>
                ) : null}

                {'list' in section && section.list ? (
                  <ul className="space-y-2 pl-5 text-[15px] leading-8 text-[#52525b]">
                    {section.list.map((item) => (
                      <li className="list-disc" key={item}>
                        {item}
                      </li>
                    ))}
                  </ul>
                ) : null}

                {'closing' in section && section.closing ? (
                  <p className="text-[15px] leading-8 text-[#52525b]">{section.closing}</p>
                ) : null}
              </section>
            ))}
          </div>
        </div>

        <div className="border-t border-[#e4e4e7] px-6 py-5 sm:px-8">
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full rounded-md bg-[#1447e6] text-white shadow-[0_12px_28px_rgba(20,71,230,0.18)] hover:bg-[#103bc4]',
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
