# pages/add/

The Add Transaction flow. Follows the standard container/view pattern (see `pages/README.md`):
`index.tsx` owns all form state (amount, category, date/time, receipt, payer override) and
assembles `AddTransactionViewProps`; `AddTransaction.desktop.tsx`/`.mobile.tsx` render it.

## Files

- `components/` — presentational components used only by this page: `AmountHero`,
  `DateTimeFields`, `MoreOptionsPanel`, `QuickFieldsRow`, `ReceiptBanner`, `ReceiptPicker`,
  one component per file, built from `components/ui/` primitives
  (`Button`, `FieldLabel`, `SelectInput`, `TextInput`, `Switch`) — check that folder before adding
  a new raw form element here.
- `categoryOptions.ts` — a data-mapping helper (`toCategoryOptions`), not a component, so it sits
  next to `index.tsx`/`types.ts` rather than in `components/`.

## Things to know

- `AddTransaction.desktop.tsx` and `.mobile.tsx` are the most duplicated pair in the codebase —
  near-identical JSX trees differing mainly in wrapper layout and a handful of prop values
  (e.g. `AmountHero boxed` vs `fullWidth`). This is a known, accepted tradeoff under the
  project's "duplicate JSX, diverge only in layout" convention — if these two pages need to
  diverge further, consider extracting the shared subtree into one component parameterized by
  the few props that differ, rather than letting the duplication grow unchecked.
- Receipt picking (`handlePickReceipt` in `index.tsx`) is a fake OCR flow — it fills in a canned
  amount/name after a timeout, not a real scan. See `components/ReceiptPicker.tsx`.

## Testing

- Container (`index.tsx`): MSW-mock the `categories`/`currencies`/transaction-create endpoints;
  test that submitting the form calls the create mutation with the right payload shape, and that
  the fake-OCR `handlePickReceipt` flow fills the expected fields after its timeout — use fake
  timers for that one rather than a real delay.
- View variants: pure render tests per `pages/README.md`'s pattern — same `AddTransactionViewProps`
  in, assert both variants render the same fields/values (this is the pair most likely to drift
  silently, per the "most duplicated pair in the codebase" note above).
- This page owns the **"can a user add a transaction"** Playwright flow from `TESTING.md` — that
  spec exercises the real submit → list round trip; don't duplicate its assertions here, just the
  component-level behavior MSW can prove faster.

## Don't

- Don't add new business logic directly to the `.desktop.tsx`/`.mobile.tsx` files or to a
  `components/` file — put it in `index.tsx` and pass it down as a prop, same as everything
  else here.
- Don't import a `components/` file from outside `pages/add/` — if it turns out to be useful
  elsewhere, promote it to `src/components/` instead of reaching across page folders.
