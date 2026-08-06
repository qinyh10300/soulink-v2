import {
  ArrowLeft,
  ArrowRight,
  Bell,
  BookOpen,
  Dumbbell,
  Heart,
  LoaderCircle,
  ShieldCheck,
  UtensilsCrossed,
  UsersRound,
  X,
  type LucideIcon,
} from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'

import { EngagementToggle } from '@/components/engagement/engagement-toggle'
import { useEngagement } from '@/contexts/engagement-context'
import { getErrorMessage } from '@/lib/api'
import type {
  EngagementScene,
  NotificationPermissionState,
} from '@/lib/engagement-api'
import { cn } from '@/lib/utils'

interface SceneOption {
  value: EngagementScene
  label: string
  detail: string
  icon: LucideIcon
  tone: string
}

const sceneOptions: SceneOption[] = [
  {
    value: 'meal',
    label: '饭搭子',
    detail: '从一顿饭开始',
    icon: UtensilsCrossed,
    tone: 'bg-[#edf3e9] text-[#4e6f43]',
  },
  {
    value: 'study',
    label: '自习',
    detail: '共同节奏与空闲',
    icon: BookOpen,
    tone: 'bg-[#e8f0f4] text-[#3c6478]',
  },
  {
    value: 'sport',
    label: '运动',
    detail: '项目和强度合拍',
    icon: Dumbbell,
    tone: 'bg-[#f5eee3] text-[#765d38]',
  },
  {
    value: 'romantic',
    label: '认真关系',
    detail: '只在你主动选择后开启',
    icon: Heart,
    tone: 'bg-[#f5e9ed] text-[#8b5063]',
  },
  {
    value: 'project',
    label: '项目队友',
    detail: '技能与角色互补',
    icon: UsersRound,
    tone: 'bg-[#eee9f2] text-[#644c75]',
  },
]

const timezoneOptions = [
  { value: 'Asia/Shanghai', label: '北京时间' },
  { value: 'Asia/Hong_Kong', label: '香港时间' },
  { value: 'Asia/Singapore', label: '新加坡时间' },
  { value: 'America/New_York', label: '纽约时间' },
]

export function OnboardingDialog() {
  const {
    preferences,
    loading,
    completeOnboarding,
    requestBrowserPermission,
    updateNotificationPreferences,
  } = useEngagement()
  const detectedTimezone = useMemo(
    () => Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Shanghai',
    [],
  )
  const [step, setStep] = useState<1 | 2>(1)
  const [selectedScenes, setSelectedScenes] = useState<EngagementScene[]>([])
  const [notificationsEnabled, setNotificationsEnabled] = useState(true)
  const [newMessages, setNewMessages] = useState(true)
  const [meetupInvites, setMeetupInvites] = useState(true)
  const [meetupReminders, setMeetupReminders] = useState(true)
  const [dailyPicks, setDailyPicks] = useState(true)
  const [systemNotifications, setSystemNotifications] = useState(false)
  const [timezone, setTimezone] = useState(detectedTimezone)
  const [saving, setSaving] = useState(false)
  const [notice, setNotice] = useState('')
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    if (!preferences) return
    setNotificationsEnabled(preferences.notificationsEnabled)
    setNewMessages(preferences.newMessages)
    setMeetupInvites(preferences.meetupInvites)
    setMeetupReminders(preferences.meetupReminders)
    setDailyPicks(preferences.dailyPicks)
    setTimezone(preferences.timezone || detectedTimezone)
    setSystemNotifications(preferences.permissionState === 'granted')
  }, [detectedTimezone, preferences])

  if (loading || !preferences || preferences.onboardingCompleted || dismissed) {
    return null
  }

  const toggleScene = (scene: EngagementScene) => {
    setSelectedScenes((current) =>
      current.includes(scene)
        ? current.filter((item) => item !== scene)
        : [...current, scene],
    )
  }

  const finish = async () => {
    if (saving || selectedScenes.length === 0) return
    setSaving(true)
    setNotice('')
    try {
      let permissionState: NotificationPermissionState = preferences.permissionState
      if (notificationsEnabled && systemNotifications) {
        permissionState = await requestBrowserPermission()
      }
      await updateNotificationPreferences({
        notificationsEnabled,
        newMessages,
        meetupInvites,
        meetupReminders,
        dailyPicks,
        permissionState,
        timezone,
      })
      await completeOnboarding(selectedScenes)
    } catch (error) {
      setNotice(getErrorMessage(error, '设置暂时无法保存'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[90] flex items-end justify-center bg-black/25 p-0 backdrop-blur-[2px] sm:items-center sm:p-5">
      <section
        aria-labelledby="soulink-onboarding-title"
        aria-modal="true"
        className="max-h-[92dvh] w-full overflow-y-auto rounded-t-lg border border-[#d9d5cf] bg-[#fbfaf8] shadow-[0_24px_80px_rgba(35,33,31,0.24)] sm:max-w-[620px] sm:rounded-lg"
        role="dialog"
      >
        <header className="sticky top-0 z-10 flex items-center justify-between border-b border-[#e2dfda] bg-[rgba(251,250,248,0.96)] px-5 py-4 backdrop-blur-lg sm:px-6">
          <div>
            <p className="text-[11px] font-semibold text-[#527164]">
              {step === 1 ? '1 / 2 · 连接场景' : '2 / 2 · 提醒偏好'}
            </p>
            <h1
              className="mt-0.5 text-lg font-semibold text-[#292724]"
              id="soulink-onboarding-title"
            >
              {step === 1 ? '你最近更想认识谁？' : '哪些事情值得提醒你？'}
            </h1>
          </div>
          <button
            aria-label="稍后设置"
            className="inline-flex h-10 w-10 items-center justify-center rounded-md text-[#817d76] hover:bg-[#efede9]"
            onClick={() => setDismissed(true)}
            title="稍后设置"
            type="button"
          >
            <X size={18} />
          </button>
        </header>

        {step === 1 ? (
          <div className="px-5 py-5 sm:px-6 sm:py-6">
            <div className="grid gap-2.5 sm:grid-cols-2">
              {sceneOptions.map((option) => {
                const Icon = option.icon
                const selected = selectedScenes.includes(option.value)
                return (
                  <button
                    aria-pressed={selected}
                    className={cn(
                      'flex min-h-18 items-center gap-3 rounded-md border px-3.5 py-3 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#527b6d]',
                      selected
                        ? 'border-[#527b6d] bg-[#eef4f1]'
                        : 'border-[#dedad4] bg-white hover:border-[#aebdb7]',
                    )}
                    key={option.value}
                    onClick={() => toggleScene(option.value)}
                    type="button"
                  >
                    <span
                      className={cn(
                        'inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md',
                        option.tone,
                      )}
                    >
                      <Icon size={18} />
                    </span>
                    <span className="min-w-0">
                      <span className="block text-sm font-semibold text-[#3f3b37]">
                        {option.label}
                      </span>
                      <span className="mt-0.5 block text-xs text-[#817d76]">
                        {option.detail}
                      </span>
                    </span>
                  </button>
                )
              })}
            </div>
            <div className="mt-5 flex items-start gap-2 border-t border-[#e5e2dd] pt-4 text-xs leading-5 text-[#77736c]">
              <ShieldCheck className="mt-0.5 shrink-0 text-[#527164]" size={15} />
              选择只影响你的推荐与校园主题；认真关系不会从浏览行为中推断。
            </div>
          </div>
        ) : (
          <div className="px-5 py-5 sm:px-6">
            <div className="divide-y divide-[#e8e5e0]">
              <EngagementToggle
                checked={notificationsEnabled}
                description="站内通知中心会保留你选择的提醒。"
                label="连接提醒"
                onChange={(checked) => {
                  setNotificationsEnabled(checked)
                  if (!checked) setSystemNotifications(false)
                }}
              />
              <EngagementToggle
                checked={systemNotifications}
                description="仅显示简短提示，不在锁屏展示聊天正文、时间或地点。"
                disabled={!notificationsEnabled}
                label="浏览器系统通知"
                onChange={setSystemNotifications}
              />
            </div>

            <div className="mt-4 grid gap-2 sm:grid-cols-2">
              {[
                ['新消息', newMessages, setNewMessages],
                ['约见邀请', meetupInvites, setMeetupInvites],
                ['临近提醒', meetupReminders, setMeetupReminders],
                ['每日精选', dailyPicks, setDailyPicks],
              ].map(([label, checked, setter]) => (
                <button
                  aria-pressed={checked as boolean}
                  className={cn(
                    'flex min-h-11 items-center justify-between rounded-md border px-3 text-sm transition-colors disabled:opacity-50',
                    checked
                      ? 'border-[#b9cbc4] bg-[#eef4f1] text-[#315c4d]'
                      : 'border-[#dedad4] bg-white text-[#77736c]',
                  )}
                  disabled={!notificationsEnabled}
                  key={label as string}
                  onClick={() => (setter as (value: boolean) => void)(!(checked as boolean))}
                  type="button"
                >
                  {label as string}
                  <Bell size={14} />
                </button>
              ))}
            </div>

            <label className="mt-5 block text-xs font-semibold text-[#5f5b55]">
              提醒时区
              <select
                className="mt-2 h-11 w-full rounded-md border border-[#d8d4cd] bg-white px-3 text-sm font-normal text-[#45413d] outline-none focus:border-[#527b6d]"
                onChange={(event) => setTimezone(event.target.value)}
                value={timezone}
              >
                {!timezoneOptions.some((option) => option.value === timezone) ? (
                  <option value={timezone}>{timezone}</option>
                ) : null}
                {timezoneOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>

            {notice ? <p className="mt-4 text-sm text-[#8c463d]">{notice}</p> : null}
          </div>
        )}

        <footer className="sticky bottom-0 flex items-center justify-between gap-3 border-t border-[#e2dfda] bg-[rgba(251,250,248,0.96)] px-5 py-4 backdrop-blur-lg sm:px-6">
          {step === 2 ? (
            <button
              className="inline-flex h-11 items-center gap-2 rounded-md px-3 text-sm font-semibold text-[#68635d] hover:bg-[#efede9]"
              onClick={() => setStep(1)}
              type="button"
            >
              <ArrowLeft size={16} />
              返回
            </button>
          ) : (
            <span className="text-xs text-[#918d86]">可多选</span>
          )}
          {step === 1 ? (
            <button
              className="inline-flex h-11 items-center gap-2 rounded-md bg-[#263b35] px-5 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:bg-[#bbb7b0]"
              disabled={selectedScenes.length === 0}
              onClick={() => setStep(2)}
              type="button"
            >
              下一步
              <ArrowRight size={16} />
            </button>
          ) : (
            <button
              className="inline-flex h-11 items-center gap-2 rounded-md bg-[#263b35] px-5 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:bg-[#bbb7b0]"
              disabled={saving}
              onClick={() => void finish()}
              type="button"
            >
              {saving ? <LoaderCircle className="animate-spin" size={16} /> : null}
              {saving ? '保存中...' : '进入 SouLink'}
            </button>
          )}
        </footer>
      </section>
    </div>
  )
}
