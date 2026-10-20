# pages/settings/passkeys/

Manage biometric (Face ID/Touch ID/...) sign-in credentials: list enrolled devices, enroll this
one, and remove any of them. Standard container/view structure (see `pages/README.md`). Entry
point is a Settings row (`Face ID & biometrics`), shown only when
`usePlatformAuthenticator()` reports a platform authenticator is available.

## Things to know

- Enrollment runs the full WebAuthn ceremony client-side: `useEnrollWebAuthnCredential`
  (`hooks/useWebAuthnCredentials.ts`) calls `registerBegin()` → `registerCredential()`
  (`lib/webauthn.ts`, prompts Face ID/Touch ID/etc.) → `registerFinish()`. A rejected ceremony
  (user cancelled, no authenticator, server rejected the attestation) surfaces as a generic
  "That didn't work — try again." — never a raw browser error message.
- Deletion asks for confirmation via `components/RemovePasskeyAction.tsx`, a `BottomSheet`-based
  drawer mirroring `pages/transactions/detail/components/DeleteTransactionAction.tsx` — this
  codebase's standard for a destructive confirmation (no shared `ConfirmDialog` abstraction exists;
  every one wraps `components/ui/BottomSheet` directly).
- `api/webauthn.ts` is the one resource file (like `api/auth.ts`) that doesn't follow the
  mock-branch (`isDesignMode()`) shape — WebAuthn needs a real browser ceremony design mode's
  "no backend" philosophy can't fake. Design mode hides the Face ID UI entirely instead (see
  `usePlatformAuthenticator`), which already reports unavailable there. The Settings row that
  links here is hidden that way, but the route itself is still directly reachable, so
  `index.tsx` also redirects to `/settings` when `isDesignMode()` — the listing query is passed
  `enabled: false` in that case so no real request fires before the redirect.

## Testing

- Container: MSW-mock `useWebAuthnCredentials`/`useEnrollWebAuthnCredential`/
  `useDeleteWebAuthnCredential`'s network calls; spy on `lib/webauthn`'s `registerCredential` to
  stand in for the real Face ID/Touch ID prompt (jsdom has no WebAuthn API at all). Cover: listing,
  the empty state, enroll happy path (begin → ceremony → finish → relist, name field clears),
  enroll failure (ceremony rejects or server rejects) shows the generic error and keeps the typed
  name, and the removal drawer opens on request, only fires the DELETE on "Remove permanently"
  (not on "Cancel"), and stays open with an inline error if the request fails.
- `lib/webauthn.test.ts` is the one place that stubs the actual global `PublicKeyCredential`/
  `navigator.credentials` — everywhere else mocks the `lib/webauthn` module functions directly.
