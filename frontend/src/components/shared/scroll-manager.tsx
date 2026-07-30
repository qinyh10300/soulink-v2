import { useEffect } from 'react'
import { useLocation } from 'react-router-dom'

export function ScrollManager() {
  const { hash, pathname } = useLocation()

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

    window.scrollTo({ top: 0, behavior: 'auto' })
  }, [hash, pathname])

  return null
}
