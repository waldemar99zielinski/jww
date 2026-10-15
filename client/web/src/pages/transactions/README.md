# pages/transactions/

Transaction history with filtering (type/user/category/name/date range). Standard
container/view structure (see `pages/README.md`); `index.tsx` owns every filter as its own
`useState` field and passes applied values + setters down.

## Files

- `components/TransactionFiltersDropdown.tsx` — the filter popover/bottom-sheet. Holds its own
  **draft** filter state separate from the page's **applied** filter state passed in via props,
  so edits don't apply until "Apply" is pressed (or the popover closes).
- `components/FilterButton.tsx` — the pill button that opens the dropdown, split into its own
  file per the one-component-per-file rule (it used to be a private function inside the
  dropdown's file).
- `components/TransactionFiltersSummary.tsx` — the compact "Showing: …" chip row shown when
  filters are active. Split from the dropdown into its own file since it's a sibling component,
  not a sub-part of it (they're both rendered independently from `Transactions.desktop.tsx`/
  `.mobile.tsx`).
- `filterLabels.ts` — the `TransactionFilterValue → label` map shared by the dropdown and the
  summary chip row; lives next to `types.ts` rather than inside either component file since both
  need it independently.

## Things to know

- Filter values map 1:1 to `api/transactions.ts`'s `TransactionFilters` — if you add a new
  filterable field, add it in both places (plus `api/mock/data.ts`'s mock filtering).

## Testing

- Container: MSW-mock `useTransactions` with a fixed set of filter combinations; assert the
  applied filters produce the right query params (see the 1:1 mapping to `TransactionFilters`
  noted above) and the right list renders.
- `TransactionFiltersDropdown`: test its draft-vs-applied state split directly with `user-event` —
  edit a draft filter, cancel/close without applying, and assert the page's applied filters (and
  thus the query) didn't change; then repeat with "Apply" and assert they did.
- This page owns the **"can a user list/view their transactions"** Playwright flow from
  `TESTING.md`, including a basic filter-narrows-the-list case.

## Don't

- Don't move draft-filter state up into `index.tsx` — it's intentionally scoped to
  `TransactionFiltersDropdown` so opening the popover, editing, and cancelling doesn't touch the
  applied filters or retrigger the transactions query until "Apply" is pressed.
- Don't merge `FilterButton`/`TransactionFiltersSummary` back into `TransactionFiltersDropdown`'s
  file — each is independently renderable and should stay its own file.
