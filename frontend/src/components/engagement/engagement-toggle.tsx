import { cn } from '@/lib/utils'

interface EngagementToggleProps {
  checked: boolean
  disabled?: boolean
  label: string
  description?: string
  onChange: (checked: boolean) => void
}

export function EngagementToggle({
  checked,
  disabled = false,
  label,
  description,
  onChange,
}: EngagementToggleProps) {
  return (
    <div className="flex min-h-12 items-center gap-3 py-2">
      <span className="min-w-0 flex-1">
        <span className="block text-sm font-medium text-[#45413d]">{label}</span>
        {description ? (
          <span className="mt-0.5 block text-xs leading-5 text-[#817d76]">
            {description}
          </span>
        ) : null}
      </span>
      <button
        aria-checked={checked}
        aria-label={label}
        className={cn(
          'relative h-7 w-12 shrink-0 rounded-full border transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#527b6d] disabled:cursor-not-allowed disabled:opacity-50',
          checked
            ? 'border-[#315c4d] bg-[#315c4d]'
            : 'border-[#c9c5be] bg-[#e8e5df]',
        )}
        disabled={disabled}
        onClick={() => onChange(!checked)}
        role="switch"
        type="button"
      >
        <span
          className={cn(
            'absolute top-0.5 h-5.5 w-5.5 rounded-full bg-white shadow-sm transition-transform',
            checked ? 'translate-x-5.5' : 'translate-x-0.5',
          )}
        />
      </button>
    </div>
  )
}
