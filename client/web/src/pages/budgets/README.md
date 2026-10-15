# pages/budgets/

**Visual prototype only — not a real feature.** Renders a `<StatusBanner>` reading "Proposed —
not built yet" and is backed entirely by `data/sampleBudgets.ts`, not any
server call or `hooks/` module. Otherwise follows the same container/view structure as every
other page (see `pages/README.md`).

## Files

- `components/BudgetCard.tsx`, `components/GoalCard.tsx` — presentational, page-local components.
  `BudgetCard` is the sole consumer of `lib/placeholderCategories.ts` — a fixed 5-category set
  that exists only for this prototype and is unrelated to real server categories (see
  `lib/categoryVisuals.ts` for those, used by every real, data-backed page).

## Testing

Low priority, matching this page's prototype status: a smoke-level render test (the
`<StatusBanner>` shows, cards render from `data/sampleBudgets.ts`) is
enough — don't invest in exhaustive container/E2E coverage for a page that isn't wired to real
data yet. Revisit once budgets are actually modeled on the server.

## Don't

- Don't copy this folder's data-sourcing approach (static sample arrays) for a new real feature —
  copy `pages/home/` or `pages/transactions/` instead, where data comes from `hooks/`.
- Don't wire this page to real API data without confirming budgets are actually being built on
  the server — check with the project owner first.
