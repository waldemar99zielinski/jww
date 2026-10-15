# internal/adapter/security

**Clean Architecture layer: Interface Adapter (outbound/driven).** Implements the `PasswordHasher`/`TokenService` ports.

**Role.** Implements the auth/user security ports: password hashing (Argon2id) and token issuing/verification (JWT). The **only** place that knows passwords are Argon2id and tokens are JWTs. Owns `security.Config` (`JWT_*`).

## Files

- `config.go` — `Config{ SecretKey, Algorithm, AccessTokenExpireMinutes }` (`JWT_SECRET_KEY` required, no default; `JWT_ALGORITHM` default `HS256`; `JWT_ACCESS_TOKEN_EXPIRE_MINUTES` default `60`); `LoadConfig()` also rejects an unsupported `Algorithm` (`jwt.GetSigningMethod` returning nil) so a typo fails at startup, not on the first login.
- `argon2_password_hasher.go` — package doc; `Argon2PasswordHasher` + `NewArgon2PasswordHasher()`; `Hash` (random salt via `crypto/rand`, PHC string) and `Verify` (re-derive + `crypto/subtle.ConstantTimeCompare`, boolean-only).
- `jwt_token_service.go` — `JwtTokenService` + `NewJwtTokenService(cfg Config)`; `Issue(userID) (string, error)`, `Verify(token) (*uuid.UUID, error)` using `jwt.WithValidMethods` to pin the algorithm.

## Rules

- **File naming:** `<scheme>_<port>.go`; config always `config.go`.
- Implicit interface conformance to `service/auth.PasswordHasher`, `service/user.PasswordHasher`, `service/auth.TokenService`.
- Errors: wrap with `%w` and a subsystem prefix (`"argon2: generating salt: %w"`, `"security: ... %w"`).
- Security primitives: `crypto/rand`, `crypto/subtle`, `jwt.WithValidMethods`. Never log/return secret material.
- Config fields use `env`/`bind`/`default` tags and `LoadConfig()`; add new vars to `server/.env.example`.

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
- `JwtTokenService`: sign/verify round-trip, wrong secret fails, expired
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
- Imports other than stdlib, `argon2`, `jwt/v5`, `uuid`, `utils/configuration`.
