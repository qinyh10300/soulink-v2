import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { createPortal } from 'react-dom'

interface WechatGroupModalProps {
  onClose: () => void
  open: boolean
}

export function WechatGroupModal({
  onClose,
  open,
}: WechatGroupModalProps) {
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
              THU-SouLink交流群
            </h3>
            <p className="mx-auto max-w-md text-base leading-7 text-[#6a5a50]">
              扫码加入微信群，和清华校内用户一起交流使用体验、产品建议与校园里的新故事
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
