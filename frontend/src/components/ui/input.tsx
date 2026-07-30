import * as React from 'react'

import { cn } from '@/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type = 'text', ...props }, ref) => {
    return (
      <input
        ref={ref}
        type={type}
        className={cn(
          'flex h-11 w-full rounded-md border border-[#d9d5cf] bg-white px-3 text-[15px] text-[#302e2b] outline-none transition-shadow placeholder:text-[#9a958d] focus:border-[#7a5a8e] focus:ring-3 focus:ring-[#5b3477]/10 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      />
    )
  },
)

Input.displayName = 'Input'

export { Input }
