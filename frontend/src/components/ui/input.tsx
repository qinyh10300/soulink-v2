import * as React from 'react'

import { cn } from '@/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type = 'text', ...props }, ref) => {
    return (
      <input
        ref={ref}
        type={type}
        className={cn(
          'flex h-11 w-full rounded-md border border-[#d4d4d8] bg-white px-3 text-[15px] text-[#18181b] outline-none transition-shadow placeholder:text-[#a1a1aa] focus:border-[#1447e6] focus:ring-3 focus:ring-[#1447e6]/10 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      />
    )
  },
)

Input.displayName = 'Input'

export { Input }
