# pages/add/

The Add Transaction flow. Follows the standard container/view pattern (see `pages/README.md`):
`index.tsx` owns all form state (amount, category, date/time, attachments, payer override, local line
items, attachment-read status) and assembles `AddTransactionViewProps`;
`AddTransaction.desktop.tsx`/`.mobile.tsx` render it.

## Files

- `prefillConfidence.ts` — the extraction review threshold (`PREFILL_CONFIDENCE_THRESHOLD = 80`)
  and `isLowConfidence()`. A proposed scalar scored strictly below it is outlined in the form.
- `components/` — presentational components used only by this page:
  `AddTransactionForm` (the shared field set + Add-only date/line-item/save slots, so the desktop
  and mobile variants cannot drift), `AttachmentBar` (compact upload/read/result status),
  `SubTransactionsPanel` (collapsible line-item summary/list), `SubTransactionModal` (focused
  add/edit dialog), `DateQuickField` — one component per file, built from `components/ui/`
  primitives (`Button`, `IconButton`, `BottomSheet`, `TextInput`, …). Check that folder before
  adding a new raw form element here.
- The other Add-form building blocks are shared with the transaction detail screen's edit mode, so
  they live in the base layer rather than here: `src/components/TransactionFormFields.tsx`,
  `AmountHero.tsx`, `DateTimeFields.tsx`, `MoreOptionsPanel.tsx`, `categoryOptions.tsx`, and the
  amount formatting/parsing helpers in `src/lib/amount.ts`. Line items are the deliberate exception:
  Add uses the page-local `SubTransactionsPanel` + `SubTransactionModal` (local drafts), while
  detail edit uses the shared inline `SubTransactionList` (per-row CRUD).

## The page in one screen

There are no tabs. Everything lives in one scrollable form:

1. **Header attachment button** — the paperclip icon opens the camera/file picker. File upload lives
   entirely behind it; there is no upload field or upload card at the top of the page. An untouched
   form simply starts with the Amount field.
2. **Attachment bar** — only appears once a file is attached. It is a single compact row showing
   upload/read progress or the scan result (`2 items added`, attachment total), with Scan/Retry/Cancel
   and a details disclosure for file management and Undo. Clicking anywhere on the row — except its
   buttons — toggles the details. Clicking a file inside the details opens the shared full-screen
   attachment lightbox (`src/components/AttachmentPreview.tsx`, the same one the transaction detail
   screen uses). It never pushes the form down when there is nothing to report.
3. **Form** — amount, name, type, category, description, date/time, payer. A successful read writes
   its values straight into these fields.
4. **Sub-transactions panel** — collapsed by default to a single summary row: the item count, their
   total, and (when they do not match the attachment total) a red difference warning. Expanding it
   reveals every row. Clicking a row opens the focused **Add/Edit sub-transaction** modal; its trash
   button removes the row from the local draft. **Review** in the attachment bar expands the panel.
5. **Sub-transaction modal** — a centered dialog with name, optional label and amount. Amounts may be
   negative for discounts/coupons but never zero. The amount uses `SignedAmountInput`, whose `±`
   button toggles the sign because iOS's decimal keypad has no minus key. It writes back to the
   page's local list, so the form's derived Amount and the panel summary update immediately.

## Extraction behavior

- A read is one combined request for every ready image (`POST
  /transaction-attachments/extraction`). A successful result is **applied automatically**: every
  proposed scalar fills its field, and every proposed line item is preloaded into the sub-transaction
  list. Each value carries an extractor confidence from 0 to 100. The explicit **Scan attachment**
  action is the confirmation; the old per-field accept/dismiss chips are gone.
- The reader's `category_hint` is matched case-insensitively against the loaded category list. A
  hint can arrive before the list for a newly proposed type has loaded, so the selected category is
  derived during render (the hint stays visible in the attachment bar until it can be matched).
- Re-scanning replaces the previously **extracted** rows but keeps rows the user added by hand
  (`SubTransactionDraft.origin` distinguishes them). A scan with no line items clears stale extracted
  rows without touching manual ones.
- A successful scan can be rolled back with **Undo scan**, which restores the exact form values and
  line items that existed before that scan. Removing the attachment afterwards keeps the applied values
  and items — they are plain form state by then.
- Each extracted value is reviewed by confidence: scalars scored **below 80** get a red outline in
  the Add form (`type`, amount/currency, name, category, description, date and time), and so do
  proposed line items that fall below it. Editing a field or line item, choosing a replacement,
  pressing **Use today** for the date/time, running a new scan or using Undo retires the outline.
  The threshold lives in `prefillConfidence.ts`; the detail screen's edit form never shows these
  outlines because only Add passes `lowConfidenceFields` / `lowConfidenceSubTransactionIds`.
- A read failure is not an upload failure: the files stay attached, the error explains that the form
  was not changed, and **Try again** retries without re-uploading. Cancelling returns to the idle
  attachment state rather than showing an error.
- When any line item exists the Amount field is the read-only Σ of the items; emptying the list
  releases it back to being typed. The attachment's own total stays visible as a checksum in both the
  attachment bar and the sub-transactions card.

## Things to know

- The amount field is a **display string**, not the wire value. It accepts only digits and a single
  comma (the dot is never a separator here), clamps the fraction to two digits while typing, and
  pads to two decimals on blur (`"5,5"` → `"5,50"`, `"5"` → `"5,00"`). On save it is converted to
  whole **minor units (cents)** with `amountToCents` (`"42,50"` → `4250`) and sent as `amount`; the
  comma and its decimals never leave the browser. `decimalStringToAmountDisplay` reformats a read's
  decimal value (`"24.90"`) into display form. Helpers live in `amount.ts`.
- The date/time picker (`DateTimeFields`) is persisted: `index.tsx` combines the `date`/`time` state
  into an RFC3339 instant with `toISODateTime` (`lib/dates`) and sends it as `created_at`, so a
  transaction can be back-dated. The field is optional on the wire — omitting it lets the server
  stamp `now()`.
- `AddTransaction.desktop.tsx` and `.mobile.tsx` keep only page chrome (header, attachment bar,
  picker); `components/AddTransactionForm.tsx` composes the shared field set with the Add-only
  date shortcut, line-item panel and save action. The field set itself and its desktop/mobile
  arrangement live in `src/components/TransactionFormFields.tsx`, also used by the detail screen's
  edit form — don't reintroduce per-field markup in either view variant.
- Attachments are real and plural: picking a file uploads it as an unattached draft right away
  (`POST /transaction-attachments`, see `api/attachments.ts` / `hooks/useAttachments.ts`). Each row
  shows its own `uploading`/`processing`/`ready`/`error` state, and every returned id is collected
  and sent together as `attachment_ids` when the transaction is created — the server links them all
  in the same unit of work. Removing a row deletes its draft, and any unsaved drafts are deleted on
  unmount. `canSave` is false while any row is still uploading or being read.
- The picker (`src/components/TransactionAttachmentPicker.tsx`, shared with the transaction
  detail page) opens directly on "Take photo"/"Choose files". Both inputs accept a multi-file
  selection and upload the original bytes; image normalization belongs to the server upload path,
  never the client.
- One shared `AbortController` cancels the grouped read from the attachment bar.
- The extraction endpoint is server-owned (see
  `server/…/transaction_attachment/extraction_handler.go`). A read that still 404s (older backend,
  unsupported media type) lands in the "couldn't read" state, so the client degrades gracefully
  without the route.

## Testing

- Container (`index.tsx`): MSW-mock the categories/currencies/transaction-create and attachment
  endpoints. Cover: submit payload shape; upload/link/remove; multi-file linking; save disabled
  while an upload or read is in flight; a successful read auto-fills scalars and preloads all line
  items; a type-changing category hint resolves after the new list loads; re-scan replaces extracted
  rows while preserving manual ones; Undo restores the pre-scan form; a failed read keeps files and
  retries; clicking a file row opens the shared attachment lightbox; the panel stays collapsed until
  expanded, rows open the edit modal, trash removes a row, and the derived amount follows.
- View variants: pure render tests per `pages/README.md`'s pattern — same
  `AddTransactionViewProps` in, assert both variants render the same fields/values. This is the pair
  most likely to drift silently, per the duplication note above.
- Component tests: `AttachmentBar` states/actions plus its details disclosure,
  `SubTransactionsPanel` collapsed summary/mismatch/expand/row actions, and `SubTransactionModal`
  create/edit validation. `SubTransactionList`'s shared behavior stays covered for the detail page.
- This page owns the **"can a user add a transaction"** Playwright flow from `TESTING.md` — that
  spec exercises the real submit → list round trip; don't duplicate its assertions here.

## Don't

- Don't add new business logic directly to the `.desktop.tsx`/`.mobile.tsx` files or to a
  `components/` file — put it in `index.tsx` and pass it down as a prop, same as everything else here.
- Don't import a `components/` file from outside `pages/add/` — if it turns out to be useful
  elsewhere, promote it to `src/components/` instead of reaching across page folders.
- Don't reintroduce a second editing surface for line items: the panel shows the summary/list, the
  modal edits one row.
