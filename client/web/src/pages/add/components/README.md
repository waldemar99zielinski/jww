# pages/add/components/

Presentational components used only by `pages/add/` (the Add Transaction flow). See
`pages/add/README.md` for the full list and conventions — summary: `PrefillProcessing`,
`QuickFieldsRow`, `TransactionAttachmentBanner`, `TransactionAttachmentPicker`, one component per
file, built from `src/components/ui/` primitives, no business logic (that lives in `../index.tsx`).
The form fields shared with the detail screen's edit mode (`AmountHero`, `DateTimeFields`,
`MoreOptionsPanel`, `categoryOptions`) were promoted to `src/components/` and are imported from
there.

## Testing

RTL render + `user-event`, props in only — no network/store needed, same as the base `components/`
layer. `TransactionAttachmentPicker` has the most internal interaction logic (the two-step
attach/prefill flow and its step reset on reopen) and is worth the most test cases here;
`PrefillProcessing` and `TransactionAttachmentBanner` need one case per status; `QuickFieldsRow`
mostly needs one render-with-props case.
