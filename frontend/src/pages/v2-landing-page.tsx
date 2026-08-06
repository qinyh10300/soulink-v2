import { motion } from 'framer-motion'
import {
  ArrowRight,
  BadgeCheck,
  BookOpen,
  BrainCircuit,
  BriefcaseBusiness,
  CalendarCheck2,
  Check,
  ChevronDown,
  Coffee,
  Dumbbell,
  Heart,
  Lightbulb,
  LockKeyhole,
  Menu,
  MessageCircle,
  Rocket,
  Search,
  ShieldCheck,
  Sparkles,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import { ProductStage } from '@/components/landing/product-stage'

const navigation = [
  { href: '#experience', label: '产品体验' },
  { href: '#scenarios', label: '连接场景' },
  { href: '#safety', label: '隐私安全' },
  { href: '#faq', label: '常见问题' },
]

const scenarioItems = [
  { icon: Coffee, label: '饭搭子' },
  { icon: BookOpen, label: '学习搭子' },
  { icon: Dumbbell, label: '运动搭子' },
  { icon: CalendarCheck2, label: '活动同伴' },
  { icon: BriefcaseBusiness, label: '项目伙伴' },
  { icon: Rocket, label: 'CTO / 联创' },
  { icon: Heart, label: '认真关系' },
]

const capabilities = [
  {
    icon: MessageCircle,
    title: '自然语言表达',
    copy: '不用填写冗长标签，像聊天一样说清楚此刻的需求。',
  },
  {
    icon: BrainCircuit,
    title: 'AI 意图理解',
    copy: '追问真正影响匹配的问题，区分必要条件与开放偏好。',
  },
  {
    icon: Search,
    title: '少量精准推荐',
    copy: '默认只给少量候选，让每一次判断都有足够的信息。',
  },
  {
    icon: Lightbulb,
    title: '推荐理由可解释',
    copy: '告诉你为什么可能合适，也诚实呈现仍然未知的部分。',
  },
  {
    icon: CalendarCheck2,
    title: '把连接带到线下',
    copy: '双方同意后协助约时间、地点和第一次见面的破冰话题。',
  },
  {
    icon: LockKeyhole,
    title: '隐私默认保护',
    copy: '身份信息按需展示，联系方式只在双方确认后揭示。',
  },
]

const faqItems = [
  {
    question: 'SouLink 和普通社交软件有什么不同？',
    answer:
      'SouLink 从「你为什么想认识一个人」开始，而不是让你无限浏览资料。AI 会理解具体场景，给出少量、可解释的推荐，决定权始终在你。',
  },
  {
    question: '为什么目前只开放清华大学？',
    answer:
      '校园身份让匹配拥有更真实的共同环境，也更容易把线上意图转化为安全、低成本的线下见面。第一阶段只服务清华校内用户。',
  },
  {
    question: 'AI 会替我自动联系别人吗？',
    answer:
      '不会。AI 只协助理解意图、整理推荐理由和规划见面。发送连接邀请、进入匹配房间以及交换联系方式都需要你的明确确认。',
  },
  {
    question: '我的信息会被公开展示吗？',
    answer:
      '不会公开形成可搜索的人物广场。只有在相关意图下进入候选范围时，系统才会按隐私设置展示必要信息。',
  },
]

const reveal = {
  initial: { opacity: 0, y: 24 },
  transition: { duration: 0.6, ease: [0.22, 1, 0.36, 1] },
  viewport: { once: true, amount: 0.2 },
  whileInView: { opacity: 1, y: 0 },
} as const

function SouLinkMark() {
  return (
    <span className="flex items-center gap-2.5">
      <span className="flex h-8 w-8 items-center justify-center rounded-md bg-[#1447e6] text-white shadow-[0_8px_20px_rgba(20,71,230,0.24)]">
        <Sparkles size={17} strokeWidth={2.2} />
      </span>
      <span className="text-[17px] font-semibold tracking-[0] text-[#09090b]">SouLink</span>
    </span>
  )
}

function IntentDemo() {
  return (
    <div className="overflow-hidden rounded-lg border border-[#e4e4e7] bg-white shadow-[0_24px_70px_rgba(9,9,11,0.08)]">
      <div className="flex h-11 items-center justify-between border-b border-[#e4e4e7] px-4">
        <div className="flex items-center gap-2 text-xs font-semibold text-[#18181b]">
          <span className="flex h-6 w-6 items-center justify-center rounded-md bg-[#1447e6] text-white">
            <Sparkles size={12} />
          </span>
          AI 连接助手
        </div>
        <span className="flex items-center gap-1.5 text-[10px] text-[#71717a]">
          <span className="h-1.5 w-1.5 rounded-full bg-[#16a36a]" />
          正在理解
        </span>
      </div>
      <div className="space-y-5 p-5 sm:p-7">
        <div className="ml-auto max-w-[82%] rounded-lg bg-[#1447e6] px-4 py-3 text-sm leading-6 text-white">
          想找一位做过 AI 产品的同学，今晚一起吃饭，顺便交流一下 Agent 产品方向。
        </div>
        <div className="flex max-w-[88%] gap-3">
          <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-[#eff4ff] text-[#1447e6]">
            <Sparkles size={14} />
          </span>
          <div>
            <p className="text-sm leading-6 text-[#3f3f46]">
              明白了。你更希望这次见面偏向轻松交流，还是寻找未来可以一起做项目的人？
            </p>
            <div className="mt-3 flex flex-wrap gap-2">
              {['轻松交流', '可能合作', '都可以'].map((item) => (
                <button
                  className="rounded-full border border-[#d4d4d8] px-3 py-1.5 text-xs text-[#52525b] transition-colors hover:border-[#1447e6] hover:text-[#1447e6]"
                  key={item}
                  type="button"
                >
                  {item}
                </button>
              ))}
            </div>
          </div>
        </div>
        <div className="border-t border-[#e4e4e7] pt-4">
          <div className="flex min-h-12 items-center justify-between rounded-lg border border-[#d4d4d8] px-4 text-xs text-[#a1a1aa]">
            继续补充你的期待…
            <span className="flex h-7 w-7 items-center justify-center rounded-md bg-[#1447e6] text-white">
              <ArrowRight size={13} />
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

function RecommendationDemo() {
  return (
    <div className="overflow-hidden rounded-lg border border-[#27272a] bg-[#09090b] text-white shadow-[0_30px_80px_rgba(9,9,11,0.18)]">
      <div className="flex h-11 items-center justify-between border-b border-white/10 px-4">
        <span className="text-xs font-semibold">今日推荐 · 1 / 3</span>
        <span className="text-[10px] text-white/38">为当前意图生成</span>
      </div>
      <div className="p-5 sm:p-7">
        <div className="flex items-start gap-4 border-b border-white/10 pb-5">
          <span className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-[#1447e6] text-sm font-semibold">
            周
          </span>
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="font-semibold">周同学</h3>
              <BadgeCheck className="text-[#6f8eff]" size={15} />
              <span className="rounded-full bg-[#153e34] px-2 py-1 text-[9px] font-medium text-[#78d7b4]">
                清华已认证
              </span>
            </div>
            <p className="mt-1 text-xs text-white/42">交叉信息院 · AI 产品方向</p>
          </div>
          <span className="text-sm font-semibold text-[#7c98ff]">92%</span>
        </div>
        <div className="divide-y divide-white/10">
          {[
            ['共同目标', '都希望交流 Agent 产品，并对未来合作保持开放'],
            ['时间重合', '今晚 18:30 后均可，活动范围接近紫荆'],
            ['仍需确认', '对方更偏技术实现，你更偏用户与产品'],
          ].map(([label, copy], index) => (
            <div className="grid gap-2 py-4 sm:grid-cols-[92px_1fr]" key={label}>
              <p className={`text-xs font-medium ${index === 2 ? 'text-[#f0ad68]' : 'text-[#7c98ff]'}`}>
                {label}
              </p>
              <p className="text-xs leading-5 text-white/58">{copy}</p>
            </div>
          ))}
        </div>
        <div className="grid grid-cols-2 gap-2 border-t border-white/10 pt-5">
          <button
            className="h-10 rounded-md border border-white/14 text-xs font-medium text-white/64 transition-colors hover:bg-white/5"
            type="button"
          >
            暂不合适
          </button>
          <button
            className="h-10 rounded-md bg-[#1447e6] text-xs font-semibold transition-colors hover:bg-[#103bc4]"
            type="button"
          >
            想认识
          </button>
        </div>
      </div>
    </div>
  )
}

function ConsentDemo() {
  return (
    <div className="overflow-hidden rounded-lg border border-[#e4e4e7] bg-white shadow-[0_24px_70px_rgba(9,9,11,0.08)]">
      <div className="border-b border-[#e4e4e7] px-5 py-4 sm:px-7">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <span className="flex h-7 w-7 items-center justify-center rounded-md bg-[#e9f8f1] text-[#087a4d]">
            <Check size={14} />
          </span>
          你们都想认识对方
        </div>
      </div>
      <div className="p-5 sm:p-7">
        <div className="relative flex items-center justify-center py-3">
          <span className="flex h-14 w-14 items-center justify-center rounded-full bg-[#eff4ff] text-sm font-semibold text-[#1447e6] ring-4 ring-white">
            你
          </span>
          <span className="-ml-2 flex h-14 w-14 items-center justify-center rounded-full bg-[#18181b] text-sm font-semibold text-white ring-4 ring-white">
            周
          </span>
        </div>
        <div className="mt-5 divide-y divide-[#e4e4e7] border-y border-[#e4e4e7]">
          <div className="flex items-center gap-3 py-4">
            <CalendarCheck2 className="text-[#1447e6]" size={17} />
            <div className="flex-1">
              <p className="text-[10px] text-[#71717a]">双方都方便</p>
              <p className="mt-1 text-sm font-medium">今天 19:00 · 紫荆园一层</p>
            </div>
          </div>
          <div className="flex items-center gap-3 py-4">
            <MessageCircle className="text-[#1447e6]" size={17} />
            <div className="flex-1">
              <p className="text-[10px] text-[#71717a]">AI 破冰建议</p>
              <p className="mt-1 text-sm font-medium">聊聊各自最想解决的 Agent 用户问题</p>
            </div>
          </div>
        </div>
        <button
          className="mt-5 flex h-10 w-full items-center justify-center gap-2 rounded-md bg-[#1447e6] text-xs font-semibold text-white transition-colors hover:bg-[#103bc4]"
          type="button"
        >
          进入匹配房间
          <ArrowRight size={14} />
        </button>
      </div>
    </div>
  )
}

export function V2LandingPage() {
  const [menuOpen, setMenuOpen] = useState(false)
  const [openFaq, setOpenFaq] = useState(0)

  return (
    <div className="min-h-screen overflow-x-hidden bg-white text-[#09090b]">
      <header className="sticky top-0 z-50 border-b border-[#e4e4e7]/80 bg-white/88 backdrop-blur-xl">
        <div className="mx-auto flex h-16 max-w-[1200px] items-center justify-between px-4 sm:px-6 lg:px-8">
          <Link aria-label="SouLink 首页" to="/">
            <SouLinkMark />
          </Link>

          <nav className="hidden items-center gap-7 text-[13px] font-medium text-[#52525b] md:flex">
            {navigation.map((item) => (
              <a className="transition-colors hover:text-[#09090b]" href={item.href} key={item.href}>
                {item.label}
              </a>
            ))}
          </nav>

          <div className="hidden items-center gap-2 md:flex">
            <Link
              className="inline-flex h-9 items-center px-3 text-[13px] font-medium text-[#3f3f46] transition-colors hover:text-[#09090b]"
              to="/login"
            >
              登录
            </Link>
            <Link
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1447e6] px-4 text-[13px] font-semibold text-white transition-colors hover:bg-[#103bc4]"
              to="/register"
            >
              快速开始
              <ArrowRight size={14} />
            </Link>
          </div>

          <button
            aria-expanded={menuOpen}
            aria-label={menuOpen ? '关闭菜单' : '打开菜单'}
            className="flex h-9 w-9 items-center justify-center rounded-md text-[#3f3f46] transition-colors hover:bg-[#f4f4f5] md:hidden"
            onClick={() => setMenuOpen((open) => !open)}
            type="button"
          >
            {menuOpen ? <X size={19} /> : <Menu size={19} />}
          </button>
        </div>

        {menuOpen ? (
          <nav className="border-t border-[#e4e4e7] bg-white px-4 pb-5 pt-3 md:hidden">
            {navigation.map((item) => (
              <a
                className="block rounded-md px-3 py-2.5 text-sm font-medium text-[#3f3f46] hover:bg-[#f4f4f5]"
                href={item.href}
                key={item.href}
                onClick={() => setMenuOpen(false)}
              >
                {item.label}
              </a>
            ))}
            <div className="mt-3 grid grid-cols-2 gap-2 border-t border-[#e4e4e7] pt-4">
              <Link
                className="flex h-10 items-center justify-center rounded-md border border-[#d4d4d8] text-sm font-medium"
                to="/login"
              >
                登录
              </Link>
              <Link
                className="flex h-10 items-center justify-center rounded-md bg-[#1447e6] text-sm font-semibold text-white"
                to="/register"
              >
                快速开始
              </Link>
            </div>
          </nav>
        ) : null}
      </header>

      <main>
        <section className="border-b border-[#e4e4e7] pt-16 sm:pt-20 lg:pt-24">
          <motion.div
            animate={{ opacity: 1, y: 0 }}
            className="mx-auto max-w-[1040px] px-4 text-center sm:px-6 lg:px-8"
            initial={{ opacity: 0, y: 18 }}
            transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1] }}
          >
            <Link
              className="inline-flex items-center gap-2 rounded-full border border-[#d4d4d8] bg-white px-3 py-1.5 text-[11px] font-medium text-[#52525b] shadow-[0_4px_14px_rgba(9,9,11,0.04)] transition-colors hover:border-[#a1a1aa]"
              to="/register"
            >
              <span className="h-1.5 w-1.5 rounded-full bg-[#1447e6]" />
              SouLink 2.0 · 清华校内开放
              <ArrowRight size={12} />
            </Link>
            <h1 className="mx-auto mt-6 max-w-[900px] text-balance text-[42px] font-semibold leading-[1.08] tracking-[0] text-[#09090b] sm:text-[62px] lg:text-[72px]">
              释放你的连接潜能
              <span className="mt-1 block text-[#1447e6]">AI 驱动，遇见对的人</span>
            </h1>
            <p className="mx-auto mt-6 max-w-[680px] text-balance text-sm leading-7 text-[#52525b] sm:text-base sm:leading-8">
              从饭搭子、学习伙伴到项目合伙人，SouLink 用 AI 理解你的真实意图，
              在清华校内给出少量、可解释、能够走向线下的匹配。
            </p>
            <div className="mt-7 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Link
                className="inline-flex h-11 w-full items-center justify-center gap-2 rounded-full bg-[#1447e6] px-6 text-sm font-semibold text-white shadow-[0_12px_28px_rgba(20,71,230,0.22)] transition-all hover:-translate-y-0.5 hover:bg-[#103bc4] sm:w-auto"
                to="/register"
              >
                告诉 AI 你想认识谁
                <ArrowRight size={15} />
              </Link>
              <a
                className="inline-flex h-11 w-full items-center justify-center rounded-full border border-[#d4d4d8] bg-white px-6 text-sm font-semibold text-[#27272a] transition-colors hover:bg-[#f4f4f5] sm:w-auto"
                href="#experience"
              >
                看看如何工作
              </a>
            </div>
            <div className="mt-7 flex items-center justify-center gap-3 text-left">
              <div className="flex -space-x-2">
                {['清', '华', 'AI'].map((label, index) => (
                  <span
                    className={`flex h-8 w-8 items-center justify-center rounded-full border-2 border-white text-[9px] font-semibold ${
                      index === 0
                        ? 'bg-[#09090b] text-white'
                        : index === 1
                          ? 'bg-[#e8eefc] text-[#1447e6]'
                          : 'bg-[#dff5eb] text-[#087a4d]'
                    }`}
                    key={label}
                  >
                    {label}
                  </span>
                ))}
              </div>
              <p className="text-[11px] leading-4 text-[#71717a]">
                为清华同学的每一次
                <br />
                真实连接而设计
              </p>
            </div>
          </motion.div>

          <ProductStage />
        </section>

        <section className="overflow-hidden border-b border-[#e4e4e7] bg-[#fafafa] py-5" id="scenarios">
          <div className="soulink-marquee-track flex w-max items-center gap-3">
            {[...scenarioItems, ...scenarioItems].map((item, index) => {
              const Icon = item.icon
              return (
                <div
                  className="flex h-10 items-center gap-2 rounded-full border border-[#e4e4e7] bg-white px-4 text-xs font-medium text-[#3f3f46]"
                  key={`${item.label}-${index}`}
                >
                  <Icon className="text-[#1447e6]" size={14} />
                  {item.label}
                </div>
              )
            })}
          </div>
        </section>

        <section className="px-4 py-20 sm:px-6 sm:py-28 lg:px-8" id="experience">
          <motion.div className="mx-auto max-w-[1000px] text-center" {...reveal}>
            <p className="text-xs font-semibold text-[#1447e6]">AI INSIDE THE MOMENT</p>
            <h2 className="mt-4 text-balance text-3xl font-semibold leading-tight sm:text-5xl">
              重新定义校内连接体验
            </h2>
            <p className="mx-auto mt-5 max-w-[680px] text-sm leading-7 text-[#71717a] sm:text-base sm:leading-8">
              AI 不替你做社交决定。它负责听懂、筛选、解释和推进，让你把注意力留给真正值得认识的人。
            </p>
          </motion.div>
        </section>

        <section className="border-y border-[#e4e4e7]">
          <div className="mx-auto grid max-w-[1200px] items-center gap-10 px-4 py-16 sm:px-6 sm:py-20 lg:grid-cols-2 lg:gap-20 lg:px-8">
            <motion.div {...reveal}>
              <span className="flex h-9 w-9 items-center justify-center rounded-md bg-[#eff4ff] text-[#1447e6]">
                <MessageCircle size={18} />
              </span>
              <p className="mt-6 text-xs font-semibold text-[#1447e6]">01 · 说出真实意图</p>
              <h2 className="mt-3 text-3xl font-semibold leading-tight sm:text-4xl">从一句人话开始，而不是一堆标签</h2>
              <p className="mt-5 max-w-[520px] text-sm leading-7 text-[#71717a] sm:text-base sm:leading-8">
                告诉 SouLink 你此刻为什么想认识一个人。AI 只追问真正影响结果的问题，并把模糊想法整理成清晰、可调整的连接意图。
              </p>
              <div className="mt-6 space-y-3">
                {['必要条件和开放偏好分开表达', '你随时可以修改或暂停意图'].map((item) => (
                  <p className="flex items-center gap-2 text-sm font-medium text-[#3f3f46]" key={item}>
                    <Check className="text-[#087a4d]" size={15} />
                    {item}
                  </p>
                ))}
              </div>
            </motion.div>
            <motion.div {...reveal} transition={{ ...reveal.transition, delay: 0.12 }}>
              <IntentDemo />
            </motion.div>
          </div>
        </section>

        <section className="border-b border-[#e4e4e7] bg-[#fafafa]">
          <div className="mx-auto grid max-w-[1200px] items-center gap-10 px-4 py-16 sm:px-6 sm:py-20 lg:grid-cols-2 lg:gap-20 lg:px-8">
            <motion.div className="lg:order-2" {...reveal}>
              <span className="flex h-9 w-9 items-center justify-center rounded-md bg-[#eff4ff] text-[#1447e6]">
                <Search size={18} />
              </span>
              <p className="mt-6 text-xs font-semibold text-[#1447e6]">02 · 理解每次推荐</p>
              <h2 className="mt-3 text-3xl font-semibold leading-tight sm:text-4xl">少一点人选，多一点判断依据</h2>
              <p className="mt-5 max-w-[520px] text-sm leading-7 text-[#71717a] sm:text-base sm:leading-8">
                SouLink 不制造无限刷人的焦虑。每次最多给出三位候选，并说明共同目标、时间条件、匹配依据和仍需确认的地方。
              </p>
              <div className="mt-6 space-y-3">
                {['推荐理由完整可见', '反馈会让后续理解更准确'].map((item) => (
                  <p className="flex items-center gap-2 text-sm font-medium text-[#3f3f46]" key={item}>
                    <Check className="text-[#087a4d]" size={15} />
                    {item}
                  </p>
                ))}
              </div>
            </motion.div>
            <motion.div className="lg:order-1" {...reveal} transition={{ ...reveal.transition, delay: 0.12 }}>
              <RecommendationDemo />
            </motion.div>
          </div>
        </section>

        <section className="border-b border-[#e4e4e7]" id="safety">
          <div className="mx-auto grid max-w-[1200px] items-center gap-10 px-4 py-16 sm:px-6 sm:py-20 lg:grid-cols-2 lg:gap-20 lg:px-8">
            <motion.div {...reveal}>
              <span className="flex h-9 w-9 items-center justify-center rounded-md bg-[#e9f8f1] text-[#087a4d]">
                <ShieldCheck size={18} />
              </span>
              <p className="mt-6 text-xs font-semibold text-[#087a4d]">03 · 双方决定是否靠近</p>
              <h2 className="mt-3 text-3xl font-semibold leading-tight sm:text-4xl">AI 做媒介，人做决定</h2>
              <p className="mt-5 max-w-[520px] text-sm leading-7 text-[#71717a] sm:text-base sm:leading-8">
                连接邀请、进入房间、交换联系方式，每一步都需要明确同意。双方愿意后，AI 再协助把见面变成一个具体、自然的校园计划。
              </p>
              <Link
                className="mt-7 inline-flex items-center gap-2 text-sm font-semibold text-[#1447e6] hover:text-[#103bc4]"
                to="/safety"
              >
                了解隐私与安全机制
                <ArrowRight size={14} />
              </Link>
            </motion.div>
            <motion.div {...reveal} transition={{ ...reveal.transition, delay: 0.12 }}>
              <ConsentDemo />
            </motion.div>
          </div>
        </section>

        <section className="px-4 py-20 sm:px-6 sm:py-28 lg:px-8">
          <motion.div className="mx-auto max-w-[1200px]" {...reveal}>
            <p className="text-xs font-semibold text-[#1447e6]">BUILT FOR REAL CONNECTIONS</p>
            <h2 className="mt-4 max-w-[820px] text-balance text-3xl font-semibold leading-tight sm:text-5xl">
              找搭子、找对象、找 CTO
              <span className="block text-[#a1a1aa]">需要的和不需要的，都想清楚了</span>
            </h2>
            <div className="mt-12 grid border-l border-t border-[#e4e4e7] sm:grid-cols-2 lg:grid-cols-3">
              {capabilities.map((item) => {
                const Icon = item.icon
                return (
                  <div className="min-h-[210px] border-b border-r border-[#e4e4e7] p-6 sm:p-8" key={item.title}>
                    <Icon className="text-[#1447e6]" size={20} />
                    <h3 className="mt-8 text-base font-semibold">{item.title}</h3>
                    <p className="mt-3 max-w-[290px] text-sm leading-6 text-[#71717a]">{item.copy}</p>
                  </div>
                )
              })}
            </div>
          </motion.div>
        </section>

        <section className="bg-[#09090b] px-4 py-20 text-white sm:px-6 sm:py-28 lg:px-8">
          <motion.div className="mx-auto max-w-[1200px]" {...reveal}>
            <div className="grid gap-10 lg:grid-cols-[1.3fr_0.7fr] lg:items-end">
              <div>
                <p className="text-xs font-semibold text-[#7c98ff]">QUALITY OVER QUANTITY</p>
                <h2 className="mt-5 max-w-[820px] text-balance text-4xl font-semibold leading-[1.12] sm:text-6xl">
                  你可以刷过很多人
                  <span className="block text-white/38">但你不必如此</span>
                </h2>
              </div>
              <p className="max-w-[480px] text-sm leading-7 text-white/52 sm:text-base sm:leading-8 lg:pb-1">
                真正的连接不需要无穷选择。SouLink 把重点放在此刻的意图、可信的背景和一次愿意发生的见面。
              </p>
            </div>

            <div className="mt-16 grid border-l border-t border-white/12 sm:grid-cols-2 lg:grid-cols-4">
              {[
                ['1', '所大学', '目前仅服务清华大学'],
                ['≤ 3', '位候选', '每轮保持少量推荐'],
                ['100%', '双向确认', '关键动作由你决定'],
                ['7', '类场景', '覆盖日常到长期关系'],
              ].map(([value, unit, copy]) => (
                <div className="min-h-[170px] border-b border-r border-white/12 p-6" key={unit}>
                  <p className="text-3xl font-semibold">
                    {value} <span className="text-sm text-white/42">{unit}</span>
                  </p>
                  <p className="mt-8 text-xs text-white/38">{copy}</p>
                </div>
              ))}
            </div>
          </motion.div>
        </section>

        <section className="border-b border-[#e4e4e7] px-4 py-20 sm:px-6 sm:py-28 lg:px-8" id="faq">
          <div className="mx-auto grid max-w-[1100px] gap-12 lg:grid-cols-[0.72fr_1.28fr] lg:gap-20">
            <motion.div {...reveal}>
              <p className="text-xs font-semibold text-[#1447e6]">FAQ</p>
              <h2 className="mt-4 text-3xl font-semibold sm:text-4xl">你可能还想知道</h2>
              <p className="mt-4 max-w-[360px] text-sm leading-7 text-[#71717a]">
                对身份、隐私或匹配方式还有疑问，可以先从这里了解。
              </p>
            </motion.div>
            <motion.div className="border-t border-[#d4d4d8]" {...reveal}>
              {faqItems.map((item, index) => {
                const isOpen = openFaq === index
                return (
                  <div className="border-b border-[#d4d4d8]" key={item.question}>
                    <button
                      aria-controls={`faq-answer-${index}`}
                      aria-expanded={isOpen}
                      className="flex w-full items-center justify-between gap-6 py-5 text-left text-sm font-semibold sm:text-base"
                      onClick={() => setOpenFaq(isOpen ? -1 : index)}
                      type="button"
                    >
                      {item.question}
                      <ChevronDown
                        className={`shrink-0 text-[#71717a] transition-transform ${isOpen ? 'rotate-180' : ''}`}
                        size={17}
                      />
                    </button>
                    <motion.div
                      animate={{ height: isOpen ? 'auto' : 0, opacity: isOpen ? 1 : 0 }}
                      className="overflow-hidden"
                      id={`faq-answer-${index}`}
                      initial={false}
                    >
                      <p className="max-w-[680px] pb-5 text-sm leading-7 text-[#71717a]">{item.answer}</p>
                    </motion.div>
                  </div>
                )
              })}
            </motion.div>
          </div>
        </section>

        <section className="px-4 py-20 text-center sm:px-6 sm:py-28 lg:px-8">
          <motion.div className="mx-auto max-w-[800px]" {...reveal}>
            <span className="mx-auto flex h-11 w-11 items-center justify-center rounded-md bg-[#1447e6] text-white shadow-[0_12px_30px_rgba(20,71,230,0.24)]">
              <Sparkles size={21} />
            </span>
            <h2 className="mt-7 text-balance text-3xl font-semibold sm:text-5xl">下一位值得认识的人，也许就在清华园</h2>
            <p className="mx-auto mt-5 max-w-[580px] text-sm leading-7 text-[#71717a] sm:text-base">
              用一句话告诉 SouLink，你此刻为什么想认识一个人。
            </p>
            <Link
              className="mt-8 inline-flex h-11 items-center gap-2 rounded-full bg-[#1447e6] px-6 text-sm font-semibold text-white shadow-[0_12px_28px_rgba(20,71,230,0.2)] transition-all hover:-translate-y-0.5 hover:bg-[#103bc4]"
              to="/register"
            >
              使用清华邮箱开始
              <ArrowRight size={15} />
            </Link>
          </motion.div>
        </section>
      </main>

      <footer className="border-t border-[#e4e4e7] bg-[#fafafa]">
        <div className="mx-auto flex max-w-[1200px] flex-col gap-8 px-4 py-8 sm:px-6 md:flex-row md:items-center md:justify-between lg:px-8">
          <div>
            <SouLinkMark />
            <p className="mt-3 text-xs text-[#71717a]">清华校内 AI 连接平台</p>
          </div>
          <div className="flex flex-wrap gap-x-6 gap-y-3 text-xs font-medium text-[#52525b]">
            <Link className="hover:text-[#09090b]" to="/safety">
              隐私与安全
            </Link>
            <Link className="hover:text-[#09090b]" to="/login">
              登录
            </Link>
            <Link className="hover:text-[#09090b]" to="/register">
              清华邮箱注册
            </Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
