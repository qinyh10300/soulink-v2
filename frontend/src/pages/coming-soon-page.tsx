import { ArrowLeft } from 'lucide-react'
import { Link } from 'react-router-dom'

import { buttonVariants } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

interface ComingSoonPageProps {
  title: string
  description: string
}

export function ComingSoonPage({ title, description }: ComingSoonPageProps) {
  return (
    <section className="px-6 pb-24 pt-36 lg:px-8">
      <div className="mx-auto max-w-3xl">
        <Card className="border-white/70 bg-white/82 shadow-soft">
          <CardContent className="space-y-6 p-8 text-center sm:p-10">
            <p className="text-sm font-semibold tracking-[0.2em] text-primary/70 uppercase">
              Roadmap
            </p>
            <h1 className="font-serif text-4xl text-foreground">{title}</h1>
            <p className="text-base leading-7 text-muted-foreground">{description}</p>
            <Link className={cn(buttonVariants({ size: 'lg' }), 'gap-2')} to="/">
              <ArrowLeft size={16} />
              返回首页
            </Link>
          </CardContent>
        </Card>
      </div>
    </section>
  )
}
