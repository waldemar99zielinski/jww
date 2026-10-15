/**
 * Design mode: browse the whole app against in-memory mock data, with no backend involved
 * and auth bypassed. Toggle with VITE_DESIGN_MODE=true in .env (requires a dev server restart,
 * same as any other Vite env var).
 */
export function isDesignMode(): boolean {
  return import.meta.env.VITE_DESIGN_MODE === 'true'
}
