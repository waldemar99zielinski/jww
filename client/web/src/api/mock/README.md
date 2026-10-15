# api/mock/

In-memory fake backend used by design mode (`VITE_DESIGN_MODE=true` in `.env`, see root
README). `data.ts` holds seed arrays (users, categories, currencies, transactions) plus a
`mockApi` object with one method per resource operation, and `mockDelay()` to simulate network
latency so loading states are visible while designing.

## Conventions

- `mockApi.<operation>` signatures should match the real `api/*.ts` function's parameters and
  filtering semantics as closely as possible — design mode exists so the whole app can be
  reviewed without a backend, and a mock that behaves differently from the real API defeats
  that purpose.
- Mutations (`createTransaction`, `updateCategory`, …) mutate the in-memory arrays directly and
  return the updated record, mirroring what the real API would persist and return.

## Testing

Not used by automated tests — this is a dev-only stand-in for design mode, not a test fixture
source (see `TESTING.md`'s "Why not reuse `src/api/mock/data.ts` for test fixtures"). Tests get
their fake data from `src/test/fixtures.ts` + MSW handlers instead, kept independent on purpose so
a design-mode data tweak can't silently break a test.

## Don't

- Don't let this file grow business logic beyond "filter/sort/mutate an in-memory array the way
  the server would" — it's a stand-in for the backend, not a second implementation of app logic.
- Don't forget to update this file when a real resource file (`api/*.ts`) gains a new filter or
  field — see the top-level `api/README.md` for why.
