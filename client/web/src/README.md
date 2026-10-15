# src/

Application source. See the root `README.md` for the full architecture overview (page
structure, state management rules, styling rules, design mode) — this file is just a map.

## Folders

- `api/` — typed fetch wrappers per server resource, plus `api/mock/` (design-mode fake backend).
- `store/` — Redux: session/auth state only, nothing else.
- `hooks/` — one module per server resource, wrapping TanStack Query.
- `lib/` — small pure utilities. No business logic, no React.
- `components/` — the **base layer**: generic, reusable, pure-display components only (`ui/` is
  the design-system primitive layer within it). No business logic, no page-specific knowledge.
- `pages/` — one folder per route (container + view-variants pattern — see root README). Each
  page that needs its own composed components keeps them in a `components/` subfolder scoped to
  that page (`pages/<page>/components/`), built from the base layer above.
- `data/` — placeholder/sample data for features not yet modeled on the server.

## Top-level files

- `App.tsx` — route table + the boot gate (waits on session restore + first-screen chunks before
  rendering anything, via `use()` on a module-level promise).
- `main.tsx` — React root, mounts `<BrowserRouter><Providers><App /></Providers></BrowserRouter>`.
- `providers.tsx` — wires the Redux store and the TanStack Query client.
- `index.css` — Tailwind entry + CSS custom properties (theme tokens consumed via `lib/cssVars.ts`).
- `vite-env.d.ts` — Vite's ambient types; don't hand-edit beyond what Vite scaffolds.

Every subfolder listed above has its own `README.md` with folder-specific conventions — read the
one for the folder you're about to touch.

## Testing

See `TESTING.md` (sibling to this folder, at `client/web/TESTING.md`) for the full testing
strategy — tooling (Vitest + RTL + MSW + Playwright), what to test per layer, and the rollout
plan. Each subfolder's own README states that folder's specific testing requirements; read both
before adding tests to a folder you haven't touched before.
