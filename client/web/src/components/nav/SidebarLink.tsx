import { Link, useLocation } from 'react-router-dom'
import { isNavItemActive, type NavItem } from '@/components/nav/navItems'

export function SidebarLink({ to, label, Icon, end }: NavItem) {
  const { pathname } = useLocation()
  const isActive = isNavItemActive(pathname, { to, end })

  return (
    <Link
      to={to}
      className={`flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13.5px] font-semibold ${
        isActive ? 'bg-accent-soft text-accent-strong' : 'text-ink-soft'
      }`}
    >
      <Icon className="h-[18px] w-[18px]" />
      {label}
    </Link>
  )
}
