export type ContactType = '' | 'wechat' | 'qq' | 'phone' | 'email' | 'other'

export interface UserProfile {
  avatar: string
  nickname: string
  email: string
  schoolName: string
  department: string
  grade: string
  contactType: ContactType
  contactValue: string
}

export interface ProfileOption {
  label: string
  value: string
}

export interface AuthSession {
  token: string
  userId: number
  email: string
  nickname: string
  loggedInAt: string
}
