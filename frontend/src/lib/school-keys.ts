export const schoolKeys = ['thu', 'pku', 'bnu', 'ruc', 'cuc', 'bfsu'] as const

export type SchoolKey = (typeof schoolKeys)[number]

export function isSchoolKey(value: string | undefined): value is SchoolKey {
  return value != null && schoolKeys.includes(value as SchoolKey)
}
