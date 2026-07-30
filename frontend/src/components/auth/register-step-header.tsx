import { Check } from 'lucide-react'

const steps = [
  { id: 1, label: '邮箱验证' },
  { id: 2, label: '设置密码' },
  { id: 3, label: '个人信息' },
]

interface RegisterStepHeaderProps {
  activeStep: number
}

export function RegisterStepHeader({ activeStep }: RegisterStepHeaderProps) {
  return (
    <div className="space-y-3">
      <div className="flex items-center gap-1.5 sm:gap-3">
        {steps.map((step, index) => {
          const isActive = step.id === activeStep
          const isComplete = step.id < activeStep

          return (
            <div key={step.id} className="flex min-w-0 flex-1 items-center gap-1.5 sm:gap-3">
              <div className="flex items-center gap-1.5 sm:gap-3">
                <div
                  className={[
                    'flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-[0.82rem] font-semibold transition-colors sm:h-8 sm:w-8 sm:text-sm',
                    isComplete
                      ? 'bg-[#6b9b63] text-white'
                      : isActive
                        ? 'bg-[#9c6b48] text-white'
                        : 'border border-[#d9ccc2] bg-white text-[#b29f92]',
                  ].join(' ')}
                >
                  {isComplete ? <Check size={15} strokeWidth={3} /> : step.id}
                </div>
                <span
                  className={[
                    'text-[0.82rem] whitespace-nowrap transition-colors sm:text-sm',
                    isComplete || isActive
                      ? 'font-medium text-[#2c2521]'
                      : 'text-[#a59387]',
                  ].join(' ')}
                >
                  {step.label}
                </span>
              </div>
              {index < steps.length - 1 ? (
                <div
                  className={[
                    'h-px min-w-2 flex-1 transition-colors sm:min-w-4',
                    step.id < activeStep ? 'bg-[#b47f5a]' : 'bg-[#ddd0c6]',
                  ].join(' ')}
                />
              ) : null}
            </div>
          )
        })}
      </div>
    </div>
  )
}
