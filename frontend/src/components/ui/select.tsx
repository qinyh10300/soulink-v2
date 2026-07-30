import * as React from 'react'
import { ChevronDown } from 'lucide-react'

import { cn } from '@/lib/utils'

const Select = React.forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement>
>(({ children, className, ...props }, ref) => {
  return (
    <div className="relative">
      <select
        ref={ref}
        className={cn(
          'flex h-13 w-full appearance-none rounded-2xl border border-[#ded3ca] bg-white/88 px-4 pr-12 text-[15px] text-[#2d2623] shadow-[0_6px_16px_rgba(95,70,55,0.04)] outline-none transition-all focus:border-[#b98b68] focus:ring-4 focus:ring-[#b98b68]/12 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDown
        className="pointer-events-none absolute right-4 top-1/2 -translate-y-1/2 text-[#b4a59a]"
        size={18}
      />
    </div>
  )
})

Select.displayName = 'Select'

export { Select }
