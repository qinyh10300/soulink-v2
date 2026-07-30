import { useEffect, useState } from 'react'

import { SendHorizonal, TriangleAlert } from 'lucide-react'

import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { getErrorMessage } from '@/lib/api'
import { deleteMyAccount, sendDeleteAccountCode } from '@/lib/auth-api'
import { cn } from '@/lib/utils'
import { VERIFICATION_CODE_COUNTDOWN_SECONDS } from '@/lib/verification-code'

interface DeleteAccountConfirmModalProps {
  expectedEmail: string
  open: boolean
  onCancel: () => void
  onDeleted: () => void
}

export function DeleteAccountConfirmModal({
  expectedEmail,
  open,
  onCancel,
  onDeleted,
}: DeleteAccountConfirmModalProps) {
  const [confirmationEmail, setConfirmationEmail] = useState('')
  const [verificationCode, setVerificationCode] = useState('')
  const [notice, setNotice] = useState('')
  const [isError, setIsError] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [isSendingCode, setIsSendingCode] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)

  useEffect(() => {
    if (!open) {
      setConfirmationEmail('')
      setVerificationCode('')
      setNotice('')
      setIsError(false)
      setCountdown(0)
      setIsSendingCode(false)
      setIsSubmitting(false)
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onCancel()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onCancel, open])

  useEffect(() => {
    if (countdown <= 0) {
      return
    }

    const timer = window.setTimeout(() => {
      setCountdown((value) => value - 1)
    }, 1000)

    return () => window.clearTimeout(timer)
  }, [countdown])

  if (!open) {
    return null
  }

  const isEmailMatch = confirmationEmail.trim() === expectedEmail

  const handleSendCode = async () => {
    if (!isEmailMatch) {
      setIsError(true)
      setNotice('请先输入当前账号邮箱，再发送验证码。')
      return
    }

    setIsSendingCode(true)

    try {
      const response = await sendDeleteAccountCode()
      setIsError(false)
      setNotice('验证码已发送到登录邮箱。确认注销前请再次检查邮箱输入。')
      setCountdown(
        response.cooldownSeconds && response.cooldownSeconds > 0
          ? response.cooldownSeconds
          : VERIFICATION_CODE_COUNTDOWN_SECONDS,
      )
    } catch (error) {
      setIsError(true)
      setNotice(getErrorMessage(error, '验证码发送失败，请稍后重试'))
    } finally {
      setIsSendingCode(false)
    }
  }

  const handleDelete = async () => {
    if (!isEmailMatch) {
      setIsError(true)
      setNotice('输入邮箱必须与当前账号邮箱完全一致。')
      return
    }

    if (!verificationCode.trim()) {
      setIsError(true)
      setNotice('请输入邮箱验证码。')
      return
    }

    setIsSubmitting(true)

    try {
      await deleteMyAccount(verificationCode.trim())
      setIsError(false)
      setNotice('账号已注销。')
      onDeleted()
    } catch (error) {
      setIsError(true)
      setNotice(getErrorMessage(error, '账号注销失败，请稍后重试'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        className="absolute inset-0 bg-[rgba(44,31,24,0.2)] backdrop-blur-[6px]"
        onClick={onCancel}
        type="button"
      />

      <div className="relative z-10 w-full max-w-[32rem] rounded-[34px] bg-[#f7f3ee] px-8 py-10 shadow-[0_28px_80px_rgba(61,40,29,0.16)] sm:px-11 sm:py-12">
        <div className="flex flex-col items-center text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-[#fff3e0] text-[#f3a241] shadow-[inset_0_0_0_1px_rgba(243,162,65,0.18)]">
            <TriangleAlert fill="currentColor" size={32} strokeWidth={1.8} />
          </div>

          <h3 className="mt-7 font-serif text-4xl text-[#3c332f] sm:text-[2.6rem]">
            确认注销账号
          </h3>

          <p className="mt-5 max-w-[26rem] text-lg leading-9 text-[#7d7570]">
            注销后账号将无法登录，历史配对和问卷数据会按平台规则匿名化保留。
            <span className="font-semibold text-[#d95d58]"> 此操作不可撤销。</span>
          </p>
        </div>

        <div className="mt-10 space-y-5">
          <div>
            <label
              className="mb-4 block text-left text-[1.05rem] text-[#b5a79d]"
              htmlFor="delete-account-email"
            >
              请输入当前账号邮箱确认：
            </label>
            <Input
              autoComplete="email"
              id="delete-account-email"
              onChange={(event) => setConfirmationEmail(event.target.value)}
              placeholder={expectedEmail}
              type="email"
              value={confirmationEmail}
            />
            {!isEmailMatch && confirmationEmail ? (
              <p className="mt-3 text-sm text-[#c85a56]">
                输入邮箱必须与当前账号邮箱完全一致。
              </p>
            ) : null}
          </div>

          <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
            <div className="min-w-0 flex-1">
              <label
                className="mb-4 block text-left text-[1.05rem] text-[#b5a79d]"
                htmlFor="delete-account-code"
              >
                邮箱验证码
              </label>
              <Input
                id="delete-account-code"
                inputMode="numeric"
                onChange={(event) => setVerificationCode(event.target.value)}
                placeholder="6 位验证码"
                value={verificationCode}
              />
            </div>
            <button
              className={cn(
                buttonVariants({ size: 'lg' }),
                'gap-2 rounded-full bg-[#8f6548] text-white hover:bg-[#7f5a40]',
              )}
              disabled={isSendingCode || countdown > 0}
              onClick={handleSendCode}
              type="button"
            >
              <SendHorizonal size={15} />
              {countdown > 0 ? `${countdown}s` : isSendingCode ? '发送中...' : '发送验证码'}
            </button>
          </div>
        </div>

        {notice ? (
          <p className={cn('mt-6 text-sm', isError ? 'text-[#c85a56]' : 'text-[#7d7570]')}>
            {notice}
          </p>
        ) : null}

        <div className="mt-10 grid gap-4 sm:grid-cols-2">
          <button
            className={cn(
              buttonVariants({ variant: 'secondary', size: 'lg' }),
              'rounded-full border-0 bg-[#eee7df] text-[1.15rem] font-semibold text-[#5f5550] shadow-none hover:bg-[#e8dfd5]',
            )}
            onClick={onCancel}
            type="button"
          >
            取消
          </button>
          <button
            className="inline-flex items-center justify-center rounded-full bg-[#c44f4c] px-6 py-3 text-[1.15rem] font-semibold text-white transition disabled:cursor-not-allowed disabled:bg-[#ddb5b3] disabled:text-white/90 hover:bg-[#b64643]"
            disabled={!isEmailMatch || isSubmitting}
            onClick={() => {
              void handleDelete()
            }}
            type="button"
          >
            {isSubmitting ? '注销中...' : '确认注销'}
          </button>
        </div>
      </div>
    </div>
  )
}
