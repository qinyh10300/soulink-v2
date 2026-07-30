import { zodResolver } from '@hookform/resolvers/zod'
import { SendHorizonal } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'

import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { getErrorMessage } from '@/lib/api'
import { resetForgotPassword, sendForgotPasswordCode } from '@/lib/auth-api'
import {
  type ForgotPasswordResetValues,
  forgotPasswordResetSchema,
  loginEmailSchema,
} from '@/lib/auth-schema'
import { cn } from '@/lib/utils'
import { VERIFICATION_CODE_COUNTDOWN_SECONDS } from '@/lib/verification-code'

interface ForgotPasswordPanelProps {
  disabled?: boolean
  email: string
  onEmailInvalid: (message: string) => void
  onEmailValid: () => void
}

export function ForgotPasswordPanel({
  disabled = false,
  email,
  onEmailInvalid,
  onEmailValid,
}: ForgotPasswordPanelProps) {
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
    if (countdown <= 0) {
      return
    }

    const timer = window.setTimeout(() => {
      setCountdown((value) => value - 1)
    }, 1000)

    return () => window.clearTimeout(timer)
  }, [countdown])

  const validateEmail = () => {
    const parsedEmail = loginEmailSchema.safeParse(email)

    if (!parsedEmail.success) {
      onEmailInvalid(parsedEmail.error.issues[0]?.message ?? '请先输入合法的清华邮箱')
      setNotice('')
      return null
    }

    onEmailValid()
    return parsedEmail.data
  }

  const handleSendCode = async () => {
    const parsedEmail = validateEmail()

    if (!parsedEmail) {
      return
    }

    setIsSendingCode(true)

    try {
      const response = await sendForgotPasswordCode(parsedEmail)
      setIsError(false)
      setNotice('验证码已发送，请查收邮箱。')
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

  const handleResetPassword = handleSubmit(async (values) => {
    const parsedEmail = validateEmail()

    if (!parsedEmail) {
      return
    }

    try {
      await resetForgotPassword({
        email: parsedEmail,
        code: values.verificationCode,
        newPassword: values.password,
      })
      setIsError(false)
      setNotice('密码已重置，请使用新密码登录。')
      reset()
    } catch (error) {
      setIsError(true)
      setNotice(getErrorMessage(error, '密码重置失败，请稍后重试'))
    }
  })

  return (
    <div
      className="space-y-4 rounded-[24px] border border-[#e5d8cf] bg-[#fffaf5]/88 px-4 py-4 shadow-[0_10px_24px_rgba(100,73,57,0.06)]"
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault()
        }
      }}
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-sm leading-6 text-[#6f6056]">将向您的邮箱发送验证码</p>
        <button
          className={cn(
            buttonVariants({ size: 'sm' }),
            'gap-2 self-start rounded-full bg-[#8f6548] text-white shadow-[0_12px_24px_rgba(105,73,51,0.14)] hover:bg-[#7f5a40] sm:self-auto',
          )}
          disabled={disabled || isSendingCode || countdown > 0}
          onClick={handleSendCode}
          type="button"
        >
          <SendHorizonal size={15} />
          {countdown > 0 ? `${countdown}s` : isSendingCode ? '发送中...' : '发送验证码'}
        </button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <label className="text-sm font-medium text-[#3a302a]" htmlFor="forgot-password-code">
            验证码
          </label>
          <Input
            id="forgot-password-code"
            inputMode="numeric"
            placeholder="6 位验证码"
            {...register('verificationCode')}
          />
          {errors.verificationCode ? (
            <p className="text-sm text-[#a35b43]">{errors.verificationCode.message}</p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label className="text-sm font-medium text-[#3a302a]" htmlFor="forgot-password-new">
            新密码
          </label>
          <Input
            id="forgot-password-new"
            placeholder="至少 8 位，包含字母和数字"
            type="password"
            {...register('password')}
          />
          {errors.password ? (
            <p className="text-sm text-[#a35b43]">{errors.password.message}</p>
          ) : null}
        </div>
      </div>

      <div className="space-y-2">
        <label className="text-sm font-medium text-[#3a302a]" htmlFor="forgot-password-confirm">
          确认新密码
        </label>
        <Input
          id="forgot-password-confirm"
          placeholder="再次输入新密码"
          type="password"
          {...register('confirmPassword')}
        />
        {errors.confirmPassword ? (
          <p className="text-sm text-[#a35b43]">{errors.confirmPassword.message}</p>
        ) : null}
      </div>

      <button
        className={cn(
          buttonVariants({ size: 'sm' }),
          'rounded-full bg-[#8f6548] text-white shadow-[0_12px_24px_rgba(105,73,51,0.14)] hover:bg-[#7f5a40]',
        )}
        disabled={disabled || isSubmitting || !isValid}
        onClick={() => {
          void handleResetPassword()
        }}
        type="button"
      >
        {isSubmitting ? '重置中...' : '重置密码'}
      </button>

      {notice ? (
        <p className={isError ? 'text-sm text-[#a35b43]' : 'text-sm text-[#6f6056]'}>
          {notice}
        </p>
      ) : null}
    </div>
  )
}
