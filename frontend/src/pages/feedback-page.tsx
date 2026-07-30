import { type FormEvent, useMemo, useState } from 'react'

import { Star } from 'lucide-react'

import { DashboardFooter } from '@/components/dashboard/dashboard-footer'
import { getErrorMessage } from '@/lib/api'
import { submitFeedback } from '@/lib/feedback-api'
import { getStoredUserProfile } from '@/lib/user-profile-storage'
import { cn } from '@/lib/utils'

type FeedbackCategory = 'match' | 'bug' | 'feature' | 'other'

const feedbackCategories: Array<{
  id: FeedbackCategory
  icon: string
  label: string
}> = [
  { id: 'match', icon: '💞', label: '配对体验' },
  { id: 'bug', icon: '🪲', label: 'Bug 报告' },
  { id: 'feature', icon: '💡', label: '功能建议' },
  { id: 'other', icon: '📨', label: '其他' },
]

export function FeedbackPage() {
  const profile = useMemo(() => getStoredUserProfile(), [])
  const [selectedCategory, setSelectedCategory] =
    useState<FeedbackCategory>('match')
  const [rating, setRating] = useState(0)
  const [details, setDetails] = useState('')
  const [contactEmail, setContactEmail] = useState(profile.email)
  const [notice, setNotice] = useState('')
  const [noticeTone, setNoticeTone] = useState<'error' | 'success'>('success')
  const [isSubmitting, setIsSubmitting] = useState(false)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (rating === 0) {
      setNoticeTone('error')
      setNotice('请先给本次体验打分。')
      return
    }

    setIsSubmitting(true)

    try {
      await submitFeedback({
        category: selectedCategory,
        rating,
        details: details.trim(),
        contactEmail: contactEmail.trim() || undefined,
      })

      setNoticeTone('success')
      setNotice('反馈已提交，感谢你的认真分享。')
      setSelectedCategory('match')
      setRating(0)
      setDetails('')
      setContactEmail(profile.email)
    } catch (error) {
      setNoticeTone('error')
      setNotice(getErrorMessage(error, '反馈提交失败，请稍后重试'))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="min-h-screen bg-[#f7f3ee]">
      <main className="px-6 pb-0 pt-16 lg:px-8 lg:pt-20">
        <div className="mx-auto w-full max-w-6xl">
          <section className="space-y-4 text-center">
            <h1 className="font-serif text-5xl text-[#2f221b] sm:text-6xl">
              反馈与建议
            </h1>
            <p className="text-lg leading-8 text-[#7c6c61] sm:text-[1.35rem]">
              你的每一条反馈，都会让  Soulink 变得更好。
            </p>
          </section>

          <form
            className="mx-auto mt-14 max-w-[46rem] rounded-[38px] border border-[#ece3da] bg-[#fbf8f4] px-8 py-10 shadow-[0_24px_60px_rgba(71,49,35,0.08)] sm:px-14 sm:py-16"
            onSubmit={handleSubmit}
          >
            <div className="space-y-14">
              <section className="space-y-5">
                <div className="text-[1.1rem] font-semibold text-[#5a4a40]">
                  反馈类型
                </div>
                <div className="flex flex-wrap gap-3">
                  {feedbackCategories.map((category) => {
                    const active = selectedCategory === category.id

                    return (
                      <button
                        className={cn(
                          'inline-flex h-13 items-center gap-2.5 rounded-full border px-6 text-[1.02rem] font-semibold transition-all hover:-translate-y-0.5',
                          active
                            ? 'border-[#bf5f5b] bg-[rgba(196,97,90,0.04)] text-[#9f4744] shadow-[inset_0_0_0_1px_rgba(191,95,91,0.06)]'
                            : 'border-[#ece4da] bg-[#f4efe9] text-[#67594f] hover:border-[#dfcfc2]',
                        )}
                        key={category.id}
                        onClick={() => setSelectedCategory(category.id)}
                        type="button"
                      >
                        <span aria-hidden="true" className="text-lg" role="img">
                          {category.icon}
                        </span>
                        <span>{category.label}</span>
                      </button>
                    )
                  })}
                </div>
              </section>

              <section className="space-y-5">
                <div className="text-[1.1rem] font-semibold text-[#5a4a40]">
                  你的满意度 <span className="text-[#b64d4a]">*</span>
                </div>
                <div className="flex items-center gap-2">
                  {Array.from({ length: 5 }, (_, index) => {
                    const value = index + 1
                    const active = value <= rating

                    return (
                      <button
                        aria-label={`评分 ${value} 星`}
                        className="transition-transform hover:scale-110"
                        key={value}
                        onClick={() => setRating(value)}
                        type="button"
                      >
                        <Star
                          className={active ? 'text-[#d9a27f]' : 'text-[#e7cdbd]'}
                          fill={active ? 'currentColor' : 'none'}
                          size={38}
                          strokeWidth={1.6}
                        />
                      </button>
                    )
                  })}
                </div>
              </section>

              <section className="space-y-5">
                <label
                  className="block text-[1.1rem] font-semibold text-[#5a4a40]"
                  htmlFor="feedback-details"
                >
                  欢迎详细说说你的体验、建议，或遇到的问题
                </label>
                <div>
                  <textarea
                    className="min-h-[16rem] w-full rounded-[24px] border border-[#e7ddd3] bg-white px-6 py-5 text-[1.05rem] leading-8 text-[#5b4d45] outline-none transition placeholder:text-[#bbb0a4] focus:border-[#d7c4b7]"
                    id="feedback-details"
                    maxLength={2000}
                    onChange={(event) => setDetails(event.target.value)}
                    placeholder="欢迎详细描述，我们会认真查看。"
                    value={details}
                  />
                  <div className="mt-4 text-right text-[1.02rem] text-[#b6a294]">
                    {details.length}/2000
                  </div>
                </div>
              </section>

              <section className="space-y-5">
                <label
                  className="block text-[1.1rem] font-semibold text-[#5a4a40]"
                  htmlFor="feedback-email"
                >
                  联系邮箱（选填）
                </label>
                <input
                  className="h-17 w-full rounded-[22px] border border-[#e7ddd3] bg-white px-6 text-[1.05rem] text-[#5b4d45] outline-none transition placeholder:text-[#bbb0a4] focus:border-[#d7c4b7]"
                  id="feedback-email"
                  onChange={(event) => setContactEmail(event.target.value)}
                  placeholder="方便我们后续联系你"
                  type="email"
                  value={contactEmail}
                />
              </section>
            </div>

            {notice ? (
              <p
                className={cn(
                  'mt-8 text-sm',
                  noticeTone === 'error' ? 'text-[#b64d4a]' : 'text-[#7b6a5f]',
                )}
              >
                {notice}
              </p>
            ) : null}

            <button
              className="mt-10 inline-flex h-17 w-full items-center justify-center rounded-full bg-[#9d433f] text-[1.5rem] font-semibold text-white transition-colors disabled:cursor-not-allowed disabled:opacity-70 hover:bg-[#8e3d39]"
              disabled={isSubmitting}
              type="submit"
            >
              {isSubmitting ? '提交中...' : '提交反馈'}
            </button>
          </form>
        </div>
      </main>

      <DashboardFooter />
    </div>
  )
}
