import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'

import {
  completeOnboarding as completeOnboardingRequest,
  fetchCampusThemes,
  fetchEngagementPreferences,
  fetchNotifications,
  markAllNotificationsRead as markAllNotificationsReadRequest,
  markNotificationDelivered as markNotificationDeliveredRequest,
  markNotificationRead as markNotificationReadRequest,
  replaceCampusThemes,
  updateNotificationPreferences as updateNotificationPreferencesRequest,
  type CampusTheme,
  type EngagementPreferences,
  type EngagementScene,
  type NotificationPermissionState,
  type NotificationPreferencesInput,
  type UserNotification,
} from '@/lib/engagement-api'
import {
  registerPWA,
  showServiceWorkerNotification,
  type BeforeInstallPromptEvent,
  type PWARegistrationController,
} from '@/lib/pwa'

interface EngagementContextValue {
  preferences: EngagementPreferences | null
  themes: CampusTheme[]
  notifications: UserNotification[]
  unreadNotifications: number
  nextNotificationCursor?: number
  loading: boolean
  installAvailable: boolean
  updateAvailable: boolean
  refreshNotifications: () => Promise<void>
  loadMoreNotifications: () => Promise<void>
  updateNotificationPreferences: (
    input: NotificationPreferencesInput,
  ) => Promise<EngagementPreferences>
  requestBrowserPermission: () => Promise<NotificationPermissionState>
  completeOnboarding: (scenes: EngagementScene[]) => Promise<EngagementPreferences>
  updateSelectedThemes: (themeKeys: string[]) => Promise<CampusTheme[]>
  markNotificationRead: (notificationId: number) => Promise<void>
  markAllNotificationsRead: () => Promise<void>
  installApp: () => Promise<boolean>
  applyUpdate: () => void
  dismissUpdate: () => void
}

const EngagementContext = createContext<EngagementContextValue | null>(null)

const fallbackPreferences: EngagementPreferences = {
  notificationsEnabled: false,
  newMessages: true,
  meetupInvites: true,
  meetupReminders: true,
  dailyPicks: true,
  permissionState: 'default',
  timezone: 'Asia/Shanghai',
  selectedScenes: [],
  onboardingCompleted: false,
}

function browserPermissionState(): NotificationPermissionState {
  if (typeof Notification === 'undefined') {
    return 'unavailable'
  }
  return Notification.permission
}

export function EngagementProvider({ children }: { children: ReactNode }) {
  const [preferences, setPreferences] = useState<EngagementPreferences | null>(null)
  const [themes, setThemes] = useState<CampusTheme[]>([])
  const [notifications, setNotifications] = useState<UserNotification[]>([])
  const [unreadNotifications, setUnreadNotifications] = useState(0)
  const [nextNotificationCursor, setNextNotificationCursor] = useState<number>()
  const [loading, setLoading] = useState(true)
  const [installPrompt, setInstallPrompt] = useState<BeforeInstallPromptEvent | null>(null)
  const [updateController, setUpdateController] =
    useState<PWARegistrationController | null>(null)
  const preferencesRef = useRef<EngagementPreferences | null>(null)

  const deliverBrowserNotifications = useCallback(
    async (items: UserNotification[], current: EngagementPreferences | null) => {
      if (
        document.visibilityState !== 'hidden' ||
        !current?.notificationsEnabled ||
        current.permissionState !== 'granted' ||
        browserPermissionState() !== 'granted'
      ) {
        return
      }

      const pending = items.filter((item) => !item.deliveredAt).slice(0, 3)
      for (const item of pending) {
        try {
          const shown = await showServiceWorkerNotification({
            id: item.id,
            title: item.title,
            body: item.body,
            actionPath: item.actionPath,
          })
          if (!shown) continue
          await markNotificationDeliveredRequest(item.id)
          const deliveredAt = new Date().toISOString()
          setNotifications((currentItems) =>
            currentItems.map((currentItem) =>
              currentItem.id === item.id
                ? { ...currentItem, deliveredAt }
                : currentItem,
            ),
          )
        } catch {
          // Delivery remains pending and can be retried on the next background refresh.
        }
      }
    },
    [],
  )

  const refreshNotifications = useCallback(async () => {
    try {
      const page = await fetchNotifications(undefined, 20)
      setNotifications(page.items)
      setUnreadNotifications(page.unreadCount)
      setNextNotificationCursor(page.nextCursor)
      await deliverBrowserNotifications(page.items, preferencesRef.current)
    } catch {
      // Notification polling never blocks the primary connection workflow.
    }
  }, [deliverBrowserNotifications])

  const loadMoreNotifications = useCallback(async () => {
    if (!nextNotificationCursor) return
    const page = await fetchNotifications(nextNotificationCursor, 20)
    setNotifications((current) => {
      const known = new Set(current.map((item) => item.id))
      return [...current, ...page.items.filter((item) => !known.has(item.id))]
    })
    setUnreadNotifications(page.unreadCount)
    setNextNotificationCursor(page.nextCursor)
  }, [nextNotificationCursor])

  useEffect(() => {
    let cancelled = false
    const load = async () => {
      const [preferencesResult, themesResult] = await Promise.allSettled([
        fetchEngagementPreferences(),
        fetchCampusThemes(),
      ])
      if (cancelled) return

      if (preferencesResult.status === 'fulfilled') {
        const nextPreferences = preferencesResult.value
        const actualPermission = browserPermissionState()
        preferencesRef.current = nextPreferences
        setPreferences(nextPreferences)
        if (nextPreferences.permissionState !== actualPermission) {
          void updateNotificationPreferencesRequest({
            notificationsEnabled: nextPreferences.notificationsEnabled,
            newMessages: nextPreferences.newMessages,
            meetupInvites: nextPreferences.meetupInvites,
            meetupReminders: nextPreferences.meetupReminders,
            dailyPicks: nextPreferences.dailyPicks,
            permissionState: actualPermission,
            timezone: nextPreferences.timezone,
          }).then((updated) => {
            preferencesRef.current = updated
            setPreferences(updated)
          }).catch(() => {})
        }
      } else {
        preferencesRef.current = fallbackPreferences
        setPreferences(fallbackPreferences)
      }
      if (themesResult.status === 'fulfilled') {
        setThemes(themesResult.value)
      }
      setLoading(false)
      void refreshNotifications()
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [refreshNotifications])

  useEffect(() => {
    const interval = window.setInterval(() => {
      void refreshNotifications()
    }, 30_000)
    const refreshWhenVisible = () => {
      if (document.visibilityState === 'visible') {
        void refreshNotifications()
      }
    }
    document.addEventListener('visibilitychange', refreshWhenVisible)
    return () => {
      window.clearInterval(interval)
      document.removeEventListener('visibilitychange', refreshWhenVisible)
    }
  }, [refreshNotifications])

  useEffect(() => {
    const handleInstallPrompt = (event: Event) => {
      event.preventDefault()
      setInstallPrompt(event as BeforeInstallPromptEvent)
    }
    const handleInstalled = () => setInstallPrompt(null)
    window.addEventListener('beforeinstallprompt', handleInstallPrompt)
    window.addEventListener('appinstalled', handleInstalled)
    void registerPWA(setUpdateController).catch(() => {})
    return () => {
      window.removeEventListener('beforeinstallprompt', handleInstallPrompt)
      window.removeEventListener('appinstalled', handleInstalled)
    }
  }, [])

  const updateNotificationPreferences = useCallback(
    async (input: NotificationPreferencesInput) => {
      const updated = await updateNotificationPreferencesRequest(input)
      preferencesRef.current = updated
      setPreferences(updated)
      await refreshNotifications()
      return updated
    },
    [refreshNotifications],
  )

  const requestBrowserPermission = useCallback(async () => {
    const current = preferencesRef.current ?? fallbackPreferences
    let permission: NotificationPermissionState = 'unavailable'
    if (typeof Notification !== 'undefined') {
      permission = await Notification.requestPermission()
    }
    await updateNotificationPreferences({
      notificationsEnabled: current.notificationsEnabled,
      newMessages: current.newMessages,
      meetupInvites: current.meetupInvites,
      meetupReminders: current.meetupReminders,
      dailyPicks: current.dailyPicks,
      permissionState: permission,
      timezone: current.timezone,
    })
    return permission
  }, [updateNotificationPreferences])

  const completeOnboarding = useCallback(async (scenes: EngagementScene[]) => {
    const updated = await completeOnboardingRequest(scenes)
    preferencesRef.current = updated
    setPreferences(updated)
    setThemes(await fetchCampusThemes())
    return updated
  }, [])

  const updateSelectedThemes = useCallback(async (themeKeys: string[]) => {
    const updated = await replaceCampusThemes(themeKeys)
    setThemes(updated)
    return updated
  }, [])

  const markNotificationRead = useCallback(async (notificationId: number) => {
    await markNotificationReadRequest(notificationId)
    const readAt = new Date().toISOString()
    setNotifications((current) =>
      current.map((item) =>
        item.id === notificationId ? { ...item, readAt } : item,
      ),
    )
    setUnreadNotifications((current) => Math.max(0, current - 1))
  }, [])

  const markAllNotificationsRead = useCallback(async () => {
    await markAllNotificationsReadRequest()
    const readAt = new Date().toISOString()
    setNotifications((current) => current.map((item) => ({ ...item, readAt })))
    setUnreadNotifications(0)
  }, [])

  const installApp = useCallback(async () => {
    if (!installPrompt) return false
    await installPrompt.prompt()
    const choice = await installPrompt.userChoice
    if (choice.outcome === 'accepted') {
      setInstallPrompt(null)
      return true
    }
    return false
  }, [installPrompt])

  const applyUpdate = useCallback(() => {
    if (!updateController) return
    navigator.serviceWorker.addEventListener(
      'controllerchange',
      () => window.location.reload(),
      { once: true },
    )
    updateController.applyUpdate()
  }, [updateController])

  const value = useMemo<EngagementContextValue>(
    () => ({
      preferences,
      themes,
      notifications,
      unreadNotifications,
      nextNotificationCursor,
      loading,
      installAvailable: Boolean(installPrompt),
      updateAvailable: Boolean(updateController),
      refreshNotifications,
      loadMoreNotifications,
      updateNotificationPreferences,
      requestBrowserPermission,
      completeOnboarding,
      updateSelectedThemes,
      markNotificationRead,
      markAllNotificationsRead,
      installApp,
      applyUpdate,
      dismissUpdate: () => setUpdateController(null),
    }),
    [
      applyUpdate,
      completeOnboarding,
      installApp,
      installPrompt,
      loadMoreNotifications,
      loading,
      markAllNotificationsRead,
      markNotificationRead,
      nextNotificationCursor,
      notifications,
      preferences,
      refreshNotifications,
      requestBrowserPermission,
      themes,
      unreadNotifications,
      updateController,
      updateNotificationPreferences,
      updateSelectedThemes,
    ],
  )

  return <EngagementContext.Provider value={value}>{children}</EngagementContext.Provider>
}

export function useEngagement() {
  const value = useContext(EngagementContext)
  if (!value) {
    throw new Error('useEngagement must be used inside EngagementProvider')
  }
  return value
}
