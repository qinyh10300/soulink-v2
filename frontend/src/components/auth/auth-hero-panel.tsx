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
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(58,41,35,0.16)_0%,rgba(78,55,45,0.2)_38%,rgba(40,25,22,0.42)_100%)]" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,216,194,0.22),transparent_28%),radial-gradient(circle_at_bottom_right,rgba(146,92,66,0.22),transparent_30%)]" />

      <Link
        className="absolute left-10 top-10 flex items-center gap-4 rounded-full border border-white/18 bg-white/10 px-5 py-3 text-white backdrop-blur-md transition-all duration-200 hover:bg-white/16 hover:shadow-[0_16px_36px_rgba(20,14,12,0.16)]"
        to="/"
      >
        <img
          alt=" Soulink"
          className="h-11 w-11 rounded-2xl border border-white/18 object-cover"
          src={brandIcon}
        />
        <div>
          <p className="text-sm font-semibold tracking-[0.18em] uppercase">
             Soulink
          </p>
          <p className="text-sm text-white/76">AI社交匹配平台</p>
        </div>
      </Link>

      <div className="absolute inset-x-10 bottom-12">
        <div className="max-w-xl space-y-5">
          <h2 className="text-balance font-serif text-6xl leading-[0.98] text-[#f4d6b7] drop-shadow-[0_14px_40px_rgba(27,16,12,0.28)]">
            {title}
          </h2>
          <p className="font-serif text-[1.5rem] italic tracking-[0.18em] text-[#f4d6b7]">
            {subtitle}
          </p>
        </div>
      </div>
    </aside>
  )
}
