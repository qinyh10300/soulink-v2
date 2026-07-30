import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { Link } from 'react-router-dom'

import { RegisterStepHeader } from '@/components/auth/register-step-header'
import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  type RegisterEmailStepValues,
  registerEmailStepSchema,
} from '@/lib/auth-schema'
import { cn } from '@/lib/utils'

interface RegisterFormProps {
  defaultEmail: string
  onContinue: (values: RegisterEmailStepValues) => Promise<void> | void
}

export function RegisterForm({ defaultEmail, onContinue }: RegisterFormProps) {
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors, isSubmitting, isValid },
  } = useForm<RegisterEmailStepValues>({
    defaultValues: {
      email: defaultEmail,
    },
    mode: 'onChange',
    resolver: zodResolver(registerEmailStepSchema),
  })

  useEffect(() => {
    reset({ email: defaultEmail })
  }, [defaultEmail, reset])

  const onSubmit = async (values: RegisterEmailStepValues) => {
    await onContinue(values)
  }

  return (
    <div className="space-y-8">
      <div className="space-y-5">
        <p className="text-sm font-semibold tracking-[0.22em] text-[#9a6d4d] uppercase">
           Soulink
        </p>
        <RegisterStepHeader activeStep={1} />
        <div className="space-y-3">
          <h1 className="font-serif text-5xl text-[#261f1b]">创建您的账号</h1>
          <p className="text-base leading-7 text-[#7e6d61]">
            使用支持的高校邮箱注册，开始寻找您的契合之人。
          </p>
        </div>
      </div>

      <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#3a302a]"
            htmlFor="register-email"
          >
            {/* 学校邮箱（目前支持清华(协和)、北大、人大、中传、北外、北师大） */}
            学校邮箱
          </label>
          <Input
            id="register-email"
            placeholder="请输入学校邮箱"
            type="email"
            {...register('email')}
          />
          {errors.email ? (
            <p className="text-sm text-[#a35b43]">{errors.email.message}</p>
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
          {isSubmitting ? '发送中...' : '发送验证码'}
        </button>
      </form>

      <p className="text-sm text-[#7e6d61]">
        已有账号？
        <Link className="ml-2 text-[#9a6d4d] hover:text-[#7d5539]" to="/login">
          立即登录
        </Link>
      </p>
    </div>
  )
}
