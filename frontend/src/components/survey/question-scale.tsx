import { cn } from '@/lib/utils'

interface QuestionScaleProps {
  min: number
  max: number
  leftLabel: string
  rightLabel: string
  value: number | null
  onChange: (value: number) => void
  disabled?: boolean
}

export function QuestionScale({
  min,
  max,
  leftLabel,
  rightLabel,
  value,
  onChange,
  disabled = false,
}: QuestionScaleProps) {
  const items = Array.from({ length: max - min + 1 }, (_, index) => min + index)
  const gridTemplateColumns = `repeat(${items.length}, minmax(0, 1fr))`

  return (
    <div className="space-y-1.5">
      <div
        className="grid w-full items-center gap-1 sm:gap-2"
        style={{ gridTemplateColumns }}
      >
        {items.map((item) => {
          const active = value === item

          return (
            <button
              className={cn(
                'justify-self-center flex h-8 w-8 items-center justify-center rounded-full border text-[13px] font-semibold transition-all sm:h-9 sm:w-9 sm:text-[14px]',
                active
                  ? 'border-[#8f6548] bg-[#8f6548] text-white shadow-[0_14px_28px_rgba(109,76,54,0.22)]'
                  : 'border-[#dfd3c9] bg-white text-[#8c7b70] hover:border-[#ba8f68] hover:text-[#7e5e4a]',
                disabled ? 'cursor-not-allowed opacity-60 hover:border-[#dfd3c9] hover:text-[#8c7b70]' : '',
              )}
              disabled={disabled}
              key={item}
              onClick={() => onChange(item)}
              type="button"
            >
              {item}
            </button>
          )
        })}
      </div>

      <div
        className="grid items-start gap-1 text-[11px] leading-[1rem] text-[#8d7b70] sm:gap-2 sm:text-[12px] sm:leading-[1.15rem]"
        style={{ gridTemplateColumns }}
      >
        <p className="col-start-1 justify-self-center max-w-[6rem] text-center sm:max-w-[7rem]">
          {leftLabel}
        </p>
        <p
          className="justify-self-center max-w-[6rem] text-center sm:max-w-[7rem]"
          style={{ gridColumnStart: items.length }}
        >
          {rightLabel}
        </p>
      </div>
    </div>
  )
}
