import brandIcon from '@/assets/soulink-logo-icon.png'
import { Link } from 'react-router-dom'

interface AuthHeroPanelProps {
  imageSrc: string
  imageSources?: {
    avif: string
    webp: string
  }
  imageAlt: string
  title: string
  subtitle: string
}

export function AuthHeroPanel({
  imageSrc,
  imageSources,
  imageAlt,
  title,
  subtitle,
}: AuthHeroPanelProps) {
  return (
    <aside className="relative hidden min-h-screen overflow-hidden lg:block">
      <picture className="absolute inset-0 block">
        {imageSources ? (
          <>
            <source srcSet={imageSources.avif} type="image/avif" />
            <source srcSet={imageSources.webp} type="image/webp" />
          </>
        ) : null}
        <img
          alt={imageAlt}
          className="h-full w-full object-cover"
          decoding="async"
          fetchPriority="high"
          height={1024}
          loading="eager"
          src={imageSrc}
          width={1536}
        />
      </picture>
      <div className="absolute inset-0 bg-[rgba(24,20,26,0.46)]" />

      <Link
        className="absolute left-10 top-10 flex items-center gap-3 rounded-md border border-white/25 bg-black/15 px-4 py-3 text-white backdrop-blur-md transition-colors hover:bg-black/25"
        to="/"
      >
        <img
          alt=" Soulink"
          className="h-9 w-9 rounded-md border border-white/20 object-cover"
          src={brandIcon}
        />
        <div>
          <p className="text-sm font-semibold">SouLink</p>
          <p className="text-xs text-white/76">清华校内 AI 匹配</p>
        </div>
      </Link>

      <div className="absolute inset-x-10 bottom-12">
        <div className="max-w-xl space-y-5">
          <h2 className="text-balance text-5xl font-semibold leading-[1.05] text-white">
            {title}
          </h2>
          <p className="text-lg text-white/78">
            {subtitle}
          </p>
        </div>
      </div>
    </aside>
  )
}
