# pages/home/

The dashboard/home screen: current-month balance, income/expense totals, category breakdown,
and the 3 most recent transactions. Standard container/view structure (see `pages/README.md`).

## Files

- `components/CategoryBreakdown.tsx`, `components/HeroBalanceCard.tsx` — presentational,
  page-local components. Both only render pre-computed values passed down as props; the
  aggregation math lives in `index.tsx` (see below).
- `pages/home/` also reuses `components/TransactionRow.tsx` from the shared base layer — it's
  one of the few components reused across two pages (`pages/home/` and `pages/transactions/`),
  which is why it lives in `src/components/` rather than here or in `pages/transactions/`.

## Things to know

- `index.tsx` computes the current month's date range once per mount (`useState(() => new
  Date())`, not `useMemo` — a stable "now" snapshot for the page's lifetime, matching the lazy
  `useState` initializer pattern used elsewhere, e.g. `providers.tsx`'s QueryClient).
- Category breakdown math (grouping expense transactions by `category_name`, computing bar
  percentages relative to the largest category) lives in `index.tsx`, not in
  `components/CategoryBreakdown.tsx` — that component only renders pre-computed values.
- Sources data via `useSummary` and `useTransactions` (`hooks/`) — never fetches directly.
- `useSummary` is called **without `user_id`**, so the balance/income/expense totals are
  household-wide (no user scope). The `useTransactions` query on this page still scopes to the
  signed-in user (`user_id: user?.id`), so the recent list and category breakdown stay per-user.

## Testing

- Container (`index.tsx`): this is where the real logic lives (current-month range, category
  breakdown aggregation/percentages), so it's the priority — MSW-mock `useSummary`/
  `useTransactions` responses and assert the computed values handed to
  `CategoryBreakdown`/`HeroBalanceCard` (via what they render) are correct, including an
  empty-transactions edge case.
- View variants + `components/`: pure render tests, pre-computed values in.
- `TransactionRow` (shared, lives in `src/components/`) is tested once there, not duplicated here.

## Don't

- Don't move the breakdown/aggregation math into the view components — keep it in `index.tsx`
  so both desktop and mobile variants render identical derived data.
