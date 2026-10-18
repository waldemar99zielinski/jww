# lib/

Small, pure, single-responsibility utilities. No React, no API calls, no app/business logic —
if a function needs a hook or touches `fetch`, it belongs in `hooks/` or `api/`, not here.

## Files

- `currency.ts` — amount formatting.
- `amount.ts` — the amount *input's* display↔wire helpers: `sanitizeAmountInput` (keystrokes),
  `normalizeAmountInput` (blur), `amountToCents` (display string → whole minor units for the API),
  and `decimalStringToAmountDisplay` (server decimal → the comma display form). Used by the shared
  `components/AmountHero.tsx` and by the Add Transaction container.
- `dates.ts` — date formatting/parsing helpers used across pages.
- `cssVars.ts` — typed helper for setting CSS custom properties from inline `style`.
- `viewport.ts` / `designMode.ts` / `authToken.ts` / `pwa.ts` — single-flag/value accessors backing
  `useViewport`, `isDesignMode()`, the stored auth token, and online/standalone status +
  service-worker registration respectively.
- `installEnvironment.ts` — the install decision tree as pure functions (`detectPlatform`,
  `isInAppWebView`, `resolveInstallUiState`): maps a captured UA + browser snapshot to what the
  Settings install row should render, so every branch is unit-testable without a browser. The
  imperative `beforeinstallprompt` API lives in `hooks/useInstallPrompt.ts`, not here.
- `categoryVisuals.ts` — resolves an icon + color for a **real, server-provided** category name
  by keyword-matching, with a deterministic hash-based fallback. This is the one to reach for
  when rendering any actual transaction/category data.
- `placeholderCategories.ts` — a **fixed, hardcoded** 5-category set used only by the
  not-yet-real Budgets page (`pages/budgets/components/BudgetCard.tsx`). Not connected to the server.
  Don't use this for anything that touches real category data — use `categoryVisuals.ts` instead.

## Testing

Pure functions, no mocking needed — the cheapest, highest-value tests in the app. Every file here
should have exhaustive unit tests covering edge cases, in particular: `currency.ts` (rounding,
negative amounts, different currency codes), `dates.ts` (timezone/boundary edge cases), and
`categoryVisuals.ts` (an exact-name match for every bootstrapped category, the keyword-fallback
path, and the final hash-based color fallback for a totally unrecognized name — its three-tier
fallback structure is exactly the kind of branching logic worth covering exhaustively).

## Don't

- Don't add a function here that depends on component state, context, or a query client.
- Don't confuse `categoryVisuals.ts` (real data) with `placeholderCategories.ts` (Budgets-page
  placeholder) — they look similar but serve different, non-overlapping consumers.
