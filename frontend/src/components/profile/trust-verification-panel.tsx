import {
  BadgeCheck,
  Camera,
  Check,
  Clock3,
  FileCheck2,
  LoaderCircle,
  RefreshCw,
  Upload,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'

import { getErrorMessage } from '@/lib/api'
import {
  appealPhotoVerification,
  fetchMyTrustProfile,
  submitPhotoVerification,
  type TrustProfile,
} from '@/lib/safety-api'

const photoStatusLabels: Record<TrustProfile['photoStatus'], string> = {
  not_submitted: '未申请',
  not_verified: '未验证',
  pending: '人工审核中',
  approved: '人工审核通过',
  rejected: '需要重新确认',
}

function TrustSignal({
  active,
  icon: Icon,
  label,
  value,
}: {
  active: boolean
  icon: typeof BadgeCheck
  label: string
  value: string
}) {
  return (
    <div className="flex min-h-20 items-center gap-3 border-b border-[#e2e4e0] px-4 py-3 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0">
      <span className={`inline-flex h-9 w-9 shrink-0 items-center justify-center rounded ${active ? 'bg-[#e4eee8] text-[#315f4f]' : 'bg-[#eceeeb] text-[#8a8f89]'}`}>
        <Icon className="h-4 w-4" />
      </span>
      <div className="min-w-0">
        <p className="text-[11px] text-[#8a8f89]">{label}</p>
        <p className="mt-0.5 truncate text-xs font-semibold text-[#424743]">{value}</p>
      </div>
    </div>
  )
}

export function TrustVerificationPanel() {
  const [profile, setProfile] = useState<TrustProfile | null>(null)
  const [photo, setPhoto] = useState<File | null>(null)
  const [appealReason, setAppealReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [notice, setNotice] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    let cancelled = false
    void fetchMyTrustProfile()
      .then((data) => {
        if (!cancelled) setProfile(data)
      })
      .catch((error: unknown) => {
        if (!cancelled) setNotice(getErrorMessage(error, '信任状态加载失败'))
      })
    return () => {
      cancelled = true
    }
  }, [])

  const upload = async () => {
    if (!photo || busy) return
    if (!['image/jpeg', 'image/png'].includes(photo.type) || photo.size > 2 * 1024 * 1024) {
      setNotice('请选择 2MB 以内的 JPEG 或 PNG 照片')
      return
    }
    setBusy(true)
    setNotice('')
    try {
      const next = await submitPhotoVerification(photo)
      setProfile(next)
      setPhoto(null)
      if (inputRef.current) inputRef.current.value = ''
      setNotice('照片已提交人工审核')
    } catch (error) {
      setNotice(getErrorMessage(error, '照片提交失败'))
    } finally {
      setBusy(false)
    }
  }

  const appeal = async () => {
    if (!profile?.photoRequestId || busy) return
    setBusy(true)
    setNotice('')
    try {
      const next = await appealPhotoVerification(profile.photoRequestId, appealReason)
      setProfile(next)
      setAppealReason('')
      setNotice('申诉已提交，照片将重新人工审核')
    } catch (error) {
      setNotice(getErrorMessage(error, '申诉提交失败'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <section className="overflow-hidden border-y border-[#d9dcd8] bg-white">
      <div className="flex items-center justify-between gap-4 border-b border-[#e2e4e0] px-5 py-4 sm:px-6">
        <div>
          <h2 className="text-base font-semibold text-[#303531]">信任状态</h2>
          <p className="mt-1 text-xs text-[#858a84]">每项信号独立展示，不合并为单一评分</p>
        </div>
        {profile?.photoStatus === 'approved' ? (
          <span className="inline-flex h-8 items-center gap-1.5 rounded bg-[#e6f0ea] px-2.5 text-xs font-semibold text-[#315f4f]">
            <Check className="h-3.5 w-3.5" />照片已核验
          </span>
        ) : null}
      </div>

      {profile ? (
        <div className="grid sm:grid-cols-2 lg:grid-cols-4">
          <TrustSignal active={profile.emailVerified} icon={BadgeCheck} label="校园邮箱" value={profile.emailVerified ? '已验证' : '未验证'} />
          <TrustSignal active={profile.profileComplete} icon={FileCheck2} label="连接资料" value={profile.profileComplete ? '已完善' : '待完善'} />
          <TrustSignal active={profile.photoStatus === 'approved'} icon={Camera} label="本人照片" value={photoStatusLabels[profile.photoStatus]} />
          <TrustSignal active={profile.recentlyActive} icon={Clock3} label="近期状态" value={profile.recentlyActive ? '14 天内活跃' : '近期未活跃'} />
        </div>
      ) : (
        <div className="flex min-h-24 items-center justify-center text-xs text-[#858a84]">
          <LoaderCircle className="mr-2 h-4 w-4 animate-spin" />正在读取信任状态
        </div>
      )}

      {profile && profile.photoStatus !== 'approved' && profile.photoStatus !== 'pending' ? (
        <div className="border-t border-[#e2e4e0] px-5 py-4 sm:px-6">
          {profile.photoStatus === 'rejected' ? (
            <div className="mb-4 rounded border border-[#ead4ce] bg-[#fff7f4] px-3 py-2 text-xs leading-5 text-[#844d43]">
              {profile.photoReviewNotes || '照片暂未通过人工审核'}
            </div>
          ) : null}
          {profile.photoStatus === 'rejected' && profile.photoRequestId ? (
            <div className="flex flex-col gap-2 sm:flex-row">
              <textarea
                aria-label="照片审核申诉理由"
                className="min-h-20 flex-1 resize-y rounded border border-[#d5d9d5] px-3 py-2 text-sm outline-none focus:border-[#668477]"
                maxLength={1000}
                onChange={(event) => setAppealReason(event.target.value)}
                placeholder="补充说明或指出审核误判"
                value={appealReason}
              />
              <button
                className="inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded border border-[#cfd5d0] px-4 text-sm font-semibold text-[#426456] hover:bg-[#edf2ef] disabled:opacity-50"
                disabled={busy || appealReason.trim().length < 8}
                onClick={() => void appeal()}
                type="button"
              >
                {busy ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
                提交申诉
              </button>
            </div>
          ) : (
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <label className="inline-flex h-10 cursor-pointer items-center justify-center gap-2 rounded border border-[#cfd5d0] bg-white px-4 text-sm font-medium text-[#4d5f56] hover:bg-[#f1f4f2]">
                <Camera className="h-4 w-4" />
                {photo?.name || '选择本人照片'}
                <input
                  accept="image/jpeg,image/png"
                  className="hidden"
                  onChange={(event) => setPhoto(event.target.files?.[0] ?? null)}
                  ref={inputRef}
                  type="file"
                />
              </label>
              <button
                className="inline-flex h-10 items-center justify-center gap-2 rounded bg-[#315f4f] px-4 text-sm font-semibold text-white hover:bg-[#274e41] disabled:opacity-45"
                disabled={!photo || busy}
                onClick={() => void upload()}
                type="button"
              >
                {busy ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <Upload className="h-4 w-4" />}
                提交人工审核
              </button>
            </div>
          )}
        </div>
      ) : null}

      {profile?.photoStatus === 'pending' ? (
        <div className="border-t border-[#e2e4e0] px-5 py-3 text-xs text-[#687069] sm:px-6">
          照片正在人工审核，结果会显示在这里。
        </div>
      ) : null}
      {notice ? <div className="border-t border-[#e2e4e0] px-5 py-3 text-xs text-[#77564f] sm:px-6">{notice}</div> : null}
    </section>
  )
}
