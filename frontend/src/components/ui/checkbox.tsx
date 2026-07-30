import { Check } from 'lucide-react'

import { cn } from '@/lib/utils'

interface CheckboxProps {
  checked: boolean
  className?: string
  disabled?: boolean
  onCheckedChange?: (checked: boolean) => void
}

export function Checkbox({
  checked,
  className,
  disabled = false,
  onCheckedChange,
}: CheckboxProps) {
  return (
    <button
      aria-checked={checked}
      className={cn(
        'inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md border border-[#bdb8b1] bg-white text-[#5b3477] transition-colors focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-[#5b3477]/12 disabled:cursor-not-allowed disabled:opacity-60',
        checked && 'border-[#76528a] bg-[#eee8f2]',
        className,
      )}
      disabled={disabled}
      onClick={() => onCheckedChange?.(!checked)}
      role="checkbox"
      type="button"
    >
      {checked ? <Check size={14} strokeWidth={3} /> : null}
    </button>
  )
}
