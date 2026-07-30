import { zodResolver } from '@hookform/resolvers/zod'
import { SendHorizonal, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'

import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { getErrorMessage } from '@/lib/api'
import { changePassword, sendChangePasswordCode } from '@/lib/auth-api'
import {
  type ForgotPasswordResetValues,
  forgotPasswordResetSchema,
} from '@/lib/auth-schema'
import { cn } from '@/lib/utils'
import { VERIFICATION_CODE_COUNTDOWN_SECONDS } from '@/lib/verification-code'

interface ChangePasswordModalProps {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export function ChangePasswordModal({
  open,
  onClose,
  onSuccess,
}: ChangePasswordModalProps) {
  const [notice, setNotice] = useState('')
  const [isError, setIsError] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [isSendingCode, setIsSendingCode] = useState(false)
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting, isValid },
  } = useForm<ForgotPasswordResetValues>({
    defaultValues: {
      verificationCode: '',
      password: '',
      confirmPassword: '',
    },
    mode: 'onChange',
    resolver: zodResolver(forgotPasswordResetSchema),
  })

  useEffect(() => {
    if (!open) {
      reset()
      setNotice('')
      setIsError(false)
      setCountdown(0)
      setIsSendingCode(false)
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)

    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open, reset])

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

  const handleSendCode = async () => {
    setIsSendingCode(true)

    try {
      const response = await sendChangePasswordCode()
      setIsError(false)
      setNotice('验证码已发送到您的登录邮箱。修改成功后需要重新登录。')
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

  const handleChangePassword = handleSubmit(async (values) => {
    try {
      await changePassword({
        code: values.verificationCode,
        newPassword: values.password,
      })
      setIsError(false)
      setNotice('密码已修改，请重新登录。')
      reset()
      onSuccess()
    } catch (error) {
      setIsError(true)
      setNotice(getErrorMessage(error, '密码修改失败，请稍后重试'))
    }
  })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 w-full max-w-lg rounded-[30px] border border-[#eadfd5] bg-[#fcf8f3] p-8 shadow-[0_26px_72px_rgba(61,40,29,0.18)]">
        <button
          className="absolute right-5 top-5 inline-flex h-10 w-10 items-center justify-center rounded-full border border-[#e8ddd3] bg-white text-[#7d6556]"
          onClick={onClose}
          type="button"
        >
          <X size={18} />
        </button>

        <div className="space-y-2">
          <h3 className="font-serif text-3xl text-[#2f221b]">修改密码</h3>
          <p className="text-base leading-7 text-[#6f6056]">
            为保证账号安全，系统会先向您的登录邮箱发送验证码。
          </p>
        </div>

        <div className="mt-6 space-y-4">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-sm leading-6 text-[#6f6056]">邮箱验证码</p>
            <button
              className={cn(
                buttonVariants({ size: 'sm' }),
                'gap-2 self-start rounded-full bg-[#8f6548] text-white shadow-[0_12px_24px_rgba(105,73,51,0.14)] hover:bg-[#7f5a40] sm:self-auto',
              )}
              disabled={isSendingCode || countdown > 0}
              onClick={handleSendCode}
              type="button"
            >
              <SendHorizonal size={15} />
              {countdown > 0 ? `${countdown}s` : isSendingCode ? '发送中...' : '发送验证码'}
            </button>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]" htmlFor="change-password-code">
              验证码
            </label>
            <Input
              id="change-password-code"
              inputMode="numeric"
              placeholder="6 位验证码"
              {...register('verificationCode')}
            />
            {errors.verificationCode ? (
              <p className="text-sm text-[#a35b43]">{errors.verificationCode.message}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]" htmlFor="change-password-new">
              新密码
            </label>
            <Input
              id="change-password-new"
              placeholder="至少 8 位，包含字母和数字"
              type="password"
              {...register('password')}
            />
            {errors.password ? (
              <p className="text-sm text-[#a35b43]">{errors.password.message}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]" htmlFor="change-password-confirm">
              确认新密码
            </label>
            <Input
              id="change-password-confirm"
              placeholder="再次输入新密码"
              type="password"
              {...register('confirmPassword')}
            />
            {errors.confirmPassword ? (
              <p className="text-sm text-[#a35b43]">{errors.confirmPassword.message}</p>
            ) : null}
          </div>
        </div>

        {notice ? (
          <p className={cn('mt-6 text-sm', isError ? 'text-[#a35b43]' : 'text-[#6f6056]')}>
            {notice}
          </p>
        ) : null}

        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'mt-8 w-full rounded-full bg-[#8f6548] text-white hover:bg-[#7f5a40]',
          )}
          disabled={isSubmitting || !isValid}
          onClick={() => {
            void handleChangePassword()
          }}
          type="button"
        >
          {isSubmitting ? '提交中...' : '确认修改'}
        </button>
      </div>
    </div>
  )
}
