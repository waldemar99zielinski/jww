# store/

Redux Toolkit store — **session/auth state only**. Server data (transactions, categories,
currencies, summaries, users) must never be duplicated here; that's TanStack Query's job (see
`hooks/`). If you're about to add a new slice, first ask: is this truly global client state that
needs to be read outside a component tree, or would a query hook / local `useState` do? Most new
state is the latter.

## Files

- `index.ts` — store setup (`configureStore`), root reducer.
- `hooks.ts` — typed `useAppDispatch`/`useAppSelector` — use these instead of the untyped
  `react-redux` hooks everywhere in the app.
- `authSlice.ts` — the one slice. Holds the current user + auth status, with three thunks:
  `restoreSession` (on boot, from a stored token), `login`, `logout`. `restoreSession` and
  `login` both short-circuit via a `condition` guard in design mode or when there's no stored
  token — check that guard before adding new async logic to this slice.

## Why auth state lives in Redux and not TanStack Query

The boot sequence (`App.tsx`) needs to dispatch `restoreSession()` and await it *before*
anything renders, outside of any component — that requires an importable, ambient store, which
TanStack Query's component-scoped cache doesn't give you. The actual bearer token itself lives
in `lib/authToken.ts`, not in this slice's state, specifically to avoid an import cycle
(`client.ts` → `store` → `authSlice` → `api/auth` → `client.ts`) — see that file's comment.

## Testing

- `authSlice.ts`: test the reducer directly (`authReducer(state, action)`) for the three thunks'
  pending/fulfilled/rejected cases — no store/React needed. The `condition` guards that
  short-circuit `restoreSession`/`login` in design mode or with no stored token are the one
  non-trivial piece of logic here and are worth a dedicated test each — a guard that silently
  skips a dispatch is exactly the sharp-edge logic worth pinning down, per `TESTING.md`'s
  philosophy.
- Don't test through a mounted component for this — a plain reducer/thunk test is faster and more
  direct.

## Don't

- Don't add a new slice for server data — use a hook in `hooks/` instead.
- Don't read `store.getState()` ad hoc outside `authSlice.ts`/`lib/authToken.ts`-style
  boot-time code — inside components, use `hooks.ts`'s typed selector hook.
