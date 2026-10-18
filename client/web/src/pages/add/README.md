# pages/add/

The Add Transaction flow. Follows the standard container/view pattern (see `pages/README.md`):
`index.tsx` owns all form state (amount, category, date/time, attachments, payer override) and
assembles `AddTransactionViewProps`; `AddTransaction.desktop.tsx`/`.mobile.tsx` render it.

## Files

- `components/` — presentational components used only by this page: `PrefillProcessing`,
  `QuickFieldsRow`, `TransactionAttachmentBanner`, `TransactionAttachmentPicker`, one component per
  file, built from `components/ui/` primitives (`Button`, `FieldLabel`, `SelectInput`, `TextInput`,
  `Switch`) — check that folder before adding a new raw form element here.
- The other Add-form building blocks are shared with the transaction detail screen's edit mode, so
  they live in the base layer rather than here: `src/components/AmountHero.tsx`,
  `DateTimeFields.tsx`, `MoreOptionsPanel.tsx`, `categoryOptions.tsx`, and the amount
  formatting/parsing helpers in `src/lib/amount.ts`.

## Things to know

- The amount field is a **display string**, not the wire value. It accepts only digits and a single
  comma (the dot is never a separator here), clamps the fraction to two digits while typing, and
  pads to two decimals on blur (`"5,5"` → `"5,50"`, `"5"` → `"5,00"`). On save it is converted to
  whole **minor units (cents)** with `amountToCents` (`"42,50"` → `4250`) and sent as `amount`; the
  comma and its decimals never leave the browser. `decimalStringToAmountDisplay` reformats the
  reader's decimal prefill (`"24.90"`) into display form. Helpers live in `amount.ts`.
- The date/time picker (`DateTimeFields`) is persisted: `index.tsx` combines the `date`/`time` state
  into an RFC3339 instant with `toISODateTime` (`lib/dates`) and sends it as `created_at`, so a
  transaction can be back-dated. The field is optional on the wire — omitting it lets the server
  stamp `now()`.
- `AddTransaction.desktop.tsx` and `.mobile.tsx` are the most duplicated pair in the codebase —
  near-identical JSX trees differing mainly in wrapper layout and a handful of prop values
  (e.g. `AmountHero boxed` vs `fullWidth`). This is a known, accepted tradeoff under the
  project's "duplicate JSX, diverge only in layout" convention — if these two pages need to
  diverge further, consider extracting the shared subtree into one component parameterized by
  the few props that differ, rather than letting the duplication grow unchecked.
- Attachments are real and plural: the page holds `attachments: TransactionAttachmentDraft[]`, and
  picking a file uploads it as an unattached draft right away
  (`POST /transaction-attachments`, see `api/attachments.ts` / `hooks/useAttachments.ts`). Each row
  shows its own `uploading`/`processing`/`ready`/`error` state, and every returned id is collected
  and sent together as `attachment_ids` when the transaction is created — the server links them all
  in the same unit of work. Removing a row deletes its draft, and any unsaved drafts are deleted on
  unmount. `canSave` is false while any row is still uploading or being read.
- The picker (`TransactionAttachmentPicker`) is a two-step sheet: step one chooses *what the file is
  for* (`attach` — store it only, or `prefill` — store it and read it), step two is today's
  "Take photo"/"Choose files". Both inputs accept a multi-file selection, so one pick can add several
  rows at once; the sheet closes after each pick and reopening it adds more. On the `prefill` flow the
  page uploads first, then calls
  `POST /transaction-attachments/{id}/extraction` (`api/attachments.ts`), shows `PrefillProcessing`
  while the read is in flight (Cancel aborts the signal), and fills only the fields the user has not
  edited — `touchedRef` marks every user change, so a read never clobbers typed input. The mapping
  policy (including case-insensitive category-name matching against the loaded list) lives in
  `applyPrefill` in `index.tsx`.
- A read failure is not an upload failure: whatever the read outcome — error, timeout, cancel or an
  empty result — the draft is already linked on save. `TransactionAttachmentBanner` reuses its
  existing box for the summary (`N fields filled…` / `Couldn't read the file`). The extraction
  endpoint is a client-side contract only; nothing in `server/` implements it yet, so a real backend
  read 404s and lands in the "couldn't read" state until the server catches up.

## Testing

- Container (`index.tsx`): MSW-mock the `categories`/`currencies` transaction-create and
  attachment endpoints; test that submitting the form calls the create mutation with the right
  payload shape, that picking a file uploads a draft (the save payload carries its
  `attachment_ids`), that picking several files links *all* their ids, that removing an attachment
  deletes the draft, that a failed upload surfaces an error state, and that save is disabled while
  an upload is in flight. For the `prefill` flow, MSW-mock the `…/extraction` endpoint too: assert
  `PrefillProcessing` shows while the read is in flight, that untouched fields (including a matched
  category) fill on success, and that a failed read keeps the file attached.
- View variants: pure render tests per `pages/README.md`'s pattern — same `AddTransactionViewProps`
  in, assert both variants render the same fields/values (this is the pair most likely to drift
  silently, per the "most duplicated pair in the codebase" note above).
- This page owns the **"can a user add a transaction"** Playwright flow from `TESTING.md` — that
  spec exercises the real submit → list round trip; don't duplicate its assertions here, just the
  component-level behavior MSW can prove faster.

## Don't

- Don't add new business logic directly to the `.desktop.tsx`/`.mobile.tsx` files or to a
  `components/` file — put it in `index.tsx` and pass it down as a prop, same as everything
  else here.
- Don't import a `components/` file from outside `pages/add/` — if it turns out to be useful
  elsewhere, promote it to `src/components/` instead of reaching across page folders.
