# routers/user

**Role.** User CRUD (protected).

## Routes

| Method | Path | Router | Handler |
|---|---|---|---|
| GET | `/v1/users` | protected | `listUsersHandler` |
| GET | `/v1/users/{userID}` | protected | `getUserHandler` |
| PUT | `/v1/users/{userID}` | protected | `updateUserHandler` |
| DELETE | `/v1/users/{userID}` | protected | `deleteUserHandler` |

`Mount(_, protected, deps)` — `public` is unused.

## Notes

- `userID` is a UUID (`uuid.Parse`).
- Update fetches the existing user, mutates `DisplayName`, then calls `Update`.
- Delete does a `GetByID` then `Delete`; returns `204`.
- Responses never include `PasswordHash` (`schema.UserResponse`).

## Testing requirement for this directory

**Location:** `tests/` subdirectory (black-box, imports this package's
exported API — see `server/TESTING.md#conventions`).

`httptest` + testify, `user.NewService` built from a local hand-written
fake `Repository` (no real Postgres) — see `server/TESTING.md`. Every
handler needs its happy path plus not-found → 404 (get/update/delete); also
assert the list/get response JSON never contains `password_hash`.
