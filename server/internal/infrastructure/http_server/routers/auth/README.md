# routers/auth

**Role.** Authentication routes.

## Routes

| Method | Path | Router | Handler | Service |
|---|---|---|---|---|
| POST | `/v1/auth/login` | public | `loginHandler` | `deps.AuthService.Login` |
| GET | `/v1/auth/me` | protected | `getMeHandler` | (user from context) |

## Notes

- `loginHandler`: `DecodeJSON[schema.LoginRequest]` → `req.Validate()` → `AuthService.Login` → `WriteJSON(200, schema.NewTokenResponse(token))`.
- `getMeHandler`: runs behind `RequireAuth`, so it just returns `*appCtx.User()` via `schema.UserToResponse`. Its deps parameter is intentionally `_`.
- This is the only resource using **both** `public` and `protected`; copy this pattern for other public+protected mixes.

## Avoid

- Registering the same path on both routers (chi panics).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, `auth.NewService` built from a local hand-written
fake `UserRepository` (no real Postgres) — see `server/TESTING.md`. Needs:
login happy path; login wrong-password and unknown-username both 401 with
the same message (anti-enumeration, mirrors the service-layer test);
`get_me` happy path and missing/malformed-token 401 with the
`WWW-Authenticate: Bearer` challenge. `RequireAuth`'s own middleware
behavior is `kernel`'s test suite's job, not this one's — here, just prove
the handlers call into it correctly.
