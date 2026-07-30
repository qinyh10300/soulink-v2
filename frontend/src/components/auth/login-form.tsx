import { zodResolver } from '@hookform/resolvers/zod'
import { AnimatePresence, motion } from 'framer-motion'
import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'

import { ForgotPasswordPanel } from '@/components/auth/forgot-password-panel'
import { FloatingToast } from '@/components/shared/floating-toast'
import { UserAgreementModal } from '@/components/auth/user-agreement-modal'
import { buttonVariants } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { getErrorMessage } from '@/lib/api'
import { login } from '@/lib/auth-api'
import { type LoginFormValues, loginSchema } from '@/lib/auth-schema'
import { resetDashboardAnnouncementSeen } from '@/lib/survey-storage'
import { saveAuthenticatedUser } from '@/lib/user-profile-storage'
import { cn } from '@/lib/utils'

const panelTransition = {
  duration: 0.22,
  ease: [0.22, 1, 0.36, 1] as const,
}

export function LoginForm() {
  const navigate = useNavigate()
  const [toastNotice, setToastNotice] = useState('')
  const [isForgotPasswordOpen, setIsForgotPasswordOpen] = useState(false)
  const [isAgreementModalOpen, setIsAgreementModalOpen] = useState(false)

  const {
    control,
    handleSubmit,
    register,
    watch,
    setValue,
    setError,
    clearErrors,
    formState: { errors, isSubmitting, isValid },
  } = useForm<LoginFormValues>({
    defaultValues: {
      email: '',
      password: '',
      agreement: false,
    },
    mode: 'onChange',
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (values: LoginFormValues) => {
    setToastNotice('')

    try {
      const result = await login(values.email, values.password)

      saveAuthenticatedUser(result.token, result.user)
      resetDashboardAnnouncementSeen()
      navigate('/dashboard')
    } catch (error) {
      setToastNotice(getErrorMessage(error, '登录失败，请稍后重试'))
    }
  }

  const handleForgotPasswordEmailInvalid = (message: string) => {
    setError('email', {
      type: 'manual',
      message,
    })
    setToastNotice('')
  }

  const handleForgotPasswordEmailValid = () => {
    clearErrors('email')
  }

  const handleAgreementConfirm = () => {
    setValue('agreement', true, {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
    clearErrors('agreement')
    setIsAgreementModalOpen(false)
  }

  const currentEmail = watch('email')

  useEffect(() => {
    if (!toastNotice) {
      return
    }

    const timer = window.setTimeout(() => {
      setToastNotice('')
    }, 3000)

    return () => window.clearTimeout(timer)
  }, [toastNotice])

  return (
    <>
      <FloatingToast message={toastNotice} />

      <div className="space-y-8">
        <div className="space-y-4">
          <p className="text-sm font-semibold tracking-[0.22em] text-[#9a6d4d] uppercase">
             Soulink
          </p>
          <div className="space-y-3">
            <h1 className="font-serif text-5xl text-[#261f1b]">欢迎回来</h1>
            <p className="text-base leading-7 text-[#7e6d61]">
              登录您的  Soulink 账号
            </p>
          </div>
        </div>

        <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-2">
            <label
              className="text-sm font-medium text-[#3a302a]"
              htmlFor="login-email"
            >
              {/* 学校邮箱（目前支持清华(协和)、北大、人大、中传、北外、北师大） */}
              学校邮箱
            </label>
            <Input
              id="login-email"
              placeholder="请输入学校邮箱"
              type="email"
              {...register('email')}
            />
            {errors.email ? (
              <p className="text-sm text-[#a35b43]">{errors.email.message}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between gap-4">
              <label
                className="text-sm font-medium text-[#3a302a]"
                htmlFor="login-password"
              >
                密码
              </label>
              <button
                className="text-sm text-[#9a6d4d] transition-colors hover:text-[#7d5539]"
                onClick={() =>
                  setIsForgotPasswordOpen((currentValue) => !currentValue)
                }
                type="button"
              >
                忘记密码？
              </button>
            </div>
            <Input id="login-password" type="password" {...register('password')} />
            {errors.password ? (
              <p className="text-sm text-[#a35b43]">{errors.password.message}</p>
            ) : null}
          </div>

          <AnimatePresence initial={false}>
            {isForgotPasswordOpen ? (
              <motion.div
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                initial={{ height: 0, opacity: 0 }}
                transition={panelTransition}
              >
                <div className="overflow-hidden">
                  <ForgotPasswordPanel
                    disabled={isSubmitting}
                    email={currentEmail}
                    onEmailInvalid={handleForgotPasswordEmailInvalid}
                    onEmailValid={handleForgotPasswordEmailValid}
                  />
                </div>
              </motion.div>
            ) : null}
          </AnimatePresence>

          <div className="space-y-2">
            <Controller
              control={control}
              name="agreement"
              render={({ field }) => (
                <label className="flex items-start gap-3 rounded-2xl border border-[#eadfd7] bg-white/58 px-4 py-3">
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                  <span className="text-sm leading-6 text-[#68584d]">
                    我已阅读并同意
                    <button
                      className="mx-1 text-[#9a6d4d] underline decoration-[#cfb29e] underline-offset-4 transition-colors hover:text-[#7d5539]"
                      onClick={(event) => {
                        event.preventDefault()
                        setIsAgreementModalOpen(true)
                      }}
                      type="button"
                    >
                      《 Soulink 用户协议》
                    </button>
                  </span>
                </label>
              )}
            />
            {errors.agreement ? (
              <p className="text-sm text-[#a35b43]">{errors.agreement.message}</p>
            ) : null}
          </div>

          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'mt-2 w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            disabled={!isValid || isSubmitting}
            type="submit"
          >
            {isSubmitting ? '登录中...' : '登录'}
          </button>
        </form>

        <p className="text-sm text-[#7e6d61]">
          还没有账号？
          <Link className="ml-2 text-[#9a6d4d] hover:text-[#7d5539]" to="/register">
            立即注册
          </Link>
        </p>
      </div>

      <UserAgreementModal
        onClose={() => setIsAgreementModalOpen(false)}
        onConfirm={handleAgreementConfirm}
        open={isAgreementModalOpen}
      />
    </>
  )
}
