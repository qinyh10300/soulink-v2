import {
  ArrowLeft,
  Ban,
  EyeOff,
  Heart,
  LockKeyhole,
  ShieldCheck,
  Trash2,
} from 'lucide-react'
import { Link } from 'react-router-dom'

const principles = [
  {
    icon: LockKeyhole,
    title: '双方同意后再揭示',
    copy: '联系方式默认隐藏。进入匹配房间后，每个人仍然单独决定是否向对方显示。',
  },
  {
    icon: EyeOff,
    title: '认真关系单独保护',
    copy: '恋爱意图不进入公开流，公开预览不展示可识别身份，且必须由用户主动开启。',
  },
  {
    icon: Ban,
    title: '随时拒绝与退出',
    copy: '可以跳过推荐、撤回兴趣、暂停匹配、隐藏联系方式，并对用户或内容进行举报。',
  },
  {
    icon: Trash2,
    title: '资料可以编辑和删除',
    copy: 'AI 画像始终可编辑、拒绝或留空；账户和意图可以主动结束或删除。',
  },
]

export function SafetyPage() {
  return (
    <div className="min-h-screen bg-[#f7f6f3] text-[#292724]">
      <header className="border-b border-[#dfdcd6] bg-[#fbfaf8]">
        <div className="mx-auto flex h-16 max-w-[1040px] items-center justify-between px-4 sm:px-6">
          <Link
            className="inline-flex items-center gap-2 text-sm font-semibold text-[#4f4b46]"
            to="/"
          >
            <ArrowLeft size={16} />
            SouLink
          </Link>
          <Link
            className="inline-flex h-9 items-center rounded-md bg-[#5b3477] px-4 text-xs font-semibold text-white"
            to="/register"
          >
            清华邮箱注册
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-[1040px] px-4 py-12 sm:px-6 sm:py-16">
        <div className="max-w-2xl">
          <p className="flex items-center gap-2 text-xs font-semibold text-[#6c4a80]">
            <ShieldCheck size={14} />
            PRIVACY & SAFETY
          </p>
          <h1 className="mt-3 text-4xl font-semibold">连接值得发生，也值得被保护。</h1>
          <p className="mt-4 text-base leading-8 text-[#706c65]">
            SouLink 把身份验证、意图可见范围、双方同意和联系方式揭示当作产品状态，而不是一段藏在页尾的承诺。
          </p>
        </div>

        <div className="mt-10 grid gap-3 sm:grid-cols-2">
          {principles.map((principle) => {
            const Icon = principle.icon
            return (
              <section
                className="rounded-lg border border-[#dfdcd6] bg-white p-5"
                key={principle.title}
              >
                <Icon size={19} className="text-[#5b3477]" />
                <h2 className="mt-4 text-base font-semibold">{principle.title}</h2>
                <p className="mt-2 text-sm leading-7 text-[#706c65]">{principle.copy}</p>
              </section>
            )
          })}
        </div>

        <section className="mt-12 rounded-lg border border-[#dfc7d0] bg-[#fff5f8] p-6">
          <div className="flex items-center gap-2 text-sm font-semibold text-[#7e3d55]">
            <Heart size={17} />
            认真关系匹配的额外边界
          </div>
          <ul className="mt-4 space-y-2 text-sm leading-6 text-[#745663]">
            <li>不在公开意图流展示可识别的恋爱需求。</li>
            <li>不使用恋爱偏好影响项目、学习或普通搭子推荐。</li>
            <li>不生成“命中注定”“完美匹配”等确定性判断。</li>
            <li>暂停或关闭后，不再进入新的认真关系推荐。</li>
          </ul>
        </section>

        <p className="mt-10 text-xs leading-6 text-[#817d76]">
          更完整的数据处理说明见仓库中的隐私文档。遇到骚扰、冒充、隐私泄露或安全风险时，请通过产品内反馈入口提交；紧急情况请优先联系学校或当地公共安全机构。
        </p>
      </main>
    </div>
  )
}
