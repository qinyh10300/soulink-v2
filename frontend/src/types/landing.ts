import type { LucideIcon } from 'lucide-react'

export interface NavItem {
  label: string
  href: string
  kind?: 'anchor' | 'route'
}

export interface HeroStat {
  label: string
  value: string
}

export interface FeatureItem {
  title: string
  description: string
  icon: LucideIcon
}

export interface FlowStep {
  title: string
  description: string
}

export interface FaqItem {
  question: string
  answer: string
}
