# pages/transactions/detail/

The transaction **detail screen**, reached by clicking a row in `pages/transactions/` (route
`/transactions/:id`). Standard container/view structure (see `pages/README.md`); it renders every
field of one `TransactionDto` plus its line items and attachments, and can switch in place into an
edit form. The read-only screen also offers a confirmed **Remove transaction** action.

## Files

- `index.tsx` — container. Reads the `:id` route param, fetches the transaction via
  `useTransaction` and the users list via `useUsers`, and turns them into a ready-to-render view
  model: a hero (`title`/`amount`/`isIncome`), a flat `items` list of already-formatted
  label/value pairs, the read-only line items, and the raw `attachments`. An invalid id is treated
  as an error without fetching. It also owns edit mode: the form state, the field handlers, and the
  `useUpdateTransaction` PATCH (see "Edit mode" below), plus the confirmed transaction delete.
- `types.ts` — `TransactionDetailViewProps` (+ `TransactionDetailItem`,
  `TransactionDetailEditProps`); the `items` shape keeps both view variants rendering the exact
  same fields — the container decides *what* is shown, the views only decide *how*. The
  `TransactionDetailEditProps` block carries the edit form's values/handlers, again computed once
  in the container.
- `TransactionDetail.desktop.tsx` / `TransactionDetail.mobile.tsx` — layout variants (centered
  card vs. full-height screen with a sticky header). Both compose the same page-local components
  below, and both swap the hero + `DetailList` for `TransactionEditForm` while editing.
- `components/TransactionHero.tsx` — category icon + name + signed amount block.
- `components/DetailList.tsx` — the label/value definition list.
- `components/DeleteTransactionAction.tsx` — the confirmation sheet for removing the whole
  transaction.
- `components/TransactionEditForm.tsx` — the edit form. One pure-display component that wraps the
  shared `src/components/TransactionFormFields.tsx` (the same field set and desktop two-column vs.
  mobile stacked arrangement as Add) and supplies the shared `SubTransactionsPanel` and
  Save/Reset actions as slots.
- `components/AttachmentList.tsx` — attachment rows. The row itself opens the lightbox (for any
  displayable type); its trailing `⋮` button opens the actions drawer.
- `components/AttachmentActions.tsx` — the actions drawer (`BottomSheet`): preview, download, and a
  two-step-confirm delete.
- `src/lib/formatBytes.ts` / `src/lib/previewKind.ts` — shared helpers: human byte sizes, and
  mapping a media type to a preview renderer (`PreviewKind`) or `null` when the browser can't show it.
- `components/SwipeToClose.tsx` — mobile-only wrapper: drag the screen to the right to dismiss it,
  with a shared animated `close` for the header button (`TransactionDetail.mobile.tsx`).

## Things to know

- It fetches the transaction by id rather than reading it from the already-loaded list query:
  a deep link / refresh lands here with an empty cache, and the by-id endpoint is the source of
  truth (`GET /v1/transactions/{transactionID}`).
- `useTransaction` is keyed `['transactions', 'detail', id]`, so `useCreateTransaction`'s
  existing `['transactions']` prefix invalidation already refreshes this screen.
- **`amount` arrives as a number, not a string.** `TransactionDto.amount` is typed `string`, but the
  server's Go `domain.Money` marshals as a bare JSON number (`42.5`). Read paths are unaffected
  because they go through `Number(...)`; the edit form seeds the input with
  `decimalStringToAmountDisplay`, which accepts both a string and a number — a numeric input there
  used to throw and make the Edit button appear dead against the real backend.

## Line items and removing the transaction

- **Read-only items** — whenever `sub_transactions` has rows, the detail screen shows the shared
  `SubTransactionsPanel`, collapsed to a count/total summary exactly like Add. Clicking it expands
  the rows (name, optional label, amount in the transaction's currency) with no add/edit/remove
  affordances. The container owns that expansion state, so entering edit mode keeps the panel as
  the user left it.
- **Remove transaction** — the button at the bottom opens `DeleteTransactionAction`, a
  `BottomSheet` confirmation whose copy states that the transaction, all of its items, and all of
  its attachments will be removed. Confirming calls `deleteTransaction()` →
  `useDeleteTransaction` (`DELETE /v1/transactions/{id}`). The server deletes every linked
  attachment (metadata and stored object) and the database's `ON DELETE CASCADE` removes the line
  items. On success the detail query is removed from the cache, the list/summary queries are
  invalidated, and the screen navigates to `/transactions` with `replace` so Back cannot reopen the
  deleted record. A failed delete keeps the sheet open with an inline error.

## Attachments (preview grants + download)

- **Add attachment** — the Attachments header has an `Add attachment` action that opens the shared
  `src/components/TransactionAttachmentPicker.tsx`. Each picked file is uploaded directly onto the
  existing transaction via `uploadTransactionAttachment()` →
  `useUploadTransactionAttachment` (`POST /v1/transactions/{id}/attachments`); the detail query is
  invalidated so the new row appears after the server responds. The button shows an uploading state
  and failures surface inline. The section renders even when there are no attachments yet.

`GET /v1/transactions/{transactionID}` is a **different shape** from list/create: it returns a
`TransactionDetailsDto` whose attachments each carry a short-lived `preview_grant`
(`TransactionAttachmentDetailDto`), a capability token the server mints only for safe inline media
(raster images and PDF; `null` for everything else). List/create responses keep the lightweight
`TransactionAttachmentDto` and never mint grants.

- **Preview** — clicking an attachment opens the shared full-screen lightbox
  (`src/components/AttachmentPreview.tsx`, also used by the Add page's attachment bar). Images
  and PDFs render straight from the server's grant URL (exchanged via
  `GET /api/v1/attachment-content?grant=…`; `attachmentPreviewUrl()` in `api/attachments.ts` builds
  it, `apiUrl()` in `api/client.ts` supplies the origin/prefix). Because that route needs no auth
  header, the URL is safe to drop straight into an `<img src>` or an `<iframe>`. Video, audio and
  text-like types have **no grant**, so the lightbox fetches their bytes through the protected
  download route and renders a blob URL (or, for text, the decoded string in a `<pre>`); anything
  else shows a "download it instead" message. `src/lib/previewKind.ts` owns that media-type→renderer map.
  The attachment *metadata* arrives with the transaction, but an image's *bytes* are fetched by the
  browser afterwards, so the row's thumbnail slot renders a spinner until the `<img>` fires `onLoad`.
  The lightbox keeps its spinner until the content is actually on screen — i.e. until the
  `<img>`/`<iframe>`/`<video>`/`<audio>` fires its load event, not merely until a source URL exists.
  In design mode there's no backend to hit.
- **Download** — the protected, transaction-scoped route
  (`GET /api/v1/transactions/{id}/attachments/{attachmentID}/download`) forces an attachment
  disposition but needs the Bearer header, so an `<a href>` can't reach it. `index.tsx`'s
  `handleDownload` fetches the bytes via `downloadTransactionAttachment()` → `apiDownload()`
  (`api/client.ts`, the `Blob`-returning twin of `apiFetch`), then saves them through a throwaway
  object-URL anchor. `downloadingId`/`downloadError` drive the drawer/lightbox busy state and the
  error line under the list.
- **Actions drawer** — each row's `⋮` button opens `AttachmentActions` (the generic `BottomSheet`):
  preview, download, and delete. Delete is a two-step confirm and calls the transaction-scoped
  `DELETE /api/v1/transactions/{id}/attachments/{attachmentID}` (`deleteTransactionAttachment()` →
  `useDeleteTransactionAttachment`), which invalidates the transaction + summary queries so the row
  disappears after the refetch.
- The container resolves raw attachments into `TransactionAttachmentItem`s (turning `preview_grant`
  into `previewUrl`) so neither view variant knows about grants, URLs, or blobs.
- While an overlay is open, `useScrollLock` freezes the page behind it — including the desktop app
  shell's inner scrolling `<main>` (marked `data-scroll-lock`), which a plain `body` overflow lock
  would miss.

## Edit mode

The header's pencil button (next to Close) switches the screen into an in-place edit form, seeded
from the fetched transaction. Every field the Add screen can set is editable: type, amount +
currency, name, category, description, date/time, and payer (attachments keep their own actions).

- **State** lives in `index.tsx` (`editing` plus one `useState` per field). `seedEditForm()` writes
  the fetched `TransactionDetailsDto` into those fields; it runs on entering edit mode and again on
  **Reset**, so Reset discards the in-progress edits without leaving the mode. Values are formatted
  exactly as the Add form expects them: the amount via `decimalStringToAmountDisplay`, the date via
  `toLocalISODate` (the viewer's calendar day — *not* UTC), and the time via `formatTime`.
- **Categories** are only fetched once editing starts (`useCategories(type, true, editing)`); the
  read-only screen never needs them. Changing the type clears the category, same as on Add.
- **Save** calls `useUpdateTransaction` (`PATCH /v1/transactions/{id}`) with the edited fields,
  including `created_at` built by `lib/dates`'s `toISODateTime`. On success it leaves edit mode; the
  mutation invalidates `['transactions']`, so the read-only fields re-render from the refetched
  transaction. A failed save keeps the form open with a "Couldn't save" line.
- **Reset / Cancel** is one button whose label follows the form's dirtiness: while every field still
  matches the loaded transaction (`isDirty === false`) it reads **Cancel** and just leaves edit mode
  (`handleCancel`); once any field differs it reads **Reset** and reverts the fields to the loaded
  values (`handleReset`, which also clears any save error via `updateTransaction.reset()`), staying
  in edit mode. `isDirty` is computed by comparing the form state against a `baseline` memo built
  from the same `editFormValues()` helper that seeds the form, so the two can't drift.
- The form markup is one page-local component (`TransactionEditForm`) shared by both view variants,
  parameterized by a `layout` prop — the same "one component, layout-only prop" shape as
  `AmountHero`'s `boxed`/`fullWidth`.
- **Line items** use the same shared `SubTransactionsPanel` + `SubTransactionModal` as Add:
  collapsed by default, expanding on click, and editing one row in the focused modal. The
  difference is persistence — each add/edit/remove on this screen is immediately a
  `POST`/`PATCH`/`DELETE` sub-transaction call owned by `index.tsx`, and a failed mutation keeps
  the modal open (or the row unchanged) with an inline error.

## Motion & the mobile swipe (and its hard limit)

A pushed screen should *arrive*, and a back gesture should *leave*. So:

- **Enter** — `TransactionDetail.mobile.tsx` slides in from the right (`.screen-enter-right` in
  `index.css`); the desktop card uses the generic `.screen-enter`. Both are gated behind
  `prefers-reduced-motion`.
- **Leave** — `SwipeToClose` owns the exit: dragging right translates the screen with the finger
  and, past ~28% of the viewport width, animates it off-screen before `onClose` (`navigate(-1)`)
  runs. The header's Close button calls the same `close()`, so button and swipe behave identically.

**The OS/browser edge-swipe back gesture cannot be intercepted, disabled, or repurposed by web
content** (iOS WKWebView and Android gesture navigation aren't scriptable). `SwipeToClose`
therefore *supplements* it rather than replacing it: it ignores touches within `EDGE_GUARD_PX` of
the left edge (that zone belongs to the system gesture) and only claims gestures that are clearly
horizontal — a mostly-vertical drag is abandoned so the page still scrolls. `[touch-action:pan-y]`
on the wrapper keeps the browser from consuming horizontal pans (and, incidentally, stops its own
horizontal overscroll hand-off inside this screen).

Because we can't animate a navigation the OS initiates, back-swipes from the edge still cut
instantly; only our own controls and the in-app swipe get the exit transition.

## Testing

- Container: MSW-mock the by-id endpoint (and a 404 for the not-found case) and assert the
  fetched transaction's fields, read-only line items, and attachments render — see the "page
  containers" section of `TESTING.md`. Also cover the attachment actions: opening the drawer
  (preview/download/delete) and that a confirmed delete hits the transaction-scoped route and drops
  the row after the refetch. Cover the whole-transaction remove flow too: the confirmation copy,
  the `DELETE` call, the navigation back to the list, and the inline error when it fails.
- `TransactionDetail.test.tsx` renders **both** view variants with identical props and asserts
  the same fields appear in each, per the desktop/mobile-divergence rule in `TESTING.md`. It also
  covers the collapsed read-only line-item panel, its expansion, and the Remove transaction
  button, plus edit mode: the header Edit button calls `onStartEdit`, and while `isEditing` each
  field renders as a control with Save/Reset.
- `index.test.tsx` covers the edit round trip: clicking Edit, changing a field, and Save issuing a
  `PATCH` with the converted payload (display commas → whole minor units); and Reset restoring the
  saved values. Line items are exercised through the shared panel and modal: expand, add/edit in
  the modal, remove from the row, and the resulting per-row CRUD calls/refetches.
- `src/components/AttachmentPreview.test.tsx` covers one renderer per media type (image via inline
  URL, PDF iframe, text/video/audio via the blob loader), the unsupported fallback, the
  loader-until-loaded state, Escape/close, and the swipe-down dismiss (past threshold closes,
  short/horizontal ones don't).
- `components/AttachmentActions.test.tsx` covers the drawer rows, the delete confirm step, and the
  busy/error states.
- `components/SwipeToClose.test.tsx` drives the gesture with `fireEvent.touch*` (rightward past
  threshold closes; short, vertical, and edge-started drags don't) and asserts `onClose` fires only
  after the exit delay, for both the swipe and the `close()` path.

## Don't

- Don't add per-field logic to the view variants — compute and format in `index.tsx` and pass
  values down through `items`. If a field needs custom markup, add it as a page-local component
  used by both variants rather than duplicating it.
- Don't re-fetch the whole transaction list to find one record; use `useTransaction(id)`.
- Don't move edit-mode state (values, handlers, the PATCH) into `TransactionEditForm` or the view
  variants — it all belongs in `index.tsx`, like every other page.
