import { ProfileAvatarUploader } from '@/components/profile/profile-avatar-uploader'
import { buttonVariants } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { majorDescriptionMaxLength } from '@/lib/auth-schema'
import { cn } from '@/lib/utils'
import type { UserProfile } from '@/types/profile'

interface ProfileFormErrors {
  nickname?: string
  department?: string
  grade?: string
}

interface ProfileFormCardProps {
  contactTypes: Array<{ label: string; value: string }>
  errors: ProfileFormErrors
  grades: Array<{ label: string; value: string }>
  notice: string
  profile: UserProfile
  onAvatarChange: (file: File) => void
  onAvatarRemove: () => void
  onChange: (field: keyof UserProfile, value: string) => void
  onSubmit: () => void
}

export function ProfileFormCard({
  contactTypes,
  errors,
  grades,
  notice,
  profile,
  onAvatarChange,
  onAvatarRemove,
  onChange,
  onSubmit,
}: ProfileFormCardProps) {
  return (
    <div className="rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.95)] p-8 shadow-[0_18px_56px_rgba(72,50,33,0.06)] sm:p-10">
      <div className="space-y-8">
        <ProfileAvatarUploader
          avatar={profile.avatar}
          nickname={profile.nickname}
          onFileChange={onAvatarChange}
          onRemove={onAvatarRemove}
        />

        <div className="grid gap-6 md:grid-cols-2">
          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]">昵称</label>
            <Input
              onChange={(event) => onChange('nickname', event.target.value)}
              placeholder="你的名字或昵称"
              value={profile.nickname}
            />
            {errors.nickname ? (
              <p className="text-sm text-[#a35b43]">{errors.nickname}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]">邮箱</label>
            <Input readOnly value={profile.email} />
            <p className="text-sm text-[#8b7a6f]">邮箱用于登录，当前不可修改。</p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]">学校</label>
            <Input readOnly value={profile.schoolName} />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]">专业介绍</label>
            <Input
              maxLength={majorDescriptionMaxLength}
              onChange={(event) => onChange('department', event.target.value)}
              placeholder="例如：国际关系 / 金融 / 临床医学"
              value={profile.department}
            />
            <p className="text-sm text-[#8b7a6f]">1 到 {majorDescriptionMaxLength} 个字，用来简要介绍你的专业方向。</p>
            {errors.department ? (
              <p className="text-sm text-[#a35b43]">{errors.department}</p>
            ) : null}
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium text-[#3a302a]">年级</label>
            <Select
              onChange={(event) => onChange('grade', event.target.value)}
              value={profile.grade}
            >
              <option value="">请选择</option>
              {grades.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
            {errors.grade ? (
              <p className="text-sm text-[#a35b43]">{errors.grade}</p>
            ) : null}
          </div>
        </div>

        <div className="rounded-[24px] border border-[#ebe1d8] bg-[#faf5ef] p-6">
          <div className="space-y-2">
            <h3 className="text-lg font-medium text-[#2f221b]">
              留给匹配对象的联系方式
            </h3>
            <p className="text-sm leading-6 text-[#c53b2f]">
              匹配成功后，对方会看到这里的联系方式；不填写时默认展示邮箱。
            </p>
          </div>

          <div className="mt-5 grid gap-5 md:grid-cols-[13rem_1fr]">
            <Select
              onChange={(event) => onChange('contactType', event.target.value)}
              value={profile.contactType}
            >
              <option value="">请选择类型</option>
              {contactTypes.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </Select>
            <Input
              onChange={(event) => onChange('contactValue', event.target.value)}
              placeholder="请输入联系方式内容"
              value={profile.contactValue}
            />
          </div>
        </div>

        <div className="flex items-center justify-between gap-4">
          <p className="text-sm text-[#8b7a6f]">{notice}</p>
          <button
            className={cn(
              buttonVariants({ size: 'lg' }),
              'rounded-full bg-[#8f6548] px-8 text-white shadow-[0_14px_32px_rgba(105,73,51,0.18)] hover:bg-[#7f5a40]',
            )}
            onClick={onSubmit}
            type="button"
          >
            保存修改
          </button>
        </div>
      </div>
    </div>
  )
}
