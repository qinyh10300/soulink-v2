import { motion } from 'framer-motion'
import { Sparkles } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'

const previewMetrics = [
  { label: '作息节奏接近', value: '91%' },
  { label: '沟通偏好兼容', value: '88%' },
  { label: '关系期待一致', value: '94%' },
]

const icebreakers = ['最近最常去的校内角落', '期末周的解压方式', '最想一起完成的一件小事']
const floatingEase = [0.22, 1, 0.36, 1] as const

export function HeroPreview() {
  return (
    <motion.div
      animate={{ y: [0, -10, 0] }}
      className="relative mx-auto max-w-[34rem]"
      transition={{
        duration: 7,
        ease: floatingEase,
        repeat: Number.POSITIVE_INFINITY,
      }}
    >
      <div className="absolute -left-4 top-6 h-28 w-28 rounded-full bg-[radial-gradient(circle,_rgba(107,143,226,0.28),_transparent_68%)] blur-2xl" />
      <div className="absolute -bottom-10 right-0 h-36 w-36 rounded-full bg-[radial-gradient(circle,_rgba(115,184,166,0.18),_transparent_68%)] blur-3xl" />

      <Card className="relative overflow-hidden border-white/70 bg-white/84 shadow-soft">
        <div className="absolute inset-0 bg-mesh opacity-50" />
        <CardContent className="relative space-y-6 p-6 sm:p-7">
          <div className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-semibold tracking-[0.18em] text-primary/70 uppercase">
                本周匹配视图
              </p>
              <p className="mt-2 text-2xl font-semibold text-foreground">
                解释式结果，而不只是一个分数
              </p>
            </div>
            <div className="rounded-full border border-white/70 bg-white/90 px-3 py-2 text-xs font-medium text-primary shadow-[0_8px_20px_rgba(39,76,155,0.08)]">
              匹配度 92
            </div>
          </div>

          <div className="grid gap-4 lg:grid-cols-[1.15fr_0.85fr]">
            <Card className="border-white/80 bg-white/82">
              <CardContent className="space-y-5 p-5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#dce7ff] text-sm font-semibold text-primary">
                      A
                    </div>
                    <div>
                      <p className="font-medium text-foreground">你</p>
                      <p className="text-sm text-muted-foreground">工科 / 大三 / 夜猫型</p>
                    </div>
                  </div>
                  <div className="h-px flex-1 bg-gradient-to-r from-[#c5d7ff] via-[#96b1f5] to-[#d4eee6]" />
                  <div className="ml-3 flex items-center gap-3">
                    <div>
                      <p className="text-right font-medium text-foreground">候选对象</p>
                      <p className="text-right text-sm text-muted-foreground">
                        文科 / 大三 / 晚睡但稳定
                      </p>
                    </div>
                    <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-[#dff0ea] text-sm font-semibold text-[#2f7465]">
                      M
                    </div>
                  </div>
                </div>

                <div className="grid gap-3 sm:grid-cols-3">
                  {previewMetrics.map((metric) => (
                    <div
                      key={metric.label}
                      className="rounded-2xl border border-white/70 bg-white/80 p-4"
                    >
                      <p className="text-xs text-muted-foreground">{metric.label}</p>
                      <p className="mt-2 text-2xl font-semibold text-foreground">
                        {metric.value}
                      </p>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            <Card className="border-white/80 bg-[#f8fbff]">
              <CardContent className="space-y-4 p-5">
                <div className="flex items-center gap-2 text-primary">
                  <Sparkles size={16} />
                  <p className="text-sm font-medium">系统给出的开场建议</p>
                </div>
                <div className="space-y-3">
                  {icebreakers.map((item) => (
                    <div
                      key={item}
                      className="rounded-2xl border border-white/80 bg-white/84 px-4 py-3 text-sm leading-6 text-foreground"
                    >
                      {item}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  )
}
