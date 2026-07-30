import bfsuCardAvif from '@/assets/optimized/cards/bfsu.avif'
import bfsuCardWebp from '@/assets/optimized/cards/bfsu.webp'
import bnuCardAvif from '@/assets/optimized/cards/bnu.avif'
import bnuCardWebp from '@/assets/optimized/cards/bnu.webp'
import cucCardAvif from '@/assets/optimized/cards/cuc.avif'
import cucCardWebp from '@/assets/optimized/cards/cuc.webp'
import pkuCardAvif from '@/assets/optimized/cards/pku.avif'
import pkuCardWebp from '@/assets/optimized/cards/pku.webp'
import rucCardAvif from '@/assets/optimized/cards/ruc.avif'
import rucCardWebp from '@/assets/optimized/cards/ruc.webp'
import thuCardAvif from '@/assets/optimized/cards/thu.avif'
import thuCardWebp from '@/assets/optimized/cards/thu.webp'
import bfsuHeroAvif from '@/assets/optimized/heroes/bfsu.avif'
import bfsuHeroWebp from '@/assets/optimized/heroes/bfsu.webp'
import bnuHeroAvif from '@/assets/optimized/heroes/bnu.avif'
import bnuHeroWebp from '@/assets/optimized/heroes/bnu.webp'
import cucHeroAvif from '@/assets/optimized/heroes/cuc.avif'
import cucHeroWebp from '@/assets/optimized/heroes/cuc.webp'
import pkuHeroAvif from '@/assets/optimized/heroes/pku.avif'
import pkuHeroWebp from '@/assets/optimized/heroes/pku.webp'
import rucHeroAvif from '@/assets/optimized/heroes/ruc.avif'
import rucHeroWebp from '@/assets/optimized/heroes/ruc.webp'
import thuHeroAvif from '@/assets/optimized/heroes/thu.avif'
import thuHeroWebp from '@/assets/optimized/heroes/thu.webp'
import type { SchoolKey } from '@/lib/school-keys'

export type OptimizedImageSource = {
  avif: string
  webp: string
}

export const campusCardImages: Record<SchoolKey, OptimizedImageSource> = {
  thu: { avif: thuCardAvif, webp: thuCardWebp },
  pku: { avif: pkuCardAvif, webp: pkuCardWebp },
  bnu: { avif: bnuCardAvif, webp: bnuCardWebp },
  ruc: { avif: rucCardAvif, webp: rucCardWebp },
  cuc: { avif: cucCardAvif, webp: cucCardWebp },
  bfsu: { avif: bfsuCardAvif, webp: bfsuCardWebp },
}

export const schoolHeroImages: Record<SchoolKey, OptimizedImageSource> = {
  thu: { avif: thuHeroAvif, webp: thuHeroWebp },
  pku: { avif: pkuHeroAvif, webp: pkuHeroWebp },
  bnu: { avif: bnuHeroAvif, webp: bnuHeroWebp },
  ruc: { avif: rucHeroAvif, webp: rucHeroWebp },
  cuc: { avif: cucHeroAvif, webp: cucHeroWebp },
  bfsu: { avif: bfsuHeroAvif, webp: bfsuHeroWebp },
}
