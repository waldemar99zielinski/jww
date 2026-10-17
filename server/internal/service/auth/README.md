# internal/service/auth

**Role.** Authentication use cases: password login and bearer-token verification. Not CRUD. Owns **no** unit of work (only single-row lookups).

## Files

- `repository.go` — `UserRepository` (smallest possible port): `GetByID`, `GetByUsername` (all `ctx`, `uow...`).
- `auth.go` — `PasswordHasher`, `TokenService`, and `Service{ repository, passwordHasher, tokenService }`; `NewService(...)`.
- `Login(ctx, username, password) (string, error)` — lookup; `user == nil || !Verify(...)` → `domain.NewAuthenticationError()`; else `tokenService.Issue(user.ID)`.
- `Authenticate(ctx, token) (*domain.User, error)` — `Verify`; any error/nil → generic `AuthenticationError`; then `GetByID`; nil → `AuthenticationError`.

## Rules

- **One generic `AuthenticationError`** for every failure (bad username/password/token). Never leak which.
- `PasswordHasher` is defined **independently** here and in `service/user` (one adapter satisfies both).
- No `uowManager`; no typed NotFound (auth failures are `AuthenticationError`, not `UserNotFoundError`).
- Imports: `domain`, `unit_of_work`, `uuid` only.

## Avoid

- Distinguishing expired vs invalid vs unknown.
- Reusing `user.Repository` here.
- Creating a UoW for a single lookup.
- Logging or returning password/token/hash material.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package auth_test`), testify, local hand-written fakes for
`UserRepository`/`PasswordHasher`/`TokenService` — see `server/TESTING.md`.
`Login` needs: happy path issues a token; wrong password AND unknown
username produce the exact same `AuthenticationError` (anti-enumeration —
assert the error *messages* match, not just that both error); a repo
failure propagates as-is (`errors.Is`), never collapsed into
`AuthenticationError`. `Authenticate` needs the same repo-error-propagates
distinction, plus invalid-token and valid-token-for-a-deleted-user both
producing `AuthenticationError`.
