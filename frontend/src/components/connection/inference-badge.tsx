import { PencilLine, Sparkles, Workflow } from 'lucide-react'
import { Link } from 'react-router-dom'

import type { InferenceMetadata } from '@/lib/connection-api'
import { cn } from '@/lib/utils'

interface InferenceBadgeProps {
  inference: InferenceMetadata
  editTo?: string
  className?: string
  compact?: boolean
}

export function InferenceBadge({
  inference,
  editTo,
  className,
  compact = false,
}: InferenceBadgeProps) {
  const fromModel = inference.source === 'model'
  const Icon = fromModel ? Sparkles : Workflow
  const confidence = Math.round(inference.confidence * 100)
  const details = [
    fromModel ? inference.model || 'OpenAI-compatible model' : 'SouLink 本地规则',
    inference.promptVersion,
    `置信度 ${confidence}%`,
  ].join(' · ')

  return (
    <span className={cn('inline-flex min-w-0 items-center gap-1.5', className)}>
      <span
        className={cn(
          'inline-flex h-6 min-w-0 items-center gap-1.5 rounded-md border px-2 text-[10px] font-medium',
          fromModel
            ? 'border-[#d4c7dc] bg-[#f6f1f8] text-[#624273]'
            : 'border-[#d4ded9] bg-[#f1f6f3] text-[#47665b]',
        )}
        title={details}
      >
        <Icon className="shrink-0" size={11} />
        <span className="truncate">
          {fromModel ? '模型辅助' : '本地规则'}
          {compact ? '' : ` · ${confidence}%`}
        </span>
      </span>
      {editTo && inference.editable ? (
        <Link
          aria-label="检查并编辑推荐依据"
          className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-[#77736c] hover:bg-[#ece9e4] hover:text-[#4e655d]"
          title="检查并编辑推荐依据"
          to={editTo}
        >
          <PencilLine size={12} />
        </Link>
      ) : null}
    </span>
  )
}
