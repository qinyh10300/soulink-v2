import { useEffect } from 'react'

import soulinkWordmark from '@/assets/soulink-interactive-white.svg'
import { schoolHeroImages } from '@/lib/school-image-assets'
import { LandingPage } from '@/pages/landing-page'

type SingaporeSchoolPageProps = {
  schoolKey: 'nus' | 'ntu' | 'fdu' | 'sjdu'
}

export function SingaporeSchoolPage({ schoolKey }: SingaporeSchoolPageProps) {
  useEffect(() => {
    const preloadTargets = [schoolHeroImages.thu.webp, soulinkWordmark]
    const appendedLinks: HTMLLinkElement[] = []

    preloadTargets.forEach((href) => {
      if (document.head.querySelector(`link[rel="preload"][href="${href}"]`)) {
        return
      }

      const preloadLink = document.createElement('link')
      preloadLink.rel = 'preload'
      preloadLink.as = 'image'
      preloadLink.href = href
      document.head.appendChild(preloadLink)
      appendedLinks.push(preloadLink)
    })

    return () => {
      appendedLinks.forEach((link) => link.remove())
    }
  }, [schoolKey])

  return <LandingPage schoolKey="thu" />
}
