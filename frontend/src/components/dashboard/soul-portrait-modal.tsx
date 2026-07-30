import { useEffect, useState } from 'react'

import {
  SoulPortraitCard,
  type SoulPortraitViewStatus,
} from '@/components/dashboard/soul-portrait-card'
import { ApiError, getErrorMessage } from '@/lib/api'
import {
  fetchQuestionnairePortrait,
  type QuestionnairePortraitResult,
} from '@/lib/questionnaire-api'

interface SoulPortraitModalProps {
  onClose: () => void
  open: boolean
  userName: string
}

export function SoulPortraitModal({
  onClose,
  open,
  userName,
}: SoulPortraitModalProps) {
  const [portrait, setPortrait] = useState<QuestionnairePortraitResult | null>(null)
  const [status, setStatus] = useState<SoulPortraitViewStatus>('idle')
  const [errorMessage, setErrorMessage] = useState('')

  useEffect(() => {
    if (!open) {
      return
    }

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose()
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose, open])

  useEffect(() => {
    if (!open) {
      return
    }

    let cancelled = false
    let retryTimer: number | null = null

    const loadPortrait = async (attempt = 0) => {
      if (cancelled) {
        return
      }

      setPortrait(null)
      setErrorMessage('')
      setStatus(attempt === 0 ? 'loading' : 'thinking')

      try {
        const result = await fetchQuestionnairePortrait()
        if (cancelled) {
          return
        }

        setPortrait(result)
        setStatus('ready')
      } catch (error) {
        if (cancelled) {
          return
        }

        if (shouldKeepThinking(error)) {
          setPortrait(null)
          setStatus('thinking')
          retryTimer = window.setTimeout(() => {
            void loadPortrait(attempt + 1)
          }, 3000)
          return
        }

        setPortrait(null)
        setStatus('error')
        setErrorMessage(getErrorMessage(error, 'Failed to load your soul portrait.'))
      }
    }

    void loadPortrait()

    return () => {
      cancelled = true
      if (retryTimer !== null) {
        window.clearTimeout(retryTimer)
      }
    }
  }, [open])

  if (!open) {
    return null
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <button
        aria-label="Close soul portrait"
        className="absolute inset-0 bg-[rgba(44,31,24,0.32)] backdrop-blur-[6px]"
        onClick={onClose}
        type="button"
      />

      <div className="relative z-10 w-full">
        <SoulPortraitCard
          errorMessage={errorMessage}
          onClose={onClose}
          portrait={portrait}
          status={status}
          userName={userName}
        />
      </div>
    </div>
  )
}

function shouldKeepThinking(error: unknown) {
  if (!(error instanceof ApiError)) {
    return false
  }

  return error.status === 409 || error.status === 425 || error.status === 503 || error.status === 504
}
