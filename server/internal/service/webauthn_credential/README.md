# internal/service/webauthn_credential

**Role.** CRUD use cases for the biometric (WebAuthn/Face ID/Touch ID/...)
credentials attached to an existing user. Never creates a user — see
`internal/service/user` for that; this package only attaches/detaches an
additional login method to an account that already has a password. The
actual WebAuthn ceremony (challenge generation, attestation/assertion
verification) lives in `adapter/security/webauthn` and
`service/workflow.WebAuthnWorkflow`, not here.

## Files

- `repository.go` — `Repository`: `Create`, `GetByID`, `GetByCredentialID`
  (login resolution), `GetByUserID` (listing), `UpdateSignCount`, `Delete`
  (all `ctx`, `uow...`).
- `ceremony.go` — the two **outbound** ports `service/workflow.WebAuthnWorkflow`
  depends on directly (this package doesn't call them itself): `Ceremony`
  (the WebAuthn Begin/Finish calls, implemented by
  `adapter/security/webauthn.Adapter`) and `ChallengeTokens` (the short-lived
  challenge-token sign/verify calls, implemented by
  `adapter/security/jwt.WebAuthnChallengeService`). Every type in both is a
  domain type, a stdlib primitive, or an opaque JSON `[]byte`/`json.RawMessage`
  — this package (and the workflow) never import the WebAuthn library or JWT
  adapter directly, matching `currency_exchange_rate.RateSource`'s precedent.
- `webauthn_credential.go` — `Service{ repository, uowManager }`; `NewService(...)`.
- `Create(ctx, userID, name, credentialID, publicKey, attestationType, aaguid, signCount, transports, uow...)`
  — assigns `ID: uuid.New()`, `CreatedAt: time.Now().UTC()`, `Validate()`, repo create.
- `List(ctx, userID, uow...)` — passthrough to `GetByUserID`.
- `GetByCredentialID(ctx, credentialID, uow...)` — passthrough; **deliberately
  returns `(nil, nil)` on a miss, never a typed `NotFoundError`** — see Rules.
- `TouchSignCount(ctx, id, signCount, lastUsedAt, uow...)` — passthrough write.
- `Delete(ctx, userID, credentialID)` — own-UoW template; 404s
  (`domain.NewWebAuthnCredentialNotFoundError`) on both "doesn't exist" and
  "belongs to a different user".

## Rules

- Follow the shared service conventions ([parent README](../README.md)).
- `GetByCredentialID` is the one method in this package that does **not**
  translate `(nil, nil)` into a typed error — the login workflow needs that
  miss to collapse into the exact same generic `AuthenticationError` as
  every other login failure (wrong password, unknown username, expired
  token); a `WebAuthnCredentialNotFoundError` here would let a caller
  distinguish "no such credential" from "wrong something else".
- `Delete` checks `existing.UserID != userID` **before** deleting, inside the
  same transaction as the existence check — never delete-then-check.

## Avoid

- Calling `user.Service.Create` from anywhere in this package — this package
  cannot create a user, by construction.
- Translating `GetByCredentialID`'s miss into a typed `NotFoundError`.
- Importing sibling services (imports: `domain`, `unit_of_work`, `uuid`, `time`).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package webauthn_credential_test`), testify, a local
hand-written fake `Repository`/`unit_of_work.Manager`/`UnitOfWork` — see
`server/TESTING.md`. Needs: `Create` sets `ID`/`CreatedAt` and never calls
the repo on a validation failure; `List` passes `userID` through; `Delete`
404s identically whether the row doesn't exist or belongs to another user,
plus `Restore`-on-every-error-path and `Finish`-exactly-once-on-success;
`GetByCredentialID`'s miss stays `(nil, nil)`, not an error.
