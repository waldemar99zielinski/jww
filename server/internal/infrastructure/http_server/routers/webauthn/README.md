# routers/webauthn

**Role.** Biometric (Face ID/Touch ID/...) credential enrollment and login routes.

## Routes

| Method | Path | Router | Handler |
|---|---|---|---|
| POST | `/v1/webauthn/credentials/register/begin` | protected | `registerBeginHandler` |
| POST | `/v1/webauthn/credentials/register/finish` | protected | `registerFinishHandler` |
| GET | `/v1/webauthn/credentials` | protected | `listCredentialsHandler` |
| DELETE | `/v1/webauthn/credentials/{credentialID}` | protected | `deleteCredentialHandler` |
| POST | `/v1/webauthn/login/begin` | public | `loginBeginHandler` |
| POST | `/v1/webauthn/login/finish` | public | `loginFinishHandler` |

## Notes

- Enrollment (`register/*`) is **always** behind `RequireAuth` — the whole
  point is that only an already-logged-in (password-verified) user can
  attach a credential to their own account, resolved from `appCtx.User().ID`,
  never a request body field. Only the login ceremony is public, since the
  whole point of this feature is signing in without a password.
- Every handler is a thin passthrough to `deps.WebAuthnWorkflow` — no
  business logic here; see `internal/service/workflow`'s README for the
  ceremony/anti-enumeration rules.
- `register/finish` and `login/finish` decode `schema.WebAuthnRegisterFinishRequest`/`WebAuthnLoginFinishRequest` + `Validate()`, same as every other write endpoint in this API.
- `login/finish`'s success response reuses `schema.TokenResponse`/`NewTokenResponse` — identical shape to a password login, so the client's post-login flow doesn't need to know which method was used.

## Avoid

- Registering a `register/*` route on `public` — enrollment without an
  authenticated caller would let anyone attach a credential to an arbitrary
  `userID`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, a real `workflow.WebAuthnWorkflow`/`user.Service`/
`webauthn_credential.Service` built from local hand-written fakes for the
repositories, `Ceremony`, and `ChallengeTokens` ports (no real Postgres or
go-webauthn) — see `server/TESTING.md` and `routers/auth/tests`' `newTestDeps`
for the precedent this mirrors (a genuinely signed token exercises the real
`kernel.RequireAuth` middleware). Needs: register begin/finish happy paths
and finish's validation/bad-JSON 400s; list only returns the caller's own
credentials; delete 404s identically for "doesn't exist" and "belongs to
another user" (and leaves the row untouched); login begin/finish happy path
assuming no `Authorization` header is required; login finish with an
unknown credential is 401; a protected route with no `Authorization` header
is 401 (just proving the wiring — `RequireAuth`'s own behavior is `kernel`'s
test suite's job).
