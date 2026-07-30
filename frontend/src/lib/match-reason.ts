import type { MatchReasonJson, MatchResult } from '@/lib/match-api'

function parseReasonCandidate(value: unknown): MatchReasonJson | null {
  if (!value) {
    return null
  }

  let payload: unknown = value
  if (typeof payload === 'string') {
    const text = payload.trim()
    if (!text.startsWith('{')) {
      return null
    }
    try {
      payload = JSON.parse(text)
    } catch {
      return null
    }
  }

  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) {
    return null
  }

  const normalized = payload as MatchReasonJson
  const hasStructuredContent =
    Boolean(normalized.match_summary) ||
    Boolean(normalized.match_reasons?.length) ||
    Boolean(normalized.date_topics?.length) ||
    Boolean(normalized.cautions?.length)

  return hasStructuredContent ? normalized : null
}

export function getStructuredMatchReason(match: MatchResult): MatchReasonJson | null {
  const candidates: unknown[] = [
    match.reasonJson,
    match.reason,
    match.reasonJson?.text,
  ]

  for (const candidate of candidates) {
    const parsed = parseReasonCandidate(candidate)
    if (parsed) {
      return parsed
    }
  }

  return null
}

export function getFallbackMatchReasonText(match: MatchResult): string {
  return match.reason?.trim() || match.reasonJson?.error?.trim() || ''
}
