import { useEffect, useState, type ImgHTMLAttributes } from 'react'

import { apiBlobRequest } from '@/lib/api'

const transparentPixel =
  'data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs='

interface AuthenticatedImageProps
  extends Omit<ImgHTMLAttributes<HTMLImageElement>, 'src'> {
  src: string
}

export function AuthenticatedImage({ src, className, ...props }: AuthenticatedImageProps) {
  const [resolvedSource, setResolvedSource] = useState(() =>
    src.startsWith('/media/') ? transparentPixel : src,
  )

  useEffect(() => {
    if (!src.startsWith('/media/')) {
      setResolvedSource(src)
      return
    }

    let cancelled = false
    let objectURL = ''
    setResolvedSource(transparentPixel)
    void apiBlobRequest(src)
      .then((blob) => {
        if (cancelled) return
        objectURL = URL.createObjectURL(blob)
        setResolvedSource(objectURL)
      })
      .catch(() => {
        if (!cancelled) setResolvedSource(transparentPixel)
      })

    return () => {
      cancelled = true
      if (objectURL) URL.revokeObjectURL(objectURL)
    }
  }, [src])

  return (
    <img
      {...props}
      className={className}
      src={resolvedSource}
    />
  )
}
