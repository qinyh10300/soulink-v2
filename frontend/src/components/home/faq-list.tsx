import { AnimatePresence, motion } from 'framer-motion'
import { ChevronDown } from 'lucide-react'
import { useState } from 'react'

import type { FaqItem } from '@/types/landing'

interface FaqListProps {
  items: FaqItem[]
}

const accordionEase = [0.22, 1, 0.36, 1] as const

export function FaqList({ items }: FaqListProps) {
  const [activeIndex, setActiveIndex] = useState(0)

  return (
    <div className="border-t border-[#ece5dd]">
      {items.map((item, index) => {
        const isActive = index === activeIndex

        return (
          <div key={item.question} className="border-b border-[#ece5dd]">
            <button
              className="flex w-full items-center justify-between gap-6 py-7 text-left transition-colors duration-200 hover:text-[#7a6452] sm:py-8"
              onClick={() => setActiveIndex(isActive ? -1 : index)}
              type="button"
            >
              <span
                className="text-[1.22rem] font-semibold leading-[1.35] tracking-[-0.02em] text-[#231d1a] sm:text-[1.5rem]"
                style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
              >
                {item.question}
              </span>
              <ChevronDown
                className={`mt-1 h-5 w-5 shrink-0 text-[#b5a79b] transition-transform duration-300 ${isActive ? 'rotate-180' : ''}`}
              />
            </button>

            <AnimatePresence initial={false}>
              {isActive ? (
                <motion.div
                  animate={{ height: 'auto', opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  initial={{ height: 0, opacity: 0 }}
                  transition={{ duration: 0.24, ease: accordionEase }}
                >
                  <div className="pb-7 pr-10 sm:pb-8">
                    <p
                      className="max-w-4xl text-[0.96rem] leading-[1.9] text-[#6f6256] sm:text-[1rem]"
                      style={{ fontFamily: '"Cormorant Garamond", "STSong", serif' }}
                    >
                      {item.answer}
                    </p>
                  </div>
                </motion.div>
              ) : null}
            </AnimatePresence>
          </div>
        )
      })}
    </div>
  )
}
