import { Check } from 'lucide-react'

import { cn } from '@/lib/utils'
import type { SurveySection, SurveySectionId } from '@/types/survey'

interface SurveySectionTabsProps {
  completedSectionIds: SurveySectionId[]
  currentSectionId: SurveySectionId
  sections: SurveySection[]
  onSelect: (sectionId: SurveySectionId) => void
}

export function SurveySectionTabs({
  completedSectionIds,
  currentSectionId,
  sections,
  onSelect,
}: SurveySectionTabsProps) {
  return (
    <div className="border-b border-[#eadfd6] bg-[rgba(249,244,239,0.88)]">
      <div className="mx-auto flex w-full max-w-[90rem] gap-2 overflow-x-auto px-4 py-3 [scrollbar-width:none] sm:gap-3 sm:px-6 sm:py-4 lg:px-10">
        {sections.map((section) => {
          const isCurrent = section.id === currentSectionId
          const isCompleted = completedSectionIds.includes(section.id)

          return (
            <button
              className={cn(
                'inline-flex shrink-0 items-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium transition-all sm:px-5 sm:py-2.5 sm:text-base lg:px-6 lg:py-3 lg:text-lg',
                isCurrent
                  ? 'bg-[#985f4b] text-white shadow-[0_12px_24px_rgba(118,78,58,0.2)]'
                  : isCompleted
                    ? 'bg-[#eef4ea] text-[#6b8b63]'
                    : 'bg-[#f3efe9] text-[#807064] hover:bg-[#ede7e0]',
              )}
              key={section.id}
              onClick={() => onSelect(section.id)}
              type="button"
            >
              {isCompleted ? <Check size={16} className="sm:h-[18px] sm:w-[18px]" /> : null}
              <span>{section.title}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
