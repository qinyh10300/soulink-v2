import { apiRequest } from '@/lib/api'

export interface ProfileContact {
  type: string
  label: string
  value: string
  updatedAt: string
}

export interface ProfileDetails {
  id: number
  email: string
  nickname: string
  schoolName: string
  department: string
  grade: string
  contactType: string
  contactValue: string
  contacts: ProfileContact[]
  createdAt: string
}

export interface ProfileUpdateResult {
  nickname: string
  schoolName: string
  department: string
  grade: string
  contactType: string
  contactValue: string
  contacts: ProfileContact[]
}

export interface PublicProfileDetails {
  id: number
  nickname: string
  schoolName: string
  department: string
  grade: string
  createdAt: string
}

export function fetchMyProfile() {
  return apiRequest<ProfileDetails>('/users/me/profile', {
    auth: true,
  })
}

export function updateMyProfile(input: {
  nickname: string
  department: string
  grade: string
  contactType: string
  contactValue: string
}) {
  return apiRequest<ProfileUpdateResult>('/users/me/profile', {
    method: 'PUT',
    auth: true,
    body: input,
  })
}

export function fetchPublicProfile(userId: number) {
  return apiRequest<PublicProfileDetails>(`/users/${userId}/profile`, {
    auth: true,
  })
}
