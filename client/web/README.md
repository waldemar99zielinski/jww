# Joint Wallet Watch — Web Client

A React + Vite client, built as a static bundle. Authenticates against the Go backend's
username/password + JWT auth (`POST /auth/login`), with protected routes gating everything
behind a signed-in session. Server data (transactions, categories, currencies, summaries)
is managed by TanStack Query; cross-cutting session state lives in Redux (`src/store`).

Every top-level page ships a mobile and a desktop layout variant (e.g. `src/pages/home/Home.mobile.tsx`
/ `Home.desktop.tsx`), picked at runtime by `useViewport()` and lazy-loaded independently via
`<ResponsiveView>` — a phone never downloads the desktop layout code, and vice versa. Each page
is also lazy-loaded at the route level (`src/App.tsx`), so the initial bundle only contains the
shell and whichever page is first visited.

## Branding

`src/components/Logo.tsx` (plus `LogoMark.tsx`/`LogoWord.tsx`) is the single source of truth for
the wordmark. `<Logo animated />` plays a one-time reveal (shows "JWW" — the bold first letters
of Joint/Wallet/Watch — then expands into the full name) and is used on the login screen;
`<Logo />` renders already-resolved for reuse elsewhere (e.g. the desktop sidebar). `<LogoMark />`
is the compact "JWW" badge used where space is tight (e.g. the mobile home header).

## Creating a user

There's no sign-up page yet. Create an account directly against the API before your first login:

```sh
curl -X POST "$VITE_API_BASE_URL/api/v1/users" \
  -H 'Content-Type: application/json' \
  -d '{"username": "waldemar", "password": "change-me-please", "display_name": "Waldemar"}'
```

## Development

```sh
npm install
npm run dev
```

Set `VITE_API_BASE_URL` in `.env` to point at the backend (see `.env.example`).

In development the client calls the Vite dev server on the **same origin** and Vite proxies
`/api/**` to the backend, so there is no CORS setup to do. The proxy target is `VITE_API_PROXY_TARGET`,
falling back to `VITE_API_BASE_URL` (and then `http://localhost:8001`) — see the `server.proxy`
block in `vite.config.ts`. Production builds are unaffected: they still call the real API origin
from `VITE_API_BASE_URL`.

### Testing on a phone (ngrok)

Installability and service workers need a secure context, so expose the dev server over HTTPS with
[ngrok](https://ngrok.com):

```sh
npm run dev                 # Vite on :5173, proxying /api to your backend
ngrok http 5173             # in another terminal — gives you an https://…ngrok-free.app URL
```

Open the ngrok URL on the device. Because `/api` is proxied through the same tunnel, the app and
the API are both reachable through that one URL — no second tunnel, no CORS.

- ngrok hostnames are allowed by default (`server.allowedHosts` in `vite.config.ts`). For any
  other tunnel host, set `VITE_DEV_ALLOWED_HOST`.
- To keep HMR alive over the HTTPS tunnel, set `VITE_DEV_HMR_CLIENT_PORT=443`.
- For the real (production) service worker and install flow, run `npm run build && npm run preview`
  and tunnel the preview port instead — `preview.proxy` forwards `/api` the same way. Build with an
  empty `VITE_API_BASE_URL` (`VITE_API_BASE_URL= npm run build`) so the built client also calls
  `/api` on its own origin.

## Build

```sh
npm run build
```

Produces a static `dist/` directory — there's no server runtime to deploy. Routing uses
`react-router-dom`'s `BrowserRouter`, so the static host must rewrite all paths to `index.html`.

## Lint

```sh
npm run lint
```

## Project map

Every folder under `src/` has its own `README.md` explaining what it's for and how to work in
it — start there before adding code to a folder you haven't touched before. High level:

- `api/` — typed fetch wrappers per resource, mirroring the server's DTOs. `api/mock/` backs
  design mode (see below).
- `store/` — Redux: session/auth state only. Server data never lives here.
- `hooks/` — one module per server resource, wrapping TanStack Query (`useQuery`/`useMutation`).
- `lib/` — small pure utilities (formatting, dates, viewport detection). No business logic.
- `components/` — the **base layer**: generic, reusable, pure-display components only
  (`ui/` is the design-system primitive layer within it). See "Component architecture" below.
- `pages/` — one folder per route. See "Page architecture" below.
- `data/` — explicitly-labeled placeholder data for features not yet modeled on the server.

## Page architecture

Every route under `pages/` (except `login/`, which has one layout for every viewport) follows
the same container/view split:

- `index.tsx` — the container. Owns all hooks (queries, mutations, local state), computes
  derived values, and assembles one `*ViewProps` object. This is the *only* place business logic
  lives for that page.
- `types.ts` — the `*ViewProps` interface shared by both view variants.
- `Name.desktop.tsx` / `Name.mobile.tsx` — pure presentational components. Same props in, same
  behavior out; they should only differ in layout (grid vs. stack, sizing, what's visible above
  the fold), never in logic. Picked at runtime by `<ResponsiveView>` (`src/components/ResponsiveView.tsx`)
  based on `useViewport()`, and each is `lazy()`-loaded independently — a phone never downloads
  the desktop chunk and vice versa.

When adding a new page, copy this shape. If a desktop/mobile pair's JSX diverges by more than
layout, that logic belongs back in `index.tsx`, not duplicated across both variants.

## Component architecture

There are two layers of components, and every new component belongs in exactly one of them:

- **`src/components/`** — the base layer. Generic, reusable, **pure-display** components: no
  business logic, no API/store/TanStack Query access, no knowledge of any specific page. They
  take data and callbacks as props and render. `src/components/ui/` is the design-system
  primitive layer within this (buttons, inputs, switches, …); the rest of `src/components/` is
  app-shell chrome (`AppShell`, nav, `ResponsiveView`, …) and a small number of components
  genuinely reused by more than one page (e.g. `TransactionRow`, used by both `pages/home/` and
  `pages/transactions/`).
- **`pages/<page>/components/`** — composed, page-specific components, built from the base
  layer above. Each one is used by exactly one page. This is also where that page's wiring
  happens: a `pages/<page>/components/Foo.tsx` component is still pure display (it receives
  props, it doesn't fetch), but *which* page's data it displays is baked into its props' shape —
  that's what makes it page-specific rather than generic. The actual fetching/mutating only ever
  happens one level up, in that page's `index.tsx` container.

**Before writing a new component, check what already exists** — first in `src/components/ui/`
and `src/components/`, then in the current page's own `components/` folder. Most new UI needs are
a composition of something that's already there; reaching for an existing primitive (or adding
one missing prop to it) beats writing a parallel one-off. If you build something in a page's
`components/` folder and later find a second page wants it too, promote it to `src/components/`
instead of copying it.

**One component per file**, as a default rule everywhere in `components/` and `pages/*/components/`.
A private subcomponent that exists purely to keep its parent's JSX readable, is never imported
anywhere else, and has no independently meaningful props may stay in the same file — otherwise
split it into its own file, even if it's small (e.g. `components/nav/NavTab.tsx` is split out of
`BottomNav.tsx`; `pages/transactions/components/FilterButton.tsx` is split out of
`TransactionFiltersDropdown.tsx`). The one deliberate, documented exception is
`src/components/icons.tsx` — a flat icon library behind one shared wrapper, where one file per
icon would hurt more than help.

## State management rules

- **Server data** (transactions, categories, currencies, summaries, users) → TanStack Query via
  a hook in `hooks/`, never Redux. Query keys start with the resource name; mutations
  `invalidateQueries` the keys they affect.
- **Session/auth state** → Redux (`src/store`), because it needs to be read outside the React
  tree (e.g. `lib/authToken.ts`, `App.tsx`'s boot gate) before the first render.
- **Everything else** (form drafts, filter drafts, UI open/closed state) → local `useState` in
  the page container, passed down as props. Don't reach for Redux or a new global store for this.

## Styling rules

- Tailwind utility classes only — no CSS-in-JS, no new stylesheets.
- Build forms and buttons from `components/ui/` primitives (`Button`, `TextInput`, `SelectInput`,
  `Switch`, `SearchableSelect`, `BottomSheet`, `DateInput`, …). Don't drop down to a raw
  `<button>`/`<input>` unless the primitive genuinely doesn't fit — if you find yourself doing
  that more than once, add the shape to `components/ui/` instead.
- Icons come from `src/components/icons.tsx` (one hand-drawn SVG set behind a single `Icon`
  wrapper). Add new icons there rather than inlining SVG elsewhere.

## Design mode (develop without a backend)

Set `VITE_DESIGN_MODE=true` in `.env` and restart the dev server to browse the whole app against
in-memory mock data (`src/api/mock/data.ts`), with auth bypassed. Every `api/*.ts` resource file
branches on `isDesignMode()` (`src/lib/designMode.ts`) to serve mock data instead of calling the
network — when adding a new API function, follow that same branch so it works in both modes.
Keep `api/mock/data.ts`'s filtering logic in sync with the real API's filter semantics (see
`api/transactions.ts`'s `TransactionFilters`) so design mode doesn't silently behave differently.

## PWA / install (online-only)

The client is installable. It ships a web app manifest and a Workbox service worker via
`vite-plugin-pwa` (configured in `vite.config.ts`), so "Joint Wallet Watch" can be added to a
phone's home screen or a desktop Dock and launched in a standalone window.

**It is not an offline app.** The service worker precaches the *app shell only* — `index.html`,
the entry JS/CSS, the icons and the manifest — and nothing else. No page/layout chunk is
precached (a phone still downloads only the layout variant it renders), and no API response is
ever cached or intercepted. The request-class contract is expressed directly in `vite.config.ts`:
the narrow `globPatterns`, `navigateFallback: 'index.html'`, and the deliberate absence of any
`runtimeCaching` block.

Because there is no offline data, a device with no connection gets a full-screen **"You're
offline"** panel covering the whole app — `/login` included — with a single *Retry*. The routed
tree stays mounted underneath but `inert`, so a half-typed transaction or a paused TanStack Query
survives a short drop instead of being destroyed by an unmount. See
`src/components/OfflineGate.tsx` and `OfflineScreen.tsx` (state comes from
`src/hooks/useOnlineStatus.ts` / `src/lib/pwa.ts`).

- **Install entry point:** Settings → "Install app", both layouts. Chromium gets the native
  install dialog (`beforeinstallprompt`); iOS/iPadOS and macOS Safari get written steps (Safari
  cannot be prompted); a browser with no install path, or an in-app webview, shows a passive row
  that explains why instead of a dead button. See `src/hooks/useInstallPrompt.ts` and
  `src/lib/installEnvironment.ts`.
- **iOS safe areas:** `index.html` sets `viewport-fit=cover`, so `env(safe-area-inset-*)` reports
  the real notch/home-indicator insets. The mobile tab bar takes its height and bottom gap from
  `--bottom-nav-height` / `--bottom-nav-bottom-gap` (`src/index.css`); the gap is half of the
  OS-reported inset — `0` on home-button iPhones, ~17px on Face ID iPhones — so the tab row
  clears the home bar without an oversized empty strip. `AppShell` clears the same values.
- **Updates are silent:** `registerType: 'autoUpdate'` activates a new deploy and reloads an open
  app without asking. A reload can discard unsaved form state (e.g. a draft in `pages/add/`).
- **Deployment prerequisites (host-side):** the production origin must serve HTTPS; serve
  `manifest.webmanifest` with a JSON content type; keep the existing SPA rewrite to
  `index.html`; do not serve `sw.js` with a long-lived cache header, or updates cannot ship.

App icon source art is `public/icons/app-icon.svg` (plus `app-icon-maskable.svg`); every PNG in
`public/icons/` is generated from those two — see `public/README.md` for the sizes, the tier
transforms and the `scripts/generate-icons.mjs` rule (manual, never part of the build).

## Known gaps / placeholders

- **Automated tests exist** — Vitest + React Testing Library + MSW (co-located `*.test.*` files,
  shared infrastructure in `src/test/`) and Playwright E2E (`tests/e2e/`). Run them with
  `npm run test` and `npm run test:e2e`. `TESTING.md` is the strategy doc; it also lists what is
  deliberately still out of scope (CI enforcement, visual regression, a coverage gate).
- `pages/budgets/` and its `components/` are a visual prototype only — gated behind a
  "Proposed, not built yet" `<StatusBanner>` and backed by `data/sampleBudgets.ts`, not the API.
