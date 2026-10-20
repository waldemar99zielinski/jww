# internal/adapter/security

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the `PasswordHasher`/`TokenService` ports.

**Role.** Security adapters. The parent package owns Argon2id password hashing; the [`jwt/`](jwt/README.md) subpackage owns all JWT configuration and token implementations.

## Files

- `argon2_password_hasher.go` — package doc; `Argon2PasswordHasher` + `NewArgon2PasswordHasher()`; `Hash` (random salt via `crypto/rand`, PHC string) and `Verify` (re-derive + `crypto/subtle.ConstantTimeCompare`, boolean-only).
- `jwt/` — `Config`, shared private codec, `AccessTokenService`, `AttachmentGrantService`, and `WebAuthnChallengeService`, split by responsibility.
- `webauthn/` — wraps `github.com/go-webauthn/webauthn` into the Begin/Finish ceremony calls; see its own [README](webauthn/README.md).

## Rules

- **File naming:** `<scheme>_<port>.go`; config always `config.go`.
- Implicit interface conformance to `service/auth.PasswordHasher`, `service/user.PasswordHasher`, `service/auth.TokenService`.
- Errors: wrap with `%w` and a subsystem prefix (`"argon2: generating salt: %w"`, `"security: ... %w"`).
- Security primitives: `crypto/rand`, `crypto/subtle`, `jwt.WithValidMethods`. Never log/return secret material.
- JWT config fields live in `jwt/config.go`; add new variables to `server/.env.example`.

## How to add

1. Add `<scheme>_<port>.go`, struct + `New...` constructor (take `Config` only if configurable).
2. Implement the exact port method signatures.
3. If new config is needed, add tagged fields + document in `.env.example`.
4. Register the constructor in `cmd/server/deps.go`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

No I/O here, so every path is a fast unit test (`package security_test`,
black-box) — no testcontainers needed:

- `Argon2PasswordHasher`: hash/verify round-trip, wrong password fails,
  two hashes of the same password differ (random salt), malformed/garbage
  hash input fails `Verify` without panicking.
- `jwt.AccessTokenService`: sign/verify round-trip, wrong secret fails, expired
  token fails, a token signed with an algorithm other than the verifier's
  pinned one is rejected even with the right secret.
- `LoadConfig`: valid env + defaults succeeds; an unsupported
  `JWT_ALGORITHM` and a missing `JWT_SECRET_KEY` both fail fast (never only
  discovered on a user's first login).

New security primitives land here with the same two requirements: a
round-trip test and a failure-mode test. See
[`server/TESTING.md`](../../../TESTING.md).

## Avoid

- `panic` on malformed hashes (return `false`); returning errors from `Verify`.
- Verifying JWTs without restricting the signing method.
- Shipping a default/example secret.
- Hashing/token logic outside this package; logging passwords/hashes/tokens.
- JWT libraries and token claims in the parent package; keep them in the `jwt` subpackage.
