# pages/login/

The login screen — `LoginPage.tsx`. Deliberately **not** split into desktop/mobile variants
(unlike every other page folder): one layout serves every viewport, since there's no meaningful
layout divergence for a single centered form. Don't restructure this into the
`index.tsx`/`types.ts`/`.desktop.tsx`/`.mobile.tsx` shape used elsewhere unless the design
actually needs to diverge by viewport.

Dispatches `login()` from `store/authSlice.ts` directly (auth is the one flow that bypasses
TanStack Query, see `store/README.md`).

There's no sign-up page — see the root README's "Creating a user" section for creating an
account via a direct API call.

## Testing

- Render with MSW mocking the login endpoint: wrong credentials show the error state and stay on
  the page; correct credentials dispatch `login()` (assert on resulting auth state) and the app
  proceeds past the boot gate.
- This page owns the **"can a user log in"** Playwright flow from `TESTING.md`, including the
  unauthenticated-redirect case for `ProtectedRoute` — that spec is the one place the real
  cookie/token round trip gets proven end to end.
