import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'

import './index.css'
import iconUrl from '@/assets/soulink-logo-icon.png'
import { router } from '@/app/router'

const favicon =
  document.querySelector<HTMLLinkElement>("link[rel='icon']") ??
  document.createElement('link')

favicon.rel = 'icon'
favicon.type = 'image/png'
favicon.href = iconUrl
document.head.appendChild(favicon)

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
