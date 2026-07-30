import {
  ArrowRight,
  Check,
  Coffee,
  HeartHandshake,
  Menu,
  Rocket,
  ShieldCheck,
  Sparkles,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import logoUrl from '@/assets/soulink-logo-icon.png'
import thuHeroAvif from '@/assets/optimized/heroes/thu.avif'
import thuHeroWebp from '@/assets/optimized/heroes/thu.webp'

const scenes = [
  {
    icon: Coffee,
    title: '今晚，找个饭搭子',
    meta: '紫荆 · 19:00',
    copy: '想边吃饭边聊聊 AI 产品，不赶时间。',
  },
  {
    icon: Rocket,
    title: '为项目找 CTO',
    meta: '每周 10 小时 · 长期',
    copy: '需要工程能力强、愿意一起验证需求的人。',
  },
  {
    icon: HeartHandshake,
    title: '认真认识一个人',
    meta: '私密推荐 · 双方自愿',
    copy: '不公开展示，不滑动浏览，慢一点了解。',
  },
]

export function V2LandingPage() {
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <div className="min-h-screen bg-[#f7f6f3] text-[#23211f]">
      <header className="fixed inset-x-0 top-0 z-40 border-b border-white/15 bg-[rgba(23,20,20,0.42)] backdrop-blur-md">
        <div className="mx-auto flex h-16 max-w-[1180px] items-center justify-between px-4 sm:px-6 lg:px-8">
          <Link className="flex items-center gap-2.5 text-white" to="/">
            <img alt="" className="h-8 w-8 rounded-md object-cover" src={logoUrl} />
            <span>
              <span className="block text-sm font-semibold">SouLink</span>
              <span className="block text-[10px] text-white/70">清华校内</span>
            </span>
          </Link>
          <nav className="hidden items-center gap-6 text-sm text-white/84 md:flex">
            <a className="hover:text-white" href="#scenes">
              校园场景
            </a>
            <a className="hover:text-white" href="#how-it-works">
              如何连接
            </a>
            <Link className="hover:text-white" to="/safety">
              隐私安全
            </Link>
            <Link className="hover:text-white" to="/login">
              登录
            </Link>
            <Link
              className="inline-flex h-9 items-center rounded-md bg-white px-4 font-semibold text-[#4d2d66]"
              to="/register"
            >
              清华邮箱注册
            </Link>
          </nav>
          <button
            aria-label={menuOpen ? '关闭菜单' : '打开菜单'}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-white hover:bg-white/10 md:hidden"
            onClick={() => setMenuOpen((open) => !open)}
            type="button"
          >
            {menuOpen ? <X size={20} /> : <Menu size={20} />}
          </button>
        </div>
        {menuOpen ? (
          <nav className="border-t border-white/15 bg-[rgba(24,21,22,0.92)] px-4 py-3 text-sm text-white md:hidden">
            <a className="block rounded-md px-3 py-2.5" href="#scenes">
              校园场景
            </a>
            <Link className="block rounded-md px-3 py-2.5" to="/safety">
              隐私安全
            </Link>
            <Link className="block rounded-md px-3 py-2.5" to="/login">
              登录
            </Link>
            <Link
              className="mt-2 flex h-10 items-center justify-center rounded-md bg-white font-semibold text-[#4d2d66]"
              to="/register"
            >
              清华邮箱注册
            </Link>
          </nav>
        ) : null}
      </header>

      <section className="relative flex min-h-[76vh] max-h-[780px] items-end overflow-hidden pt-16">
        <picture className="absolute inset-0 block">
          <source srcSet={thuHeroAvif} type="image/avif" />
          <source srcSet={thuHeroWebp} type="image/webp" />
          <img
            alt="清华校园"
            className="h-full w-full object-cover object-center"
            decoding="async"
            fetchPriority="high"
            src={thuHeroWebp}
          />
        </picture>
        <div className="absolute inset-0 bg-[rgba(21,17,22,0.48)]" />

        <div className="relative z-10 mx-auto w-full max-w-[1180px] px-4 pb-12 text-white sm:px-6 sm:pb-16 lg:px-8">
          <p className="mb-4 inline-flex items-center gap-2 text-sm font-medium text-white/80">
            <ShieldCheck size={16} />
            仅限清华身份验证用户
          </p>
          <h1 className="text-5xl font-semibold leading-none sm:text-6xl lg:text-7xl">
            SouLink
          </h1>
          <p className="mt-5 max-w-2xl text-xl font-medium leading-8 text-white sm:text-2xl">
            在清华，找到下一位值得认识的人。
          </p>
          <p className="mt-3 max-w-xl text-sm leading-7 text-white/76 sm:text-base">
            说清楚你为什么想认识一个人。SouLink 帮你整理意图、给出少量解释充分的推荐，并在双方愿意后把连接带到校园里。
          </p>
          <div className="mt-7 flex flex-wrap gap-3">
            <Link
              className="inline-flex h-11 items-center gap-2 rounded-md bg-white px-5 text-sm font-semibold text-[#4d2d66] hover:bg-[#f2edf5]"
              to="/register"
            >
              告诉 SouLink 你想认识谁
              <ArrowRight size={16} />
            </Link>
            <Link
              className="inline-flex h-11 items-center rounded-md border border-white/40 px-5 text-sm font-semibold text-white hover:bg-white/10"
              to="/login"
            >
              已有账号
            </Link>
          </div>
        </div>
      </section>

      <section className="border-b border-[#dfdcd6] bg-[#fbfaf8] py-6">
        <div className="mx-auto grid max-w-[1180px] grid-cols-1 gap-4 px-4 text-sm text-[#5f5b55] sm:grid-cols-3 sm:px-6 lg:px-8">
          <p className="flex items-center gap-2">
            <Check size={16} className="text-[#2d6a53]" />
            意图优先，不公开陈列人
          </p>
          <p className="flex items-center gap-2">
            <Check size={16} className="text-[#2d6a53]" />
            默认最多三条高相关推荐
          </p>
          <p className="flex items-center gap-2">
            <Check size={16} className="text-[#2d6a53]" />
            双方同意后才揭示联系方式
          </p>
        </div>
      </section>

      <section className="mx-auto max-w-[1180px] px-4 py-16 sm:px-6 lg:px-8" id="scenes">
        <div className="max-w-2xl">
          <p className="text-xs font-semibold text-[#6c4a80]">TODAY AT TSINGHUA</p>
          <h2 className="mt-2 text-3xl font-semibold text-[#292724]">从一个具体理由开始</h2>
          <p className="mt-3 text-sm leading-7 text-[#706c65]">
            不是先看一张脸，而是先知道为什么你们可能值得聊一次。
          </p>
        </div>
        <div className="mt-8 grid gap-3 md:grid-cols-3">
          {scenes.map((scene) => {
            const Icon = scene.icon
            return (
              <article
                className="rounded-lg border border-[#dfdcd6] bg-white p-5"
                key={scene.title}
              >
                <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#eee8f2] text-[#5b3477]">
                  <Icon size={18} />
                </span>
                <h3 className="mt-4 text-base font-semibold text-[#302e2b]">{scene.title}</h3>
                <p className="mt-1 text-xs font-medium text-[#6c4a80]">{scene.meta}</p>
                <p className="mt-3 text-sm leading-6 text-[#706c65]">{scene.copy}</p>
              </article>
            )
          })}
        </div>
      </section>

      <section className="border-y border-[#dfdcd6] bg-[#efede9] py-16" id="how-it-works">
        <div className="mx-auto max-w-[1180px] px-4 sm:px-6 lg:px-8">
          <div className="grid gap-10 lg:grid-cols-[0.72fr_1.28fr]">
            <div>
              <p className="flex items-center gap-2 text-xs font-semibold text-[#6c4a80]">
                <Sparkles size={14} />
                AI INSIDE THE MOMENT
              </p>
              <h2 className="mt-2 text-3xl font-semibold text-[#292724]">AI 做媒介，人做决定</h2>
            </div>
            <ol className="grid gap-px overflow-hidden rounded-lg border border-[#d5d1cb] bg-[#d5d1cb] sm:grid-cols-3">
              {[
                ['01', '说出意图', '一句自然语言，补充最多三个真正重要的问题。'],
                ['02', '理解推荐', '看到为什么可能合适，也看到仍然不确定的地方。'],
                ['03', '双方同意', '进入匹配房间，约时间地点，再决定是否交换联系。'],
              ].map(([number, title, copy]) => (
                <li className="bg-[#fbfaf8] p-5" key={number}>
                  <span className="text-xs font-bold text-[#6c4a80]">{number}</span>
                  <h3 className="mt-3 text-base font-semibold text-[#302e2b]">{title}</h3>
                  <p className="mt-2 text-sm leading-6 text-[#706c65]">{copy}</p>
                </li>
              ))}
            </ol>
          </div>
        </div>
      </section>

      <footer className="bg-[#262329] py-8 text-white/70">
        <div className="mx-auto flex max-w-[1180px] flex-col gap-4 px-4 text-xs sm:flex-row sm:items-center sm:justify-between sm:px-6 lg:px-8">
          <p>SouLink · 清华校内 AI 关系匹配</p>
          <div className="flex gap-5">
            <Link className="hover:text-white" to="/safety">
              隐私与安全
            </Link>
            <Link className="hover:text-white" to="/login">
              登录
            </Link>
          </div>
        </div>
      </footer>
    </div>
  )
}
