# api/

Thin, typed fetch wrappers — one file per server resource (`attachments.ts`, `auth.ts`,
`categories.ts`, `currencies.ts`, `summary.ts`, `transactions.ts`, `users.ts`), plus `types.ts`
for the shared DTO shapes and `client.ts` for the one HTTP chokepoint. DTO shapes intentionally
mirror the server's `server/internal/adapter/schema/*.go` structs — keep them in sync when the
server's schema changes.

## Conventions

Every resource function follows the same two-line shape:

```ts
export function listThings(filters: ThingFilters): Promise<ThingDto[]> {
  if (isDesignMode()) return mockDelay(mockApi.listThings(filters))
  return apiFetch<ThingDto[]>(`/things${toQueryString(filters)}`)
}
```

- `isDesignMode()` (`lib/designMode.ts`) gates every function — when adding a new endpoint,
  add its mock counterpart in `mock/data.ts` and branch the same way, so design mode (develop
  without a backend) keeps working for every resource.
- `apiFetch<T>` (`client.ts`) is the only place that attaches the auth header, unwraps the
  `{data: T}` response envelope, and converts non-2xx responses into a typed `ApiError`. Call
  it instead of `fetch` directly.
- Downloads are the one exception to the JSON envelope: `apiDownload` (`client.ts`) mirrors
  `apiFetch`'s auth header/error handling but returns a `Blob`. `apiUrl` builds the full URL for
  raw browser-driven requests that can't go through `apiFetch` (an `<img src>` or an `<iframe>`) —
  e.g. the grant-authenticated attachment preview route.
- Multipart uploads (`attachments.ts`) pass a `FormData` body to `apiFetch`; it detects this and
  omits the JSON `Content-Type` so the browser sends its own multipart boundary. The file part
  must be named `file` (the server's contract).
- `toQueryString` (`client.ts`) is the shared query-string builder for filter objects — use it
  rather than hand-building query strings.
- `auth.ts` is the one file that doesn't follow the mock-branch shape — auth is bypassed one
  layer up, in `store/authSlice.ts`'s thunks, not here.

## Testing

- `client.ts`'s pure logic — `toQueryString` (empty object, `undefined` values, non-string types)
  and `readErrorMessage`'s JSON/plain-text fallback — is worth unit testing directly, no mocking
  needed.
- Resource functions themselves (the `listThings`-shaped wrappers) aren't unit tested in
  isolation; they're exercised through `hooks/` tests via MSW instead. The MSW handlers in
  `src/test/handlers.ts` must mirror each resource function's path/method/query-param shape
  exactly — a handler that silently diverges from the real endpoint makes hook tests pass against
  a backend contract that doesn't exist. See `TESTING.md`'s "hooks" section.

## Don't

- Don't put request/response shaping logic in components or hooks — it belongs in the resource
  file, close to the fetch call.
- Don't let a new filter field exist on the real API's filter type without also handling it in
  `mock/data.ts` — a mock that silently ignores a filter makes design-mode review misleading.
