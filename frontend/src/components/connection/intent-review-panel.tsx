import {
  Check,
  ChevronDown,
  Clock3,
  Eye,
  EyeOff,
  LoaderCircle,
  LockKeyhole,
  MapPin,
  PencilLine,
  Save,
  Send,
  ShieldCheck,
} from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'

import {
  IntentCategoryIcon,
  intentCategoryMeta,
  intentCategoryOrder,
} from '@/components/connection/intent-category'
import type {
  IntentInput,
  IntentVisibility,
  MatchingProfile,
  ProjectMemo,
} from '@/lib/connection-api'
import { cn } from '@/lib/utils'

const visibilityOptions: Array<{
  value: IntentVisibility
  label: string
  description: string
  icon: typeof Eye
}> = [
  {
    value: 'private',
    label: '仅 AI 搜索',
    description: '不进入公开流',
    icon: EyeOff,
  },
  {
    value: 'candidates',
    label: '仅候选可见',
    description: '只向推荐的人展示',
    icon: LockKeyhole,
  },
  {
    value: 'feed',
    label: '进入意图流',
    description: '校内同学可发现',
    icon: Eye,
  },
]

const fieldClassName =
  'mt-1.5 w-full rounded-md border border-[#d8d5ce] bg-white px-3 text-sm text-[#2b2926] outline-none transition-colors placeholder:text-[#9b968e] focus:border-[#55786c] focus:ring-3 focus:ring-[#55786c]/10'

function emptyProjectMemo(): ProjectMemo {
  return {
    building: '',
    stage: '',
    weeklyTime: '',
    compensation: '',
    requiredSkills: [],
    niceToHaveSkills: [],
    workingStyle: '',
    riskAppetite: '',
  }
}

function splitTags(value: string) {
  return value
    .split(/[,，、\n]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

interface IntentReviewPanelProps {
  draft: IntentInput
  profile: MatchingProfile | null
  submitting: 'draft' | 'active' | null
  onChange: <K extends keyof IntentInput>(
    key: K,
    value: IntentInput[K],
  ) => void
  onProjectChange: <K extends keyof ProjectMemo>(
    key: K,
    value: ProjectMemo[K],
  ) => void
  onSubmit: (status: 'draft' | 'active') => void
  className?: string
}

export function IntentReviewPanel({
  draft,
  profile,
  submitting,
  onChange,
  onProjectChange,
  onSubmit,
  className,
}: IntentReviewPanelProps) {
  const [detailsOpen, setDetailsOpen] = useState(false)
  const meta = intentCategoryMeta[draft.category]
  const romanticNeedsOptIn =
    draft.category === 'romantic' && profile && !profile.romanticOptIn
  const canSubmit =
    Boolean(draft.title.trim() && draft.description.trim()) &&
    !romanticNeedsOptIn

  const changeCategory = (category: IntentInput['category']) => {
    onChange('category', category)
    if (category === 'romantic') {
      onChange('visibility', 'candidates')
    }
    onChange(
      'projectMemo',
      category === 'project' || category === 'cofounder'
        ? draft.projectMemo ?? emptyProjectMemo()
        : null,
    )
  }

  return (
    <div className={cn('flex h-full min-h-0 flex-col bg-[#fbfaf8]', className)}>
      <header className="flex h-16 shrink-0 items-center justify-between border-b border-[#dfdcd6] px-5">
        <div>
          <p className="text-[11px] font-semibold text-[#817d76]">意图草稿</p>
          <p className="mt-0.5 text-sm font-semibold text-[#2e2b28]">
            发布前确认
          </p>
        </div>
        <span className="inline-flex items-center gap-1.5 text-[11px] font-medium text-[#477062]">
          <span className="h-1.5 w-1.5 rounded-full bg-[#4f846f]" />
          实时更新
        </span>
      </header>

      <div className="min-h-0 flex-1 overflow-y-auto">
        <section className="border-b border-[#e2dfd9] px-5 py-5">
          <div className="flex items-start gap-3">
            <IntentCategoryIcon category={draft.category} size={17} />
            <div className="min-w-0 flex-1">
              <span className="text-xs font-semibold text-[#5d5364]">
                {meta.label}
              </span>
              <h2 className="mt-1 text-[17px] font-semibold leading-6 text-[#282623]">
                {draft.title || '尚未命名的意图'}
              </h2>
            </div>
          </div>

          <p className="mt-4 text-sm leading-6 text-[#65615b]">
            {draft.description || '继续对话后，这里会形成一段清晰的需求描述。'}
          </p>

          {(draft.timeWindow || draft.location) && (
            <div className="mt-4 flex flex-wrap gap-x-4 gap-y-2 text-xs text-[#77736c]">
              {draft.timeWindow ? (
                <span className="inline-flex items-center gap-1.5">
                  <Clock3 size={14} />
                  {draft.timeWindow}
                </span>
              ) : null}
              {draft.location ? (
                <span className="inline-flex items-center gap-1.5">
                  <MapPin size={14} />
                  {draft.location}
                </span>
              ) : null}
            </div>
          )}

          {draft.constraints.length > 0 ? (
            <div className="mt-4 flex flex-wrap gap-1.5">
              {draft.constraints.slice(0, 5).map((constraint) => (
                <span
                  className="rounded-md bg-[#efede9] px-2 py-1 text-[11px] text-[#65615b]"
                  key={constraint}
                >
                  {constraint}
                </span>
              ))}
            </div>
          ) : null}
        </section>

        <section className="border-b border-[#e2dfd9] px-5 py-4">
          <div className="space-y-2 text-xs text-[#66625c]">
            <p className="flex items-start gap-2">
              <Check className="mt-0.5 shrink-0 text-[#477062]" size={14} />
              <span>
                目标：{draft.goal || '认识一个适合继续交流的人'}
              </span>
            </p>
            <p className="flex items-start gap-2">
              <Check className="mt-0.5 shrink-0 text-[#477062]" size={14} />
              <span>
                下一步：{draft.desiredNextStep || '先进行一次低压力沟通'}
              </span>
            </p>
          </div>
        </section>

        <section>
          <button
            aria-expanded={detailsOpen}
            className="flex w-full items-center justify-between px-5 py-4 text-left text-sm font-semibold text-[#3b3834] hover:bg-[#f5f3ef]"
            onClick={() => setDetailsOpen((open) => !open)}
            type="button"
          >
            <span className="inline-flex items-center gap-2">
              <PencilLine size={15} />
              编辑完整信息
            </span>
            <ChevronDown
              className={cn(
                'text-[#817d76] transition-transform',
                detailsOpen && 'rotate-180',
              )}
              size={17}
            />
          </button>

          {detailsOpen ? (
            <div className="space-y-5 border-t border-[#e2dfd9] px-5 py-5">
              <label className="block text-xs font-semibold text-[#5f5b55]">
                标题
                <input
                  className={cn(fieldClassName, 'h-10')}
                  maxLength={80}
                  onChange={(event) => onChange('title', event.target.value)}
                  value={draft.title}
                />
              </label>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                具体描述
                <textarea
                  className={cn(fieldClassName, 'min-h-28 resize-y py-2.5 leading-6')}
                  maxLength={1200}
                  onChange={(event) =>
                    onChange('description', event.target.value)
                  }
                  value={draft.description}
                />
              </label>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                连接目标
                <input
                  className={cn(fieldClassName, 'h-10')}
                  maxLength={200}
                  onChange={(event) => onChange('goal', event.target.value)}
                  value={draft.goal}
                />
              </label>

              <div>
                <span className="text-xs font-semibold text-[#5f5b55]">
                  意图类型
                </span>
                <div className="mt-2 grid grid-cols-3 gap-2">
                  {intentCategoryOrder.map((category) => {
                    const categoryMeta = intentCategoryMeta[category]
                    const Icon = categoryMeta.icon
                    return (
                      <button
                        className={cn(
                          'flex min-h-14 flex-col items-center justify-center gap-1 rounded-md border px-1.5 py-2 text-[11px] font-medium transition-colors',
                          draft.category === category
                            ? 'border-[#68887d] bg-[#eaf2ee] text-[#315c4d]'
                            : 'border-[#dedad4] bg-white text-[#6f6b65] hover:border-[#aebdb7]',
                        )}
                        key={category}
                        onClick={() => changeCategory(category)}
                        type="button"
                      >
                        <Icon className={categoryMeta.iconClassName} size={15} />
                        {categoryMeta.shortLabel}
                      </button>
                    )
                  })}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  时间窗口
                  <input
                    className={cn(fieldClassName, 'h-10')}
                    onChange={(event) =>
                      onChange('timeWindow', event.target.value)
                    }
                    placeholder="今晚 / 本周"
                    value={draft.timeWindow}
                  />
                </label>
                <label className="block text-xs font-semibold text-[#5f5b55]">
                  校内地点
                  <input
                    className={cn(fieldClassName, 'h-10')}
                    onChange={(event) =>
                      onChange('location', event.target.value)
                    }
                    placeholder="紫荆 / 图书馆"
                    value={draft.location}
                  />
                </label>
              </div>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                重要偏好
                <input
                  className={cn(fieldClassName, 'h-10')}
                  onChange={(event) =>
                    onChange('constraints', splitTags(event.target.value))
                  }
                  placeholder="用逗号分隔"
                  value={draft.constraints.join('，')}
                />
              </label>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                明确不合适的情况
                <input
                  className={cn(fieldClassName, 'h-10')}
                  onChange={(event) =>
                    onChange('dealBreakers', splitTags(event.target.value))
                  }
                  placeholder="可留空"
                  value={draft.dealBreakers.join('，')}
                />
              </label>

              <label className="block text-xs font-semibold text-[#5f5b55]">
                希望的下一步
                <input
                  className={cn(fieldClassName, 'h-10')}
                  maxLength={200}
                  onChange={(event) =>
                    onChange('desiredNextStep', event.target.value)
                  }
                  value={draft.desiredNextStep}
                />
              </label>

              {(draft.category === 'project' ||
                draft.category === 'cofounder') &&
              draft.projectMemo ? (
                <div className="space-y-4 border-y border-[#dedbd5] py-5">
                  <div>
                    <h3 className="text-sm font-semibold text-[#3f3b37]">
                      项目 memo
                    </h3>
                    <p className="mt-1 text-[11px] text-[#817d76]">
                      用于判断能力、投入和协作方式是否匹配
                    </p>
                  </div>

                  <label className="block text-xs font-semibold text-[#5f5b55]">
                    正在做什么
                    <textarea
                      className={cn(fieldClassName, 'min-h-20 resize-y py-2.5')}
                      onChange={(event) =>
                        onProjectChange('building', event.target.value)
                      }
                      value={draft.projectMemo.building}
                    />
                  </label>

                  <div className="grid grid-cols-2 gap-3">
                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      当前阶段
                      <input
                        className={cn(fieldClassName, 'h-10')}
                        onChange={(event) =>
                          onProjectChange('stage', event.target.value)
                        }
                        value={draft.projectMemo.stage}
                      />
                    </label>
                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      每周投入
                      <input
                        className={cn(fieldClassName, 'h-10')}
                        onChange={(event) =>
                          onProjectChange('weeklyTime', event.target.value)
                        }
                        value={draft.projectMemo.weeklyTime}
                      />
                    </label>
                  </div>

                  <label className="block text-xs font-semibold text-[#5f5b55]">
                    权益或补偿预期
                    <input
                      className={cn(fieldClassName, 'h-10')}
                      onChange={(event) =>
                        onProjectChange('compensation', event.target.value)
                      }
                      value={draft.projectMemo.compensation}
                    />
                  </label>

                  <div className="grid grid-cols-2 gap-3">
                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      必备技能
                      <input
                        className={cn(fieldClassName, 'h-10')}
                        onChange={(event) =>
                          onProjectChange(
                            'requiredSkills',
                            splitTags(event.target.value),
                          )
                        }
                        value={draft.projectMemo.requiredSkills.join('，')}
                      />
                    </label>
                    <label className="block text-xs font-semibold text-[#5f5b55]">
                      加分技能
                      <input
                        className={cn(fieldClassName, 'h-10')}
                        onChange={(event) =>
                          onProjectChange(
                            'niceToHaveSkills',
                            splitTags(event.target.value),
                          )
                        }
                        value={draft.projectMemo.niceToHaveSkills.join('，')}
                      />
                    </label>
                  </div>

                  <label className="block text-xs font-semibold text-[#5f5b55]">
                    协作方式
                    <textarea
                      className={cn(fieldClassName, 'min-h-20 resize-y py-2.5')}
                      onChange={(event) =>
                        onProjectChange('workingStyle', event.target.value)
                      }
                      value={draft.projectMemo.workingStyle}
                    />
                  </label>

                  <label className="block text-xs font-semibold text-[#5f5b55]">
                    风险偏好
                    <textarea
                      className={cn(fieldClassName, 'min-h-20 resize-y py-2.5')}
                      onChange={(event) =>
                        onProjectChange('riskAppetite', event.target.value)
                      }
                      value={draft.projectMemo.riskAppetite}
                    />
                  </label>
                </div>
              ) : null}

              <div>
                <span className="text-xs font-semibold text-[#5f5b55]">
                  谁可以看见
                </span>
                <div className="mt-2 space-y-2">
                  {visibilityOptions.map((option) => {
                    const Icon = option.icon
                    const disabled =
                      draft.category === 'romantic' &&
                      option.value !== 'candidates'
                    return (
                      <button
                        className={cn(
                          'flex w-full items-center gap-3 rounded-md border px-3 py-2.5 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-45',
                          draft.visibility === option.value
                            ? 'border-[#68887d] bg-[#eaf2ee]'
                            : 'border-[#dedad4] bg-white hover:border-[#aebdb7]',
                        )}
                        disabled={disabled}
                        key={option.value}
                        onClick={() => onChange('visibility', option.value)}
                        type="button"
                      >
                        <Icon className="shrink-0 text-[#5f6f69]" size={15} />
                        <span className="min-w-0">
                          <span className="block text-xs font-semibold text-[#45413d]">
                            {option.label}
                          </span>
                          <span className="mt-0.5 block text-[11px] text-[#817d76]">
                            {option.description}
                          </span>
                        </span>
                      </button>
                    )
                  })}
                </div>
              </div>
            </div>
          ) : null}
        </section>

        {romanticNeedsOptIn ? (
          <section className="mx-5 my-4 rounded-md border border-[#e5cad3] bg-[#fff5f8] px-4 py-3">
            <p className="text-sm font-medium text-[#7e3d55]">
              认真关系匹配尚未开启
            </p>
            <p className="mt-1 text-xs leading-5 text-[#8e6473]">
              这是一个单独的私密选择，不会显示在公开意图流。
            </p>
            <Link
              className="mt-2 inline-flex text-xs font-semibold text-[#7e3d55]"
              to="/profile"
            >
              前往 AI 档案开启
            </Link>
          </section>
        ) : null}

        <section className="mx-5 my-4 flex items-start gap-2 text-[11px] leading-5 text-[#77736c]">
          <ShieldCheck className="mt-0.5 shrink-0 text-[#477062]" size={14} />
          只有双方都表达兴趣后，才会创建连接房间。
        </section>
      </div>

      <footer className="shrink-0 border-t border-[#dfdcd6] bg-[#fbfaf8] p-4">
        <button
          className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-md bg-[#263b35] px-4 text-sm font-semibold text-white transition-colors hover:bg-[#1d302a] disabled:cursor-not-allowed disabled:opacity-45"
          disabled={!canSubmit || submitting !== null}
          onClick={() => onSubmit('active')}
          type="button"
        >
          {submitting === 'active' ? (
            <LoaderCircle className="animate-spin" size={16} />
          ) : (
            <Send size={16} />
          )}
          发布并开始匹配
        </button>
        <button
          className="mt-2 inline-flex h-9 w-full items-center justify-center gap-2 rounded-md text-xs font-semibold text-[#625e58] hover:bg-[#efede9] disabled:opacity-45"
          disabled={!canSubmit || submitting !== null}
          onClick={() => onSubmit('draft')}
          type="button"
        >
          {submitting === 'draft' ? (
            <LoaderCircle className="animate-spin" size={14} />
          ) : (
            <Save size={14} />
          )}
          保存草稿
        </button>
      </footer>
    </div>
  )
}
