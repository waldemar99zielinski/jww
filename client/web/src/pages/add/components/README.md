# pages/add/components/

Presentational components used only by `pages/add/` (the Add Transaction flow). See
`pages/add/README.md` for the full list and conventions — summary: `AmountHero`, `DateTimeFields`,
`MoreOptionsPanel`, `QuickFieldsRow`, `ReceiptBanner`, `ReceiptPicker`, one
component per file, built from `src/components/ui/` primitives, no business logic (that lives in
`../index.tsx`).

## Testing

RTL render + `user-event`, props in only — no network/store needed, same as the base `components/`
layer. `ReceiptPicker` has the most internal interaction logic (the fake-timeout OCR fill) and is
worth the most test cases here; `AmountHero`/`QuickFieldsRow`/`DateTimeFields`/`ReceiptBanner`
mostly need one render-with-props case each.
