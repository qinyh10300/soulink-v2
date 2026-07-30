export function getAvatarFallbackLabel(name?: string, fallback = 'S') {
  const trimmedName = name?.trim()

  if (!trimmedName) {
    return fallback
  }

  const [firstCharacter] = Array.from(trimmedName)

  if (!firstCharacter) {
    return fallback
  }

  return /^[a-z]$/i.test(firstCharacter) ? firstCharacter.toUpperCase() : firstCharacter
}
