# Testing Strategy

This document describes this app's testing strategy: the tooling, what to test at each layer of
the architecture in `README.md`, the conventions, and how the suite was rolled out. The tooling
below is now **implemented** — co-located `*.test.*` files live beside the code they cover, shared
infrastructure is in `src/test/`, and Playwright specs are in `tests/e2e/`. Run it all with
`npm run test` and `npm run test:e2e`.

Tech choices below (Vitest, React Testing Library, MSW, Playwright) have been discussed and
approved — see decisions inline. Nothing here should be treated as silently approving further
tool additions beyond what's listed; any new testing-adjacent dependency still needs sign-off
first, per the project's existing "ask before tech choices" rule. (The one production dependency
this feature set added, `vite-plugin-pwa`, was chosen explicitly under that same rule for
installability and the offline shell — see the root README's "PWA / install" section.)

## Philosophy

Match test effort to where regressions actually happen in this codebase, not to a coverage
number:

- **Pure logic** (`lib/`, `store/`, the derived-value computations inside page `index.tsx`
  containers) is cheap to test exhaustively and where the sharpest bugs hide (date math, currency
  formatting, filter logic, category icon lookup, reducers).
- **Components** should be tested the way a user uses them — render, query by role/label, fire
  events, assert on what's visible. Never assert on internal state or implementation detail.
- **The two architectural layers that are *pure display*** (`src/components/`,
  `pages/*/Name.desktop.tsx` / `Name.mobile.tsx`) are tested with props in, rendered output out —
  no network, no store, no router needed.
- **Page containers** (`pages/*/index.tsx`) are the one place business logic and data-fetching
  meet, so they're tested with the network mocked (MSW) and real Redux/TanStack Query providers,
  asserting on behavior (what renders, what gets called) rather than re-implementing the hooks.
- **End-to-end tests are expensive and few, and purely functional.** They answer "can a user log
  in / list transactions / add a transaction / manage categories" — not a restatement of what
  component tests already prove, and never an assertion on appearance (layout, styling,
  responsiveness stay out of E2E's job entirely).
- **No enforced coverage percentage.** A line-coverage gate rewards tests written to satisfy the
  gate, not tests that catch real regressions. Coverage is reported for visibility (and as a
  "something got missed" smell detector), not graded in CI.

## Tooling

| Concern | Choice | Why |
|---|---|---|
| Unit / component test runner | **Vitest** | Reuses `vite.config.ts` as-is (same plugins, same `@/` alias, same env handling) — zero parallel config to maintain. Jest would need its own transform/module-resolution setup duplicating what Vite already does. |
| DOM environment | **jsdom** | Covers all of this app's components (plain DOM forms/lists/Tailwind, no canvas/WebGL). Vitest's newer Browser Mode (real browser) is reserved for cases jsdom can't express — not needed here today. |
| Component queries | **@testing-library/react** + **@testing-library/user-event** + **@testing-library/jest-dom** | De facto standard; queries by role/label force accessible markup instead of brittle implementation-detail assertions. |
| Network mocking | **MSW (Mock Service Worker) v2** | Intercepts at the `fetch` layer, so `api/client.ts`'s real request/error-handling code runs in tests exactly as in production — unlike mocking `api/*.ts` functions directly, which would skip that code path entirely. Independent of `src/api/mock/data.ts` (that's a dev-only "design mode" stand-in — see below). |
| Coverage | **@vitest/coverage-v8** | V8's remapped coverage is now as accurate as Istanbul without Istanbul's Babel-instrumentation overhead. Report-only, no CI gate (see Philosophy). |
| E2E | **Playwright**, run against the real backend | Faster, more parallel-friendly, and less browser-locked than Cypress for a small engineering-led team. Runs against the `development/docker-compose.yml` stack (Postgres + Go server) rather than `VITE_DESIGN_MODE` — the entire point of E2E here is proving the real HTTP/auth/DB path works, which a mocked backend can't do. |

### Why not reuse `src/api/mock/data.ts` for test fixtures

That file exists to make the app browsable without a backend (`VITE_DESIGN_MODE`) and is
explicitly a dev tool, not a correctness contract (see `src/api/mock/README.md`'s "Don't" list).
Tying test assertions to it would mean a change to design-mode's in-memory arrays (fixing a
demo-data typo, say) could silently break unrelated tests. MSW handlers get their fixtures from a
small `src/test/fixtures.ts` shaped like the real API DTOs (`api/types.ts`) — a second, intentional
source of test data, not a reuse of the dev-mode one.

## Setup

```sh
npm install -D vitest jsdom @testing-library/react @testing-library/user-event @testing-library/jest-dom @vitest/coverage-v8 msw
npm install -D @playwright/test
npx playwright install --with-deps chromium
```

Add a `test` block to the existing `vite.config.ts` (Vitest reads it directly — no separate
config file needed):

```ts
/// <reference types="vitest/config" />
import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import { configDefaults } from 'vitest/config'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: { '@': path.resolve(__dirname, './src') },
    },
    server: {
      allowedHosts: env.VITE_DEV_ALLOWED_HOST ? [env.VITE_DEV_ALLOWED_HOST] : undefined,
    },
    test: {
      environment: 'jsdom',
      setupFiles: ['./src/test/setup.ts'],
      // `describe`/`it`/`expect` imported explicitly, not injected as globals — matches this
      // repo's no-magic-globals style. One consequence: React Testing Library's own auto-cleanup
      // only hooks into a *global* `afterEach`, so it silently never runs under this setting —
      // `src/test/setup.ts` calls RTL's `cleanup()` from an explicitly-imported `afterEach`
      // instead. Don't drop that call if you ever touch setup.ts.
      globals: false,
      // `tests/e2e/` is Playwright's own runner, not Vitest — without this, Vitest's default
      // include glob also picks up `tests/e2e/*.spec.ts` and crashes calling Playwright's
      // `test.describe()` outside Playwright's runner.
      exclude: [...configDefaults.exclude, 'tests/e2e/**'],
      coverage: {
        provider: 'v8',
        reporter: ['text', 'html'],
      },
    },
  }
})
```

New `package.json` scripts:

```json
"test": "vitest run",
"test:watch": "vitest",
"test:coverage": "vitest run --coverage",
"test:e2e": "playwright test"
```

Playwright gets its own config at `client/web/playwright.config.ts` (separate test runner, so it
lives outside `src/`), pointed at the dev server and expecting the docker-compose backend to
already be running:

```ts
import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:5173',
    reuseExistingServer: true,
  },
  use: { baseURL: 'http://localhost:5173' },
})
```

Document the manual precondition in `tests/e2e/README.md`: `docker compose -f ../../development/docker-compose.yml up -d`
(run from `client/web/`, same as every other script here — not from inside `tests/e2e/`) before
`npm run test:e2e`, plus how the suite provisions its test user (see Playwright section below).

## File layout & naming

Co-located, matching this repo's existing pattern of keeping a thing and its documentation
together (every folder already has its own `README.md` beside the code it describes):

- `src/lib/dates.ts` → `src/lib/dates.test.ts`
- `src/hooks/useCategories.ts` → `src/hooks/useCategories.test.ts`
- `src/components/TransactionRow.tsx` → `src/components/TransactionRow.test.tsx`
- `src/pages/add/index.tsx` → `src/pages/add/index.test.tsx`

Shared test infrastructure lives in `src/test/`:

- `src/test/setup.ts` — jest-dom matchers, MSW server lifecycle (`beforeAll`/`afterEach`/`afterAll`)
- `src/test/server.ts` — MSW `setupServer(...handlers)`
- `src/test/handlers.ts` — one MSW handler per `api/*.ts` resource, mirroring its routes
- `src/test/fixtures.ts` — fixture objects shaped like `api/types.ts` DTOs
- `src/test/renderWithProviders.tsx` — custom RTL `render()` wrapping `<Providers>` +
  `<MemoryRouter>`, for anything that needs Redux/TanStack Query/routing context

E2E specs live in a top-level `tests/e2e/` folder (sibling to `src/`), since Playwright is a
distinct runner with its own lifecycle: `tests/e2e/login.spec.ts`, `tests/e2e/add-transaction.spec.ts`, etc.

## What to test, by layer

**`lib/`** — plain unit tests, no rendering, no mocking. Pure functions in, assertions on output.
Priority: `dates.ts`, `currency.ts`, `categoryVisuals.ts` (the exact-name → icon map plus the
keyword/hash fallback), `cssVars.ts` if it has branching logic.

**`store/`** — `authSlice` is a handful of reducers/selectors; test the reducer directly
(`authReducer(state, action)`) rather than mounting a whole store. Low priority given how little
logic lives here per the state-management rules.

**`hooks/`** — `renderHook` from `@testing-library/react`, wrapped in a test `QueryClientProvider`
(fresh `QueryClient` per test, `retry: false` and `gcTime: 0` so failures surface immediately
instead of retrying/caching across tests), network mocked via MSW. Assert on `data`/`isError`
states and that mutations call `invalidateQueries` with the right keys (e.g.
`useCreateCategory`/`useUpdateCategory` in `useCategories.ts` invalidating `['categories', type]`).

**`components/` (base layer) and `pages/*/components/` (page-specific)** — RTL render with props,
`user-event` for interaction, assert via `getByRole`/`getByLabelText`/`getByText` — not
`getByTestId` unless a role-based query genuinely isn't available (e.g. a purely decorative
element). These are pure-display per the architecture rules, so no MSW/Redux/Router needed unless
the component itself takes a router-dependent prop (e.g. `SidebarLink`, `NavTab`).

**`pages/*/Name.desktop.tsx` / `Name.mobile.tsx`** — same as above: pure props-in/render-out,
since the architecture rule requires these to never diverge from their container in logic. A
divergence caught by a test here (e.g. desktop variant silently missing a field the mobile variant
shows) is exactly the kind of bug this layer split is supposed to prevent.

**`pages/*/index.tsx` containers** — the integration point. Render through
`renderWithProviders`, MSW mocks the network, assert on end-to-end behavior: given these mocked
API responses, does the right `*ViewProps` reach the view (verified via what the view renders),
do interactions trigger the right mutation calls, does an error response show the right state.
This is where most business-logic bugs will surface, since this is "the only place business logic
lives for that page" per the architecture rules.

**`ResponsiveView` / `useViewport`** — test both branches explicitly: mock the viewport
hook/media query to force `'mobile'` and `'desktop'` and assert the correct lazy variant mounts.

**`components/icons.tsx`** — not tested directly (purely decorative SVG paths); covered
indirectly wherever a component renders using it.

## Playwright: pure functionality, critical flows only

E2E here tests **functionality, not appearance** — every spec asserts on behavior a user depends
on (did the right data get saved/shown, did navigation go where it should), never on pixels,
layout, or styling. No screenshot/visual diffing, no asserting on CSS classes or exact spacing —
that's what unit/component tests and (if ever added) a dedicated visual-regression tool are for.
Queries stay role/label-based (`getByRole`, `getByLabel`), same as the RTL conventions above, so a
spec keeps working across a restyle as long as the underlying behavior hasn't changed.

Keep this to the handful of flows where only a real browser + real backend adds value — each
phrased as "can a user actually do X":

1. **Can a user log in?** — good credentials land on the authenticated shell; bad credentials
   show an error and keep them on the login screen; an unauthenticated visit to a protected route
   redirects to login.
2. **Can a user list/view their transactions?** — after seeding a known set of transactions for
   the test user, the transactions page lists them with the right amounts/categories/dates, and
   basic filtering (e.g. by category or date range) narrows the list correctly.
3. **Can a user add a transaction?** — fill the form, submit, see it appear in the transaction
   list and reflected in the home summary (proves the full round trip: auth header, persistence,
   cache invalidation, rendering).
4. **Can a user manage categories?** — create/rename/deactivate a category in Settings and see it
   reflected in the Add Transaction category picker.
5. **Can a user manage currencies?** — add/select a currency in Settings and see it available
   where currency is chosen elsewhere (e.g. Add Transaction).

Each spec needs its own isolated test user/data rather than sharing fixtures with other specs (to
stay parallel-safe) — provision via the same `POST /api/v1/users` endpoint the root README already
documents for manual account creation, called from a Playwright `globalSetup` or per-spec
`beforeAll`.

Explicitly **not** covered by E2E: anything already proven by a page-container test (form
validation edge cases, filter combinations, empty states) and anything visual (responsive
layout/breakpoints, Tailwind styling, icon rendering) — re-proving functional behavior against a
real browser and real DB just adds slow, flaky duplicates of fast, deterministic tests, and visual
concerns belong to a different kind of test entirely.

## Manual verification: PWA install & offline

Install and offline behaviour cannot be fully automated (real home screens, real Safari dialogs,
real launchers). The automated coverage for this feature is the unit/hook/component tests beside
`src/lib/installEnvironment.ts`, `src/hooks/useInstallPrompt.ts`, `src/components/Offline*.tsx`
and `src/pages/settings/`, plus `tests/e2e/offline.spec.ts` — which covers the offline gate and
focus restoration in a real browser (via `context.setOffline`). It does **not** cover the install
flow or the service worker; those are checked by hand against `npm run preview`, once, per
release that touches the PWA layer. Each row is a five-minute check; record the outcome in the PR.

| Platform | Check | Pass condition |
|---|---|---|
| Android Chrome | Open the preview build over HTTPS (or a LAN tunnel), use the app ~30 s, open Settings | "Install app" appears; the dialog shows the wallet icon and "Joint Wallet Watch"; the installed launcher icon is not letterboxed into a white circle |
| Android Chrome (installed) | Launch from the launcher with airplane mode on | The app opens and shows the offline screen — no browser error page, no blank screen |
| iOS Safari | Settings → Install app | The sheet's steps match what the OS actually shows, including the "Open as Web App" toggle |
| iOS (installed) | Launch from the home screen; then airplane mode and relaunch | No Safari chrome; content not clipped under the status bar; the offline screen appears |
| macOS Safari 17+ | Settings → Install app → follow the inline steps | "Add to Dock" produces a Dock entry with the correct icon; the window opens standalone |
| Desktop Chrome/Edge | Address-bar install icon **and** the Settings row | Both paths work; the installed window is standalone; the row is hidden once installed |
| Any target | While installed and online, navigate the app and watch Cache Storage in DevTools | Only the shell entries and icons are present; no `/api/` or API-origin entries exist |

After any change to the build config, run `npm run build` and read the injected precache manifest
in `dist/sw.js`: it must list `index.html`, the entry `assets/app-*.js`, the entry CSS, the icons
and the manifest — and must **not** list any page or layout chunk (`Home.mobile-*.js`,
`Settings.desktop-*.js`, `pages/*/index.tsx`'s `index-*.js`, …). A malformed list is the
difference between a working offline screen and a browser error page.

## Rollout plan

Suggested order — each phase is independently shippable:

1. **Tooling setup** — install deps, add Vitest config, `src/test/` scaffolding (setup, MSW
   server/handlers, `renderWithProviders`). Write one trivial test per new piece to prove it
   works end to end (e.g. one `lib/dates.test.ts`, one MSW-backed hook test).
2. **`lib/` + `store/`** — fast, isolated, no mocking infrastructure needed beyond what's already
   set up. Good first real coverage pass.
3. **`hooks/`** — proves the MSW setup against every resource (`categories`, `currencies`,
   `transactions`, `summary`, `auth`, `users`).
4. **`components/` base layer + `ui/`** — the most reused code; bugs here ripple widest.
5. **Page containers + view variants**, one page at a time, starting with `pages/add/` (most
   complex form logic) and `pages/transactions/` (filtering logic).
6. **Playwright critical flows** — once the above exist, so E2E failures are more likely to mean
   "real integration bug" rather than "a bug component tests would've caught faster."

## Explicitly out of scope (for now)

- **CI enforcement** — not covered by this doc; running these locally (and in editors via Vitest
  watch mode) is the near-term plan. Wiring a CI workflow to run `lint` / `tsc -b` / `test` /
  `test:e2e` on PRs is a natural follow-up once the suite exists, but is a separate decision.
- **Visual regression testing** (e.g. Chromatic, Playwright screenshot diffing) — not proposed
  here; revisit if the hand-rolled icon/Tailwind styling layer proves to regress silently often
  enough to justify it.
- **Numeric coverage gate** — see Philosophy above.
