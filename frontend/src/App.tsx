import { Outlet } from 'react-router-dom'

import { AppLayout } from '@/components/layout/app-layout'

export default function App() {
  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}
