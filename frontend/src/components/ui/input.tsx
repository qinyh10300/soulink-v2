import * as React from 'react'

import { cn } from '@/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type = 'text', ...props }, ref) => {
    return (
      <input
        ref={ref}
        type={type}
        className={cn(
          'flex h-13 w-full rounded-2xl border border-[#ded3ca] bg-white/88 px-4 text-[15px] text-[#2d2623] shadow-[0_6px_16px_rgba(95,70,55,0.04)] outline-none transition-all placeholder:text-[#b4a59a] focus:border-[#b98b68] focus:ring-4 focus:ring-[#b98b68]/12 disabled:cursor-not-allowed disabled:opacity-60',
          className,
        )}
        {...props}
      />
    )
  },
)

Input.displayName = 'Input'

export { Input }
