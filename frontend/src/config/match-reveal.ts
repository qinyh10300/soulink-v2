// 统一揭晓时间配置：前端会在这个时间点之前隐藏最近一次匹配结果请求。
const GLOBAL_MATCH_RESULT_REVEAL_AT = '2026-04-16T20:00:00+08:00'

export function getGlobalMatchResultRevealAt(fallback?: string) {
  const value = GLOBAL_MATCH_RESULT_REVEAL_AT.trim()
  if (value) {
    return value
  }

  return fallback
}

export function hasReachedGlobalMatchRevealAt(fallback?: string) {
  const revealAt = getGlobalMatchResultRevealAt(fallback)
  if (!revealAt) {
    return false
  }

  return new Date(revealAt).getTime() <= Date.now()
}
