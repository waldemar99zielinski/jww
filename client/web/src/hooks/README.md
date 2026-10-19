# hooks/

One module per server resource, each wrapping TanStack Query. This is the *only* place
`useQuery`/`useMutation` should be called from directly — page containers (`pages/*/index.tsx`)
call these hooks, never TanStack Query's hooks directly.

## Conventions

- Query hooks: `queryKey: [resourceName, ...params]`, `queryFn` delegates straight to the
  matching `api/*.ts` function. See `useCategories`, `useCurrencies`, `useSummary`,
  `useTransactions`, `useUsers` for the pattern.
- Mutation hooks: on `onSuccess`, `queryClient.invalidateQueries` the query key(s) the mutation
  affects, using the *response* data to build the key where relevant (e.g.
  `invalidateQueries({ queryKey: ['categories', category.type] })`) rather than trusting
  caller-supplied input — the response is the source of truth for what actually changed.
- `useDefaultCurrency.ts` and `useAuth.ts` are composed hooks (deriving from currencies data /
  Redux auth state respectively) rather than direct API wrappers — that's fine for a hook that
  combines existing state, just don't let business logic beyond simple derivation creep in here.
- `useViewport.ts`, `useOnlineStatus.ts` and `useInstallPrompt.ts` are the hooks with no API/store
  backing — `useViewport` and `useOnlineStatus` are `useSyncExternalStore` wrappers over a browser
  flag (breakpoint, connectivity); `useInstallPrompt` wraps Chromium's `beforeinstallprompt`/
  `appinstalled` events. All three are browser-state hooks, not server resources.
- `useScrollLock.ts` freezes background scrolling while an overlay is open. It locks the document
  plus any `[data-scroll-lock]` element (the desktop shell scrolls an inner `<main>`, not the body),
  and is reference-counted so stacked overlays don't unlock each other early.
- `useAttachments.ts` wraps the attachment endpoints: the draft mutations
  (`useUploadDraftAttachment`, `useDeleteDraftAttachment`) invalidate nothing — a draft has no list
  of its own, and `useCreateTransaction` already invalidates the transaction/summary queries when
  the draft is linked on save — while `useUploadTransactionAttachment` (adding a file to an existing
  transaction) and `useDeleteTransactionAttachment` (deleting a *linked* attachment) invalidate
  `['transactions']` + `['summary']`, since the detail screen shows the rows they change.
- `useTransactions.ts` exposes `useTransactions(filters)` (the list, key `['transactions', filters]`),
  `useTransaction(id)` (one record for the detail screen, key `['transactions', 'detail', id]`),
  `useUpdateTransaction` (a PATCH mutation), and `useDeleteTransaction` (a DELETE mutation that
  removes the detail query from the cache before invalidating the list + summary, since refetching
  a deleted detail would 404) — so the `['transactions']` prefix invalidation refreshes both
  queries.

## Adding a new resource hook

1. Add the fetch function to `api/<resource>.ts` (and its mock counterpart).
2. Add a hook here following the query-key convention above.
3. Only call it from a page container, never from a presentational component.

## Testing

`renderHook` from `@testing-library/react`, wrapped in a test `QueryClient` (`retry: false,
gcTime: 0`) and MSW-mocked network — see `TESTING.md`'s "hooks" section for the full pattern.
Priority: query hooks resolving to the right `data`/`isError` state from a mocked response, and
mutation hooks calling `invalidateQueries` with the exact key their `onSuccess` handler builds —
especially where it's built from the *response* rather than caller input (e.g.
`useCreateCategory`/`useUpdateCategory` using `category.type`), per the convention above. That
response-not-input rule is exactly the kind of thing worth pinning down with a test.

## Don't

- Don't give a mutation's input type a field the mutation never uses — if the API call doesn't
  need it, don't require callers to pass it just because the response happens to include it.
