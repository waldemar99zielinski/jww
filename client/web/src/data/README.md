# data/

Explicitly-labeled placeholder/sample data for features that aren't modeled on the server yet.
Every file here starts with a comment saying so — keep that convention for any new placeholder
data you add.

## Files

- `sampleBudgets.ts` — feeds `pages/budgets/` only, which is gated behind a visible "Proposed —
  not built yet" `<StatusBanner>`. Low risk: nobody can mistake it for a real feature.

## Testing

Not unit tested — these are static arrays, not logic. Don't write a test asserting on
`sampleBudgets.ts`'s exact contents (that's the implementation-detail testing `TESTING.md`'s
philosophy explicitly avoids); test the *components* that consume them (`pages/budgets/`) instead.

## Don't

- Don't import from `data/` into anything that isn't explicitly a known-placeholder feature —
  if real data is needed, add a server resource + `api/`/`hooks/` module instead.
- Don't add new placeholder data without a header comment stating it's a placeholder and what
  page(s) consume it, so the next person who finds it understands the risk immediately.
