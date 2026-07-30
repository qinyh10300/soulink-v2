import { useEffect, useState } from 'react'
import { ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { MatchRevealCountdown } from '@/components/home/match-reveal-countdown'
import { LandingSharedSections } from '@/components/home/landing-shared-sections'
import { ContactUsModal } from '@/components/layout/contact-us-modal'
import { buttonVariants } from '@/components/ui/button'
import {
  fetchAllSchoolRegisteredUserCounts,
  type RegisteredUserCountSchoolKey,
  type SchoolRegisteredUserCounts,
} from '@/lib/meta-api'
import { cn } from '@/lib/utils'
import { aboutContactItems } from '@/mock/about'

type CampusCard = {
  key: Extract<RegisteredUserCountSchoolKey, 'thu' | 'pku' | 'bnu' | 'ruc' | 'cuc' | 'bfsu' | 'ustb'>
  name: string
  englishName: string
  summary: string
  href: string
}

type SingaporeCampusCard = {
  key: Extract<RegisteredUserCountSchoolKey, 'nus' | 'ntu'>
  name: string
  englishName: string
  summary: string
  href: string
}

type ShanghaiCampusCard = {
  key: Extract<RegisteredUserCountSchoolKey, 'fdu' | 'sjtu' | 'tju'>
  name: string
  englishName: string
  summary: string
  href: string
}

type HubeiCampusCard = {
  key: Extract<RegisteredUserCountSchoolKey, 'zuel'>
  name: string
  englishName: string
  summary: string
  href: string
}

type HunanCampusCard = {
  key: Extract<RegisteredUserCountSchoolKey, 'csu'>
  name: string
  englishName: string
  summary: string
  href: string
}

const beijingCampusCards: CampusCard[] = [
  {
    key: 'thu',
    name: '清华大学',
    englishName: 'Tsinghua University',
    summary:
      '从图书馆到操场，从夜里灯光到清晨薄雾，这里更适合认真表达，也适合慢慢靠近。',
    href: '/schools/thu',
  },
  {
    key: 'pku',
    name: '北京大学',
    englishName: 'Peking University',
    summary:
      '从燕园树影到未名湖畔，适合把克制、诗意与理解写成一场更有分量的相遇。',
    href: '/schools/pku',
  },
  {
    key: 'bnu',
    name: '北京师范大学',
    englishName: 'Beijing Normal University',
    summary:
      '更温和、更有倾听感的校园气质，适合把细腻、稳定和长期主义放进相遇的开头。',
    href: '/schools/bnu',
  },
  {
    key: 'ruc',
    name: '中国人民大学',
    englishName: 'Renmin University of China',
    summary:
      '在人文与公共讨论的气息里，适合遇见表达清晰、价值感明确，也愿意认真回应的人。',
    href: '/schools/ruc',
  },
  {
    key: 'cuc',
    name: '中国传媒大学',
    englishName: 'Communication University of China',
    summary:
      '在创作、表达与叙事交织的校园里，适合遇见灵感充沛，也愿意真诚靠近的人。',
    href: '/schools/cuc',
  },
  {
    key: 'bfsu',
    name: '北京外国语大学',
    englishName: 'Beijing Foreign Studies University',
    summary:
      '在多元语言与文化相遇的日常里，适合遇见愿意理解彼此、认真交流的人。',
    href: '/schools/bfsu',
  },
  {
    key: 'ustb',
    name: '北京科技大学',
    englishName: 'University of Science and Technology Beijing',
    summary: '在工程底色与校园温度并存的氛围中，适合遇见稳定真诚、愿意长期沟通的人。',
    href: '/login',
  },
]

const singaporeCampusCards: SingaporeCampusCard[] = [
  {
    key: 'nus',
    name: '新加坡国立大学',
    englishName: 'National University of Singapore',
    summary:
      '在海风、雨林和高密度学术社区之间，适合遇见视野开阔、节奏稳定，也愿意认真了解彼此的人。',
    href: '/login',
  },
  {
    key: 'ntu',
    name: '南洋理工大学',
    englishName: 'Nanyang Technological University',
    summary:
      '在创新、工程与跨学科表达交汇的校园里，适合遇见理性清晰，同时保有温度与想象力的人。',
    href: '/login',
  },
]

const shanghaiCampusCards: ShanghaiCampusCard[] = [
  {
    key: 'fdu',
    name: '复旦大学',
    englishName: 'Fudan University',
    summary:
      '在开放、流动与人文气质并存的校园里，适合遇见表达自洽、视野开阔，也愿意认真建立连接的人。',
    href: '/login',
  },
  {
    key: 'sjtu',
    name: '上海交通大学',
    englishName: 'Shanghai Jiao Tong University',
    summary:
      '在理工、创新与高强度成长节奏中，适合遇见目标明确、思路清晰，同时也保有温度的人。',
    href: '/login',
  },
  {
    key: 'tju',
    name: '同济大学',
    englishName: 'Tongji University',
    summary: '在人文与理工融合的校园气质里，适合遇见思路清晰、也愿意细腻倾听的人。',
    href: '/login',
  },
]

const hubeiCampusCards: HubeiCampusCard[] = [
  {
    key: 'zuel',
    name: '中南财经政法大学',
    englishName: 'Zhongnan University of Economics and Law',
    summary:
      '在法学、财经与人文交汇的校园氛围里，适合遇见表达清晰、认真坦诚，也愿意共同建立稳定连接的人。',
    href: '/login',
  },
]

const hunanCampusCards: HunanCampusCard[] = [
  {
    key: 'csu',
    name: '中南大学',
    englishName: 'Central South University',
    summary: '在多学科交汇与务实成长节奏中，适合遇见表达清晰、真诚可靠的人。',
    href: '/login',
  },
]

const campusEnglishNameClass =
  'text-[0.72rem] uppercase leading-[1.45] tracking-[0.16em] text-[#9d7d67] break-words'

function formatRegisteredUserCountLabel(counts: SchoolRegisteredUserCounts, schoolKey: RegisteredUserCountSchoolKey) {
  const rawRegisteredUserCount = counts[schoolKey]
  if (rawRegisteredUserCount == null) {
    return '实时注册人数加载中'
  }
  return `实时注册人数 ${rawRegisteredUserCount.toLocaleString('zh-CN')} 人`
}

export function AlliancePage() {
  const [contactModalOpen, setContactModalOpen] = useState(false)
  const [registeredUserCounts, setRegisteredUserCounts] = useState<SchoolRegisteredUserCounts>({})
  const contactEmail = aboutContactItems[0]?.value ?? 'thusoulink@163.com'
  const totalRegisteredUserCount = Object.values(registeredUserCounts).reduce(
    (sum, count) => sum + (count ?? 0),
    0,
  )
  const totalRegisteredUserCountLabel =
    Object.keys(registeredUserCounts).length === 0
      ? '实时注册总人数：加载中'
      : `实时注册总人数：${totalRegisteredUserCount.toLocaleString('zh-CN')} 人`

  useEffect(() => {
    let cancelled = false

    const loadRegisteredUserCounts = async () => {
      const counts = await fetchAllSchoolRegisteredUserCounts().catch(() => ({}))

      if (!cancelled) {
        setRegisteredUserCounts(counts)
      }
    }

    void loadRegisteredUserCounts()

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <>
      <div className="relative overflow-hidden bg-[#f8f4ee] text-[#241d1a]">
        <div className="pointer-events-none absolute inset-x-0 top-0 h-[32rem] bg-[radial-gradient(circle_at_top,_rgba(114,124,183,0.16),_transparent_58%)]" />
        <div className="pointer-events-none absolute left-[-10rem] top-24 h-[22rem] w-[22rem] rounded-full bg-[radial-gradient(circle,_rgba(204,178,157,0.24),_transparent_68%)] blur-3xl" />
        <div className="pointer-events-none absolute right-[-8rem] top-16 h-[24rem] w-[24rem] rounded-full bg-[radial-gradient(circle,_rgba(169,170,207,0.18),_transparent_72%)] blur-3xl" />

        <section className="relative px-6 pb-14 pt-36 lg:px-8 lg:pb-18 lg:pt-44">
          <div className="mx-auto w-full max-w-7xl">
            <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(24rem,34rem)] lg:items-center lg:gap-12">
              <div className="space-y-6 text-center lg:text-left">
                <h1
                  className="text-[3.5rem] leading-[0.88] tracking-[-0.07em] text-[#231c18] sm:text-[4.8rem] lg:text-[6rem]"
                  style={{ fontFamily: '"Bodoni Moda", "Cormorant Garamond", "STSong", serif' }}
                >
                  <span className="block bg-[linear-gradient(135deg,#241b18_0%,#6d5b98_45%,#9a7db9_100%)] bg-clip-text text-transparent">
                    Soulink
                  </span>
                  <span className="mt-1 block bg-[linear-gradient(135deg,#6d5b98_0%,#9a7db9_48%,#c89f9d_100%)] bg-clip-text text-transparent">
                    高校联盟
                  </span>
                </h1>
                <p className="mx-auto max-w-2xl text-lg leading-8 text-[#6d5f56] sm:text-[1.08rem] lg:mx-0">
                  从不同校园的氛围、故事和节奏出发，进入属于各自学校的 Soulink 页面。先选择校园，再开始更具体的相遇。
                </p>
                <p className="mx-auto max-w-2xl text-lg leading-8 text-[#6d5f56] sm:text-[1.08rem] lg:mx-0">
                  在 Soulink 高校联盟中，参与者将填写一份深度问卷，并可指定希望匹配的对象来自哪所高校。每周固定时间，Soulink 会进行 AI 深度匹配，并在网站上公布匹配结果。
                </p>
                <div className="flex flex-wrap items-center justify-center gap-3 lg:justify-start">
                  <Link
                    className={cn(
                      buttonVariants({ variant: 'secondary' }),
                      'inline-flex rounded-full border-[#d8c5b3] bg-white/70 px-6 py-6 text-sm tracking-[0.14em] text-[#7b6353] shadow-none hover:bg-white/90',
                    )}
                    to="/about#about-alliance"
                  >
                    关于高校联盟
                  </Link>
                  <div className="inline-flex items-center rounded-full border border-[#e4d5c7] bg-white/72 px-5 py-3 text-sm text-[#7b6353] shadow-[0_10px_24px_rgba(119,92,70,0.08)]">
                    <span className="whitespace-nowrap tracking-[0.06em]">{totalRegisteredUserCountLabel}</span>
                  </div>
                </div>
              </div>

              <MatchRevealCountdown className="mx-auto w-full max-w-[34rem]" />
            </div>
          </div>
        </section>

        <section className="relative px-6 pb-20 lg:px-8 lg:pb-24">
          <div className="mx-auto w-full max-w-7xl space-y-5">
            <div className="space-y-2 text-center">
              <p className="text-sm uppercase tracking-[0.24em] text-[#9a816d]">Campus Entry</p>
              <h2
                className="text-[2rem] tracking-[-0.05em] text-[#261f1a] sm:text-[2.6rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                选择你的校园
              </h2>
            </div>

            <div className="pt-2">
              <p
                className="text-left text-[1.6rem] font-semibold italic tracking-[0.14em] text-[#9d7d67] sm:text-[1.95rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                北京·Beijing
              </p>
            </div>

            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
              {beijingCampusCards.map((school) => {
                const rawRegisteredUserCount = registeredUserCounts[school.key]
                const registeredUserCount = rawRegisteredUserCount ?? 0
                const registeredUserCountLabel =
                  rawRegisteredUserCount == null
                    ? '实时注册人数加载中'
                    : `实时注册人数 ${registeredUserCount.toLocaleString('zh-CN')} 人`
                const cardContent = (
                  <article className="group relative flex h-full min-h-[8.5rem] flex-col overflow-hidden rounded-[26px] border border-white/70 bg-white shadow-[0_16px_42px_rgba(111,84,61,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_54px_rgba(111,84,61,0.12)]">
                    <div className="relative flex flex-1 flex-col gap-3 p-5">
                      <div className="space-y-1.5">
                        <div className="space-y-1">
                          <p className={campusEnglishNameClass}>
                            {school.englishName}
                          </p>
                          <h3
                            className="text-[1.85rem] leading-[0.95] tracking-[-0.04em] text-[#241d1a]"
                            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                          >
                            {school.name}
                          </h3>
                        </div>

                      </div>

                      <div className="flex items-center justify-between gap-3 pt-0.5">
                        <span className="inline-flex items-center gap-2 text-sm font-medium text-[#7c5f4f]">
                          进入校园
                          <ArrowRight className="transition group-hover:translate-x-1" size={16} />
                        </span>
                        <span className="ml-auto inline-flex shrink-0 items-center justify-end rounded-full bg-[#f6efe6] px-3 py-1 text-[0.72rem] font-medium tracking-[0.08em] text-[#8a6a57]">
                          {registeredUserCountLabel}
                        </span>
                      </div>
                    </div>
                  </article>
                )

                return (
                  <Link className="block" key={school.key} to={school.href}>
                    {cardContent}
                  </Link>
                )
              })}
            </div>

            <div className="pt-3">
              <p
                className="text-left text-[1.6rem] font-semibold italic tracking-[0.14em] text-[#9d7d67] sm:text-[1.95rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                上海·Shanghai
              </p>
            </div>

            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
              {shanghaiCampusCards.map((school) => (
                <Link className="block" key={school.key} to={school.href}>
                  <article className="group relative flex h-full min-h-[8.5rem] flex-col overflow-hidden rounded-[26px] border border-white/70 bg-white shadow-[0_16px_42px_rgba(111,84,61,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_54px_rgba(111,84,61,0.12)]">
                    <div className="relative flex flex-1 flex-col gap-3 p-5">
                      <div className="space-y-1.5">
                        <div className="space-y-1">
                          <p className={campusEnglishNameClass}>
                            {school.englishName}
                          </p>
                          <h3
                            className="text-[1.85rem] leading-[0.95] tracking-[-0.04em] text-[#241d1a]"
                            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                          >
                            {school.name}
                          </h3>
                        </div>

                      </div>

                      <div className="flex items-center justify-between gap-3 pt-0.5">
                        <span className="inline-flex items-center gap-2 text-sm font-medium text-[#7c5f4f]">
                          进入校园
                          <ArrowRight className="transition group-hover:translate-x-1" size={16} />
                        </span>
                        <span className="ml-auto inline-flex shrink-0 items-center justify-end rounded-full bg-[#f6efe6] px-3 py-1 text-[0.72rem] font-medium tracking-[0.08em] text-[#8a6a57]">
                          {formatRegisteredUserCountLabel(registeredUserCounts, school.key)}
                        </span>
                      </div>
                    </div>
                  </article>
                </Link>
              ))}
            </div>

            <div className="pt-3">
              <p
                className="text-left text-[1.6rem] font-semibold italic tracking-[0.14em] text-[#9d7d67] sm:text-[1.95rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                湖北·Hubei
              </p>
            </div>

            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
              {hubeiCampusCards.map((school) => (
                <Link className="block" key={school.key} to={school.href}>
                  <article className="group relative flex h-full min-h-[8.5rem] flex-col overflow-hidden rounded-[26px] border border-white/70 bg-white shadow-[0_16px_42px_rgba(111,84,61,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_54px_rgba(111,84,61,0.12)]">
                    <div className="relative flex flex-1 flex-col gap-3 p-5">
                      <div className="space-y-1.5">
                        <div className="space-y-1">
                          <p className={campusEnglishNameClass}>
                            {school.englishName}
                          </p>
                          <h3
                            className="text-[1.85rem] leading-[0.95] tracking-[-0.04em] text-[#241d1a]"
                            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                          >
                            {school.name}
                          </h3>
                        </div>

                      </div>

                      <div className="flex items-center justify-between gap-3 pt-0.5">
                        <span className="inline-flex items-center gap-2 text-sm font-medium text-[#7c5f4f]">
                          进入校园
                          <ArrowRight className="transition group-hover:translate-x-1" size={16} />
                        </span>
                        <span className="ml-auto inline-flex shrink-0 items-center justify-end rounded-full bg-[#f6efe6] px-3 py-1 text-[0.72rem] font-medium tracking-[0.08em] text-[#8a6a57]">
                          {formatRegisteredUserCountLabel(registeredUserCounts, school.key)}
                        </span>
                      </div>
                    </div>
                  </article>
                </Link>
              ))}
            </div>

            <div className="pt-3">
              <p
                className="text-left text-[1.6rem] font-semibold italic tracking-[0.14em] text-[#9d7d67] sm:text-[1.95rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                湖南·Hunan
              </p>
            </div>

            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
              {hunanCampusCards.map((school) => (
                <Link className="block" key={school.key} to={school.href}>
                  <article className="group relative flex h-full min-h-[8.5rem] flex-col overflow-hidden rounded-[26px] border border-white/70 bg-white shadow-[0_16px_42px_rgba(111,84,61,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_54px_rgba(111,84,61,0.12)]">
                    <div className="relative flex flex-1 flex-col gap-3 p-5">
                      <div className="space-y-1.5">
                        <div className="space-y-1">
                          <p className={campusEnglishNameClass}>
                            {school.englishName}
                          </p>
                          <h3
                            className="text-[1.85rem] leading-[0.95] tracking-[-0.04em] text-[#241d1a]"
                            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                          >
                            {school.name}
                          </h3>
                        </div>

                      </div>

                      <div className="flex items-center justify-between gap-3 pt-0.5">
                        <span className="inline-flex items-center gap-2 text-sm font-medium text-[#7c5f4f]">
                          进入校园
                          <ArrowRight className="transition group-hover:translate-x-1" size={16} />
                        </span>
                        <span className="ml-auto inline-flex shrink-0 items-center justify-end rounded-full bg-[#f6efe6] px-3 py-1 text-[0.72rem] font-medium tracking-[0.08em] text-[#8a6a57]">
                          {formatRegisteredUserCountLabel(registeredUserCounts, school.key)}
                        </span>
                      </div>
                    </div>
                  </article>
                </Link>
              ))}
            </div>

            <div className="pt-3">
              <p
                className="text-left text-[1.6rem] font-semibold italic tracking-[0.14em] text-[#9d7d67] sm:text-[1.95rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                新加坡·Singapore
              </p>
            </div>

            <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
              {singaporeCampusCards.map((school) => (
                <Link className="block" key={school.key} to={school.href}>
                  <article className="group relative flex h-full min-h-[8.5rem] flex-col overflow-hidden rounded-[26px] border border-white/70 bg-white shadow-[0_16px_42px_rgba(111,84,61,0.09)] transition duration-300 hover:-translate-y-1 hover:shadow-[0_22px_54px_rgba(111,84,61,0.12)]">
                    <div className="relative flex flex-1 flex-col gap-3 p-5">
                      <div className="space-y-1.5">
                        <div className="space-y-1">
                          <p className={campusEnglishNameClass}>
                            {school.englishName}
                          </p>
                          <h3
                            className="text-[1.85rem] leading-[0.95] tracking-[-0.04em] text-[#241d1a]"
                            style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                          >
                            {school.name}
                          </h3>
                        </div>

                      </div>

                      <div className="flex items-center justify-between gap-3 pt-0.5">
                        <span className="inline-flex items-center gap-2 text-sm font-medium text-[#7c5f4f]">
                          进入校园
                          <ArrowRight className="transition group-hover:translate-x-1" size={16} />
                        </span>
                        <span className="ml-auto inline-flex shrink-0 items-center justify-end rounded-full bg-[#f6efe6] px-3 py-1 text-[0.72rem] font-medium tracking-[0.08em] text-[#8a6a57]">
                          {formatRegisteredUserCountLabel(registeredUserCounts, school.key)}
                        </span>
                      </div>
                    </div>
                  </article>
                </Link>
              ))}
            </div>

            <div className="mt-10 rounded-[30px] border border-[#eadccd] bg-[linear-gradient(135deg,rgba(255,250,244,0.96)_0%,rgba(249,241,233,0.92)_100%)] px-6 py-7 shadow-[0_18px_44px_rgba(111,84,61,0.08)] sm:px-8">
              <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
                <div className="space-y-2">
                  <p className="text-sm uppercase tracking-[0.24em] text-[#9b816f]">Open Invitation</p>
                  <h3
                    className="text-[1.9rem] leading-[1.08] tracking-[-0.04em] text-[#241d1a] sm:text-[2.3rem]"
                    style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                  >
                    欢迎更多高校入驻 Soulink 高校联盟
                  </h3>
                  <p className="max-w-3xl text-[0.98rem] leading-7 text-[#6d5f56]">
                    如果你也期待让本校同学拥有更认真、更真诚的连接方式，Soulink 欢迎更多校园一起加入这条高校联盟的路径。
                  </p>
                </div>
                <button
                  className={cn(
                    buttonVariants({ variant: 'secondary' }),
                    'inline-flex rounded-full border-[#d8c5b3] bg-white/80 px-6 py-6 text-sm tracking-[0.14em] text-[#7b6353] shadow-none hover:bg-white',
                  )}
                  onClick={() => setContactModalOpen(true)}
                  type="button"
                >
                  联系我们加入联盟
                </button>
              </div>
            </div>
          </div>
        </section>

        <LandingSharedSections />
      </div>

      <ContactUsModal
        contactEmail={contactEmail}
        onClose={() => setContactModalOpen(false)}
        open={contactModalOpen}
      />
    </>
  )
}
