import {
  ArrowUp,
  CalendarDays,
  Check,
  ChevronRight,
  Clock3,
  Coffee,
  Heart,
  MessageCircle,
  MoreHorizontal,
  Search,
  ShieldCheck,
  Sparkles,
  Users,
} from 'lucide-react'
import type { ReactNode } from 'react'

type PreviewWindowProps = {
  children: ReactNode
  className?: string
  title: string
}

function PreviewWindow({ children, className = '', title }: PreviewWindowProps) {
  return (
    <div
      className={`overflow-hidden rounded-lg border border-white/10 bg-[#09090b] shadow-[0_36px_90px_rgba(9,9,11,0.2)] ${className}`}
    >
      <div className="flex h-10 items-center gap-2 border-b border-white/10 bg-[#18181b] px-3 text-white/45">
        <span className="h-2 w-2 rounded-full bg-white/18" />
        <span className="h-2 w-2 rounded-full bg-white/18" />
        <span className="h-2 w-2 rounded-full bg-white/18" />
        <span className="ml-2 min-w-0 truncate text-[10px] font-medium text-white/55">
          {title}
        </span>
      </div>
      {children}
    </div>
  )
}

function DiscoverPreview() {
  const people = [
    { initials: '周', label: 'AI 产品 / 交叉信息院', score: '92%' },
    { initials: '林', label: '羽毛球 / 经管学院', score: '88%' },
    { initials: '陈', label: '创业 / 美术学院', score: '84%' },
  ]

  return (
    <PreviewWindow className="h-[390px]" title="SouLink · 今日推荐">
      <div className="flex h-[350px] flex-col bg-[#111113] p-4 text-white">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-[10px] text-white/42">08 月 06 日 · 清华园</p>
            <p className="mt-1 text-sm font-semibold">值得认识的人</p>
          </div>
          <button
            aria-label="搜索"
            className="flex h-7 w-7 items-center justify-center rounded-md border border-white/10 text-white/60"
            type="button"
          >
            <Search size={13} />
          </button>
        </div>

        <div className="mt-4 border-y border-white/10 py-3">
          <div className="flex items-center gap-2 text-[10px] text-white/50">
            <Sparkles className="text-[#5b7cff]" size={12} />
            根据「找一位懂 AI 产品的晚饭搭子」
          </div>
        </div>

        <div className="divide-y divide-white/10">
          {people.map((person, index) => (
            <div className="flex items-center gap-3 py-3" key={person.label}>
              <span
                className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-xs font-semibold ${
                  index === 0
                    ? 'bg-[#1447e6] text-white'
                    : 'bg-white/10 text-white/70'
                }`}
              >
                {person.initials}
              </span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-xs font-medium">{person.label}</p>
                <p className="mt-1 text-[9px] text-white/38">共同话题 3 · 本周有空</p>
              </div>
              <span className="text-[10px] font-semibold text-[#7c98ff]">{person.score}</span>
              <ChevronRight className="text-white/28" size={13} />
            </div>
          ))}
        </div>

        <div className="mt-auto flex items-center gap-2 border-t border-white/10 pt-3 text-[9px] text-white/38">
          <ShieldCheck size={11} />
          少量推荐 · 理由可解释 · 双向确认
        </div>
      </div>
    </PreviewWindow>
  )
}

function AssistantPreview() {
  return (
    <PreviewWindow className="h-[500px]" title="SouLink AI · 新建连接">
      <div className="flex h-[460px] bg-[#09090b] text-white">
        <aside className="hidden w-[29%] shrink-0 flex-col border-r border-white/10 bg-[#111113] p-3 sm:flex">
          <div className="flex items-center gap-2 px-1">
            <span className="flex h-6 w-6 items-center justify-center rounded-md bg-[#1447e6]">
              <Sparkles size={13} />
            </span>
            <span className="text-[11px] font-semibold">SouLink</span>
          </div>
          <button
            className="mt-4 flex h-8 items-center gap-2 rounded-md border border-white/12 px-2.5 text-[10px] text-white/72"
            type="button"
          >
            <MessageCircle size={12} />
            新建连接
          </button>
          <p className="mt-5 px-1 text-[8px] font-semibold text-white/28">最近</p>
          <div className="mt-2 space-y-1">
            {['找今晚的饭搭子', '认识一位前端合伙人', '周末一起逛展'].map((item, index) => (
              <div
                className={`truncate rounded-md px-2 py-2 text-[9px] ${
                  index === 0 ? 'bg-white/8 text-white/80' : 'text-white/42'
                }`}
                key={item}
              >
                {item}
              </div>
            ))}
          </div>
          <div className="mt-auto flex items-center gap-2 border-t border-white/10 px-1 pt-3">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-white/12 text-[8px]">
              清
            </span>
            <div className="min-w-0">
              <p className="truncate text-[9px] font-medium">清华认证用户</p>
              <p className="text-[8px] text-white/30">在线</p>
            </div>
          </div>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex h-10 items-center justify-between border-b border-white/8 px-4">
            <span className="text-[9px] text-white/36">AI 连接助手</span>
            <MoreHorizontal size={14} className="text-white/34" />
          </div>
          <div className="flex flex-1 flex-col items-center justify-center px-4 pb-2 sm:px-6">
            <span className="flex h-9 w-9 items-center justify-center rounded-md bg-[#1447e6] shadow-[0_8px_24px_rgba(20,71,230,0.32)]">
              <Sparkles size={18} />
            </span>
            <h3 className="mt-4 text-center text-lg font-semibold sm:text-xl">今天想认识谁？</h3>
            <p className="mt-2 max-w-[310px] text-center text-[10px] leading-5 text-white/42">
              说说你的场景、时间和期待，SouLink 会帮你把模糊念头整理成一次可以发生的连接。
            </p>

            <div className="mt-5 grid w-full max-w-[390px] grid-cols-1 gap-2 sm:grid-cols-3">
              {[
                [Coffee, '今晚饭搭子'],
                [Users, '项目合伙人'],
                [Heart, '认真认识'],
              ].map(([Icon, label]) => {
                const SuggestionIcon = Icon as typeof Coffee
                return (
                  <button
                    className="flex min-h-16 flex-row items-center gap-2 rounded-md border border-white/10 bg-white/[0.025] p-2.5 text-left text-[9px] text-white/64 transition-colors hover:border-white/20 hover:bg-white/[0.05] sm:flex-col sm:items-start"
                    key={label as string}
                    type="button"
                  >
                    <SuggestionIcon className="text-[#7290ff]" size={13} />
                    {label as string}
                  </button>
                )
              })}
            </div>
          </div>

          <div className="px-3 pb-3 sm:px-5 sm:pb-4">
            <div className="rounded-lg border border-white/14 bg-[#18181b] p-2.5 shadow-[0_14px_40px_rgba(0,0,0,0.28)]">
              <p className="min-h-8 text-[10px] leading-5 text-white/68">
                想找一位做过 AI 产品的同学，今晚在紫荆吃饭聊聊…
              </p>
              <div className="mt-2 flex items-center justify-between">
                <div className="flex gap-1.5">
                  <span className="rounded-full bg-white/8 px-2 py-1 text-[8px] text-white/42">
                    今晚
                  </span>
                  <span className="rounded-full bg-white/8 px-2 py-1 text-[8px] text-white/42">
                    紫荆
                  </span>
                </div>
                <button
                  aria-label="发送"
                  className="flex h-6 w-6 items-center justify-center rounded-md bg-[#1447e6] text-white"
                  type="button"
                >
                  <ArrowUp size={12} />
                </button>
              </div>
            </div>
            <p className="mt-2 text-center text-[7px] text-white/24">
              AI 可能会犯错，重要信息请自行确认
            </p>
          </div>
        </div>
      </div>
    </PreviewWindow>
  )
}

function MatchRoomPreview() {
  return (
    <PreviewWindow className="h-[390px]" title="SouLink · 匹配房间">
      <div className="flex h-[350px] flex-col bg-[#111113] p-4 text-white">
        <div className="flex items-start justify-between">
          <div>
            <span className="rounded-full bg-[#143b31] px-2 py-1 text-[8px] font-semibold text-[#79d7b4]">
              双方已同意
            </span>
            <h3 className="mt-3 text-sm font-semibold">第一次见面计划</h3>
          </div>
          <span className="flex h-8 w-8 items-center justify-center rounded-full bg-[#1447e6] text-[10px] font-semibold">
            林
          </span>
        </div>

        <div className="mt-4 divide-y divide-white/10 border-y border-white/10">
          <div className="flex gap-3 py-3">
            <CalendarDays className="mt-0.5 text-[#7c98ff]" size={14} />
            <div>
              <p className="text-[9px] text-white/38">推荐时间</p>
              <p className="mt-1 text-[11px] font-medium">周四 19:00 - 20:00</p>
            </div>
          </div>
          <div className="flex gap-3 py-3">
            <Coffee className="mt-0.5 text-[#7c98ff]" size={14} />
            <div>
              <p className="text-[9px] text-white/38">推荐地点</p>
              <p className="mt-1 text-[11px] font-medium">紫荆园 · 一层东侧</p>
            </div>
          </div>
          <div className="flex gap-3 py-3">
            <Clock3 className="mt-0.5 text-[#7c98ff]" size={14} />
            <div>
              <p className="text-[9px] text-white/38">破冰话题</p>
              <p className="mt-1 text-[11px] font-medium">你们都关注 AI Agent 产品</p>
            </div>
          </div>
        </div>

        <div className="mt-auto grid grid-cols-2 gap-2">
          <button
            className="flex h-8 items-center justify-center gap-1.5 rounded-md border border-white/12 text-[9px] text-white/62"
            type="button"
          >
            <MessageCircle size={11} />
            先聊两句
          </button>
          <button
            className="flex h-8 items-center justify-center gap-1.5 rounded-md bg-[#1447e6] text-[9px] font-semibold"
            type="button"
          >
            <Check size={11} />
            确认计划
          </button>
        </div>
      </div>
    </PreviewWindow>
  )
}

export function ProductStage() {
  return (
    <>
      <div className="relative mx-auto mt-12 hidden h-[570px] max-w-[1240px] md:block lg:mt-16">
        <div className="absolute left-[1.5%] top-[112px] z-10 w-[43%] -rotate-[6deg] transition-transform duration-500 hover:z-30 hover:rotate-0 hover:scale-[1.02]">
          <DiscoverPreview />
        </div>
        <div className="absolute left-1/2 top-0 z-20 w-[58%] max-w-[720px] -translate-x-1/2 transition-transform duration-500 hover:scale-[1.015]">
          <AssistantPreview />
        </div>
        <div className="absolute right-[1.5%] top-[112px] z-10 w-[43%] rotate-[6deg] transition-transform duration-500 hover:z-30 hover:rotate-0 hover:scale-[1.02]">
          <MatchRoomPreview />
        </div>
      </div>

      <div className="mx-auto mt-10 max-w-[620px] px-4 md:hidden">
        <AssistantPreview />
      </div>
    </>
  )
}
