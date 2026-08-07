import { CalendarDays, Eye, Sparkles } from 'lucide-react'

import type { ProfilePrivacy } from '@/lib/profile-api'
import { cn } from '@/lib/utils'

interface ProfilePrivacyPanelProps {
  value: ProfilePrivacy
  onChange: (value: ProfilePrivacy) => void
}

interface PrivacySwitchProps {
  checked: boolean
  disabled?: boolean
  label: string
  onChange: (checked: boolean) => void
}

function PrivacySwitch({ checked, disabled, label, onChange }: PrivacySwitchProps) {
  return (
    <button
      aria-checked={checked}
      aria-label={label}
      className={cn(
        'relative h-6 w-11 shrink-0 rounded-full border transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#527b6d]',
        checked ? 'border-[#315f50] bg-[#315f50]' : 'border-[#c9c5bf] bg-[#dedbd5]',
        disabled && 'cursor-not-allowed opacity-45',
      )}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      role="switch"
      type="button"
    >
      <span
        className={cn(
          'absolute top-0.5 h-4.5 w-4.5 rounded-full bg-white shadow-sm transition-transform',
          checked ? 'translate-x-5' : 'translate-x-0.5',
        )}
      />
    </button>
  )
}

export function ProfilePrivacyPanel({ value, onChange }: ProfilePrivacyPanelProps) {
  const rows = [
    {
      key: 'profilePublic' as const,
      icon: Eye,
      title: '公开个人主页',
      description: '允许已登录的清华用户通过名片与主页链接查看基本资料。',
      disabled: false,
    },
    {
      key: 'schedulePublic' as const,
      icon: CalendarDays,
      title: '公开时间表',
      description: '主页显示本周忙闲；私人日程仍只显示为“忙碌”。',
      disabled: !value.profilePublic,
    },
    {
      key: 'preferencesPublic' as const,
      icon: Sparkles,
      title: '公开个人偏好',
      description: '主页显示自我介绍、技能、兴趣与常用协作方式。',
      disabled: !value.profilePublic,
    },
  ]

  return (
    <section className="overflow-hidden rounded-lg border border-[#dfd9d2] bg-white">
      <div className="border-b border-[#e8e3dd] px-6 py-5">
        <h2 className="text-lg font-semibold text-[#302e2b]">主页与隐私</h2>
        <p className="mt-1 text-sm text-[#79736d]">每一项都由你决定，默认不会公开。</p>
      </div>
      <div className="divide-y divide-[#ebe7e2]">
        {rows.map(({ key, icon: Icon, title, description, disabled }) => (
          <div className="flex items-center gap-4 px-6 py-4" key={key}>
            <span className="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-[#eaf1ed] text-[#376251]">
              <Icon className="h-4 w-4" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="block text-sm font-semibold text-[#393632]">{title}</span>
              <span className="mt-0.5 block text-xs leading-5 text-[#817b74]">{description}</span>
            </span>
            <PrivacySwitch
              checked={value[key]}
              disabled={disabled}
              label={title}
              onChange={(checked) => {
                if (key === 'profilePublic' && !checked) {
                  onChange({
                    profilePublic: false,
                    schedulePublic: false,
                    preferencesPublic: false,
                  })
                  return
                }
                onChange({ ...value, [key]: checked })
              }}
            />
          </div>
        ))}
      </div>
    </section>
  )
}
