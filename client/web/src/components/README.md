# components/

The **base layer**: generic, reusable, pure-display components with no business logic and no
knowledge of any specific page or feature. If a component needs `hooks/`, `api/`, `store/`, or
page-specific types to do its job, it does not belong here — it belongs in that page's own
`pages/<page>/components/` folder instead (see `pages/README.md`).

**Before adding a new component anywhere in the app, check here first.** A large share of what
a new screen needs usually already exists as a primitive in `ui/` or a shared piece here —
reusing it keeps the design consistent and avoids a second near-duplicate implementation.

## What belongs here vs. in a page's `components/`

| | `src/components/` (this folder) | `pages/<page>/components/` |
|---|---|---|
| Used by | 2+ pages, or app-shell chrome | exactly one page |
| Knows about | nothing page-specific — props only | that page's `types.ts`, its view props |
| Logic | pure display/formatting | still pure display — business logic stays in the page's `index.tsx` container, never in a component file |

If a component you're adding to a page folder turns out to be useful elsewhere too, promote it
here instead of copying it.

## Layout

- `ui/` — the design-system primitive layer (buttons, inputs, switches, bottom sheets, …). See
  its own README for the one-component-per-file rule and the full primitive list.
- `nav/` — the app's navigation chrome (`BottomNav`/`SidebarNav` and their row sub-components,
  `navItems.ts`), shared across every authenticated page via `AppShell`.
- Top-level files — app-shell and cross-page components:
  - `AppShell.tsx` — picks the desktop/mobile shell (sidebar vs. bottom nav) for all authenticated
    routes.
  - `BootSplash.tsx`, `PageFallback.tsx` — loading states shown during the boot gate / lazy route
    loads (see root README's page architecture section).
  - `DesignModeBadge.tsx`, `ProtectedRoute.tsx`, `ResponsiveView.tsx` — app-shell/routing
    infrastructure.
  - `OfflineGate.tsx` / `OfflineScreen.tsx` — the blocking "You're offline" state. `OfflineGate`
    wraps the whole app (in `App.tsx`), marks the routed tree `inert` while offline and owns
    focus/announcement; `OfflineScreen` is the pure panel it renders. App-shell chrome, so it
    lives here rather than in a page.
  - `Logo.tsx` / `LogoMark.tsx` / `LogoWord.tsx` — the brand wordmark system. `LogoWord` is a
    private helper split into its own file (per the one-component-per-file rule) but is only ever
    imported by `Logo.tsx` — don't reach for it directly.
  - `Section.tsx`, `StatusBanner.tsx`, `ProgressBar.tsx` — small generic display primitives used
    across pages but one level above `ui/` (slightly more opinionated than a raw form control).
  - `TransactionRow.tsx` — genuinely reused across two pages (`pages/home/`, `pages/transactions/`)
    at once, which is why it lives here rather than in either page's own `components/` folder.
  - `icons.tsx` — **the one exception** to one-component-per-file: a hand-drawn SVG icon set
    behind a single shared `Icon` wrapper. Splitting 24+ one-line icon components into 24+ files
    would hurt navigability more than it helps — add new icons here, don't inline SVG elsewhere.

## Testing

Render with React Testing Library, props in, assert on rendered output — no network/store/router
mocking needed for anything here (pure display, no `hooks`/`api`/`store` access). Query by role/
label over `getByTestId`. `TransactionRow.tsx` is the one component here that renders real
category/transaction data shapes — give it a case per transaction type (income vs expense) and a
case exercising `lib/categoryVisuals.ts`'s fallback path (an unrecognized category name).

## Rules (see root README for the full project-wide rules)

- One component per file, default. A private subcomponent that exists purely to keep a parent's
  JSX readable, is never imported anywhere else, and has no independently meaningful props may
  stay in the same file — otherwise split it out (see `icons.tsx` as the deliberate, documented
  exception, and `Logo.tsx`/`LogoWord.tsx` as the normal case).
- No business logic, no hooks that touch `api/`/`store/`/TanStack Query. Components here take
  data and callbacks as props and render — nothing else. Wiring to real data happens in a page's
  `index.tsx` container.
