import { AlertTriangle, Send, ShieldAlert, X } from 'lucide-react'

import type { MessageSafetyAssessment } from '@/lib/chat-api'

const categoryLabels: Record<MessageSafetyAssessment['categories'][number], string> = {
  spam: '垃圾推广',
  contact_inducement: '联系方式诱导',
  harassment: '骚扰或冒犯',
  threat: '威胁表达',
  privacy: '隐私风险',
  coercion: '强迫或施压',
}

export function MessageRiskReviewDialog({
  assessment,
  content,
  sending,
  onCancel,
  onConfirm,
}: {
  assessment: MessageSafetyAssessment
  content: string
  sending: boolean
  onCancel: () => void
  onConfirm: () => void
}) {
  return (
    <div className="absolute inset-0 z-40 flex items-end justify-center bg-black/25 p-0 sm:items-center sm:p-5">
      <section
        aria-modal="true"
        className="w-full max-w-md border border-[#ddd8d2] bg-white shadow-xl sm:rounded-md"
        role="dialog"
      >
        <header className="flex items-center border-b border-[#e4e1dc] px-4 py-3.5">
          {assessment.level === 'high' ? (
            <ShieldAlert className="h-5 w-5 text-[#a24b43]" />
          ) : (
            <AlertTriangle className="h-5 w-5 text-[#9a6a36]" />
          )}
          <h2 className="ml-2 text-sm font-semibold text-[#34312e]">发送前重新检查</h2>
          <button
            aria-label="关闭风险复核"
            className="ml-auto inline-flex h-8 w-8 items-center justify-center rounded text-[#77736d] hover:bg-[#f0eeea]"
            disabled={sending}
            onClick={onCancel}
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        </header>

        <div className="px-4 py-4">
          <div className="flex flex-wrap gap-1.5">
            {assessment.categories.map((category) => (
              <span className="rounded bg-[#f6ebe7] px-2 py-1 text-[11px] font-semibold text-[#8b4e43]" key={category}>
                {categoryLabels[category]}
              </span>
            ))}
          </div>
          <ul className="mt-3 space-y-1.5 text-xs leading-5 text-[#605b55]">
            {assessment.reasons.map((reason) => <li key={reason}>• {reason}</li>)}
          </ul>
          <div className="mt-4 max-h-28 overflow-y-auto rounded border border-[#e2ded8] bg-[#faf9f7] px-3 py-2 text-sm leading-6 text-[#494641]">
            {content}
          </div>
          <p className="mt-2 text-[10px] text-[#92908b]">
            {assessment.inferenceSource === 'model' ? 'AI 与规则复核' : '本地规则复核'} · {Math.round(assessment.confidence * 100)}%
          </p>
        </div>

        <footer className="grid grid-cols-2 gap-2 border-t border-[#e4e1dc] px-4 py-3">
          <button
            className="h-9 rounded border border-[#d5d1cb] text-xs font-semibold text-[#5f5a54] hover:bg-[#f1efeb] disabled:opacity-50"
            disabled={sending}
            onClick={onCancel}
            type="button"
          >
            返回修改
          </button>
          <button
            className="inline-flex h-9 items-center justify-center gap-1.5 rounded bg-[#8d4a43] text-xs font-semibold text-white hover:bg-[#793e38] disabled:opacity-50"
            disabled={sending}
            onClick={onConfirm}
            type="button"
          >
            <Send className="h-3.5 w-3.5" />
            仍然发送
          </button>
        </footer>
      </section>
    </div>
  )
}
