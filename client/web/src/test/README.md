# test/

Shared test infrastructure (not a source folder — nothing here ships in the app bundle). See the
root `TESTING.md` for the full strategy this backs.

- `setup.ts` — Vitest's global setup file (wired via `vite.config.ts`'s `test.setupFiles`):
  jest-dom matchers, MSW server lifecycle (`listen`/`resetHandlers`/`close`), and an explicit RTL
  `cleanup()` in `afterEach`. That last one is required, not cosmetic: RTL's own auto-cleanup only
  registers against a *global* `afterEach`, and this project runs Vitest with `test.globals: false`
  — without the explicit call, one test's rendered DOM leaks into the next within the same file.
- `server.ts` — the MSW `setupServer` instance, built from `handlers.ts`.
- `handlers.ts` — one MSW handler per `api/*.ts` endpoint, mirroring its real path/method/query
  params. Kept deliberately independent of `api/mock/data.ts` (that's design mode's dev-only fake
  backend) — see `api/mock/README.md` and `TESTING.md` for why.
- `fixtures.ts` — fixture objects shaped like `api/types.ts` DTOs, consumed by `handlers.ts` and
  directly by tests that need a fixture value to assert against.
- `renderWithProviders.tsx` — a custom React Testing Library `render()` that wraps `<ReduxProvider>`
  + `<QueryClientProvider>` + `<MemoryRouter>`, each freshly constructed per call (a fresh
  `QueryClient` with `retry: false, gcTime: 0` so failures surface immediately instead of
  retrying/caching across tests, and a fresh Redux store so tests don't share state). Use this
  instead of RTL's plain `render()` for anything that needs those providers (page containers,
  anything using `hooks/`, anything routed).
- `normalizeText.ts` — strips the non-breaking space `lib/currency.ts`'s `Intl.NumberFormat`
  output embeds before the currency symbol. RTL's `getByText`/`findByText` normalize the *DOM
  side*'s whitespace but never the query string you pass in, so asserting rendered currency text
  against a raw `formatAmount(...)` result never matches without wrapping the expected string in
  this first.

## Don't

- Don't import `api/mock/data.ts` fixtures into `fixtures.ts` or a test — design mode's mock data
  is a dev tool, not a test-correctness contract, and the two are intentionally allowed to drift.
- Don't add a new MSW handler without keeping its path/params in sync with the real `api/*.ts`
  function it mirrors — a handler that diverges from the real endpoint makes tests pass against a
  backend contract that doesn't exist.
