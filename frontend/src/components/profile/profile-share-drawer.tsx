import {
  CalendarDays,
  Check,
  Copy,
  ExternalLink,
  Eye,
  LockKeyhole,
  Share2,
  Sparkles,
  X,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { AuthenticatedImage } from '@/components/media/authenticated-image'
import { getErrorMessage } from '@/lib/api'
import { fetchMyProfile, type ProfileDetails } from '@/lib/profile-api'
import { cn } from '@/lib/utils'

interface ProfileShareDrawerProps {
  open: boolean
  userId: number
  onClose: () => void
}

export function ProfileShareDrawer({ open, userId, onClose }: ProfileShareDrawerProps) {
  const [profile, setProfile] = useState<ProfileDetails | null>(null)
  const [notice, setNotice] = useState('')
  const [copied, setCopied] = useState(false)
  const shareURL = useMemo(
    () => (typeof window === 'undefined' ? `/profile/${userId}` : `${window.location.origin}/profile/${userId}`),
    [userId],
  )

  useEffect(() => {
    if (!open) return
    let cancelled = false
    void fetchMyProfile()
      .then((data) => {
        if (cancelled) return
        setProfile(data)
        setNotice('')
      })
      .catch((error) => {
        if (!cancelled) setNotice(getErrorMessage(error, '名片加载失败'))
      })

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      cancelled = true
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [onClose, open])

  if (!open) return null

  const canShare = Boolean(profile?.privacy.profilePublic)

  const copyLink = async () => {
    if (!canShare) return
    try {
      await navigator.clipboard.writeText(shareURL)
      setCopied(true)
      setNotice('主页链接已复制')
      window.setTimeout(() => setCopied(false), 1800)
    } catch {
      setNotice('复制失败，请在主页地址栏复制链接')
    }
  }

  const shareCard = async () => {
    if (!profile || !canShare) return
    if (!navigator.share) {
      await copyLink()
      return
    }
    try {
      await navigator.share({
        title: `${profile.nickname} 的 SouLink 名片`,
        text: `${profile.nickname} · ${profile.department} · 清华大学`,
        url: shareURL,
      })
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return
      setNotice('系统分享未完成，可以改用复制链接')
    }
  }

  return (
    <div className="fixed inset-0 z-[75]">
      <button
        aria-label="关闭名片抽屉"
        className="absolute inset-0 cursor-default bg-black/25 backdrop-blur-[1px]"
        onClick={onClose}
        type="button"
      />
      <aside
        aria-label="分享我的名片"
        aria-modal="true"
        className="absolute inset-y-0 right-0 flex w-full max-w-[420px] flex-col border-l border-[#dad7d1] bg-[#fbfaf8] shadow-[-18px_0_48px_rgba(34,32,29,0.14)]"
        role="dialog"
      >
        <header className="flex h-16 shrink-0 items-center justify-between border-b border-[#e1ded8] px-5">
          <div className="flex items-center gap-2">
            <Share2 className="h-4.5 w-4.5 text-[#42695b]" />
            <h2 className="text-base font-semibold text-[#292724]">分享我的名片</h2>
          </div>
          <button
            aria-label="关闭"
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-[#76716a] hover:bg-[#eeece8]"
            onClick={onClose}
            title="关闭"
            type="button"
          >
            <X className="h-4.5 w-4.5" />
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-6">
          {profile ? (
            <>
              <Link
                className="block overflow-hidden rounded-lg border border-[#d9d7d1] bg-white shadow-[0_8px_24px_rgba(34,32,29,0.07)] transition-transform hover:-translate-y-0.5"
                onClick={onClose}
                to={`/profile/${userId}`}
              >
                <div className="h-20 bg-[#dfe9e3]" />
                <div className="px-5 pb-5">
                  {profile.avatarUrl ? (
                    <AuthenticatedImage
                      alt={profile.nickname}
                      className="-mt-8 h-16 w-16 rounded-lg border-4 border-white bg-[#d1ddd6] object-cover"
                      src={profile.avatarUrl}
                    />
                  ) : (
                    <span className="-mt-8 flex h-16 w-16 items-center justify-center rounded-lg border-4 border-white bg-[#d1ddd6] text-2xl font-semibold text-[#315d4f]">
                      {profile.nickname.trim().slice(0, 1) || '同'}
                    </span>
                  )}
                  <div className="mt-3 flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h3 className="truncate text-xl font-semibold text-[#292b28]">{profile.nickname}</h3>
                      <p className="mt-1 text-sm leading-5 text-[#70756f]">
                        {[profile.schoolName, profile.department, profile.grade].filter(Boolean).join(' · ')}
                      </p>
                    </div>
                    <ExternalLink className="mt-1 h-4 w-4 shrink-0 text-[#7e847e]" />
                  </div>
                  <div className="mt-5 grid grid-cols-3 border-t border-[#eceae6] pt-4 text-center text-[11px] text-[#71766f]">
                    <span className="flex flex-col items-center gap-1.5">
                      <Eye className="h-4 w-4 text-[#527164]" />
                      个人主页
                    </span>
                    <span className={cn('flex flex-col items-center gap-1.5', !profile.privacy.schedulePublic && 'opacity-40')}>
                      <CalendarDays className="h-4 w-4 text-[#527164]" />
                      时间表
                    </span>
                    <span className={cn('flex flex-col items-center gap-1.5', !profile.privacy.preferencesPublic && 'opacity-40')}>
                      <Sparkles className="h-4 w-4 text-[#527164]" />
                      个人偏好
                    </span>
                  </div>
                </div>
              </Link>

              {!canShare ? (
                <div className="mt-4 flex gap-3 rounded-md border border-[#dfd4c6] bg-[#f8f1e8] p-4">
                  <LockKeyhole className="mt-0.5 h-4 w-4 shrink-0 text-[#856346]" />
                  <div>
                    <p className="text-sm font-semibold text-[#59483a]">个人主页尚未公开</p>
                    <p className="mt-1 text-xs leading-5 text-[#7b6a5b]">开启后，其他已登录的清华用户才能通过名片查看主页。</p>
                    <Link
                      className="mt-2 inline-flex text-xs font-semibold text-[#43695b] hover:underline"
                      onClick={onClose}
                      to="/account"
                    >
                      前往隐私设置
                    </Link>
                  </div>
                </div>
              ) : null}
            </>
          ) : null}

          {notice || !profile ? (
            <p className="mt-4 text-center text-xs text-[#77736d]">
              {notice || '正在读取名片...'}
            </p>
          ) : null}
        </div>

        <footer className="grid shrink-0 grid-cols-[1fr_auto] gap-2 border-t border-[#e1ded8] bg-white p-4 pb-[calc(1rem_+_env(safe-area-inset-bottom))]">
          <button
            className="inline-flex h-11 items-center justify-center gap-2 rounded-md bg-[#263b35] px-4 text-sm font-semibold text-white hover:bg-[#1d302a] disabled:cursor-not-allowed disabled:bg-[#aaa8a3]"
            disabled={!canShare}
            onClick={() => void shareCard()}
            type="button"
          >
            <Share2 className="h-4 w-4" />
            分享名片
          </button>
          <button
            aria-label="复制主页链接"
            className="inline-flex h-11 w-11 items-center justify-center rounded-md border border-[#cbc8c2] text-[#53504b] hover:bg-[#f1efeb] disabled:cursor-not-allowed disabled:opacity-40"
            disabled={!canShare}
            onClick={() => void copyLink()}
            title="复制主页链接"
            type="button"
          >
            {copied ? <Check className="h-4 w-4 text-[#337052]" /> : <Copy className="h-4 w-4" />}
          </button>
        </footer>
      </aside>
    </div>
  )
}
