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
          'flex h-11 w-full appearance-none rounded-md border border-[#d9d5cf] bg-white px-3 pr-10 text-[15px] text-[#302e2b] outline-none transition-shadow focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDown
        className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[#8d8982]"
        size={18}
      />
    </div>
  )
})

Select.displayName = 'Select'

export { Select }
