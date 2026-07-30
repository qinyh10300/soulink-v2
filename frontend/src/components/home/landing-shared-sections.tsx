import { motion, type Variants } from 'framer-motion'
import { MailOpen, MapPinned, ScrollText } from 'lucide-react'

import { FaqList } from '@/components/home/faq-list'
import { SectionHeading } from '@/components/home/section-heading'
import { Card, CardContent } from '@/components/ui/card'
import { faqItems } from '@/mock/landing'

const sectionEase = [0.22, 1, 0.36, 1] as const

const containerVariant: Variants = {
  hidden: { opacity: 0, y: 16 },
  show: {
    opacity: 1,
    y: 0,
    transition: {
      staggerChildren: 0.08,
      duration: 0.45,
      ease: sectionEase,
    },
  },
}

const itemVariant: Variants = {
  hidden: { opacity: 0, y: 18 },
  show: { opacity: 1, y: 0, transition: { duration: 0.45, ease: sectionEase } },
}

const stepIcons = [ScrollText, MailOpen, MapPinned] as const
const howItWorksSteps = [
  {
    points: [
      '52 道深度问卷，关于真诚、边界、节奏与期待。',
      '我们会根据这些真实的回答，为你寻找更相投的人。',
    ],
    title: '写下你自己',
  },
  {
    points: [
      '每周四晚 20:00 到 21:00，结果会陆续更新。',
      '没有反复滑动，只有一次认真送达，和关于你们相遇的解释。',
    ],
    title: '每周一次，拆开来信',
  },
  {
    points: [
      '我们为两颗愿意靠近的心打开联系，让相遇有机会继续发生。',
      '从教学楼前的暮色，到操场边的晚风，故事交给你们自己开始。',
    ],
    title: '去见见 TA',
  },
] as const

export function LandingSharedSections() {
  return (
    <div className="space-y-0">
      <section className="scroll-mt-28 px-6 py-16 lg:px-8 lg:py-24" id="features">
        <div className="mx-auto w-full max-w-5xl space-y-8">
          <div className="mx-auto max-w-4xl space-y-4 text-center">
            <p
              className="text-[0.92rem] uppercase tracking-[0.28em] text-[#b59d87]"
              style={{ fontFamily: '"Cormorant Garamond", "Times New Roman", serif' }}
            >
              How It Works
            </p>
            <h2
              className="text-[2.6rem] leading-[1.04] tracking-[-0.05em] text-[#231d1a] sm:text-[4.2rem] lg:text-[3.1rem]"
              style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
            >
              三步，开启你的<span className="text-[#a28af4]">故事</span>
            </h2>
          </div>

          <motion.div
            className="grid gap-6 lg:grid-cols-3"
            initial="hidden"
            transition={{ staggerChildren: 0.08 }}
            viewport={{ once: true, amount: 0.2 }}
            whileInView="show"
            variants={containerVariant}
          >
            {howItWorksSteps.map((step, index) => {
              const Icon = stepIcons[index] ?? MapPinned
              const stepNumber = `${index + 1}`.padStart(2, '0')

              return (
                <motion.div key={stepNumber} variants={itemVariant}>
                  <Card className="h-full overflow-hidden rounded-[2rem] border border-white/80 bg-[linear-gradient(180deg,rgba(255,255,255,0.97)_0%,rgba(255,252,248,0.99)_100%)] shadow-[0_18px_44px_rgba(151,123,94,0.06)] transition-all duration-300 hover:-translate-y-1.5 hover:shadow-[0_24px_54px_rgba(151,123,94,0.1)]">
                    <CardContent className="flex h-full flex-col gap-8 p-8 sm:p-10">
                      <div className="flex items-center justify-between gap-4">
                        <span
                          className={`block text-[2.8rem] leading-none tracking-[-0.04em] sm:text-[3.2rem] ${
                            stepNumber === '01'
                              ? 'text-[#efc3cd]'
                              : stepNumber === '02'
                                ? 'text-[#d9caf5]'
                                : 'text-[#ecd5be]'
                          }`}
                          style={{ fontFamily: '"Playfair Display", "Times New Roman", serif' }}
                        >
                          {stepNumber}
                        </span>
                        <div
                          className={`flex h-12 w-12 items-center justify-center rounded-2xl ${
                            stepNumber === '01'
                              ? 'bg-[#f8ebe4] text-[#bf7f6f]'
                              : stepNumber === '02'
                                ? 'bg-[#efe8da] text-[#a07c57]'
                                : 'bg-[#e7eef0] text-[#62808a]'
                          }`}
                        >
                          <Icon size={22} />
                        </div>
                      </div>

                      <div className="space-y-5">
                        <h3
                          className="text-[1.5rem] font-semibold leading-[1.35] tracking-[-0.02em] text-[#231d1a]"
                          style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                        >
                          {step.title}
                        </h3>
                        <div className="space-y-5">
                          {step.points.map((point) => (
                            <div key={point} className="flex items-start gap-4">
                              <span className="mt-3 h-2.5 w-2.5 shrink-0 rounded-full bg-[#8cbcaf]" />
                              <p
                                className="text-[0.96rem] leading-[1.9] text-[#6f6256]"
                                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                              >
                                {point}
                              </p>
                            </div>
                          ))}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </motion.div>
              )
            })}
          </motion.div>
        </div>
      </section>

      <section className="scroll-mt-28 px-6 py-16 lg:px-8 lg:py-24" id="faq">
        <div className="mx-auto w-full max-w-5xl space-y-10">
          <SectionHeading align="center" description="" eyebrow="" title="常见问题" />
          <div className="mx-auto max-w-3xl">
            <FaqList items={faqItems} />
          </div>
        </div>
      </section>
    </div>
  )
}
