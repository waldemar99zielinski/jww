import type { ComponentType, SVGProps } from 'react'
import { IconGear, IconHome, IconList, IconPlus, IconTarget } from '@/components/icons'

export interface NavItem {
  to: string
  label: string
  Icon: ComponentType<SVGProps<SVGSVGElement>>
  end?: boolean
}

export const primaryNavItems: NavItem[] = [
  { to: '/', label: 'Home', Icon: IconHome, end: true },
  { to: '/transactions', label: 'History', Icon: IconList },
  { to: '/budgets', label: 'Budgets', Icon: IconTarget },
  { to: '/settings', label: 'Settings', Icon: IconGear },
]

export const addNavItem: NavItem = { to: '/add', label: 'Add', Icon: IconPlus }

/** Mirrors react-router's own NavLink `end` matching, for nav links built from `NavItem` by hand. */
export function isNavItemActive(pathname: string, { to, end }: Pick<NavItem, 'to' | 'end'>): boolean {
  return end ? pathname === to : pathname === to || pathname.startsWith(`${to}/`)
}
