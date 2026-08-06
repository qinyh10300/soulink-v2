import { zodResolver } from '@hookform/resolvers/zod'
import { AnimatePresence, motion } from 'framer-motion'
import { KeyRound, ShieldCheck, UserRound } from 'lucide-react'
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

const demoMode = import.meta.env.VITE_DEMO_MODE === 'true'

interface DemoAccount {
  id: 'member' | 'admin'
  label: string
  email: string
  password: string
  redirectTo: string
}

const configuredDemoAccounts = [
  {
    id: 'member',
    label: '普通演示账号',
    email: import.meta.env.VITE_DEMO_EMAIL ?? '',
    password: import.meta.env.VITE_DEMO_PASSWORD ?? '',
    redirectTo: '/home',
  },
  {
    id: 'admin',
    label: 'Hack 管理员',
    email: import.meta.env.VITE_DEMO_ADMIN_EMAIL ?? '',
    password: import.meta.env.VITE_DEMO_ADMIN_PASSWORD ?? '',
    redirectTo: '/admin/safety',
  },
] satisfies DemoAccount[]

const demoAccounts = configuredDemoAccounts.filter(
  (account) => account.email && account.password,
)

export function LoginForm() {
  const navigate = useNavigate()
  const [toastNotice, setToastNotice] = useState('')
  const [isForgotPasswordOpen, setIsForgotPasswordOpen] = useState(false)
  const [isAgreementModalOpen, setIsAgreementModalOpen] = useState(false)
  const [quickLoginAccount, setQuickLoginAccount] = useState<DemoAccount['id'] | null>(
    null,
  )

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

  const authenticate = async (email: string, password: string, redirectTo: string) => {
    const result = await login(email, password)

    saveAuthenticatedUser(result.token, result.user)
    resetDashboardAnnouncementSeen()
    navigate(redirectTo)
  }

  const onSubmit = async (values: LoginFormValues) => {
    setToastNotice('')

    try {
      await authenticate(values.email, values.password, '/home')
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

  const fillDemoAccount = (account: DemoAccount) => {
    setValue('email', account.email, {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
    setValue('password', account.password, {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
    setValue('agreement', true, {
      shouldDirty: true,
      shouldTouch: true,
      shouldValidate: true,
    })
    clearErrors()
  }

  const handleQuickLogin = async (account: DemoAccount) => {
    setToastNotice('')
    setQuickLoginAccount(account.id)
    fillDemoAccount(account)

    try {
      await authenticate(account.email, account.password, account.redirectTo)
    } catch (error) {
      setToastNotice(getErrorMessage(error, '快捷登录失败，请稍后重试'))
      setQuickLoginAccount(null)
    }
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
          <p className="text-xs font-semibold text-[#1447e6]">
            SOULINK · 清华校内
          </p>
          <div className="space-y-3">
            <h1 className="text-3xl font-semibold text-[#09090b]">欢迎回来</h1>
            <p className="text-sm leading-7 text-[#71717a]">
              登录后继续你的意图和匹配房间。
            </p>
          </div>
        </div>

        {demoMode && demoAccounts.length > 0 ? (
          <section className="rounded-lg border border-[#cbd8ff] bg-[#f4f7ff] p-3">
            <div className="mb-2 flex items-center gap-2 text-[#243a75]">
              <KeyRound size={15} />
              <span className="text-xs font-semibold">本地快捷登录</span>
            </div>
            <div className="divide-y divide-[#dbe3f8]">
              {demoAccounts.map((account) => {
                const AccountIcon = account.id === 'admin' ? ShieldCheck : UserRound
                const isLoggingIn = quickLoginAccount === account.id

                return (
                  <div className="flex min-w-0 items-center gap-3 py-2.5" key={account.id}>
                    <span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-white text-[#1447e6]">
                      <AccountIcon size={16} />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-xs font-semibold text-[#243a75]">
                        {account.label}
                      </span>
                      <span className="mt-0.5 block break-all font-mono text-[10px] leading-4 text-[#66729a]">
                        {account.email} · {account.password}
                      </span>
                    </span>
                    <button
                      className="h-8 w-[76px] shrink-0 rounded-md border border-[#b8c8f5] bg-white text-[11px] font-semibold text-[#1447e6] hover:bg-[#eff4ff] disabled:cursor-wait disabled:opacity-60"
                      disabled={quickLoginAccount !== null || isSubmitting}
                      onClick={() => void handleQuickLogin(account)}
                      type="button"
                    >
                      {isLoggingIn ? '登录中...' : '一键登录'}
                    </button>
                  </div>
                )
              })}
            </div>
          </section>
        ) : null}

        <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-2">
            <label
              className="text-sm font-medium text-[#3f3f46]"
              htmlFor="login-email"
            >
              清华邮箱
            </label>
            <Input
              id="login-email"
              placeholder="请输入清华邮箱"
              type="email"
              {...register('email')}
            />
            {errors.email ? (
              <p className="text-sm text-[#b42318]">{errors.email.message}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between gap-4">
              <label
                className="text-sm font-medium text-[#3f3f46]"
                htmlFor="login-password"
              >
                密码
              </label>
              <button
                className="text-sm text-[#1447e6] transition-colors hover:text-[#103bc4]"
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
              <p className="text-sm text-[#b42318]">{errors.password.message}</p>
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
                    disabled={isSubmitting || quickLoginAccount !== null}
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
                <label className="flex items-start gap-3 rounded-md border border-[#e4e4e7] bg-[#fafafa] px-4 py-3">
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                  <span className="text-sm leading-6 text-[#52525b]">
                    我已阅读并同意
                    <button
                      className="mx-1 text-[#1447e6] underline decoration-[#a9baf0] underline-offset-4 transition-colors hover:text-[#103bc4]"
                      onClick={(event) => {
                        event.preventDefault()
                        setIsAgreementModalOpen(true)
                      }}
                      type="button"
                    >
                      《SouLink 用户协议》
                    </button>
                  </span>
                </label>
              )}
            />
            {errors.agreement ? (
              <p className="text-sm text-[#b42318]">{errors.agreement.message}</p>
            ) : null}
          </div>

          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'mt-2 w-full rounded-md bg-[#1447e6] text-white hover:bg-[#103bc4]',
            )}
            disabled={!isValid || isSubmitting || quickLoginAccount !== null}
            type="submit"
          >
            {isSubmitting ? '登录中...' : '登录'}
          </button>
        </form>

        <p className="text-sm text-[#71717a]">
          还没有账号？
          <Link className="ml-2 text-[#1447e6] hover:text-[#103bc4]" to="/register">
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
