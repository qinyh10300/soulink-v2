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
          'flex h-11 w-full appearance-none rounded-md border border-[#d4d4d8] bg-white px-3 pr-10 text-[15px] text-[#18181b] outline-none transition-shadow focus:border-[#1447e6] focus:ring-3 focus:ring-[#1447e6]/10 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDown
        className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[#71717a]"
        size={18}
      />
    </div>
  )
})

Select.displayName = 'Select'

export { Select }
