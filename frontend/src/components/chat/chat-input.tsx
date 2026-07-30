import { Send } from 'lucide-react'
import { useState } from 'react'

import { Button } from '@/components/ui/button'

interface Props {
  onSend: (content: string) => void
  disabled?: boolean
}

export function ChatInput({ onSend, disabled }: Props) {
  const [text, setText] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = text.trim()
    if (!trimmed) return
    onSend(trimmed)
    setText('')
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit(e)
    }
  }

  return (
    <form
      className="flex items-end gap-2 border-t border-[#eee4db] bg-[rgba(255,251,247,0.96)] px-4 py-3"
      onSubmit={handleSubmit}
    >
      <textarea
        className="max-h-24 min-h-[40px] flex-1 resize-none rounded-2xl border border-[#eee4db] bg-white px-4 py-2.5 text-sm text-[#2f221b] outline-none transition-colors placeholder:text-[#7e6d61]/50 focus:border-[#c7b09e]"
        disabled={disabled}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="输入消息..."
        rows={1}
        value={text}
      />
      <Button
        className="h-10 w-10 shrink-0 rounded-full p-0"
        disabled={disabled || !text.trim()}
        type="submit"
      >
        <Send className="h-4 w-4" />
      </Button>
    </form>
  )
}
