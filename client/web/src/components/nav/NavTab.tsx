import { Link, useLocation } from 'react-router-dom'
import { isNavItemActive, type NavItem } from '@/components/nav/navItems'

export function NavTab({ to, label, Icon, end }: NavItem) {
  const { pathname } = useLocation()
  const isActive = isNavItemActive(pathname, { to, end })

  return (
    <Link
      to={to}
      className={`flex flex-col items-center gap-[3px] rounded-lg px-1 py-1.5 text-[10.5px] font-semibold ${
        isActive ? 'text-accent' : 'text-muted'
      }`}
    >
      <Icon className="h-[21px] w-[21px]" />
      {label}
    </Link>
  )
}
