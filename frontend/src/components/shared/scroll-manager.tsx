import { useEffect, useLayoutEffect } from 'react'
import { useLocation } from 'react-router-dom'

export function ScrollManager() {
  const { hash, pathname } = useLocation()

  useLayoutEffect(() => {
    const previousScrollRestoration = window.history.scrollRestoration
    window.history.scrollRestoration = 'manual'

    return () => {
      window.history.scrollRestoration = previousScrollRestoration
    }
  }, [])

  useEffect(() => {
    if (hash) {
      const id = hash.replace('#', '')
      const element = document.getElementById(id)

      if (element) {
        window.setTimeout(() => {
          element.scrollIntoView({ behavior: 'smooth', block: 'start' })
        }, 80)
      }

      return
    }

    const resetScroll = () => window.scrollTo({ top: 0, behavior: 'auto' })
    resetScroll()

    const animationFrame = window.requestAnimationFrame(resetScroll)
    const timeout = window.setTimeout(resetScroll, 120)

    return () => {
      window.cancelAnimationFrame(animationFrame)
      window.clearTimeout(timeout)
    }
  }, [hash, pathname])

  return null
}
