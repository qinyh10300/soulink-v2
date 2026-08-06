export const connectionPaths = {
  root: '/connect',
  discover: '/connect/discover',
  intents: '/connect/intents',
  newIntent: '/connect/new',
  groups: '/connect/groups',
  rooms: '/connect/rooms',
} as const

export function connectionIntentPath(intentId: number | string) {
  return `/connect/intents/${intentId}`
}

export function connectionRoomPath(roomId: number | string) {
  return `/connect/rooms/${roomId}`
}

export function contextProfilePath(
  userId: number | string,
  intentId: number | string,
) {
  return `/profile/${userId}?intentId=${intentId}`
}

export function isConnectionPath(pathname: string) {
  return pathname === connectionPaths.root || pathname.startsWith(`${connectionPaths.root}/`)
}
