import { Link } from 'react-router-dom'
import { Logo } from '@/components/Logo'
import { addNavItem, primaryNavItems } from '@/components/nav/navItems'
import { SidebarLink } from '@/components/nav/SidebarLink'

export function SidebarNav() {
  return (
    <nav className="flex h-screen w-60 flex-none flex-col gap-1 border-r border-line bg-surface px-3 py-6">
      <Logo size="sm" className="mb-5 px-2" />

      <Link
        to={addNavItem.to}
        className="mb-4 flex items-center justify-center gap-2 rounded-xl bg-accent py-2.5 text-[13px] font-bold text-on-accent"
      >
        <addNavItem.Icon className="h-4 w-4" />
        Add transaction
      </Link>

      {primaryNavItems.map((item) => (
        <SidebarLink key={item.to} {...item} />
      ))}
    </nav>
  )
}
