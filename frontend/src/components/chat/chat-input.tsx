import { Send } from 'lucide-react'
import { useState } from 'react'

interface Props {
  onSend: (content: string) => boolean | Promise<boolean>
  disabled?: boolean
  initialText?: string
}

export function ChatInput({ onSend, disabled, initialText = '' }: Props) {
  const [text, setText] = useState(initialText)
  const [sending, setSending] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = text.trim()
    if (!trimmed || sending) return
    setSending(true)
    try {
      if (await onSend(trimmed)) setText('')
    } finally {
      setSending(false)
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      void handleSubmit(e)
    }
  }

  return (
    <form
      className="flex items-end gap-2 border-t border-[#dfe1de] bg-[#f7f8f6] px-4 py-3"
      onSubmit={handleSubmit}
    >
      <textarea
        className="max-h-24 min-h-10 flex-1 resize-none rounded-md border border-[#d4d7d3] bg-white px-3 py-2.5 text-sm text-[#292b28] outline-none transition-colors placeholder:text-[#91958f] focus:border-[#559268]"
        disabled={disabled || sending}
        maxLength={4000}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="输入消息..."
        rows={1}
        value={text}
      />
      <button
        aria-label="发送消息"
        className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md bg-[#2f7543] text-white hover:bg-[#28653a] disabled:bg-[#b8c5bb]"
        disabled={disabled || sending || !text.trim()}
        title="发送"
        type="submit"
      >
        <Send className="h-4 w-4" />
      </button>
    </form>
  )
}
