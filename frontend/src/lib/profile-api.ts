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
  avatarUrl: string
  privacy: ProfilePrivacy
  createdAt: string
}

export interface ProfilePrivacy {
  profilePublic: boolean
  schedulePublic: boolean
  preferencesPublic: boolean
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
  avatarUrl: string
  schedulePublic: boolean
  preferencesPublic: boolean
  preferences?: PublicProfilePreferences
  createdAt: string
}

export interface PublicProfilePreferences {
  bio: string
  skills: string[]
  interests: string[]
  availability: string
  preferredScenes: string[]
  collaborationStyle: string
}

export interface ProfileMediaAsset {
  key: string
  url: string
  mimeType: string
  byteSize: number
  width: number
  height: number
  isAnimated: boolean
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
    body: { ...input },
  })
}

export function updateMyProfilePrivacy(input: ProfilePrivacy) {
  return apiRequest<ProfilePrivacy>('/users/me/profile/privacy', {
    method: 'PUT',
    auth: true,
    body: { ...input },
  })
}

export function uploadMyAvatar(file: File) {
  const form = new FormData()
  form.append('file', file)
  return apiRequest<ProfileMediaAsset>('/users/me/avatar', {
    method: 'POST',
    auth: true,
    body: form,
  })
}

export function deleteMyAvatar() {
  return apiRequest<{ deleted: boolean }>('/users/me/avatar', {
    method: 'DELETE',
    auth: true,
  })
}

export function fetchPublicProfile(userId: number) {
  return apiRequest<PublicProfileDetails>(`/users/${userId}/profile`, {
    auth: true,
  })
}
