import { ArrowLeft, UserRound } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import { fetchPublicProfile, type PublicProfileDetails } from '@/lib/profile-api'

function formatDate(value: string) {
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(new Date(value))
}

function avatarLabel(name: string) {
  return name.trim().slice(0, 1)
}

export function PublicProfilePage() {
  const { userId } = useParams()
  const numericUserId = Number(userId)
  const [profile, setProfile] = useState<PublicProfileDetails | null>(null)
  const [notice, setNotice] = useState('正在加载个人主页...')

  const loadProfile = useCallback(async () => {
    if (!Number.isFinite(numericUserId) || numericUserId <= 0) {
      setNotice('用户不存在')
      return
    }

    try {
      const profileData = await fetchPublicProfile(numericUserId)
      setProfile(profileData)
      setNotice('')
    } catch (error) {
      setNotice(getErrorMessage(error, '个人主页加载失败，请稍后重试'))
    }
  }, [numericUserId])

  useEffect(() => {
    void loadProfile()
  }, [loadProfile])

  return (
    <div className="min-h-screen bg-[#f7f2ec] px-4 pb-12 pt-6 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-3xl space-y-6">
        <div className="flex items-center gap-3">
          <Link
            className="inline-flex h-9 w-9 items-center justify-center rounded-full border border-[#eee4db] bg-white/80 text-[#5f4d42] transition-colors hover:bg-white"
            to="/plaza"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <h1 className="text-xl font-semibold text-[#2f221b]">个人主页</h1>
        </div>

        {profile ? (
          <section className="overflow-hidden rounded-[32px] border border-[#eee4db] bg-[rgba(255,251,247,0.96)] shadow-[0_18px_54px_rgba(72,50,33,0.06)]">
            <div className="h-32 bg-[radial-gradient(circle_at_top_left,#e2c4a8,#f8efe7_46%,#fffaf5)]" />
            <div className="px-6 pb-6">
              <div className="-mt-12 flex h-24 w-24 items-center justify-center rounded-full border-4 border-[#fffaf5] bg-[#f1e2d4] text-4xl font-semibold text-[#8f5c42] shadow-[0_14px_30px_rgba(72,50,33,0.12)]">
                {avatarLabel(profile.nickname) || <UserRound className="h-8 w-8" />}
              </div>
              <div className="mt-4 space-y-2">
                <h2 className="text-3xl font-semibold text-[#2f221b]">{profile.nickname}</h2>
                <p className="text-sm text-[#8a7467]">
                  {[profile.schoolName, profile.department, profile.grade].filter(Boolean).join(' · ')}
                </p>
                <p className="text-sm text-[#a08d81]">
                  加入于 {formatDate(profile.createdAt)}
                </p>
              </div>
            </div>
          </section>
        ) : null}

        {notice ? (
          <div className="rounded-2xl border border-[#eadfd5] bg-white/75 px-4 py-3 text-sm text-[#7b675b]">
            {notice}
          </div>
        ) : null}
      </div>
    </div>
  )
}
