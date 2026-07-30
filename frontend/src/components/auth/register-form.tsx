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
        <p className="text-xs font-semibold text-[#6c4a80]">
          SOULINK · 清华校内
        </p>
        <RegisterStepHeader activeStep={1} />
        <div className="space-y-3">
          <h1 className="text-3xl font-semibold text-[#292724]">创建账号</h1>
          <p className="text-sm leading-7 text-[#77736c]">
            使用清华邮箱验证身份，开始一次有明确理由的连接。
          </p>
        </div>
      </div>

      <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#4f4b46]"
            htmlFor="register-email"
          >
            清华邮箱
          </label>
          <Input
            id="register-email"
            placeholder="请输入清华邮箱"
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
            'w-full rounded-md bg-[#5b3477] text-white hover:bg-[#4c2b65]',
          )}
          disabled={!isValid || isSubmitting}
          type="submit"
        >
          {isSubmitting ? '发送中...' : '发送验证码'}
        </button>
      </form>

      <p className="text-sm text-[#77736c]">
        已有账号？
        <Link className="ml-2 text-[#5b3477] hover:text-[#45265e]" to="/login">
          立即登录
        </Link>
      </p>
    </div>
  )
}
