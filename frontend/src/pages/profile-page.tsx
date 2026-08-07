import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { DashboardFooter } from '@/components/dashboard/dashboard-footer'
import { AccountManagementCard } from '@/components/profile/account-management-card'
import { ChangePasswordModal } from '@/components/profile/change-password-modal'
import { DeleteAccountConfirmModal } from '@/components/profile/delete-account-confirm-modal'
import { ProfileFormCard } from '@/components/profile/profile-form-card'
import { ProfilePrivacyPanel } from '@/components/profile/profile-privacy-panel'
import { TrustVerificationPanel } from '@/components/profile/trust-verification-panel'
import { getErrorMessage } from '@/lib/api'
import { majorDescriptionMaxLength } from '@/lib/auth-schema'
import { fetchContactTypes, fetchGrades } from '@/lib/meta-api'
import {
  deleteMyAvatar,
  fetchMyProfile,
  type ProfilePrivacy,
  updateMyProfile,
  updateMyProfilePrivacy,
  uploadMyAvatar,
} from '@/lib/profile-api'
import { clearSurveyDraft } from '@/lib/survey-storage'
import {
  clearAuthSession,
  clearUserProfile,
  getStoredUserProfile,
  saveUserProfile,
} from '@/lib/user-profile-storage'
import type { ContactType, ProfileOption, UserProfile } from '@/types/profile'

interface ProfileFormErrors {
  nickname?: string
  department?: string
  grade?: string
}

function toOptions(values: string[]): ProfileOption[] {
  return values.map((value) => ({
    label: value,
    value,
  }))
}

export function ProfilePage() {
  const navigate = useNavigate()
  const [profile, setProfile] = useState<UserProfile>(() => getStoredUserProfile())
  const [errors, setErrors] = useState<ProfileFormErrors>({})
  const [notice, setNotice] = useState('正在加载个人资料...')
  const [grades, setGrades] = useState<ProfileOption[]>([])
  const [contactTypes, setContactTypes] = useState<Array<{ label: string; value: string }>>([])
  const [changePasswordOpen, setChangePasswordOpen] = useState(false)
  const [deleteAccountOpen, setDeleteAccountOpen] = useState(false)
  const [privacy, setPrivacy] = useState<ProfilePrivacy>({
    profilePublic: false,
    schedulePublic: false,
    preferencesPublic: false,
  })
  const [pendingAvatar, setPendingAvatar] = useState<File | null>(null)
  const [removeAvatar, setRemoveAvatar] = useState(false)
  const [privacySaving, setPrivacySaving] = useState(false)
  const [privacyNotice, setPrivacyNotice] = useState('设置尚未同步')

  useEffect(() => {
    let cancelled = false

    const loadProfilePage = async () => {
      try {
        const profileData = await fetchMyProfile()

        const [gradeData, contactTypeData] = await Promise.all([
          fetchGrades(),
          fetchContactTypes(),
        ])

        if (cancelled) {
          return
        }

        setGrades(toOptions(gradeData))
        setContactTypes(contactTypeData)

        const nextProfile: UserProfile = {
          ...getStoredUserProfile(profileData.email),
          email: profileData.email,
          nickname: profileData.nickname,
          schoolName: profileData.schoolName,
          department: profileData.department,
          grade: profileData.grade,
          contactType: (profileData.contactType || '') as ContactType,
          contactValue: profileData.contactValue || '',
          avatar: profileData.avatarUrl || '',
        }

        setProfile(nextProfile)
        setPrivacy(profileData.privacy)
        setPrivacyNotice('隐私设置已同步')
        saveUserProfile(nextProfile)
        setNotice('您可以随时更新公开给匹配对象的信息。')
      } catch (error) {
        if (cancelled) {
          return
        }

        setNotice(getErrorMessage(error, '个人资料加载失败，请稍后重试'))
      }
    }

    void loadProfilePage()

    return () => {
      cancelled = true
    }
  }, [])

  const handleChange = (field: keyof UserProfile, value: string) => {
    setProfile((current) => ({
      ...current,
      [field]: value,
    }))
    setErrors((current) => ({
      ...current,
      [field]: undefined,
    }))
  }

  const handleAvatarChange = (file: File) => {
    setPendingAvatar(file)
    setRemoveAvatar(false)
    const reader = new FileReader()

    reader.onload = () => {
      setProfile((current) => ({
        ...current,
        avatar: typeof reader.result === 'string' ? reader.result : current.avatar,
      }))
    }

    reader.readAsDataURL(file)
  }

  const handleAvatarRemove = () => {
    setPendingAvatar(null)
    setRemoveAvatar(true)
    setProfile((current) => ({ ...current, avatar: '' }))
  }

  const handleSave = async () => {
    const nextErrors: ProfileFormErrors = {}

    if (!profile.nickname.trim()) {
      nextErrors.nickname = '请输入昵称'
    }

    if (!profile.department.trim()) {
      nextErrors.department = '请输入专业介绍'
    } else if (profile.department.trim().length > majorDescriptionMaxLength) {
      nextErrors.department = `专业介绍不能超过 ${majorDescriptionMaxLength} 个字`
    }

    if (!profile.grade.trim()) {
      nextErrors.grade = '请选择年级'
    }

    setErrors(nextErrors)

    if (Object.keys(nextErrors).length > 0) {
      setNotice('请先完善必填信息后再保存。')
      return
    }

    if ((profile.contactType === '') !== (profile.contactValue.trim() === '')) {
      setNotice('请同时填写联系方式类型和内容。')
      return
    }

    try {
      const updated = await updateMyProfile({
        nickname: profile.nickname.trim(),
        department: profile.department,
        grade: profile.grade,
        contactType: profile.contactType,
        contactValue: profile.contactValue.trim(),
      })

      let avatar = profile.avatar
      if (pendingAvatar) {
        const uploaded = await uploadMyAvatar(pendingAvatar)
        avatar = uploaded.url
      } else if (removeAvatar) {
        await deleteMyAvatar()
        avatar = ''
      }

      const nextProfile: UserProfile = {
        ...profile,
        avatar,
        nickname: updated.nickname,
        schoolName: updated.schoolName,
        department: updated.department,
        grade: updated.grade,
        contactType: (updated.contactType || '') as ContactType,
        contactValue: updated.contactValue || '',
      }

      setProfile(nextProfile)
      saveUserProfile(nextProfile)
      window.location.reload()
    } catch (error) {
      setNotice(getErrorMessage(error, '保存失败，请稍后重试'))
    }
  }

  const handlePrivacySave = async () => {
    try {
      setPrivacySaving(true)
      setPrivacyNotice('正在保存...')
      const updated = await updateMyProfilePrivacy(privacy)
      setPrivacy(updated)
      setPrivacyNotice('隐私设置已保存')
    } catch (error) {
      setPrivacyNotice(getErrorMessage(error, '隐私设置保存失败'))
    } finally {
      setPrivacySaving(false)
    }
  }

  const handleDeleteAccount = () => {
    clearAuthSession()
    clearUserProfile()
    clearSurveyDraft()
    navigate('/login')
  }

  const handlePasswordChanged = () => {
    clearAuthSession()
    clearUserProfile()
    clearSurveyDraft()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-[#f6efe8]">
      <main className="px-6 pb-0 pt-10 lg:px-8 lg:pt-14">
        <div className="mx-auto w-full max-w-3xl space-y-10">
          <section className="space-y-3">
            <h1 className="font-serif text-5xl text-[#2f221b]">个人资料</h1>
            <p className="max-w-2xl text-lg leading-8 text-[#7f6c60]">
              管理你的个人信息，让匹配更准确。
            </p>
          </section>

          <ProfileFormCard
            contactTypes={contactTypes}
            errors={errors}
            grades={grades}
            notice={notice}
            onAvatarChange={handleAvatarChange}
            onAvatarRemove={handleAvatarRemove}
            onChange={handleChange}
            onSubmit={() => {
              void handleSave()
            }}
            profile={profile}
          />

          <ProfilePrivacyPanel
            notice={privacyNotice}
            onChange={(next) => {
              setPrivacy(next)
              setPrivacyNotice('有未保存的更改')
            }}
            onSave={() => void handlePrivacySave()}
            saving={privacySaving}
            value={privacy}
          />

          <TrustVerificationPanel />

          <AccountManagementCard
            onDeleteAccount={() => setDeleteAccountOpen(true)}
            onOpenChangePassword={() => setChangePasswordOpen(true)}
          />
        </div>
      </main>

      <DashboardFooter />

      <ChangePasswordModal
        onClose={() => setChangePasswordOpen(false)}
        onSuccess={handlePasswordChanged}
        open={changePasswordOpen}
      />
      <DeleteAccountConfirmModal
        expectedEmail={profile.email}
        onCancel={() => setDeleteAccountOpen(false)}
        onDeleted={handleDeleteAccount}
        open={deleteAccountOpen}
      />
    </div>
  )
}
