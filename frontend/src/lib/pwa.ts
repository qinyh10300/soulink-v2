export interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>
  userChoice: Promise<{ outcome: 'accepted' | 'dismissed'; platform: string }>
}

export interface PWARegistrationController {
  registration: ServiceWorkerRegistration
  applyUpdate: () => void
}

export async function registerPWA(
  onUpdateAvailable: (controller: PWARegistrationController) => void,
) {
  if (!('serviceWorker' in navigator)) {
    return null
  }

  const registration = await navigator.serviceWorker.register('/sw.js', {
    scope: '/',
  })

  const announceWaitingWorker = () => {
    if (!registration.waiting) {
      return
    }
    onUpdateAvailable({
      registration,
      applyUpdate: () => registration.waiting?.postMessage({ type: 'SKIP_WAITING' }),
    })
  }

  announceWaitingWorker()
  registration.addEventListener('updatefound', () => {
    const worker = registration.installing
    if (!worker) return
    worker.addEventListener('statechange', () => {
      if (worker.state === 'installed' && navigator.serviceWorker.controller) {
        announceWaitingWorker()
      }
    })
  })

  return registration
}

export async function showServiceWorkerNotification(input: {
  id: number
  title: string
  body: string
  actionPath: string
}) {
  if (!('serviceWorker' in navigator) || Notification.permission !== 'granted') {
    return false
  }

  const registration = await navigator.serviceWorker.ready
  await registration.showNotification(input.title, {
    body: input.body,
    icon: '/pwa/icon-192.png',
    badge: '/pwa/badge-96.png',
    tag: `soulink-${input.id}`,
    data: { actionPath: input.actionPath, notificationId: input.id },
  })
  return true
}
