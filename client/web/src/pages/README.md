# pages/

One folder per route. See the root `README.md`'s "Page architecture" and "Component
architecture" sections for the full pattern — summary:

- `index.tsx` — container: owns hooks/state, assembles a `*ViewProps` object. All business logic
  lives here.
- `types.ts` — the shared `*ViewProps` interface.
- `Name.desktop.tsx` / `Name.mobile.tsx` — pure presentational layout variants, picked at
  runtime by `<ResponsiveView>` based on viewport, each independently `lazy()`-loaded.
- `components/` (when the page needs its own composed components) — presentational components
  used only by this page, composed from `src/components/ui/` (and other shared base components).
  See the root README's "Component architecture" section before adding one: check whether
  something in `src/components/` already covers it, keep one component per file, and never put
  business logic in a component here — only in `index.tsx`.

`pages/login/` is the one exception — a single `LoginPage.tsx` with no desktop/mobile split
(deliberate: one login screen regardless of viewport, don't "fix" it to match the others).

## Folders

- `add/` — Add Transaction flow.
- `budgets/` — visual-only prototype (see its own README).
- `home/` — dashboard / home screen.
- `login/` — the one non-split page.
- `settings/` — settings root + `categories/` and `currencies/` sub-pages.
- `transactions/` — transaction history + filters, plus `transactions/detail/` for viewing one
  transaction.

## Adding a new page

1. Create `pages/<name>/{index.tsx,types.ts,Name.desktop.tsx,Name.mobile.tsx}` following an
   existing folder (e.g. `pages/home/`) as a template. Add a `components/` subfolder only once
   you have a component that's genuinely specific to this page.
2. Register the route (lazy-imported) in `src/App.tsx`.
3. If it needs a nav entry, add it to `components/nav/navItems.ts` — both nav components pick it
   up automatically.

## Testing

See `TESTING.md` for the full breakdown; summary for this layer:

- **`index.tsx` containers** — render through a test helper wrapping real Redux/TanStack Query
  providers (MSW-mocked network), assert on behavior: given these mocked responses, does the
  right data reach the view, do handlers call the right mutation. This is where most
  business-logic bugs surface, since it's the only place business logic lives for a page.
- **`.desktop.tsx`/`.mobile.tsx`** — pure props-in/render-out tests, no providers needed. Since
  the architecture rule forbids these from diverging in logic, a test catching a desktop/mobile
  content mismatch is catching exactly the bug this layer split exists to prevent.
- **`components/`** — same as the base layer: RTL render + `user-event`, no mocking.
- A page may also own one Playwright flow if it's one of the critical ones listed in
  `TESTING.md`'s E2E section — check there before adding a new E2E spec for a page.

## Don't

- Don't put page-specific business logic in the `.desktop.tsx`/`.mobile.tsx` files or in a
  `components/` component — if the same computation or handler needs to exist in both view
  variants, it belongs in `index.tsx` instead.
- Don't duplicate a component that already exists in `src/components/` — import it instead of
  writing a page-local copy.
