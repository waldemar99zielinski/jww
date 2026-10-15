import type { CSSProperties } from 'react'

/** Lets a style object carry custom properties (e.g. `--tint`) past TypeScript's excess-property check. */
export function cssVars(vars: Record<string, string | number>): CSSProperties {
  return vars as CSSProperties
}
