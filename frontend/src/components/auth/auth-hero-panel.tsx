import { ArrowUp, Check, ShieldCheck, Sparkles } from 'lucide-react'
import { Link } from 'react-router-dom'

interface AuthHeroPanelProps {
  title: string
  subtitle: string
}

export function AuthHeroPanel({ title, subtitle }: AuthHeroPanelProps) {
  return (
    <aside className="relative hidden min-h-screen overflow-hidden bg-[#09090b] px-10 py-9 text-white lg:flex lg:flex-col xl:px-14 xl:py-11">
      <Link className="relative z-10 flex w-fit items-center gap-2.5" to="/">
        <span className="flex h-8 w-8 items-center justify-center rounded-md bg-[#1447e6] shadow-[0_8px_24px_rgba(20,71,230,0.32)]">
          <Sparkles size={17} />
        </span>
        <span className="text-[17px] font-semibold">SouLink</span>
      </Link>

      <div className="relative z-10 my-auto max-w-[680px] py-12">
        <p className="inline-flex items-center gap-2 rounded-full border border-white/12 px-3 py-1.5 text-[10px] font-medium text-white/56">
          <span className="h-1.5 w-1.5 rounded-full bg-[#6f8eff]" />
          清华校内 AI 连接平台
        </p>
        <h2 className="mt-6 max-w-[620px] text-balance text-4xl font-semibold leading-[1.1] xl:text-5xl">
          {title}
        </h2>
        <p className="mt-5 max-w-[540px] text-sm leading-7 text-white/48 xl:text-base xl:leading-8">
          {subtitle}
        </p>

        <div className="mt-10 max-w-[620px] overflow-hidden rounded-lg border border-white/12 bg-[#111113] shadow-[0_36px_90px_rgba(0,0,0,0.32)]">
          <div className="flex h-10 items-center justify-between border-b border-white/10 px-4">
            <span className="flex items-center gap-2 text-[10px] font-medium text-white/58">
              <Sparkles className="text-[#7c98ff]" size={12} />
              SouLink AI
            </span>
            <span className="flex items-center gap-1.5 text-[9px] text-white/30">
              <span className="h-1.5 w-1.5 rounded-full bg-[#28b77b]" />
              正在理解
            </span>
          </div>
          <div className="space-y-5 p-5 xl:p-6">
            <div className="ml-auto max-w-[82%] rounded-lg bg-[#1447e6] px-4 py-3 text-xs leading-5 text-white">
              想认识一位也在做 AI 产品的同学，今晚一起吃饭聊聊。
            </div>
            <div className="flex max-w-[88%] gap-3">
              <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-white/8 text-[#7c98ff]">
                <Sparkles size={13} />
              </span>
              <p className="text-xs leading-6 text-white/62">
                好的。你更期待轻松交流，还是希望对方也对未来合作保持开放？
              </p>
            </div>
            <div className="flex items-center justify-between rounded-lg border border-white/12 bg-[#18181b] px-4 py-3 text-[10px] text-white/30">
              继续补充你的期待…
              <span className="flex h-6 w-6 items-center justify-center rounded-md bg-[#1447e6] text-white">
                <ArrowUp size={12} />
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="relative z-10 flex flex-wrap items-center gap-x-6 gap-y-2 border-t border-white/10 pt-5 text-[10px] text-white/34">
        <span className="flex items-center gap-1.5">
          <ShieldCheck size={12} />
          清华身份验证
        </span>
        <span className="flex items-center gap-1.5">
          <Check size={12} />
          双向同意后连接
        </span>
      </div>
    </aside>
  )
}
