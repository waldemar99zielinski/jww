# internal/adapter/security/webauthn

**Clean Architecture layer: Interface Adapter (outbound/driven).**

**Role.** Wraps `github.com/go-webauthn/webauthn` into the Begin/Finish
ceremony calls `service/webauthn_credential.Ceremony` declares, translating
between our stored `domain.WebAuthnCredential` rows and the library's own
types. Nothing outside this package ever imports go-webauthn directly — the
library's `*protocol.CredentialCreation`/`*protocol.CredentialAssertion` and
`*webauthn.SessionData` are marshaled to plain `[]byte` JSON at this
package's exported boundary, so `Ceremony` (defined in
`service/webauthn_credential`, which this package implements implicitly)
never needs to import go-webauthn either.

## Files

- `config.go` — `Config` (`WEBAUTHN_RP_ID`, `WEBAUTHN_RP_DISPLAY_NAME`,
  `WEBAUTHN_RP_ORIGINS`, the last comma-split via `RPOriginsList()` —
  mirrors `http_server.Config.CorsOriginsList`). Both `RPID`/`RPOrigins` are
  `bind:"required"`: there is no sane default, so a missing value fails
  server startup rather than silently accepting the wrong origin.
- `user_adapter.go` — `webAuthnUser` (unexported) implements the library's
  `webauthn.User` interface over a `domain.User` + its
  `[]domain.WebAuthnCredential`; `toLibraryCredential`/`fromLibraryCredential`
  convert our stored fields to/from the library's `webauthn.Credential`.
- `webauthn.go` — `Adapter{ lib *webauthn.WebAuthn }`, `New(cfg)`.
  `BeginRegistration`/`FinishRegistration` (enrollment, always called by an
  already-authenticated user) and `BeginLogin`/`FinishLogin` (discoverable
  login, via `service/webauthn_credential.DiscoverableUserHandler`). Every
  `Begin*` marshals the library's options/session to `[]byte` JSON; every
  `Finish*` unmarshals the session back and wraps the raw credential JSON
  body in an `http.Request` purely to satisfy the library's signature — no
  network involved.

## Rules

- Resident keys (`ResidentKeyRequirementRequired`) + required user
  verification — this is what makes login genuinely usernameless/biometric,
  not just "possession of the device".
- Attestation preference is `none` — no FIDO Metadata Service dependency;
  this app trusts the platform's own UV gate, not an attestation chain.
- `BeginRegistration` excludes the user's existing credentials automatically
  — the library builds the exclusion list from `WebAuthnCredentials()`, so
  callers only need to pass the existing rows through.

## Avoid

- Importing `github.com/go-webauthn/webauthn` anywhere outside this package.
- Returning the library's raw `*webauthn.Credential`/`*webauthn.SessionData`
  types from `Adapter`'s exported surface beyond what the workflow needs —
  keep the boundary in plain `[]byte`/`string`/`domain` types.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

No I/O here, so every path is a fast unit test (`package webauthn_test`):
`LoadConfig`'s required/default-value behavior; `New(cfg)` succeeds for a
valid config; `BeginRegistration` produces an `excludeCredentials` list
containing the caller's existing credential IDs and carries the configured
`AuthenticatorSelection`/`AttestationPreference`; `BeginLogin` produces a
non-empty challenge; `FinishRegistration`/`FinishLogin` reject malformed
credential JSON (the request-wrapping path) without needing a real
authenticator response — a full ceremony round-trip needs a real browser
authenticator and is covered by manual verification, not an automated test.
