# tests/e2e/

Playwright specs — see root `TESTING.md`'s "Playwright: pure functionality, critical flows only"
section for what belongs here and why. Every spec asserts on behavior only (did the right data
get saved/shown, did navigation go where it should), never on appearance.

## Prerequisites

1. **Postgres + the real Go backend running.** Start the dev stack:
   ```sh
   docker compose -f ../../development/docker-compose.yml up -d
   ```
   then run the Go server separately (see `server/README.md`) against it. Playwright's own
   `webServer` config only starts the Vite dev server (`npm run dev`) — it does not start Postgres
   or the backend for you.
2. **A pre-existing test account**, via two env vars:
   ```sh
   export PLAYWRIGHT_TEST_USERNAME=...
   export PLAYWRIGHT_TEST_PASSWORD=...
   ```
3. If your backend isn't on the default `http://localhost:8001` (check your `.env`'s
   `VITE_API_BASE_URL` — Playwright's Node process doesn't read `.env` automatically), also set:
   ```sh
   export PLAYWRIGHT_API_BASE_URL=http://localhost:8001   # match your .env
   ```

Then: `npm run test:e2e`.

### Test user — why not self-register

Every spec was designed to register a fresh, disposable user per run (`POST /api/v1/users`, as
documented in the root README's "Creating a user" section). While implementing this suite we
found that endpoint **isn't actually wired to a route** on the server — `SignupWorkflow` is built
in `cmd/server/deps.go` but no router mounts it (confirmed: it's referenced nowhere else in
`server/internal/infrastructure/http_server/routers/`). That's a real bug affecting real
onboarding too, not just this test suite — flagged separately, not fixed in this round.

Until it's fixed, provision the one account these env vars point to by whatever means currently
works in your environment (the dev DB may already have a seeded user outside migrations — ask the
project owner — or insert one directly). Once `POST /api/v1/users` is wired up, these specs should
switch back to self-registering a fresh user per run (see the git history of `support/testUser.ts`
for the shape that approach had — `uniqueTestUser`/`registerTestUser`) for better isolation.

## Shared data caveat

Categories and currencies are **global**, not per-user (see
`server/migrations/00006_seed_transaction_categories.sql` — no `user_id` column). Categories are
pre-seeded so every test user already sees them; currencies are not, so `support/currencies.ts`'s
`ensureCurrency` creates one if none exist. The `categories.spec.ts` and `currencies.spec.ts`
specs each create a uniquely-named category/currency (via `support/testUser.ts`'s `uniqueLabel`)
rather than cleaning up after themselves — there's no delete endpoint for either, so this data
accumulates across runs. Run this suite against a disposable/dev database, not production.

## Files

- `support/testUser.ts` — `getTestUser()` reads the env vars above; `uniqueLabel()` generates
  collision-free names for data a spec creates.
- `support/auth.ts` — `loginViaUi()`, the shared login-form interaction.
- `support/apiAuth.ts` — `getAccessToken()`/`authHeader()`, for specs that need to seed data via
  the API directly (faster and more reliable than doing it through the UI) before driving the UI.
- `support/categories.ts`, `support/currencies.ts`, `support/transactions.ts` — typed API helpers
  for seeding/looking up fixture data for a spec.
- `login.spec.ts`, `transactions.spec.ts`, `add-transaction.spec.ts`, `categories.spec.ts`,
  `currencies.spec.ts` — the five critical flows from `TESTING.md`.
