import { isDesignMode } from '@/lib/designMode'

/** A persistent reminder that everything on screen is mocked — no backend, nothing saved. */
export function DesignModeBadge() {
  if (!isDesignMode()) return null

  return (
    <div className="pointer-events-none fixed inset-x-0 top-0 z-50 flex justify-center">
      <span className="rounded-b-lg bg-critical px-3 py-1 text-[11px] font-bold tracking-wide text-on-accent">
        DESIGN MODE — mocked data, no backend
      </span>
    </div>
  )
}
