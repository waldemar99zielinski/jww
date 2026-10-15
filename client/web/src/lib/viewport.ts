export type Viewport = 'mobile' | 'desktop'

/** Matches Tailwind's `lg` breakpoint, so CSS and layout-selection agree on where "desktop" starts. */
export const DESKTOP_BREAKPOINT_PX = 1024

export function getViewport(): Viewport {
  return window.innerWidth >= DESKTOP_BREAKPOINT_PX ? 'desktop' : 'mobile'
}
