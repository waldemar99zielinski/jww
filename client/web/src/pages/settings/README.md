# pages/settings/

Settings root (account info, currency summary, logout) plus two nested sub-pages:
`categories/` and `currencies/`. The root page follows the standard container/view structure
(see `pages/README.md`); it's a thin container — just `useAuth()` + `useDefaultCurrency()` fed
straight into `SettingsViewProps`, no derived logic.

## Files

- `components/SettingsRow.tsx` — the shared row layout (label + control/value + optional
  chevron) used by the settings root's list. Scoped here because every current consumer lives
  under `pages/settings/` (the root page); if `categories/`/`currencies/` ever need it too,
  it's still reachable via `@/pages/settings/components/SettingsRow` without re-exporting.

## Sub-pages

- `categories/` — manage transaction categories (create, toggle active, per expense/income type).
- `currencies/` — manage currencies (create, mark one as the base currency).

Both are separately routed in `App.tsx` (`/settings/categories`, `/settings/currencies`) and
each is its own self-contained container/view folder — don't merge their logic back into the
settings root container.

## Testing

Root container is thin (two hooks straight into view props) — one render test covering both
populated and loading/error states is enough; no dedicated container-logic tests needed here. See
`categories/README.md` and `currencies/README.md` for their own (more involved) sub-pages.

## Don't

- Don't add category/currency management logic to `pages/settings/index.tsx` — that belongs in
  the respective sub-page.
