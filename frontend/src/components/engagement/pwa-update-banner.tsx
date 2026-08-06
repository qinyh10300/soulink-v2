import { RefreshCw, X } from 'lucide-react'

import { useEngagement } from '@/contexts/engagement-context'

export function PWAUpdateBanner() {
  const { updateAvailable, applyUpdate, dismissUpdate } = useEngagement()
  if (!updateAvailable) return null

  return (
    <div className="fixed inset-x-3 bottom-[calc(4.75rem_+_env(safe-area-inset-bottom))] z-[85] mx-auto flex max-w-xl items-center gap-3 rounded-md border border-[#cdd8d3] bg-[#f4f8f6] p-3 shadow-[0_14px_40px_rgba(35,33,31,0.16)] lg:bottom-5">
      <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-[#315c4d] text-white">
        <RefreshCw size={16} />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block text-sm font-semibold text-[#35433e]">SouLink 已有新版本</span>
        <span className="mt-0.5 block text-xs text-[#6d7873]">由你决定何时刷新当前页面。</span>
      </span>
      <button
        className="inline-flex h-9 items-center rounded-md bg-[#263b35] px-3 text-xs font-semibold text-white hover:bg-[#1d302a]"
        onClick={applyUpdate}
        type="button"
      >
        立即更新
      </button>
      <button
        aria-label="稍后更新"
        className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-[#77736c] hover:bg-white"
        onClick={dismissUpdate}
        title="稍后"
        type="button"
      >
        <X size={16} />
      </button>
    </div>
  )
}
