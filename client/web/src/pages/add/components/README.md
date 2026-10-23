# pages/add/components/

Presentational components used only by `pages/add/` (the Add Transaction flow). See
`pages/add/README.md` for the full behavior. One component per file, built from
`src/components/ui/` primitives, no business logic (that lives in `../index.tsx`).

- `AttachmentBar` — compact attachment status shown only once a file is attached: upload/read progress,
  item count and attachment total after a scan, and an expandable details section with the file list,
  warnings, unresolved category hint and Undo. Clicking a file row opens the shared
  `src/components/AttachmentPreview.tsx` lightbox. The empty upload surface is gone; the header's
  **attachment** button starts the flow.
- `DateQuickField` — the date shortcut that reveals the date/time fields. Uploading lives behind the
  header attachment button, whose picker is the shared `src/components/TransactionAttachmentPicker.tsx`.

The form fields and line-item UI shared with the detail screen (`AmountHero`, `DateTimeFields`,
`MoreOptionsPanel`, `SubTransactionsPanel`, `SubTransactionModal`, `categoryOptions`) live in
`src/components/` and are imported from there.

## Testing

RTL render + `user-event`, props in only — no network/store needed, same as the base `components/`
layer. `AttachmentBar` covers its empty (render nothing), ready, reading, success and failure states,
including the details disclosure. `DateQuickField` needs one render-with-props case. The shared
panel, modal, and picker have their own tests under `src/components/`.
