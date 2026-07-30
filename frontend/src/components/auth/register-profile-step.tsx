import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'

import { RegisterStepHeader } from '@/components/auth/register-step-header'
import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import {
  majorDescriptionMaxLength,
  type RegisterProfileStepValues,
  registerProfileStepSchema,
} from '@/lib/auth-schema'
import { cn } from '@/lib/utils'

interface RegisterProfileStepProps {
  defaultValues: RegisterProfileStepValues
  grades: string[]
  isOptionsLoading: boolean
  onBack: () => void
  onComplete: (values: RegisterProfileStepValues) => Promise<void> | void
}

export function RegisterProfileStep({
  defaultValues,
  grades,
  isOptionsLoading,
  onBack,
  onComplete,
}: RegisterProfileStepProps) {
  const {
    handleSubmit,
    register,
    reset,
    formState: { errors, isSubmitting, isValid },
  } = useForm<RegisterProfileStepValues>({
    defaultValues,
    mode: 'onChange',
    resolver: zodResolver(registerProfileStepSchema),
  })

  useEffect(() => {
    reset(defaultValues)
  }, [defaultValues, reset])

  const onSubmit = async (values: RegisterProfileStepValues) => {
    await onComplete(values)
  }

  return (
    <div className="space-y-8">
      <div className="space-y-5">
        <p className="text-xs font-semibold text-[#6c4a80]">
          SOULINK · 清华校内
        </p>
        <RegisterStepHeader activeStep={3} />
      </div>

      <div className="space-y-4">
        <p className="text-xs font-semibold text-[#6c4a80]">
          Step 3
        </p>
        <h1 className="text-3xl font-semibold text-[#292724]">完善基础资料</h1>
        <p className="text-sm leading-7 text-[#77736c]">
          填写基本信息，让匹配更准确。
        </p>
      </div>

      <form className="space-y-5" onSubmit={handleSubmit(onSubmit)}>
        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#4f4b46]"
            htmlFor="register-nickname"
          >
            昵称
          </label>
          <Input
            id="register-nickname"
            placeholder="您的名字或昵称"
            {...register('nickname')}
          />
          {errors.nickname ? (
            <p className="text-sm text-[#a35b43]">{errors.nickname.message}</p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#4f4b46]"
            htmlFor="register-department"
          >
            专业介绍
          </label>
          <Input
            id="register-department"
            maxLength={majorDescriptionMaxLength}
            placeholder="例如：国际关系 / 金融 / 临床医学"
            {...register('department')}
          />
          <p className="text-sm text-[#8b7a6f]">1 到 {majorDescriptionMaxLength} 个字，用来简要介绍你的专业方向。</p>
          {errors.department ? (
            <p className="text-sm text-[#a35b43]">{errors.department.message}</p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label
            className="text-sm font-medium text-[#4f4b46]"
            htmlFor="register-grade"
          >
            年级
          </label>
          <Select
            id="register-grade"
            defaultValue=""
            disabled={isOptionsLoading}
            {...register('grade')}
          >
            <option value="" disabled>
              {isOptionsLoading ? '加载中...' : '请选择'}
            </option>
            {grades.map((grade) => (
              <option key={grade} value={grade}>
                {grade}
              </option>
            ))}
          </Select>
          {errors.grade ? (
            <p className="text-sm text-[#a35b43]">{errors.grade.message}</p>
          ) : null}
        </div>

        <button
          className={cn(
            buttonVariants({ size: 'lg' }),
            'w-full rounded-md bg-[#5b3477] text-white hover:bg-[#4c2b65]',
          )}
          disabled={!isValid || isSubmitting || isOptionsLoading}
          type="submit"
        >
          {isSubmitting ? '注册中...' : '完成注册'}
        </button>
      </form>

      <button
        className="text-sm text-[#77736c] transition-colors hover:text-[#5b3477]"
        onClick={onBack}
        type="button"
      >
        返回上一步
      </button>
    </div>
  )
}
