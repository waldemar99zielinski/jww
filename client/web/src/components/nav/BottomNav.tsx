import { Link } from 'react-router-dom'
import { NavTab } from '@/components/nav/NavTab'
import { addNavItem, primaryNavItems } from '@/components/nav/navItems'

const leftTabs = primaryNavItems.slice(0, 2)
const rightTabs = primaryNavItems.slice(2)

export function BottomNav() {
  return (
    <nav className="tabbar fixed inset-x-0 bottom-0 z-10 flex h-[72px] items-center justify-around border-t border-line pb-[env(safe-area-inset-bottom)]">
      {leftTabs.map((tab) => (
        <NavTab key={tab.to} {...tab} />
      ))}

      <Link
        to={addNavItem.to}
        aria-label="Add transaction"
        className="relative top-[-14px] flex h-[50px] w-[50px] flex-none items-center justify-center rounded-full bg-accent text-on-accent shadow-[var(--shadow-md)]"
      >
        <addNavItem.Icon className="h-[22px] w-[22px]" />
        <span className="absolute bottom-[-15px] text-[9.5px] font-semibold text-muted">Add</span>
      </Link>

      {rightTabs.map((tab) => (
        <NavTab key={tab.to} {...tab} />
      ))}
    </nav>
  )
}
