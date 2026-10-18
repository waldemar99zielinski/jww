# internal/adapter/security/jwt

**Role.** All JWT configuration, shared signing/parsing mechanics, and purpose-specific token adapters.

## Layout

- `config.go` — shared `JWT_*` environment configuration and fail-fast algorithm validation.
- `codec.go` — private signing/parsing implementation shared by every token purpose; always pins the configured algorithm.
- `access_token.go` — user access tokens with the shared `jww` issuer and the `jww-access` audience.
- `attachment_grant.go` — five-minute attachment capabilities with the shared `jww` issuer, the `jww-attachment-content` audience, attachment scope, object version, and preview/download purpose claims.
- `tests/` — black-box configuration, round-trip, expiry, algorithm, secret, and cross-audience rejection tests.

Token-purpose code belongs in its own file. New token types must use the shared codec but define a unique audience and strict claims so one token type can never be accepted as another.
