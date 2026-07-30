import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'

import { RegisterStepHeader } from '@/components/auth/register-step-header'
import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  type RegisterPasswordStepValues,
  registerPasswordStepSchema,
} from '@/lib/auth-schema'
import { cn } from '@/lib/utils'

interface RegisterPasswordStepProps {
  countdown: number
  defaultValues: RegisterPasswordStepValues
  email: string
  onBack: () => void
  onContinue: (values: RegisterPasswordStepValues) => Promise<void> | void
  onResend: () => Promise<void> | void
}

export function RegisterPasswordStep({
  countdown,
  defaultValues,
  email,
  onBack,
  onContinue,
  onResend,
}: RegisterPasswordStepProps) {
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors, isSubmitting, isValid },
  } = useForm<RegisterPasswordStepValues>({
    defaultValues,
    mode: 'onChange',
    resolver: zodResolver(registerPasswordStepSchema),
  })

  useEffect(() => {
    reset(defaultValues)
  }, [defaultValues, reset])

  const onSubmit = async (values: RegisterPasswordStepValues) => {
    await onContinue(values)
  }

  return (
    <div className="space-y-8">
      <div className="space-y-5">
        <p className="text-sm font-semibold tracking-[0.22em] text-[#9a6d4d] uppercase">
           Soulink
        </p>
        <RegisterStepHeader activeStep={2} />
      </div>

      <div className="space-y-4">
        <p className="text-sm font-semibold tracking-[0.22em] text-[#9a6d4d] uppercase">
          Step 2
        </p>
        <h1 className="font-serif text-5xl text-[#261f1b]">验证并设置密码</h1>
        <p className="text-base leading-7 text-[#7e6d61]">
          验证码已发送至 <span className="font-medium text-[#5d493d]">{email}</span>
        </p>
      </div>

      <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#3a302a]"
            htmlFor="register-code"
          >
            验证码
          </label>
          <div className="grid gap-3 sm:grid-cols-[1fr_6.5rem]">
            <Input
              id="register-code"
              inputMode="numeric"
              placeholder="6 位验证码"
              {...register('verificationCode')}
            />
            <button
              className={cn(
                buttonVariants({ variant: 'secondary', size: 'default' }),
                'rounded-2xl border-[#eadfd7] bg-white/76 text-[#8f6548] shadow-none hover:bg-white',
                (countdown > 0 || isSubmitting) &&
                  'cursor-default text-[#b2a194] hover:bg-white/76',
              )}
              disabled={countdown > 0 || isSubmitting}
              onClick={() => {
                void onResend()
              }}
              type="button"
            >
              {countdown > 0 ? `${countdown}s` : '重新发送'}
            </button>
          </div>
          {errors.verificationCode ? (
            <p className="text-sm text-[#a35b43]">
              {errors.verificationCode.message}
            </p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#3a302a]"
            htmlFor="register-password"
          >
            设置密码
          </label>
          <Input
            id="register-password"
            placeholder="至少 8 位，包含字母和数字"
            type="password"
            {...register('password')}
          />
          {errors.password ? (
            <p className="text-sm text-[#a35b43]">{errors.password.message}</p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#3a302a]"
            htmlFor="register-confirm-password"
          >
            确认密码
          </label>
          <Input
            id="register-confirm-password"
            placeholder="再次输入密码"
            type="password"
            {...register('confirmPassword')}
          />
          {errors.confirmPassword ? (
            <p className="text-sm text-[#a35b43]">
              {errors.confirmPassword.message}
            </p>
          ) : null}
        </div>

        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'w-full rounded-full bg-[#8f6548] text-white shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
          )}
          disabled={!isValid || isSubmitting}
          type="submit"
        >
          {isSubmitting ? '验证中...' : '下一步'}
        </button>
      </form>

      <button
        className="text-sm text-[#8e7b6d] transition-colors hover:text-[#725744]"
        onClick={onBack}
        type="button"
      >
        返回修改邮箱
      </button>
    </div>
  )
}
