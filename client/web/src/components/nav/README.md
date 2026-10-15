# components/nav/

Navigation chrome shared across the authenticated app shell — lives in `src/components/` (not a
page folder) because it's used by `AppShell.tsx` for every page, not any single one.

- `navItems.ts` — the single source of truth for the primary nav items (`to`/`label`/`Icon`/
  `end`) and the standalone "Add" action, plus `isNavItemActive()` — the shared active-route
  check (mirrors `react-router`'s own `NavLink` `end` semantics). Add a new nav destination by
  editing `primaryNavItems` once; both nav variants below pick it up automatically.
- `BottomNav.tsx` + `NavTab.tsx` — mobile tab bar. `NavTab` is `BottomNav`'s per-item row,
  split into its own file (one component per file — see `components/ui/README.md`'s convention).
- `SidebarNav.tsx` + `SidebarLink.tsx` — desktop sidebar nav, same split pattern.

## Testing

- `isNavItemActive()` is a pure function — unit test it directly against `navItems.ts`'s `end`
  semantics (exact-match vs prefix-match routes), independent of rendering anything.
- `BottomNav`/`NavTab` and `SidebarNav`/`SidebarLink` need a router context to test active-state
  styling/`aria-current` — render through a `MemoryRouter` (see `TESTING.md`) at a given path and
  assert the matching item is marked active.

## Don't

- Don't reimplement active-route matching locally in a new nav component — import
  `isNavItemActive` from `navItems.ts`.
- Don't hardcode route paths/labels/icons outside `navItems.ts` — add a new `NavItem` entry
  instead, so both desktop and mobile nav pick it up automatically.
- Don't inline a nav row component back into `BottomNav.tsx`/`SidebarNav.tsx` — keep the
  one-component-per-file split so each row can be read and changed independently.
