import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

import { cn } from '@/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-full text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default:
          'bg-primary text-primary-foreground shadow-[0_12px_24px_rgba(39,76,155,0.18)] hover:-translate-y-0.5 hover:bg-[#1f428a]',
        secondary:
          'border border-border bg-white/84 text-secondary-foreground shadow-[0_8px_24px_rgba(15,23,42,0.06)] hover:-translate-y-0.5 hover:bg-white',
        ghost:
          'bg-transparent text-muted-foreground hover:bg-white/70 hover:text-foreground',
      },
      size: {
        default: 'h-11 px-5',
        lg: 'h-12 px-6 text-[15px]',
        sm: 'h-9 px-4 text-sm',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => {
    return (
      <button
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  },
)
Button.displayName = 'Button'

export { Button, buttonVariants }
