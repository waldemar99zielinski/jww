import type { ComponentType, ReactNode, SVGProps } from 'react'
import { Link } from 'react-router-dom'
import { IconChevronRight } from '@/components/icons'

interface SettingsRowProps {
  Icon: ComponentType<SVGProps<SVGSVGElement>>
  label: string
  trailing?: ReactNode
  to?: string
  /** When given (and `to` isn't), the row becomes a real, full-width `<button>`. */
  onClick?: () => void
}

export function SettingsRow({ Icon, label, trailing, to, onClick }: SettingsRowProps) {
  const content = (
    <>
      <Icon className="h-[18px] w-[18px] text-muted" />
      <span className="flex-1">{label}</span>
      {trailing ?? ((to || onClick) && <IconChevronRight className="h-[18px] w-[18px] text-muted" />)}
    </>
  )

  if (to) {
    return (
      <li className="border-b border-line last:border-b-0">
        <Link to={to} className="flex items-center gap-3 py-3.5 text-[13.5px] font-semibold">
          {content}
        </Link>
      </li>
    )
  }

  if (onClick) {
    return (
      <li className="border-b border-line last:border-b-0">
        <button
          type="button"
          onClick={onClick}
          className="flex w-full items-center gap-3 py-3.5 text-left text-[13.5px] font-semibold"
        >
          {content}
        </button>
      </li>
    )
  }

  return (
    <li className="flex items-center gap-3 border-b border-line py-3.5 text-[13.5px] font-semibold last:border-b-0">
      {content}
    </li>
  )
}
