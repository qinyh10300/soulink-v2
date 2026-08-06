import { getAuthSession } from '@/lib/user-profile-storage'

// 生产：与站点同域，由 Nginx 反代到后端（见 nginx.conf 的 /api/）。勿默认 localhost。
// 本地 `npm run dev`：vite 将 /api 代理到后端（见 vite.config.ts）。
// 需要指向其它地址时构建前设置 VITE_API_BASE_URL，例如 https://api.example.com/api/v1
const DEFAULT_API_BASE_URL = '/api/v1'
// const DEFAULT_API_BASE_URL = 'http://127.0.0.1:8080/api/v1'

interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}

interface ApiRequestOptions extends Omit<RequestInit, 'body' | 'headers'> {
  auth?: boolean
  body?: BodyInit | Record<string, unknown> | undefined
  headers?: HeadersInit
}

export class ApiError extends Error {
  code: number
  status: number
  data: unknown

  constructor(message: string, status: number, code: number, data: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.data = data
  }
}

export function resolveApiBaseUrl() {
  const baseUrl = import.meta.env.VITE_API_BASE_URL?.trim() || DEFAULT_API_BASE_URL

  return baseUrl.endsWith('/') ? baseUrl.slice(0, -1) : baseUrl
}

export async function apiBlobRequest(path: string) {
  const token = getAuthSession()?.token
  if (!token) {
    throw new ApiError('请先登录后再继续操作', 401, 4011, null)
  }
  const response = await fetch(`${resolveApiBaseUrl()}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
    cache: 'no-store',
  })
  if (!response.ok) {
    let payload: ApiEnvelope<unknown> | null = null
    try {
      payload = (await response.json()) as ApiEnvelope<unknown>
    } catch {
      // Binary endpoints may return an empty error body.
    }
    throw new ApiError(
      payload?.message || `Request failed with status ${response.status}`,
      response.status,
      payload?.code ?? -1,
      payload?.data ?? null,
    )
  }
  return response.blob()
}

function buildRequestBody(body: ApiRequestOptions['body']) {
  if (body === undefined || body instanceof FormData || typeof body === 'string') {
    return body
  }

  return JSON.stringify(body)
}

export async function apiRequest<T>(
  path: string,
  options: ApiRequestOptions = {},
) {
  const { auth = false, body, headers: initialHeaders, ...rest } = options
  const headers = new Headers(initialHeaders)

  if (
    body !== undefined &&
    !(body instanceof FormData) &&
    typeof body !== 'string' &&
    !headers.has('Content-Type')
  ) {
    headers.set('Content-Type', 'application/json')
  }

  if (auth) {
    const token = getAuthSession()?.token

    if (!token) {
      throw new ApiError('请先登录后再继续操作', 401, 4011, null)
    }

    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${resolveApiBaseUrl()}${path}`, {
    ...rest,
    body: buildRequestBody(body),
    headers,
  })

  let payload: ApiEnvelope<T> | null = null

  try {
    payload = (await response.json()) as ApiEnvelope<T>
  } catch {
    // Keep payload empty when the response has no JSON body.
  }

  if (!response.ok || !payload || payload.code !== 0) {
    throw new ApiError(
      payload?.message || `Request failed with status ${response.status}`,
      response.status,
      payload?.code ?? -1,
      payload?.data ?? null,
    )
  }

  return payload.data
}

export function getErrorMessage(
  error: unknown,
  fallback = '操作失败，请稍后重试',
) {
  if (error instanceof ApiError) {
    return error.message
  }

  if (error instanceof Error && error.message) {
    return error.message
  }

  return fallback
}
