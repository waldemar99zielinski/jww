# internal/service/user

**Role.** User CRUD. Hashes passwords via an injected `PasswordHasher`. Owns transactions for `Update`/`Delete`.

## Files

- `repository.go` — `Repository`: `Create`, `GetByID(ctx, uuid.UUID, uow...)`, `List`, `Update`, `Delete`.
- `user.go` — `PasswordHasher` (`Hash`, `Verify`) and `Service{ repository, passwordHasher, uowManager }`; `NewService(repo, hasher, uowManager)`. The hasher is an independently owned duplicate of the `auth` port (one adapter satisfies both).
- `Create(ctx, username, password, displayName string, uow...)` — `passwordHasher.Hash(password)` **first** (fail before building the entity), then `domain.User{ID: uuid.New(), PasswordHash: hash, CreatedAt: time.Now().UTC()}`, `Validate()`, repo create.
- `GetByID` — nil → `domain.NewUserNotFoundError(id)`.
- `List` — passthrough.
- `Update`/`Delete` — own-UoW template, no `uow` param.

## Rules

- Hashing happens here, never in a handler or gateway; the plaintext password is never stored on the entity (`PasswordHash` only).
- `Create` accepts an optional `uow`, so a caller like `workflow.SignupWorkflow` can include it in a larger transaction.

## Avoid

- Sharing the `PasswordHasher` interface with `service/auth` (each package owns its port).
- Returning `UserNotFoundError` on auth paths (that's `AuthenticationError`).
- Imports beyond `domain`, `unit_of_work`, `uuid`, `time`.

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

Black-box (`package user_test`), testify, local hand-written fakes for
`Repository`/`PasswordHasher`/`unit_of_work.Manager` — see
`server/TESTING.md`. `Create` needs: the hasher output lands in
`PasswordHash` (never the raw password); a hasher error never reaches the
repo; a validation failure never reaches the repo. `GetByID` needs the
`(nil, nil)` → `NotFoundError` translation. `Update`/`Delete` need
`Restore`-on-every-error-path and `Finish`-exactly-once-on-success.
