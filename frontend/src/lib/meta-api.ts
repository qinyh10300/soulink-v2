import { apiRequest } from '@/lib/api'
import type { SchoolKey } from '@/lib/school-keys'

export interface RegisteredUserCountResult {
  registeredUserCount: number
}

export type RegisteredUserCountSchoolKey =
  | SchoolKey
  | 'ustb'
  | 'csu'
  | 'tju'
  | 'sjtu'
  | 'fdu'
  | 'nus'
  | 'ntu'
  | 'zuel'

export type SchoolRegisteredUserCounts = Partial<Record<RegisteredUserCountSchoolKey, number>>

export function fetchDepartments(email?: string) {
  const params = new URLSearchParams()

  if (email?.trim()) {
    params.set('email', email.trim())
  }

  const query = params.size > 0 ? `?${params.toString()}` : ''

  return apiRequest<string[]>(`/meta/departments${query}`)
}

export function fetchGrades() {
  return apiRequest<string[]>('/meta/grades')
}

export function fetchContactTypes() {
  return apiRequest<Array<{ label: string; value: string }>>('/meta/contact-types')
}

export function fetchRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/registered-user-count')
}

export function fetchPkuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/pku-registered-user-count')
}

export function fetchBnuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/bnu-registered-user-count')
}

export function fetchRucRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/ruc-registered-user-count')
}

export function fetchCucRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/cuc-registered-user-count')
}

export function fetchBfsuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/bfsu-registered-user-count')
}

export function fetchUstbRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/ustb-registered-user-count')
}

export function fetchCsuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/csu-registered-user-count')
}

export function fetchTjuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/tju-registered-user-count')
}

export function fetchSjtuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/sjtu-registered-user-count')
}

export function fetchFduRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/fdu-registered-user-count')
}

export function fetchNusRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/nus-registered-user-count')
}

export function fetchNtuRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/ntu-registered-user-count')
}

export function fetchZuelRegisteredUserCount() {
  return apiRequest<RegisteredUserCountResult>('/meta/zuel-registered-user-count')
}

const schoolRegisteredUserCountFetchers: Record<
  RegisteredUserCountSchoolKey,
  () => Promise<RegisteredUserCountResult>
> = {
  thu: fetchRegisteredUserCount,
  pku: fetchPkuRegisteredUserCount,
  bnu: fetchBnuRegisteredUserCount,
  ruc: fetchRucRegisteredUserCount,
  cuc: fetchCucRegisteredUserCount,
  bfsu: fetchBfsuRegisteredUserCount,
  ustb: fetchUstbRegisteredUserCount,
  csu: fetchCsuRegisteredUserCount,
  tju: fetchTjuRegisteredUserCount,
  sjtu: fetchSjtuRegisteredUserCount,
  fdu: fetchFduRegisteredUserCount,
  nus: fetchNusRegisteredUserCount,
  ntu: fetchNtuRegisteredUserCount,
  zuel: fetchZuelRegisteredUserCount,
}

export function fetchSchoolRegisteredUserCount(schoolKey: SchoolKey) {
  return schoolRegisteredUserCountFetchers[schoolKey]()
}

export async function fetchAllSchoolRegisteredUserCounts(): Promise<SchoolRegisteredUserCounts> {
  const entries = Object.entries(schoolRegisteredUserCountFetchers) as Array<
    [RegisteredUserCountSchoolKey, () => Promise<RegisteredUserCountResult>]
  >
  const settled = await Promise.allSettled(entries.map(([, fetcher]) => fetcher()))

  return settled.reduce<SchoolRegisteredUserCounts>((counts, result, index) => {
    if (result.status === 'fulfilled') {
      counts[entries[index][0]] = result.value.registeredUserCount
    }

    return counts
  }, {})
}
