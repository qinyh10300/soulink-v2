import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface AccountManagementCardProps {
  onDeleteAccount: () => void
  onOpenChangePassword: () => void
}

export function AccountManagementCard({
  onDeleteAccount,
  onOpenChangePassword,
}: AccountManagementCardProps) {
  return (
    <div className="rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.95)] p-8 shadow-[0_18px_56px_rgba(72,50,33,0.06)] sm:p-10">
      <div className="space-y-6">
        <div className="flex flex-col gap-5 rounded-[26px] border border-[#efe4db] bg-[#faf5ef] px-6 py-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-2">
            <h3 className="text-xl font-medium text-[#2f221b]">修改密码</h3>
            <p className="text-sm leading-6 text-[#817166]">
              定期修改密码以保护账号安全
            </p>
          </div>
          <button
            className={cn(
              buttonVariants({ variant: 'secondary', size: 'lg' }),
              'rounded-full border-[#e5d9cf] bg-white text-[#6f5f54]',
            )}
            onClick={onOpenChangePassword}
            type="button"
          >
            修改密码
          </button>
        </div>

        <div className="flex flex-col gap-5 rounded-[26px] border border-[#f0d9d5] bg-[#fdf4f3] px-6 py-6 sm:flex-row sm:items-center sm:justify-between">
          <div className="space-y-2">
            <h3 className="text-xl font-medium text-[#7d433d]">注销账号</h3>
            <p className="text-sm leading-6 text-[#8b6b67]">
              注销后您的数据将被匿名化保留（详见隐私协议）
            </p>
          </div>
          <button
            className="inline-flex h-12 items-center justify-center rounded-full border border-[#d79d97] bg-white px-6 text-[15px] font-medium text-[#a3564d] transition-colors hover:bg-[#fff5f3]"
            onClick={onDeleteAccount}
            type="button"
          >
            注销账号
          </button>
        </div>
      </div>
    </div>
  )
}
