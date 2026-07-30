import type { ReactNode } from 'react'

interface SectionHeadingProps {
  eyebrow: string
  title: string
  description: string
  align?: 'left' | 'center'
  action?: ReactNode
}

export function SectionHeading({
  eyebrow,
  title,
  description,
  align = 'left',
  action,
}: SectionHeadingProps) {
  const centered = align === 'center'

  return (
    <div
      className={
        centered
          ? 'mx-auto max-w-2xl space-y-4 text-center'
          : 'max-w-2xl space-y-4'
      }
    >
      <p className="text-sm font-semibold tracking-[0.2em] text-primary/70 uppercase">
        {eyebrow}
      </p>
      <div className="space-y-3">
        <h2 className="text-balance font-serif text-3xl leading-tight text-foreground sm:text-4xl">
          {title}
        </h2>
        <p className="text-balance text-base leading-7 text-muted-foreground sm:text-lg">
          {description}
        </p>
      </div>
      {action ? <div>{action}</div> : null}
    </div>
  )
}
