import { createPortal } from 'react-dom'

interface FloatingToastProps {
  message: string
}

export function FloatingToast({ message }: FloatingToastProps) {
  if (!message || typeof document === 'undefined') {
    return null
  }

  return createPortal(
    <div className="pointer-events-none fixed inset-x-0 top-[max(0.25rem,env(safe-area-inset-top))] z-[70] flex justify-center px-4">
      <div className="rounded-full border border-[#e4d5c9] bg-[#fff8f2] px-5 py-2.5 text-sm text-[#7b5f4f] shadow-[0_10px_24px_rgba(71,49,35,0.16)]">
        {message}
      </div>
    </div>,
    document.body,
  )
}

