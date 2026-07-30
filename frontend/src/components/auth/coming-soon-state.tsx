import { ArrowLeft, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ComingSoonStateProps {
  onBack: () => void
}

export function ComingSoonState({ onBack }: ComingSoonStateProps) {
  return (
    <div className="space-y-8 text-center">
      <div className="space-y-3">
        <p className="text-sm font-semibold tracking-[0.18em] text-[#9c6b48] uppercase">
          Step 1 Completed
        </p>
        <h1 className="font-serif text-4xl text-[#261f1b]">该功能正在开发中</h1>
        <p className="text-base leading-7 text-[#7d6b60]">
          注册完整流程稍后开放，敬请期待。当前已保留邮箱验证入口与整体流程结构。
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <Link
          className={cn(
            buttonVariants({ variant: 'secondary', size: 'lg' }),
            'gap-2 rounded-full border-[#ded2c7] bg-white text-[#2b241f] shadow-none hover:bg-white',
          )}
          to="/login"
        >
          <ArrowLeft size={16} />
          返回登录
        </Link>
        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'gap-2 rounded-full bg-[#8f6548] shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
          )}
          onClick={onBack}
          type="button"
        >
          返回注册
          <ArrowRight size={16} />
        </button>
      </div>
    </div>
  )
}
