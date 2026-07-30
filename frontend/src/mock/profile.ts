import type { ContactType, ProfileOption, UserProfile } from '@/types/profile'

export const profileDepartments: ProfileOption[] = []

export const profileGrades: ProfileOption[] = [
  { label: '大一', value: '大一' },
  { label: '大二', value: '大二' },
  { label: '大三', value: '大三' },
  { label: '大四', value: '大四' },
  { label: '研一', value: '研一' },
  { label: '研二', value: '研二' },
  { label: '研三', value: '研三' },
  { label: '博一', value: '博一' },
  { label: '博二', value: '博二' },
  { label: '博三及以上', value: '博三及以上' },
]

export const contactTypeOptions: Array<{ label: string; value: ContactType }> = [
  { label: '微信', value: 'wechat' },
  { label: 'QQ', value: 'qq' },
  { label: '手机号', value: 'phone' },
  { label: '其他', value: 'other' },
]

export function createDefaultProfile(email: string): UserProfile {
  return {
    avatar: '',
    nickname: '',
    email,
    schoolName: '',
    department: '',
    grade: '',
    contactType: '',
    contactValue: '',
  }
}
