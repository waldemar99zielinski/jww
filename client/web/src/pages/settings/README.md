# pages/settings/

Settings root (account info, currency summary, install entry point, logout) plus two nested
sub-pages: `categories/` and `currencies/`. The root page follows the standard container/view
structure (see `pages/README.md`): `useAuth()` + `useDefaultCurrency()` feed `SettingsViewProps`
directly, and `useInstallPrompt()` is resolved into the `install` view props (see "Install row"
below) rather than exposing the hook to the views.

## Files

- `components/SettingsRow.tsx` — the shared row layout (label + control/value + optional
  chevron) used by the settings root's list. Scoped here because every current consumer lives
  under `pages/settings/` (the root page); if `categories/`/`currencies/` ever need it too,
  it's still reachable via `@/pages/settings/components/SettingsRow` without re-exporting.
  Pass `onClick` (without `to`) to render an actionable full-width `<button>` row — that's the
  "Install app" row.
- `components/InstallInstructions.tsx` — the per-platform install steps. One component, two
  presentations: a `BottomSheet` when `compact` (mobile) and an inline card when not (desktop);
  the copy is identical in both, only the surface differs.

## Install row

The root page renders an "Install app" row between "Export data" and "About Joint Wallet Watch".
The container (`index.tsx`) calls `useInstallPrompt()` and resolves everything into an
`InstallViewProps` object on `SettingsViewProps`, so both view variants just render the result:
actionable (Chromium prompt / Safari instructions), passive with a reason (unsupported browser or
in-app webview), or hidden (already installed, or Chromium has not offered a prompt yet). See the
root README's "PWA / install" section.

## Sub-pages

- `categories/` — manage transaction categories (create, toggle active, per expense/income type).
- `currencies/` — manage currencies (create, mark one as the base currency).

Both are separately routed in `App.tsx` (`/settings/categories`, `/settings/currencies`) and
each is its own self-contained container/view folder — don't merge their logic back into the
settings root container.

## Testing

Root container is thin (a few hooks feeding view props; install state resolution is the only
derived logic) — one render test covering the populated state, plus install-row state coverage,
is enough. See `categories/README.md` and `currencies/README.md` for their own (more involved)
sub-pages.

## Don't

- Don't add category/currency management logic to `pages/settings/index.tsx` — that belongs in
  the respective sub-page.
