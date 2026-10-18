import { Outlet } from 'react-router-dom'
import { DesignModeBadge } from '@/components/DesignModeBadge'
import { BottomNav } from '@/components/nav/BottomNav'
import { SidebarNav } from '@/components/nav/SidebarNav'
import { useViewport } from '@/hooks/useViewport'

export function AppShell() {
  const viewport = useViewport()

  if (viewport === 'desktop') {
    return (
      <div className="flex h-screen overflow-hidden bg-paper text-ink">
        <DesignModeBadge />
        <SidebarNav />
        <main data-scroll-lock className="no-scrollbar min-w-0 flex-1 overflow-y-auto overflow-x-hidden">
          <div className="mx-auto w-full max-w-5xl px-10 py-10">
            <Outlet />
          </div>
        </main>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-paper text-ink">
      <DesignModeBadge />
      <main className="mx-auto max-w-md pb-[calc(72px+env(safe-area-inset-bottom)+12px)] standalone:pb-[calc(88px+env(safe-area-inset-bottom)+12px)]">
        <Outlet />
      </main>
      <BottomNav />
    </div>
  )
}
