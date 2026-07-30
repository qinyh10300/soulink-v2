import { motion } from 'framer-motion'
import { useEffect, useState } from 'react'

import bnuStoryImage1 from '@/assets/story/BNU-1.png'
import bnuStoryImage2 from '@/assets/story/BNU-2.png'
import bnuStoryImage3 from '@/assets/story/BNU-3.png'
import bnuStoryImage4 from '@/assets/story/BNU-4.png'
import bfsuStoryImage1 from '@/assets/story/BFSU-1.png'
import bfsuStoryImage2 from '@/assets/story/BFSU-2.png'
import bfsuStoryImage3 from '@/assets/story/BFSU-3.png'
import bfsuStoryImage4 from '@/assets/story/BFSU-4.png'
import cucStoryImage1 from '@/assets/story/CUC-1.jpg'
import cucStoryImage2 from '@/assets/story/CUC-2.jpg'
import cucStoryImage3 from '@/assets/story/CUC-3.jpg'
import cucStoryImage4 from '@/assets/story/CUC-4.jpg'
import pkuStoryImage1 from '@/assets/story/PKU-1.png'
import pkuStoryImage2 from '@/assets/story/PKU-2.png'
import pkuStoryImage3 from '@/assets/story/PKU-3.png'
import pkuStoryImage4 from '@/assets/story/PKU-4.png'
import rucStoryImage1 from '@/assets/story/RUC-1.jpg'
import rucStoryImage2 from '@/assets/story/RUC-2.jpg'
import rucStoryImage3 from '@/assets/story/RUC-3.jpg'
import rucStoryImage4 from '@/assets/story/RUC-4.jpg'
import storyImage1 from '@/assets/story/THU-1.png'
import storyImage2 from '@/assets/story/THU-2.png'
import storyImage3 from '@/assets/story/THU-3.png'
import storyImage4 from '@/assets/story/THU-4.png'
import { LandingHero } from '@/components/home/landing-hero'
import { schoolHeroImages } from '@/lib/school-image-assets'
import type { SchoolKey } from '@/lib/school-keys'

const sectionEase = [0.22, 1, 0.36, 1] as const
const HERO_SLIDE_INTERVAL_MS = 10000

function SectionDivider() {
  return <div className="mx-auto h-px w-full max-w-6xl bg-[#efe8df]" aria-hidden="true" />
}

const storyContent = {
  thu: {
    poemLines: [
      '东操有风，西荷塘有月，',
      '北馆灯亮，南园声未歇。',
      '光影流转，人群往复，',
      '心意明灭，却少一人可解。',
    ],
  },
  pku: {
    poemLines: [
      '塔影临湖，书灯映夜，',
      '燕园有风，行人不止。',
      '四时流转，草木自深，',
      '人海之中，我们不断错身。',
    ],
  },
  bnu: {
    poemLines: [
      '木铎有声，京师有风，',
      '银杏渐黄，灯火渐明。',
      '四季轮转，人海往复，',
      '也想遇见一位同频的同行者。',
    ],
  },
  ruc: {
    poemLines: [
      '观点在这里相遇，',
      '热望在这里生长。',
      '在表达与倾听之间，',
      '也想认真遇见彼此。',
    ],
  },
  cuc: {
    poemLines: [
      '镜头记录光影，',
      '声音延伸想象。',
      '故事尚未落幕，人海之中，',
      '仍想遇见一人共讲。',
    ],
  },
  bfsu: {
    poemLines: [
      '风过校园，书页轻翻，',
      '灯火深处，夜色未阑。',
      '世间语言何其千万，',
      '却仍想与你，轻声细谈。',
    ],
  },
} as const

const heroSlides = [
  {
    count: undefined,
    image: schoolHeroImages.thu,
    key: 'thu',
    mobileImageHeight: '72svh',
    mobileImageLeft: '50%',
    objectPosition: 'center 42%',
    showRegistrationCount: true,
    tagline: '水木之间，灵魂相逢',
    title: 'THUers',
  },
  {
    count: 68,
    image: schoolHeroImages.pku,
    key: 'pku',
    mobileImageHeight: '70svh',
    mobileImageLeft: '42%',
    objectPosition: 'center 50%',
    showRegistrationCount: true,
    tagline: '未名湖畔，灵魂相逢',
    title: 'PKUers',
  },
  {
    count: 32,
    image: schoolHeroImages.bnu,
    key: 'bnu',
    mobileImageHeight: '70svh',
    mobileImageLeft: '50%',
    objectPosition: 'center 48%',
    showRegistrationCount: false,
    tagline: '京师学脉，理想相续',
    title: 'BNUers',
  },
  {
    count: undefined,
    image: schoolHeroImages.ruc,
    key: 'ruc',
    mobileImageHeight: '54svh',
    mobileImageLeft: '50%',
    objectPosition: 'center 20%',
    showRegistrationCount: false,
    tagline: '人文热望，真诚相遇',
    title: 'RUCers',
  },
  {
    count: undefined,
    image: schoolHeroImages.cuc,
    key: 'cuc',
    mobileImageHeight: '62svh',
    mobileImageLeft: '50%',
    objectPosition: 'center 34%',
    showRegistrationCount: false,
    tagline: '镜头之外，认真相遇',
    title: 'CUCers',
  },
  {
    count: undefined,
    image: schoolHeroImages.bfsu,
    key: 'bfsu',
    mobileImageHeight: '66svh',
    mobileImageLeft: '50%',
    objectPosition: 'center 52%',
    showRegistrationCount: false,
    tagline: '语通世界，心向彼此',
    title: 'BFSUers',
  },
] as const

interface LandingPageProps {
  schoolKey?: SchoolKey
}

export function LandingPage({ schoolKey }: LandingPageProps = {}) {
  const thuStoryImagesLeft = [storyImage1, storyImage2]
  const thuStoryImagesRight = [storyImage3, storyImage4]
  const pkuStoryImagesLeft = [pkuStoryImage1, pkuStoryImage2]
  const pkuStoryImagesRight = [pkuStoryImage3, pkuStoryImage4]
  const bnuStoryImagesLeft = [bnuStoryImage1, bnuStoryImage2]
  const bnuStoryImagesRight = [bnuStoryImage3, bnuStoryImage4]
  const bfsuStoryImagesLeft = [bfsuStoryImage1, bfsuStoryImage2]
  const bfsuStoryImagesRight = [bfsuStoryImage3, bfsuStoryImage4]
  const cucStoryImagesLeft = [cucStoryImage1, cucStoryImage2]
  const cucStoryImagesRight = [cucStoryImage3, cucStoryImage4]
  const rucStoryImagesLeft = [rucStoryImage1, rucStoryImage2]
  const rucStoryImagesRight = [rucStoryImage3, rucStoryImage4]
  const initialSchool = schoolKey ?? 'thu'
  const activeHeroSlides = schoolKey
    ? heroSlides.filter((slide) => slide.key === schoolKey)
    : heroSlides
  const [hoveredStoryImage, setHoveredStoryImage] = useState<string | null>(null)
  const [activeHeroSlide, setActiveHeroSlide] = useState(() =>
    schoolKey ? 0 : Math.max(heroSlides.findIndex((slide) => slide.key === initialSchool), 0),
  )
  const [activeStory, setActiveStory] = useState<SchoolKey>(initialSchool)

  const goToNextHeroSlide = () => {
    if (activeHeroSlides.length <= 1) {
      return
    }
    setActiveHeroSlide((currentSlide) => (currentSlide + 1) % activeHeroSlides.length)
  }

  const goToPreviousHeroSlide = () => {
    if (activeHeroSlides.length <= 1) {
      return
    }
    setActiveHeroSlide((currentSlide) =>
      (currentSlide - 1 + activeHeroSlides.length) % activeHeroSlides.length,
    )
  }

  useEffect(() => {
    setActiveStory(initialSchool)
    setActiveHeroSlide(0)
  }, [initialSchool])

  useEffect(() => {
    if (activeHeroSlides.length <= 1) {
      return
    }
    const timer = window.setInterval(() => {
      goToNextHeroSlide()
    }, HERO_SLIDE_INTERVAL_MS)

    return () => window.clearInterval(timer)
  }, [activeHeroSlides.length])

  const currentHeroSlide = activeHeroSlides[activeHeroSlide] ?? activeHeroSlides[0]
  const currentStory = storyContent[activeStory]
  const rucPoemLines = [
    '明德楼前，风过未停，',
    '林荫深处，人影匆匆。',
    '世事渐近，理想未冷，',
    '人海之中，总与一人错逢。',
  ] as const
  const displayPoemLines = activeStory === 'ruc' ? rucPoemLines : currentStory.poemLines
  const storyImagesBySchool = {
    thu: { left: thuStoryImagesLeft, right: thuStoryImagesRight },
    pku: { left: pkuStoryImagesLeft, right: pkuStoryImagesRight },
    bnu: { left: bnuStoryImagesLeft, right: bnuStoryImagesRight },
    ruc: { left: rucStoryImagesLeft, right: rucStoryImagesRight },
    cuc: { left: cucStoryImagesLeft, right: cucStoryImagesRight },
    bfsu: { left: bfsuStoryImagesLeft, right: bfsuStoryImagesRight },
  } as const
  const storyImagesLeft = storyImagesBySchool[activeStory].left
  const storyImagesRight = storyImagesBySchool[activeStory].right
  const locationLabelBySchool: Record<SchoolKey, string> = {
    thu: '清华·情人坡',
    pku: '北大·未名湖',
    bnu: '北师大·邱季端',
    ruc: '人大·人文广场',
    cuc: '中传·文化广场',
    bfsu: '北外·西院林荫',
  }
  const storyLabelBySchool: Record<SchoolKey, string> = {
    thu: 'Tsinghua Story',
    pku: 'Peking Story',
    bnu: 'BNU Story',
    ruc: 'RUC Story',
    cuc: 'CUC Story',
    bfsu: 'BFSU Story',
  }
  const currentHeroTagline =
    currentHeroSlide.key === 'pku' ? '未名湖畔，知己相遇' : currentHeroSlide.tagline
  const currentLocationLabel =
    currentHeroSlide.key === 'ruc' ? '人大·明德楼' : locationLabelBySchool[currentHeroSlide.key]

  return (
    <div className="relative isolate bg-[#fdfbf8]">
      <div className="relative z-10">
        <div className="relative">
          <div className="pointer-events-none absolute inset-0 z-0 overflow-hidden md:hidden">
            {activeHeroSlides.map((slide, index) => {
              const isActive = index === activeHeroSlide

              return (
                <motion.div
                  key={`${slide.key}-mobile`}
                  animate={{ opacity: isActive ? 1 : 0 }}
                  className="absolute inset-0"
                  initial={false}
                  transition={{ duration: 0.9, ease: sectionEase }}
                >
                  <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,246,236,0.32),transparent_46%),linear-gradient(180deg,rgba(37,28,25,0.04)_0%,rgba(37,28,25,0.18)_100%)]" />
                  <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(24,16,12,0)_0%,rgba(24,16,12,0)_24%,rgba(24,16,12,0.02)_60%,rgba(24,16,12,0.02)_100%)]" />
                  <picture>
                    <source srcSet={slide.image.avif} type="image/avif" />
                    <source srcSet={slide.image.webp} type="image/webp" />
                    <img
                      alt=""
                      className={`absolute top-16 max-w-none -translate-x-1/2 opacity-96 ${slide.key === 'bfsu' ? 'scale-[0.92]' : ''}`}
                      decoding="async"
                      fetchPriority={isActive ? 'high' : 'low'}
                      loading={isActive ? 'eager' : 'lazy'}
                      src={slide.image.webp}
                      style={{
                        height: slide.mobileImageHeight,
                        left: slide.mobileImageLeft,
                        maskImage:
                          'linear-gradient(180deg, rgba(0,0,0,1) 0%, rgba(0,0,0,1) 78%, rgba(0,0,0,0.72) 90%, transparent 100%)',
                        maxHeight: '82%',
                        WebkitMaskImage:
                          'linear-gradient(180deg, rgba(0,0,0,1) 0%, rgba(0,0,0,1) 78%, rgba(0,0,0,0.72) 90%, transparent 100%)',
                      }}
                    />
                  </picture>
                  <div className="absolute inset-x-0 bottom-0 h-[30%] bg-[linear-gradient(180deg,rgba(253,251,248,0)_0%,rgba(253,251,248,0.36)_52%,#fdfbf8_100%)]" />
                </motion.div>
              )
            })}
          </div>

          <div className="pointer-events-none absolute inset-0 z-0 hidden overflow-hidden md:block">
            {activeHeroSlides.map((slide, index) => {
              const offset =
                index === activeHeroSlide
                  ? '0%'
                  : index < activeHeroSlide
                    ? '-100%'
                    : '100%'

              return (
                <picture key={slide.key}>
                  <source srcSet={slide.image.avif} type="image/avif" />
                  <source srcSet={slide.image.webp} type="image/webp" />
                  <motion.img
                    alt=""
                    animate={{ opacity: 0.7, x: offset }}
                    className={`absolute left-1/2 top-1/2 h-auto min-h-full w-auto min-w-full max-w-none -translate-x-1/2 -translate-y-1/2 object-cover ${slide.key === 'bfsu' ? 'scale-[0.88]' : ''}`}
                    decoding="async"
                    fetchPriority={index === activeHeroSlide ? 'high' : 'low'}
                    initial={false}
                    loading={index === activeHeroSlide ? 'eager' : 'lazy'}
                    src={slide.image.webp}
                    style={{ objectPosition: slide.objectPosition }}
                    transition={{ duration: 1.15, ease: sectionEase }}
                  />
                </picture>
              )
            })}
          </div>

          <div className="relative z-10">
            <LandingHero
              countOverride={currentHeroSlide.count}
              currentLocationLabel={currentLocationLabel}
              currentSchoolKey={schoolKey}
              currentTagline={currentHeroTagline}
              currentTitle={currentHeroSlide.title}
              onNextSlide={goToNextHeroSlide}
              onPreviousSlide={goToPreviousHeroSlide}
              showNextArrow={!schoolKey && activeHeroSlides.length > 1}
              showPreviousArrow={!schoolKey && activeHeroSlides.length > 1}
              showRegistrationCount={schoolKey ? true : currentHeroSlide.showRegistrationCount}
            />
          </div>
        </div>

        <div className="space-y-0 bg-[#fdfbf8]/96">
          <section className="px-6 py-10 lg:px-8 lg:py-14" id="story">
            <div className="mx-auto flex w-full max-w-6xl justify-center">
              <div className="relative w-full max-w-[76rem]">
                <motion.div
                  className="absolute left-0 top-1/2 hidden -translate-y-1/2 lg:flex lg:w-[18rem] lg:flex-col lg:gap-6 xl:w-[19rem]"
                  initial={{ opacity: 0, x: -18 }}
                  transition={{ duration: 1.1, delay: 0.95, ease: sectionEase }}
                  viewport={{ once: true, amount: 0.35 }}
                  whileInView={{ opacity: 1, x: 0 }}
                >
                  {storyImagesLeft.map((image) => (
                    <div
                      key={image}
                      className="relative h-[12rem] w-[14.5rem] overflow-hidden rounded-[1.5rem] transition-transform duration-500 ease-out hover:z-10 hover:scale-[1.04] xl:h-[13rem] xl:w-[15.5rem]"
                      onMouseEnter={() => setHoveredStoryImage(image)}
                      onMouseLeave={() => setHoveredStoryImage(null)}
                      style={{
                        WebkitMaskImage:
                          hoveredStoryImage === image
                            ? 'none'
                            : 'radial-gradient(circle at center, rgba(0,0,0,0.96) 44%, rgba(0,0,0,0.56) 72%, transparent 100%)',
                        maskImage:
                          hoveredStoryImage === image
                            ? 'none'
                            : 'radial-gradient(circle at center, rgba(0,0,0,0.96) 44%, rgba(0,0,0,0.56) 72%, transparent 100%)',
                      }}
                    >
                      <img
                        alt=""
                        className="h-full w-full select-none object-cover transition-[filter,opacity] duration-500 ease-out"
                        decoding="async"
                        loading="lazy"
                        src={image}
                        style={{
                          filter: hoveredStoryImage === image ? 'blur(0px)' : 'blur(1px)',
                          opacity: 1,
                        }}
                      />
                    </div>
                  ))}
                </motion.div>

                <motion.div
                  className="absolute right-0 top-1/2 hidden -translate-y-1/2 lg:flex lg:w-[18rem] lg:flex-col lg:gap-6 xl:w-[19rem]"
                  initial={{ opacity: 0, x: 18 }}
                  transition={{ duration: 1.1, delay: 1.05, ease: sectionEase }}
                  viewport={{ once: true, amount: 0.35 }}
                  whileInView={{ opacity: 1, x: 0 }}
                >
                  {storyImagesRight.map((image) => (
                    <div
                      key={image}
                      className="relative ml-auto h-[12rem] w-[14.5rem] overflow-hidden rounded-[1.5rem] transition-transform duration-500 ease-out hover:z-10 hover:scale-[1.04] xl:h-[13rem] xl:w-[15.5rem]"
                      onMouseEnter={() => setHoveredStoryImage(image)}
                      onMouseLeave={() => setHoveredStoryImage(null)}
                      style={{
                        WebkitMaskImage:
                          hoveredStoryImage === image
                            ? 'none'
                            : 'radial-gradient(circle at center, rgba(0,0,0,0.96) 44%, rgba(0,0,0,0.56) 72%, transparent 100%)',
                        maskImage:
                          hoveredStoryImage === image
                            ? 'none'
                            : 'radial-gradient(circle at center, rgba(0,0,0,0.96) 44%, rgba(0,0,0,0.56) 72%, transparent 100%)',
                      }}
                    >
                      <img
                        alt=""
                        className="h-full w-full select-none object-cover transition-[filter,opacity] duration-500 ease-out"
                        decoding="async"
                        loading="lazy"
                        src={image}
                        style={{
                          filter: hoveredStoryImage === image ? 'blur(0px)' : 'blur(1px)',
                          opacity: 0.95,
                        }}
                      />
                    </div>
                  ))}
                </motion.div>

                <div className="mx-auto w-full max-w-[40rem] overflow-hidden px-8 py-8 sm:px-10 sm:py-10">
                  <div className="pointer-events-none absolute inset-0 opacity-70">
                    <div className="absolute inset-x-12 top-6 h-px bg-[linear-gradient(90deg,transparent,rgba(210,191,170,0.78),transparent)]" />
                    <div className="absolute inset-x-12 bottom-6 h-px bg-[linear-gradient(90deg,transparent,rgba(210,191,170,0.62),transparent)]" />
                    <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(255,255,255,0.52),transparent_38%),repeating-linear-gradient(180deg,rgba(222,211,198,0.06)_0px,rgba(222,211,198,0.06)_1px,transparent_1px,transparent_18px)]" />
                  </div>

                  <div className="relative space-y-8 text-center">
                    <div className="space-y-4">
                      <div className="flex flex-col items-center gap-4">
                        <div className="flex flex-wrap items-center justify-center gap-1 rounded-[1.5rem] border border-[#e6d7c7] bg-[#fbf5ee]/95 p-1 shadow-[0_10px_24px_rgba(167,136,104,0.08)]">
                          {!schoolKey
                            ? ([
                                { key: 'thu', label: 'Hear THU Story' },
                                { key: 'pku', label: 'Hear PKU Story' },
                                { key: 'bnu', label: 'Hear BNU Story' },
                                { key: 'ruc', label: 'Hear RUC Story' },
                                { key: 'cuc', label: 'Hear CUC Story' },
                                { key: 'bfsu', label: 'Hear BFSU Story' },
                              ] as const).map((storyOption) => {
                                const isActive = activeStory === storyOption.key

                                return (
                                  <button
                                    key={storyOption.key}
                                    aria-pressed={isActive}
                                    className={`inline-flex items-center justify-center rounded-full px-4 py-2 text-[0.82rem] tracking-[0.08em] transition sm:px-5 ${
                                      isActive
                                        ? 'bg-[#ead7c2] text-[#7a5f49] shadow-[0_8px_18px_rgba(169,131,94,0.16)]'
                                        : 'bg-transparent text-[#ab8b71] hover:bg-[#f4e8db] hover:text-[#896a52]'
                                    }`}
                                    onClick={() => setActiveStory(storyOption.key)}
                                    type="button"
                                  >
                                    {storyOption.label}
                                  </button>
                                )
                              })
                            : (
                              <div className="inline-flex items-center justify-center rounded-full bg-[#f4e9de] px-5 py-2 text-[0.82rem] tracking-[0.12em] text-[#8d6f5b]">
                                {storyLabelBySchool[activeStory]}
                              </div>
                            )}
                        </div>
                        <p
                          className="text-[1.55rem] italic leading-none tracking-[0.16em] text-[#b79d86] sm:text-[2.1rem]"
                          style={{ fontFamily: '"Cormorant Garamond", "Times New Roman", serif' }}
                        >
                          Our Story
                        </p>
                      </div>
                      <div className="mx-auto flex w-full max-w-[8rem] items-center gap-3">
                        <div className="h-px flex-1 bg-[#eadfce]" />
                        <div className="h-1.5 w-1.5 rounded-full bg-[#d6c4b1]" />
                        <div className="h-px flex-1 bg-[#eadfce]" />
                      </div>
                    </div>

                    <div className="space-y-8 text-[#433a33]">
                      <p className="font-serif text-[1.12rem] leading-[1.95] tracking-[0.02em] sm:text-[1.26rem]">
                        {displayPoemLines[0]}
                        <br />
                        {displayPoemLines[1]}
                        <br />
                        {displayPoemLines[2]}
                        <br />
                        {displayPoemLines[3]}
                      </p>

                      <p className="font-serif text-[1.12rem] leading-[1.95] tracking-[0.02em] text-[#5f5449] sm:text-[1.26rem]">
                        SouLink 不再追逐仓促的相逢，
                        <br />
                        只想留下一隅静处，
                        <br />
                        让真实慢慢浮现。
                      </p>

                      <p className="font-serif text-[1.12rem] leading-[1.95] tracking-[0.02em] text-[#5f5449] sm:text-[1.26rem]">
                        我愿坦然展开自我，
                        <br />
                        也愿温柔接住对方。
                        <br />
                        以真心，换真心。
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <SectionDivider />
        </div>
      </div>
    </div>
  )
}
