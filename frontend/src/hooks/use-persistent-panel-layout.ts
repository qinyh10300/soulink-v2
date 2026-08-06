import { useCallback, useRef, useState } from 'react'
import type { Layout } from 'react-resizable-panels'

interface PersistentPanelConfig {
  defaultSize?: number
  maxSize?: number
  minSize?: number
  rememberSize?: boolean
}

type PersistentPanelConfigs = Record<string, PersistentPanelConfig>

interface PanelPreference {
  collapsed: boolean
  size: number
}

type PanelPreferences = Record<string, PanelPreference>

function clampSize(value: number, config: PersistentPanelConfig) {
  return Math.min(
    config.maxSize ?? Number.POSITIVE_INFINITY,
    Math.max(config.minSize ?? 0, value),
  )
}

function defaultPreferences(configs: PersistentPanelConfigs) {
  return Object.fromEntries(
    Object.entries(configs).map(([id, config]) => [
      id,
      {
        collapsed: false,
        size: config.defaultSize ?? 0,
      },
    ]),
  ) as PanelPreferences
}

function readPreferences(key: string, configs: PersistentPanelConfigs) {
  const defaults = defaultPreferences(configs)

  try {
    const raw = window.localStorage.getItem(key)
    if (!raw) {
      return defaults
    }

    const parsed = JSON.parse(raw) as Partial<PanelPreferences>
    return Object.fromEntries(
      Object.entries(configs).map(([id, config]) => {
        const stored = parsed[id]
        const size =
          typeof stored?.size === 'number' && Number.isFinite(stored.size)
            ? clampSize(stored.size, config)
            : defaults[id].size

        return [
          id,
          {
            collapsed:
              typeof stored?.collapsed === 'boolean'
                ? stored.collapsed
                : false,
            size,
          },
        ]
      }),
    ) as PanelPreferences
  } catch {
    return defaults
  }
}

export function usePersistentPanelLayout(
  key: string,
  configs: PersistentPanelConfigs,
) {
  const [initialPreferences] = useState(() => readPreferences(key, configs))
  const configsRef = useRef(configs)
  const preferencesRef = useRef(initialPreferences)
  const groupElementRef = useRef<HTMLDivElement>(null)

  const getDefaultSize = useCallback((id: string) => {
    const config = configsRef.current[id]
    const preference = preferencesRef.current[id]

    if (!config || !preference) {
      return undefined
    }
    if (preference.collapsed) {
      return 0
    }
    return config.rememberSize ? preference.size : undefined
  }, [])

  const isInitiallyCollapsed = useCallback(
    (id: string) => preferencesRef.current[id]?.collapsed ?? false,
    [],
  )

  const onLayoutChanged = useCallback(
    (layout: Layout) => {
      const groupWidth = groupElementRef.current?.getBoundingClientRect().width
      const allPanelsVisible = Object.keys(configsRef.current).every(
        (id) => (layout[id] ?? 0) > 0.01,
      )
      const next = { ...preferencesRef.current }

      Object.entries(configsRef.current).forEach(([id, config]) => {
        const percentage = layout[id] ?? 0
        const current = next[id]
        next[id] = {
          ...current,
          collapsed: percentage <= 0.01,
        }

        if (
          config.rememberSize &&
          allPanelsVisible &&
          groupWidth &&
          percentage > 0.01
        ) {
          next[id].size = clampSize(
            (groupWidth * percentage) / 100,
            config,
          )
        }
      })

      preferencesRef.current = next
      try {
        window.localStorage.setItem(key, JSON.stringify(next))
      } catch {
        // Layout persistence is optional when storage is unavailable.
      }
    },
    [key],
  )

  const reset = useCallback(() => {
    const defaults = defaultPreferences(configsRef.current)
    preferencesRef.current = defaults
    try {
      window.localStorage.setItem(key, JSON.stringify(defaults))
    } catch {
      // Layout persistence is optional when storage is unavailable.
    }
  }, [key])

  return {
    getDefaultSize,
    groupElementRef,
    isInitiallyCollapsed,
    onLayoutChanged,
    reset,
  }
}
