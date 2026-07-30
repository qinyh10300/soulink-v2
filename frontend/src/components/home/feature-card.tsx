import { ArrowUpRight } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import type { FeatureItem } from '@/types/landing'

interface FeatureCardProps {
  feature: FeatureItem
}

export function FeatureCard({ feature }: FeatureCardProps) {
  const Icon = feature.icon

  return (
    <Card className="group h-full border-white/70 bg-white/76 shadow-card transition-all duration-300 hover:-translate-y-1.5 hover:border-white hover:shadow-soft">
      <CardContent className="flex h-full flex-col gap-5 p-7">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-accent text-primary">
          <Icon size={24} />
        </div>
        <div className="space-y-3">
          <div className="flex items-center justify-between gap-4">
            <h3 className="text-xl font-semibold text-foreground">{feature.title}</h3>
            <ArrowUpRight
              className="text-muted-foreground transition-transform duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5"
              size={18}
            />
          </div>
          <p className="text-sm leading-7 text-muted-foreground">
            {feature.description}
          </p>
        </div>
      </CardContent>
    </Card>
  )
}
