import {
  AlertCircle,
  Check,
  ChevronRight,
  ClipboardCheck,
  Eye,
  EyeOff,
  Heart,
  LoaderCircle,
  LockKeyhole,
  PauseCircle,
  Save,
  ShieldCheck,
  Sparkles,
  UserRound,
  Wrench,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { getErrorMessage } from '@/lib/api'
import { RecommendationPreferencesCard } from '@/components/connection/recommendation-preferences-card'
import {
  fetchMatchingProfile,
  updateMatchingProfile,
  type IntentCategory,
  type MatchingProfile,
} from '@/lib/connection-api'
import { fetchSafetyCapabilities } from '@/lib/safety-api'
import { cn } from '@/lib/utils'
import { getStoredUserProfile } from '@/lib/user-profile-storage'

type PreviewMode = 'general' | 'project' | 'romantic' | 'feed'

const emptyProfile: MatchingProfile = {
  bio: '',
  skills: [],
  interests: [],
  availability: '',
  preferredScenes: [],
  collaborationStyle: '',
  boundaries: '',
  aiPortrait: '',
  aiPortraitApproved: false,
  romanticOptIn: false,
  matchingPaused: false,
  hiddenCategories: [],
}

const previewModes: Array<{ value: PreviewMode; label: string }> = [
  { value: 'general', label: '普通搭子' },
  { value: 'project', label: '项目协作' },
  { value: 'romantic', label: '认真关系' },
  { value: 'feed', label: '意图流' },
]

const categoryPrivacyOptions: Array<{ value: IntentCategory; label: string }> = [
  { value: 'meal', label: '饭搭子' },
  { value: 'study', label: '学习搭子' },
  { value: 'sport', label: '运动搭子' },
  { value: 'event', label: '活动同伴' },
  { value: 'conversation', label: '认真聊天' },
  { value: 'project', label: '项目伙伴' },
  { value: 'cofounder', label: 'CTO / 联创' },
]

function tagsToText(values: string[]) {
  return values.join('，')
}

function textToTags(value: string) {
  return value
    .split(/[,，、\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function profilePreviewText(profile: MatchingProfile, mode: PreviewMode) {
  if (mode === 'project') {
    const skills = profile.skills.slice(0, 4).join('、') || '技能待补充'
    return `${skills}。${profile.collaborationStyle || '协作方式待补充'}`
  }
  if (mode === 'romantic') {
    if (!profile.romanticOptIn) {
      return '认真关系匹配未开启，任何关系偏好都不会被用于推荐。'
    }
    return profile.aiPortrait || '画像尚未确认，只会使用你主动填写并批准的内容。'
  }
  if (mode === 'feed') {
    return `${profile.bio || '一位清华同学'}。联系方式会在双方同意前保持隐藏。`
  }
  return (
    profile.aiPortrait ||
    [profile.bio, profile.availability].filter(Boolean).join(' ') ||
    '补充一些资料后，这里会显示别人看到的简短预览。'
  )
}

export function ConnectionProfilePage() {
  const identity = useMemo(() => getStoredUserProfile(), [])
  const [profile, setProfile] = useState<MatchingProfile>(emptyProfile)
  const [previewMode, setPreviewMode] = useState<PreviewMode>('general')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState('')
  const [saved, setSaved] = useState(false)
  const [isSafetyAdmin, setIsSafetyAdmin] = useState(false)

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      const [profileResult, capabilitiesResult] = await Promise.allSettled([
        fetchMatchingProfile(),
        fetchSafetyCapabilities(),
      ])
      if (cancelled) {
        return
      }
      if (profileResult.status === 'fulfilled') {
        setProfile(profileResult.value)
      } else {
        setNotice(getErrorMessage(profileResult.reason, 'AI 档案暂时无法加载'))
      }
      setIsSafetyAdmin(
        capabilitiesResult.status === 'fulfilled' &&
          capabilitiesResult.value.isAdmin,
      )
      setLoading(false)
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [])

  const update = <K extends keyof MatchingProfile>(
    key: K,
    value: MatchingProfile[K],
  ) => {
    setProfile((current) => ({ ...current, [key]: value }))
    setSaved(false)
  }

  const saveProfile = async () => {
    if (saving) {
      return
    }
    setSaving(true)
    setNotice('')
    try {
      setProfile(await updateMatchingProfile(profile))
      setSaved(true)
      window.setTimeout(() => setSaved(false), 1800)
    } catch (error) {
      setNotice(getErrorMessage(error, 'AI 档案暂时无法保存'))
    } finally {
      setSaving(false)
    }
  }

  const toggleHiddenCategory = (category: IntentCategory) => {
    const hidden = profile.hiddenCategories.includes(category)
    update(
      'hiddenCategories',
      hidden
        ? profile.hiddenCategories.filter((item) => item !== category)
        : [...profile.hiddenCategories, category],
    )
  }

  if (loading) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center text-sm text-[#77736c]">
        <LoaderCircle className="mr-2 animate-spin" size={18} />
        正在读取你的 AI 档案...
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-[1100px] px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
      <header className="flex flex-col gap-4 border-b border-[#dfdcd6] pb-5 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="flex items-center gap-1.5 text-xs font-medium text-[#6c4a80]">
            <Sparkles size={13} />
            MATCHING PROFILE
          </p>
          <h1 className="mt-1 text-2xl font-semibold text-[#292724]">AI 档案与隐私</h1>
          <p className="mt-1 text-sm text-[#77736c]">你决定系统可以理解什么，也决定别人能看到什么。</p>
        </div>
        <button
          className="inline-flex h-10 items-center justify-center gap-2 self-start rounded-md bg-[#5b3477] px-4 text-sm font-semibold text-white hover:bg-[#4c2b65] disabled:opacity-50"
          disabled={saving}
          onClick={() => void saveProfile()}
          type="button"
        >
          {saving ? (
            <LoaderCircle className="animate-spin" size={16} />
          ) : saved ? (
            <Check size={16} />
          ) : (
            <Save size={16} />
          )}
          {saving ? '保存中...' : saved ? '已保存' : '保存档案'}
        </button>
      </header>

      {notice ? (
        <div className="mt-5 flex items-start gap-2 rounded-md border border-[#ead4ce] bg-[#fff6f3] px-4 py-3 text-sm text-[#8c463d]">
          <AlertCircle className="mt-0.5 shrink-0" size={16} />
          {notice}
        </div>
      ) : null}

      <div className="grid gap-7 py-7 lg:grid-cols-[minmax(0,1.2fr)_minmax(300px,0.8fr)]">
        <main className="min-w-0 space-y-6">
          <section className="rounded-lg border border-[#dfdcd6] bg-white">
            <div className="flex items-center gap-3 border-b border-[#e5e2dd] px-5 py-4">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#e9f1ed] text-[#2d6a53]">
                <ShieldCheck size={18} />
              </span>
              <div>
                <h2 className="text-sm font-semibold text-[#37332f]">清华身份</h2>
                <p className="text-xs text-[#77736c]">已验证，公开预览不会显示邮箱</p>
              </div>
            </div>
            <div className="grid gap-4 px-5 py-4 sm:grid-cols-2">
              <div>
                <p className="text-[11px] text-[#8a867f]">昵称</p>
                <p className="mt-1 text-sm font-medium text-[#48443f]">
                  {identity.nickname || '清华同学'}
                </p>
              </div>
              <div>
                <p className="text-[11px] text-[#8a867f]">院系与年级</p>
                <p className="mt-1 text-sm font-medium text-[#48443f]">
                  {identity.department || '待补充'} · {identity.grade || '待补充'}
                </p>
              </div>
            </div>
          </section>

          <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
            <div className="flex items-center gap-2">
              <UserRound size={17} className="text-[#5b3477]" />
              <h2 className="text-sm font-semibold text-[#37332f]">匹配资料</h2>
            </div>
            <div className="mt-5 space-y-5">
              <label className="block text-xs font-semibold text-[#5f5b55]">
                一段不表演的自我介绍
                <textarea
                  className="mt-2 min-h-28 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fcfbf9] p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                  maxLength={1000}
                  onChange={(event) => update('bio', event.target.value)}
                  placeholder="现在在做什么、喜欢怎样认识一个新朋友。"
                  value={profile.bio}
                />
              </label>

              <div className="grid gap-4 sm:grid-cols-2">
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  技能与领域
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    onChange={(event) =>
                      update('skills', textToTags(event.target.value))
                    }
                    placeholder="前端，AI，产品"
                    value={tagsToText(profile.skills)}
                  />
                </label>
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  兴趣
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    onChange={(event) =>
                      update('interests', textToTags(event.target.value))
                    }
                    placeholder="跑步，电影，创业"
                    value={tagsToText(profile.interests)}
                  />
                </label>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  通常什么时候有空
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    onChange={(event) => update('availability', event.target.value)}
                    placeholder="工作日晚间，周末下午"
                    value={profile.availability}
                  />
                </label>
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  偏好的校园场景
                  <input
                    className="mt-2 h-10 w-full rounded-md border border-[#d9d5cf] bg-[#fcfbf9] px-3 text-sm outline-none focus:border-[#7a5a8e]"
                    onChange={(event) =>
                      update('preferredScenes', textToTags(event.target.value))
                    }
                    placeholder="紫荆，图书馆，东操"
                    value={tagsToText(profile.preferredScenes)}
                  />
                </label>
              </div>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                协作与沟通方式
                <textarea
                  className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fcfbf9] p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                  onChange={(event) =>
                    update('collaborationStyle', event.target.value)
                  }
                  placeholder="例如：先对齐目标，再快速做一个能验证的版本。"
                  value={profile.collaborationStyle}
                />
              </label>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                边界与不希望发生的事
                <textarea
                  className="mt-2 min-h-20 w-full resize-y rounded-md border border-[#d9d5cf] bg-[#fcfbf9] p-3 text-sm leading-6 outline-none focus:border-[#7a5a8e]"
                  onChange={(event) => update('boundaries', event.target.value)}
                  placeholder="例如：不希望未经同意交换联系方式。"
                  value={profile.boundaries}
                />
              </label>
            </div>
          </section>

          <section className="rounded-lg border border-[#d9d0df] bg-[#f1ecf4] p-5">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Sparkles size={17} className="text-[#5b3477]" />
                <h2 className="text-sm font-semibold text-[#4d2d66]">AI 画像</h2>
              </div>
              <label className="inline-flex items-center gap-2 text-xs font-medium text-[#66566f]">
                <input
                  checked={profile.aiPortraitApproved}
                  className="h-4 w-4 accent-[#5b3477]"
                  onChange={(event) =>
                    update('aiPortraitApproved', event.target.checked)
                  }
                  type="checkbox"
                />
                批准用于匹配
              </label>
            </div>
            <textarea
              className="mt-4 min-h-32 w-full resize-y rounded-md border border-[#cec0d6] bg-white/80 p-3 text-sm leading-7 text-[#594b61] outline-none focus:border-[#76528a]"
              maxLength={1000}
              onChange={(event) => {
                update('aiPortrait', event.target.value)
                if (!event.target.value.trim()) {
                  update('aiPortraitApproved', false)
                }
              }}
              placeholder="这段画像始终可以编辑、拒绝或留空。"
              value={profile.aiPortrait}
            />
            <p className="mt-2 text-[11px] leading-5 text-[#766a7b]">
              只写你愿意确认的显式特征，不生成诊断、创伤或敏感身份推断。
            </p>
          </section>
        </main>

        <aside className="min-w-0 space-y-5">
          <section className="rounded-lg border border-[#dfdcd6] bg-[#fbfaf8] p-5">
            <div className="flex items-center gap-2">
              <Eye size={16} className="text-[#5b3477]" />
              <h2 className="text-sm font-semibold text-[#37332f]">别人会看到什么</h2>
            </div>
            <div className="mt-4 grid grid-cols-2 gap-1 rounded-md bg-[#ebe8e3] p-1">
              {previewModes.map((mode) => (
                <button
                  className={cn(
                    'h-8 rounded-md text-[11px] font-medium transition-colors',
                    previewMode === mode.value
                      ? 'bg-white text-[#4d2d66] shadow-sm'
                      : 'text-[#706c65]',
                  )}
                  key={mode.value}
                  onClick={() => setPreviewMode(mode.value)}
                  type="button"
                >
                  {mode.label}
                </button>
              ))}
            </div>
            <div className="mt-4 rounded-md border border-[#e0ddd7] bg-white p-4">
              <div className="flex items-center gap-2">
                <span className="inline-flex h-8 w-8 items-center justify-center rounded-md bg-[#e9f1ed] text-sm font-semibold text-[#2d6a53]">
                  {(identity.nickname || '同学').slice(0, 1)}
                </span>
                <span>
                  <span className="block text-sm font-semibold text-[#37332f]">
                    {previewMode === 'romantic' ? '一位同学' : identity.nickname || '清华同学'}
                  </span>
                  <span className="flex items-center gap-1 text-[10px] text-[#77736c]">
                    <ShieldCheck size={10} />
                    清华身份已验证
                  </span>
                </span>
              </div>
              <p className="mt-4 text-xs leading-6 text-[#69655f]">
                {profilePreviewText(profile, previewMode)}
              </p>
              {previewMode !== 'romantic' && profile.skills.length > 0 ? (
                <div className="mt-3 flex flex-wrap gap-1.5">
                  {profile.skills.slice(0, 4).map((skill) => (
                    <span
                      className="rounded-md bg-[#efede9] px-2 py-1 text-[10px] text-[#69655f]"
                      key={skill}
                    >
                      {skill}
                    </span>
                  ))}
                </div>
              ) : null}
            </div>
          </section>

          <section className="rounded-lg border border-[#dfdcd6] bg-white p-5">
            <div className="flex items-center gap-2">
              <LockKeyhole size={16} className="text-[#5b3477]" />
              <h2 className="text-sm font-semibold text-[#37332f]">匹配开关</h2>
            </div>
            <div className="mt-4 space-y-3">
              <button
                className={cn(
                  'flex w-full items-center justify-between rounded-md border px-3 py-3 text-left',
                  profile.romanticOptIn
                    ? 'border-[#dfc0cc] bg-[#fff3f7]'
                    : 'border-[#dedad4] bg-[#fbfaf8]',
                )}
                onClick={() => update('romanticOptIn', !profile.romanticOptIn)}
                type="button"
              >
                <span className="flex items-center gap-2.5">
                  <Heart size={16} className="text-[#9c4662]" />
                  <span>
                    <span className="block text-xs font-semibold text-[#4f4b46]">
                      认真关系匹配
                    </span>
                    <span className="mt-0.5 block text-[10px] text-[#817d76]">
                      独立、私密、可随时关闭
                    </span>
                  </span>
                </span>
                <span
                  className={cn(
                    'relative h-6 w-11 rounded-full transition-colors',
                    profile.romanticOptIn ? 'bg-[#a14765]' : 'bg-[#c9c5bf]',
                  )}
                >
                  <span
                    className={cn(
                      'absolute top-1 h-4 w-4 rounded-full bg-white transition-transform',
                      profile.romanticOptIn ? 'translate-x-6' : 'translate-x-1',
                    )}
                  />
                </span>
              </button>

              <button
                className={cn(
                  'flex w-full items-center justify-between rounded-md border px-3 py-3 text-left',
                  profile.matchingPaused
                    ? 'border-[#e3c9be] bg-[#fff5f1]'
                    : 'border-[#dedad4] bg-[#fbfaf8]',
                )}
                onClick={() => update('matchingPaused', !profile.matchingPaused)}
                type="button"
              >
                <span className="flex items-center gap-2.5">
                  <PauseCircle size={16} className="text-[#90604c]" />
                  <span>
                    <span className="block text-xs font-semibold text-[#4f4b46]">
                      暂停所有新匹配
                    </span>
                    <span className="mt-0.5 block text-[10px] text-[#817d76]">
                      现有房间不受影响
                    </span>
                  </span>
                </span>
                <span
                  className={cn(
                    'relative h-6 w-11 rounded-full transition-colors',
                    profile.matchingPaused ? 'bg-[#96634f]' : 'bg-[#c9c5bf]',
                  )}
                >
                  <span
                    className={cn(
                      'absolute top-1 h-4 w-4 rounded-full bg-white transition-transform',
                      profile.matchingPaused ? 'translate-x-6' : 'translate-x-1',
                    )}
                  />
                </span>
              </button>

              <div className="border-t border-[#e4e1dc] pt-4">
                <div className="flex items-center gap-2">
                  <EyeOff size={15} className="text-[#716c66]" />
                  <p className="text-xs font-semibold text-[#4f4b46]">隐藏意图类别</p>
                </div>
                <div className="mt-3 grid grid-cols-2 gap-x-3 gap-y-2">
                  {categoryPrivacyOptions.map((option) => (
                    <label
                      className="flex min-h-9 cursor-pointer items-center gap-2 rounded-md border border-[#dedad4] bg-[#fbfaf8] px-2.5 text-[11px] text-[#5f5b55]"
                      key={option.value}
                    >
                      <input
                        checked={profile.hiddenCategories.includes(option.value)}
                        className="h-3.5 w-3.5 accent-[#5b3477]"
                        onChange={() => toggleHiddenCategory(option.value)}
                        type="checkbox"
                      />
                      {option.label}
                    </label>
                  ))}
                </div>
              </div>
            </div>
          </section>

          <RecommendationPreferencesCard />

          <section className="rounded-lg border border-[#dfdcd6] bg-[#f5f3ef]">
            <Link
              className="flex items-center gap-3 border-b border-[#e1ded8] px-4 py-3.5 hover:bg-white/60"
              to="/account"
            >
              <Wrench size={16} className="text-[#716c66]" />
              <span className="min-w-0 flex-1">
                <span className="block text-xs font-semibold text-[#4f4b46]">账户设置</span>
                <span className="mt-0.5 block text-[10px] text-[#817d76]">
                  基础资料、密码与账号管理
                </span>
              </span>
              <ChevronRight size={15} className="text-[#8d8982]" />
            </Link>
            <Link
              className={cn(
                'flex items-center gap-3 px-4 py-3.5 hover:bg-white/60',
                isSafetyAdmin && 'border-b border-[#e1ded8]',
              )}
              to="/safety"
            >
              <ShieldCheck size={16} className="text-[#527164]" />
              <span className="min-w-0 flex-1">
                <span className="block text-xs font-semibold text-[#4f4b46]">隐私与安全</span>
                <span className="mt-0.5 block text-[10px] text-[#817d76]">
                  联系揭示、举报与数据边界
                </span>
              </span>
              <ChevronRight size={15} className="text-[#8d8982]" />
            </Link>
            {isSafetyAdmin ? (
              <Link
                className="flex items-center gap-3 px-4 py-3.5 hover:bg-white/60"
                to="/admin/safety"
              >
                <ClipboardCheck size={16} className="text-[#6c4a80]" />
                <span className="min-w-0 flex-1">
                  <span className="block text-xs font-semibold text-[#4f4b46]">
                    运营与安全
                  </span>
                  <span className="mt-0.5 block text-[10px] text-[#817d76]">
                    核查举报并更新处理状态
                  </span>
                </span>
                <ChevronRight size={15} className="text-[#8d8982]" />
              </Link>
            ) : null}
          </section>
        </aside>
      </div>
    </div>
  )
}
