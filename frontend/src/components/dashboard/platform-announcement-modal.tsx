import { useEffect } from 'react'
import { Megaphone, X } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface PlatformAnnouncementModalProps {
  open: boolean
  onClose: () => void
}

const announcementParagraphs = [
  '欢迎来到  Soulink。',
  ' Soulink 现已上线基础核心功能，平台当前已接入真实业务接口，用户可正常使用注册、登录、个人资料编辑、问卷保存、配对结果查看及反馈提交等服务。相关数据将按照平台规则进行处理，以保障基本功能的稳定运行与使用体验。',
  '与此同时，部分功能仍在持续建设与优化中，包括跨校联动、灵魂画像、Agent 红娘等模块。上述功能目前将继续以占位入口、静态页面或暂未开放状态展示，后续将在完成功能测试与体验优化后逐步开放。已展示但尚未开放的功能，不影响平台现阶段核心服务的正常使用。',
  '涉及账号安全的重要操作，如修改密码与注销账号，平台将通过您的登录邮箱发送验证码进行身份校验。为确保操作顺利完成，请使用可正常接收邮件的邮箱账号。若您当前处于演示环境或测试环境中，建议优先使用可稳定收取邮件的测试账号进行体验。',
  '如您在使用过程中遇到异常情况、页面问题或功能建议，欢迎通过平台反馈入口提交，我们会持续跟进并优化产品体验。',
  '感谢您的理解与支持，祝您在  Soulink 获得真诚、愉快的连接体验。',
]

export function PlatformAnnouncementModal({
  open,
  onClose,
}: PlatformAnnouncementModalProps) {
  useEffect(() => {
    if (!open) {
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open])

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
      <button
        aria-label="关闭平台公告"
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[7px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 flex max-h-[min(88vh,50rem)] w-full max-w-4xl flex-col overflow-hidden rounded-[34px] border border-[#eadfd5] bg-[#fcf8f3] shadow-[0_32px_90px_rgba(62,40,30,0.22)]">
        <button
          aria-label="关闭"
          className="absolute right-5 top-5 inline-flex h-11 w-11 items-center justify-center rounded-full border border-[#ece0d7] bg-white/80 text-[#8b7160] transition-colors hover:bg-white"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="px-8 pb-4 pt-10 text-center sm:px-12">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-[linear-gradient(135deg,#f7d0d8,#f2b3c0)] text-[#8f4e5b] shadow-[0_16px_32px_rgba(156,91,106,0.18)]">
            <Megaphone size={28} />
          </div>
          <h2 className="mt-6 font-serif text-4xl text-[#2c211c]">平台公告</h2>
        </div>

        <div className="flex-1 overflow-y-auto px-8 pb-8 sm:px-12">
          <div className="rounded-[26px] border border-[#f1e6dc] bg-[#f7f1ea] px-6 py-6 text-left shadow-[inset_0_1px_0_rgba(255,255,255,0.6)] sm:px-8">
            <div className="space-y-5 text-[17px] leading-8 text-[#53453d]">
              {announcementParagraphs.map((paragraph) => (
                <p key={paragraph}>{paragraph}</p>
              ))}
            </div>
          </div>
        </div>

        <div className="border-t border-[#eee2d8] px-8 py-6 sm:px-12">
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            onClick={onClose}
            type="button"
          >
            我知道了
          </button>
        </div>
      </div>
    </div>
  )
}
