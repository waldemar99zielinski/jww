import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface SectionProps {
  title: string
  action?: { label: string; to: string }
  children: ReactNode
}

export function Section({ title, action, children }: SectionProps) {
  return (
    <section className="flex flex-col gap-2.5">
      <div className="flex items-baseline justify-between">
        <h3 className="m-0 text-[13px] font-semibold tracking-wide">{title}</h3>
        {action && (
          <Link to={action.to} className="text-xs font-semibold text-accent underline underline-offset-2">
            {action.label}
          </Link>
        )}
      </div>
      {children}
    </section>
  )
}
