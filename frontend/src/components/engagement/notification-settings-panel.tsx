import {
  Bell,
  Check,
  Download,
  LoaderCircle,
  Save,
  ShieldCheck,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { EngagementToggle } from '@/components/engagement/engagement-toggle'
import { useEngagement } from '@/contexts/engagement-context'
import { getErrorMessage } from '@/lib/api'
import type { NotificationPermissionState } from '@/lib/engagement-api'
import { cn } from '@/lib/utils'

const permissionLabels: Record<NotificationPermissionState, string> = {
  default: '尚未询问',
  granted: '浏览器已允许',
  denied: '浏览器已拒绝',
  unavailable: '当前浏览器不支持',
}

const timezoneOptions = [
  { value: 'Asia/Shanghai', label: '北京时间' },
  { value: 'Asia/Hong_Kong', label: '香港时间' },
  { value: 'Asia/Singapore', label: '新加坡时间' },
  { value: 'America/New_York', label: '纽约时间' },
]

export function NotificationSettingsPanel() {
  const {
    preferences,
    themes,
    installAvailable,
    installApp,
    requestBrowserPermission,
    updateNotificationPreferences,
    updateSelectedThemes,
  } = useEngagement()
  const [notificationsEnabled, setNotificationsEnabled] = useState(false)
  const [newMessages, setNewMessages] = useState(true)
  const [meetupInvites, setMeetupInvites] = useState(true)
  const [meetupReminders, setMeetupReminders] = useState(true)
  const [dailyPicks, setDailyPicks] = useState(true)
  const [timezone, setTimezone] = useState('Asia/Shanghai')
  const [selectedThemeKeys, setSelectedThemeKeys] = useState<string[]>([])
  const [permissionState, setPermissionState] =
    useState<NotificationPermissionState>('default')
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    if (!preferences) return
    setNotificationsEnabled(preferences.notificationsEnabled)
    setNewMessages(preferences.newMessages)
    setMeetupInvites(preferences.meetupInvites)
    setMeetupReminders(preferences.meetupReminders)
    setDailyPicks(preferences.dailyPicks)
    setTimezone(preferences.timezone)
    setPermissionState(preferences.permissionState)
  }, [preferences])

  useEffect(() => {
    setSelectedThemeKeys(themes.filter((theme) => theme.selected).map((theme) => theme.key))
  }, [themes])

  if (!preferences) return null

  const save = async () => {
    if (saving) return
    setSaving(true)
    setSaved(false)
    setNotice('')
    try {
      await Promise.all([
        updateNotificationPreferences({
          notificationsEnabled,
          newMessages,
          meetupInvites,
          meetupReminders,
          dailyPicks,
          permissionState,
          timezone,
        }),
        updateSelectedThemes(selectedThemeKeys),
      ])
      setSaved(true)
      window.setTimeout(() => setSaved(false), 1800)
    } catch (error) {
      setNotice(getErrorMessage(error, '提醒设置暂时无法保存'))
    } finally {
      setSaving(false)
    }
  }

  const askPermission = async () => {
    setNotice('')
    try {
      const nextState = await requestBrowserPermission()
      setPermissionState(nextState)
      if (nextState === 'denied') {
        setNotice('浏览器已拒绝系统通知，站内通知中心仍可正常使用。')
      }
    } catch (error) {
      setNotice(getErrorMessage(error, '浏览器通知授权暂时无法更新'))
    }
  }

  const toggleTheme = (key: string) => {
    setSelectedThemeKeys((current) =>
      current.includes(key) ? current.filter((item) => item !== key) : [...current, key],
    )
  }

  return (
    <section
      className="rounded-lg border border-[#dfdcd6] bg-white"
      id="notification-settings"
    >
      <header className="flex flex-col gap-3 border-b border-[#e5e2dd] px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <span className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-[#e8f0f4] text-[#416879]">
            <Bell size={17} />
          </span>
          <div>
            <h2 className="text-sm font-semibold text-[#37332f]">提醒与校园主题</h2>
            <p className="mt-0.5 text-xs text-[#77736c]">
              {permissionLabels[permissionState]}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {installAvailable ? (
            <button
              className="inline-flex h-9 items-center gap-2 rounded-md border border-[#d8d4cd] bg-white px-3 text-xs font-semibold text-[#5f5b55] hover:bg-[#f2f0ec]"
              onClick={() => void installApp()}
              type="button"
            >
              <Download size={14} />
              安装 SouLink
            </button>
          ) : null}
          <button
            className="inline-flex h-9 items-center gap-2 rounded-md bg-[#315c4d] px-3 text-xs font-semibold text-white hover:bg-[#284f42] disabled:opacity-50"
            disabled={saving}
            onClick={() => void save()}
            type="button"
          >
            {saving ? (
              <LoaderCircle className="animate-spin" size={14} />
            ) : saved ? (
              <Check size={14} />
            ) : (
              <Save size={14} />
            )}
            {saving ? '保存中...' : saved ? '已保存' : '保存设置'}
          </button>
        </div>
      </header>

      <div className="grid gap-0 lg:grid-cols-2">
        <div className="px-5 py-4 lg:border-r lg:border-[#e5e2dd]">
          <EngagementToggle
            checked={notificationsEnabled}
            description="关闭后不再生成新的消息、约见和精选提醒。"
            label="连接提醒"
            onChange={setNotificationsEnabled}
          />
          <div className="mt-2 grid gap-2 sm:grid-cols-2">
            {[
              ['新消息', newMessages, setNewMessages],
              ['约见邀请', meetupInvites, setMeetupInvites],
              ['临近提醒', meetupReminders, setMeetupReminders],
              ['每日精选', dailyPicks, setDailyPicks],
            ].map(([label, checked, setter]) => (
              <button
                aria-pressed={checked as boolean}
                className={cn(
                  'flex min-h-11 items-center justify-between rounded-md border px-3 text-xs font-semibold disabled:opacity-45',
                  checked
                    ? 'border-[#b9cbc4] bg-[#eef4f1] text-[#315c4d]'
                    : 'border-[#dedad4] bg-[#faf9f7] text-[#817d76]',
                )}
                disabled={!notificationsEnabled}
                key={label as string}
                onClick={() => (setter as (value: boolean) => void)(!(checked as boolean))}
                type="button"
              >
                {label as string}
                {checked ? <Check size={13} /> : null}
              </button>
            ))}
          </div>

          <div className="mt-4 flex flex-col gap-3 border-t border-[#ebe8e3] pt-4 sm:flex-row sm:items-end">
            <label className="min-w-0 flex-1 text-xs font-semibold text-[#5f5b55]">
              提醒时区
              <select
                className="mt-2 h-10 w-full rounded-md border border-[#d8d4cd] bg-white px-3 text-xs font-normal outline-none focus:border-[#527b6d]"
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
            {permissionState !== 'granted' ? (
              <button
                className="inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-md border border-[#cfd7d3] px-3 text-xs font-semibold text-[#315c4d] hover:bg-[#eef4f1]"
                onClick={() => void askPermission()}
                type="button"
              >
                <ShieldCheck size={14} />
                允许系统通知
              </button>
            ) : null}
          </div>
        </div>

        <div className="border-t border-[#e5e2dd] px-5 py-4 lg:border-t-0">
          <h3 className="text-xs font-semibold text-[#5f5b55]">关注的校园主题</h3>
          <p className="mt-1 text-xs leading-5 text-[#817d76]">
            首页优先呈现你保留的快捷场景，不改变有限推荐原则。
          </p>
          <div className="mt-3 grid gap-2 sm:grid-cols-2">
            {themes.map((theme) => {
              const selected = selectedThemeKeys.includes(theme.key)
              return (
                <button
                  aria-pressed={selected}
                  className={cn(
                    'min-h-14 rounded-md border px-3 py-2 text-left transition-colors',
                    selected
                      ? 'border-[#b9cbc4] bg-[#eef4f1]'
                      : 'border-[#dedad4] bg-[#faf9f7] hover:bg-white',
                  )}
                  key={theme.key}
                  onClick={() => toggleTheme(theme.key)}
                  type="button"
                >
                  <span className="flex items-center justify-between gap-2 text-xs font-semibold text-[#45413d]">
                    {theme.title}
                    {selected ? <Check className="text-[#315c4d]" size={13} /> : null}
                  </span>
                  <span className="mt-1 block text-[10px] text-[#817d76]">
                    {theme.scheduleHint}
                  </span>
                </button>
              )
            })}
          </div>
        </div>
      </div>

      {notice ? (
        <p className="border-t border-[#ead4ce] bg-[#fff6f3] px-5 py-3 text-xs text-[#8c463d]">
          {notice}
        </p>
      ) : null}
    </section>
  )
}
