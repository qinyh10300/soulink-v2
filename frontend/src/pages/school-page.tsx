import { useEffect } from 'react'
import { Navigate, useParams } from 'react-router-dom'

import soulinkWordmark from '@/assets/soulink-interactive-white.svg'
import { schoolHeroImages } from '@/lib/school-image-assets'
import { isSchoolKey, type SchoolKey } from '@/lib/school-keys'
import { LandingPage } from '@/pages/landing-page'

const schoolHeroPreloadImages: Record<SchoolKey, string> = {
  thu: schoolHeroImages.thu.webp,
  pku: schoolHeroImages.pku.webp,
  bnu: schoolHeroImages.bnu.webp,
  ruc: schoolHeroImages.ruc.webp,
  cuc: schoolHeroImages.cuc.webp,
  bfsu: schoolHeroImages.bfsu.webp,
}

export function SchoolPage() {
  const { schoolKey } = useParams()
  const resolvedSchoolKey = isSchoolKey(schoolKey) ? schoolKey : null

  useEffect(() => {
    if (!resolvedSchoolKey) {
      return
    }

    const preloadTargets = [schoolHeroPreloadImages[resolvedSchoolKey], soulinkWordmark]
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
  }, [resolvedSchoolKey])

  if (!resolvedSchoolKey) {
    return <Navigate replace to="/" />
  }

  return <LandingPage schoolKey={resolvedSchoolKey} />
}
