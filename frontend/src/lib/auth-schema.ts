import { z } from 'zod'

export const supportedEmailSuffixes = [
  '@mails.tsinghua.edu.cn',
  '@mail.tsinghua.edu.cn',
  '@tsinghua.org.cn',
  '@mail.tsinghua.org.cn',
  '@tsinghua.edu.cn',
  '@student.pumc.edu.cn',
  '@pku.edu.cn',
  '@stu.pku.edu.cn',
  '@bimba.pku.edu.cn',
  '@mail.bnu.edu.cn',
  '@mails.cuc.edu.cn',
  '@cuc.edu.cn',
  '@ruc.edu.cn',
  '@mail.bfsu.edu.cn',
  '@bfsu.edu.cn',
  '@mail.xs.ustb.edu.cn',
  '@csu.edu.cn',
  '@tongji.edu.cn',
  '@stu.zuel.edu.cn',
  '@u.nus.edu',
  '@e.ntu.edu.sg',
  '@m.fudan.edu.cn',
  '@sjtu.edu.cn',
] as const

export const majorDescriptionMaxLength = 20

function isSupportedSchoolEmail(email: string) {
  const normalizedEmail = email.trim().toLowerCase()

  return supportedEmailSuffixes.some((suffix) => normalizedEmail.endsWith(suffix))
}

function createSchoolEmailSchema(invalidDomainMessage: string) {
  return z
    .string()
    .trim()
    .min(1, '邮箱不能为空')
    .email('邮箱格式错误')
    .refine((email) => isSupportedSchoolEmail(email), {
      message: invalidDomainMessage,
    })
}

export const loginEmailSchema = createSchoolEmailSchema(
  '仅支持清华、北大、人大、中传、北外、北师大、北科大、中南、同济及协和邮箱登录',
)
export const registerEmailSchema =
  createSchoolEmailSchema('仅支持清华、北大、人大、中传、北外、北师大、北科大、中南、同济及协和邮箱注册')

export const loginSchema = z.object({
  email: loginEmailSchema,
  password: z.string().trim().min(1, '密码不能为空'),
  agreement: z.boolean().refine((value) => value, {
    message: '请先阅读并同意用户协议',
  }),
})

export const registerEmailStepSchema = z.object({
  email: registerEmailSchema,
})

export const registerPasswordStepSchema = z
  .object({
    verificationCode: z
      .string()
      .trim()
      .min(1, '验证码不能为空')
      .regex(/^\d{6}$/, '验证码必须为 6 位数字'),
    password: z
      .string()
      .trim()
      .min(1, '密码不能为空')
      .min(8, '密码至少 8 位')
      .regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, '密码需同时包含字母和数字'),
    confirmPassword: z.string().trim().min(1, '请再次输入密码'),
  })
  .refine((values) => values.password === values.confirmPassword, {
    path: ['confirmPassword'],
    message: '两次输入的密码不一致',
  })

export const forgotPasswordResetSchema = z
  .object({
    verificationCode: z
      .string()
      .trim()
      .min(1, '请输入验证码')
      .regex(/^\d{6}$/, '验证码必须为 6 位数字'),
    password: z
      .string()
      .trim()
      .min(1, '密码不能为空')
      .min(8, '密码至少 8 位')
      .regex(/^(?=.*[A-Za-z])(?=.*\d).+$/, '密码需同时包含字母和数字'),
    confirmPassword: z.string().trim().min(1, '请再次输入密码'),
  })
  .refine((values) => values.password === values.confirmPassword, {
    path: ['confirmPassword'],
    message: '两次输入的密码不一致',
  })

export const registerProfileStepSchema = z.object({
  nickname: z.string().trim().min(1, '昵称不能为空'),
  department: z
    .string()
    .trim()
    .min(1, '请输入专业介绍')
    .max(majorDescriptionMaxLength, `专业介绍不能超过 ${majorDescriptionMaxLength} 个字`),
  grade: z.string().trim().min(1, '请选择年级'),
})

export const registerSchema = registerEmailStepSchema

export type LoginFormValues = z.infer<typeof loginSchema>
export type ForgotPasswordResetValues = z.infer<typeof forgotPasswordResetSchema>
export type RegisterEmailStepValues = z.infer<typeof registerEmailStepSchema>
export type RegisterPasswordStepValues = z.infer<
  typeof registerPasswordStepSchema
>
export type RegisterProfileStepValues = z.infer<
  typeof registerProfileStepSchema
>
