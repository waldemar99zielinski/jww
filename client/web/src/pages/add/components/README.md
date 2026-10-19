# pages/add/components/

Presentational components used only by `pages/add/` (the Add Transaction flow). See
`pages/add/README.md` for the full behavior. One component per file, built from
`src/components/ui/` primitives, no business logic (that lives in `../index.tsx`).

- `AttachmentBar` — compact attachment status shown only once a file is attached: upload/read progress,
  item count and attachment total after a scan, and an expandable details section with the file list,
  warnings, unresolved category hint and Undo. Clicking a file row opens the shared
  `src/components/AttachmentPreview.tsx` lightbox. The empty upload surface is gone; the header's
  **attachment** button starts the flow.
- `SubTransactionsPanel` — collapsed by default to a count/total summary with an attachment-total
  mismatch warning; expanding reveals the rows. Row click opens the edit modal, the trash button
  removes it, and one Add action opens the create modal.
- `SubTransactionModal` — centered dialog (`src/components/ui/Modal`) for one line item: name,
  optional label and amount, with validation and a Cancel/Save footer.
- `DateQuickField` — the date shortcut that reveals the date/time fields. Uploading lives behind the
  header attachment button, whose picker is the shared `src/components/TransactionAttachmentPicker.tsx`.

The form fields shared with the detail screen's edit mode (`AmountHero`, `DateTimeFields`,
`MoreOptionsPanel`, `SubTransactionList`, `categoryOptions`) live in `src/components/` and are
imported from there.

## Testing

RTL render + `user-event`, props in only — no network/store needed, same as the base `components/`
layer. `AttachmentBar` covers its empty (render nothing), ready, reading, success and failure states,
including the details disclosure. `SubTransactionsPanel` covers the collapsed summary, mismatch,
expand and row actions. `SubTransactionModal` covers create/edit values and validation.
`DateQuickField` needs one render-with-props case. The shared picker has its own test under
`src/components/`.
