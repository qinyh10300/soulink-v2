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
        <p className="text-xs font-semibold text-[#1447e6] uppercase">
          Step 1 Completed
        </p>
        <h1 className="text-4xl font-semibold text-[#09090b]">该功能正在开发中</h1>
        <p className="text-base leading-7 text-[#71717a]">
          注册完整流程稍后开放，敬请期待。当前已保留邮箱验证入口与整体流程结构。
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-2">
        <Link
          className={cn(
            buttonVariants({ variant: 'secondary', size: 'lg' }),
            'gap-2 rounded-md border-[#d4d4d8] bg-white text-[#27272a] shadow-none hover:bg-[#f4f4f5]',
          )}
          to="/login"
        >
          <ArrowLeft size={16} />
          返回登录
        </Link>
        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'gap-2 rounded-md bg-[#1447e6] shadow-[0_12px_28px_rgba(20,71,230,0.18)] hover:bg-[#103bc4]',
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
