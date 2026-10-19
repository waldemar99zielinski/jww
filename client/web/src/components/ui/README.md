# components/ui/

The design-system primitive layer: `Button`, `IconButton`, `TextInput`, `SelectInput`,
`InlineSelect`, `SignedAmountInput`, `FieldLabel`, `DateInput`, `SearchableSelect`,
`SegmentedControl`, `Switch`, `BottomSheet`, `Modal`, `Loader`, plus `classes.ts` (the `cx()`
class-join helper).

Every form control and button in the app is expected to route through this layer — it's the
single place styling, accessibility (`role`/`aria-*`), and interaction behavior for a given
control type are defined.

**Check this folder before building a new input/control/popover shape anywhere in the app.**
If something close already exists, extend it with a prop rather than writing a parallel
one-off — see "Adding a new primitive" below for when a feature-specific shape should actually
graduate here instead.

## Conventions

- **One component per file.** Each primitive gets its own file named after it
  (`Button.tsx`, `IconButton.tsx`, `TextInput.tsx`, …) — this folder used to bundle multiple
  unrelated primitives per file (e.g. a single `Input.tsx` with four exports); it doesn't anymore,
  and new primitives shouldn't restart that pattern.
- Components here are generic: no knowledge of any specific feature/page, props only (no
  feature-specific business logic, no direct API/store access).
- Accessibility is handled here once, not re-implemented per call site — e.g. `Switch` carries
  `role="switch" aria-checked aria-label`, `Loader` carries `role="status" aria-label`,
  `BottomSheet` manages `aria-hidden`/`inert` when closed. Keep that bar when adding new
  primitives.
- `Button`'s `variant="custom"` exists for one-off layout needs while still getting `<button>`
  semantics (type, disabled styling hook) for free — prefer it over a raw `<button>` even when
  none of the styled variants fit.

## Adding a new primitive

If you notice the same bespoke control (a specific input shape, a specific popover pattern)
appearing in more than one page's `components/` folder, that's the signal to extract it here
instead of leaving copies behind — this is how `DateInput` ended up here, moved out of
`pages/transactions/components/TransactionFiltersDropdown.tsx` once it was clear the shape was
reusable.

## Testing

This is the most reused layer in the app, so it carries the highest bar: every primitive gets at
least one RTL test exercising its documented accessibility contract (e.g. `Switch`'s
`role="switch" aria-checked`, `Loader`'s `role="status"`, `BottomSheet`'s `aria-hidden`/`inert`
toggle when closed, `Modal`'s `role="dialog" aria-modal` + Escape/close) plus its core interaction
via `user-event` (click/type/keyboard), not just a render snapshot. A bug introduced here silently breaks every page composing it, so these tests are
worth writing before the page-level tests that depend on them.

## Don't

- Don't import anything from `pages/`, `hooks/`, or `api/` into this folder — primitives take
  data and callbacks as props only.
- Don't reach for a raw `<input>`/`<button>`/`<select>` in feature code when one of these covers
  the case.
- Don't put more than one component in a new file here — split it, even if the two feel related
  (see `Button.tsx`/`IconButton.tsx` as separate files despite `IconButton` wrapping `Button`).
