import { useEffect } from 'react'
import { Mail, X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ContactUsModalProps {
  contactEmail: string
  open: boolean
  onClose: () => void
}

export function ContactUsModal({
  contactEmail,
  open,
  onClose,
}: ContactUsModalProps) {
  useEffect(() => {
    if (!open) {
      return
    }

    const originalOverflow = document.body.style.overflow

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    document.body.style.overflow = 'hidden'
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.overflow = originalOverflow
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [onClose, open])

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center p-4 sm:p-6">
      <button
        aria-label="关闭联系我们弹窗"
        className="absolute inset-0 bg-[rgba(48,34,28,0.4)] backdrop-blur-[5px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[min(98vh,43rem)] w-full max-w-3xl flex-col overflow-hidden rounded-[32px] border border-[#eadfd5] bg-[#fdf9f4] shadow-[0_32px_88px_rgba(59,39,28,0.22)]">
        <button
          aria-label="关闭"
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#eadfd5] bg-white/86 text-[#8b7160] transition-colors hover:bg-white"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="border-b border-[#efe5dc] px-7 pb-6 pt-8 sm:px-10">
          <div className="flex items-center gap-4">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[linear-gradient(135deg,#f6ddd3,#f0c4bb)] text-[#8f5d47] shadow-[0_12px_24px_rgba(143,93,71,0.14)]">
              <Mail size={22} />
            </div>
            <div className="min-w-0">
              <h2 className="font-serif text-3xl text-[#2f221b] sm:text-4xl">
                联系我们
              </h2>
              <p className="mt-1 text-sm text-[#8a7667]">产品建议、反馈与合作沟通</p>
            </div>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto px-7 py-7 sm:px-10">
          <section className="rounded-[24px] border border-[#efe4db] bg-[rgba(255,255,255,0.72)] px-5 py-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)] sm:px-6">
            <h3 className="font-serif text-2xl text-[#2f221b]">联系说明</h3>
            <div className="mt-4 space-y-3 text-[15px] leading-7 text-[#5d5048] sm:text-base">
              <p>如果你有建议、反馈或合作想法，欢迎随时来信联系我们。我们会尽快查看并回复与产品体验、平台使用和合作相关的问题。</p>
            </div>
          </section>

          <section className="mt-6 rounded-[24px] border border-[#efe4db] bg-[rgba(255,255,255,0.72)] px-5 py-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)] sm:px-6">
            <h3 className="font-serif text-2xl text-[#2f221b]">入驻高校联盟</h3>
            <div className="mt-4 space-y-3 text-[15px] leading-7 text-[#5d5048] sm:text-base">
              <p>如果你也期待让本校同学拥有更认真、更真诚的连接方式，Soulink 欢迎更多校园一起加入这条高校联盟的路径。欢迎邮箱联系我们，并附上你所在高校的邮箱后缀信息。</p>
            </div>
          </section>

          <section className="mt-6 rounded-[24px] border border-[#efe4db] bg-[rgba(255,255,255,0.72)] px-5 py-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)] sm:px-6">
            <h3 className="font-serif text-2xl text-[#2f221b]">加入Soulink团队</h3>
            <div className="mt-4 space-y-4 text-[15px] leading-7 text-[#5d5048] sm:text-base">
              <p>我们正在招募 SouLink 高校联盟在各个校园的运营负责人或校园大使，感兴趣的同学欢迎联系我们。</p>
              <div className="space-y-2">
                <p className="font-medium text-[#6b5648]">加入我们，你将获得：</p>
                <ul className="space-y-2 pl-5 text-[#5d5048] marker:text-[#b58e76]">
                  <li>1. 参与一款产品从构想到落地的完整实践经验</li>
                  <li>2. 有机会搭建属于自己的校园社群，拓展优质人脉</li>
                  <li>3. 参与 SouLink 品牌线上线下活动的策划与执行</li>
                  <li>4. 提升团队协作、资源整合与项目推动能力</li>
                </ul>
              </div>
            </div>
          </section>

          <section className="mt-6 rounded-[24px] border border-[#efe4db] bg-[rgba(255,255,255,0.72)] px-5 py-5 shadow-[inset_0_1px_0_rgba(255,255,255,0.64)] sm:px-6">
            <h3 className="font-serif text-2xl text-[#2f221b]">联系邮箱</h3>
            <p className="mt-4 text-[15px] leading-7 text-[#5d5048] sm:text-base">
              <a
                className="text-[#8f6548] underline decoration-[#d4b5a0] underline-offset-4 hover:text-[#744d35]"
                href={`mailto:${contactEmail}`}
              >
                {contactEmail}
              </a>
            </p>
          </section>
        </div>

        <div className="border-t border-[#efe5dc] px-7 py-5 sm:px-10">
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_30px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            onClick={onClose}
            type="button"
          >
            关闭
          </button>
        </div>
      </div>
    </div>
  )
}
