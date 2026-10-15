# pages/settings/currencies/

Manage currencies: list them, create a new one (3-letter code + display name), and mark one as
the base currency. Standard container/view structure (see `pages/README.md`).

## Things to know

- The first currency ever created is automatically flagged `isBase: true` (`index.tsx`'s
  `handleCreate`: `isBase: (currencies?.length ?? 0) === 0`) — there's always exactly one base
  currency once any currency exists.
- `handleMakeBase` re-sends the currency's existing `name` alongside `isBase: true` — same
  full-field-set update pattern as `pages/settings/categories/`.

## Testing

- Container: MSW-mock the currencies list/create endpoints; test the two behaviors called out
  above directly — `isBase: true` auto-set on the first-ever currency, and `handleMakeBase`
  resending the existing `name` alongside `isBase: true`.
- Also test `handleCreate`'s validation gate (3-char code, non-empty name) rejecting invalid input
  before it ever reaches the mutation.
- This page owns the **"can a user manage currencies"** Playwright flow from `TESTING.md`.

## Don't

- Don't allow creating a currency without going through `handleCreate`'s validation (3-char code,
  non-empty name) — the view components should keep relying on the container to gate the
  create action (`creating`/disabled state) rather than re-validating independently.
