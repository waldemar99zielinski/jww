# scripts/bootstrap

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).** Standalone tool — not part of the server binary.

**Role.** Seed the local development test accounts straight into the database, using the same `gateway` repository, Argon2id `security` hasher, and `user.Service` the server uses, so each account logs in through `/v1/auth/login` like any other. Unlike [`signup`](../signup/README.md), which creates one account from its parameters, this tool always seeds the same fixed set of throwaway logins.

The accounts live in code (`testAccounts` in `bootstrap.go`); each password is identical to its username:

| Username | Password | Display name |
|---|---|---|
| `julcia` | `julcia` | `Julcia` |
| `wozi` | `wozi` | `Wozi` |
| `baby` | `baby` | `Baby` |

## Files

- `bootstrap.go` — the `testAccounts` to seed and `Seed`, which reuses [`scripts/signup`](../signup/README.md)'s `Client` to create each account (skipping ones that already exist).
- `cmd/main.go` — the command: resolves the database URL, configures the logger, opens the pgx pool, and calls `Seed`.

## Usage

From the `server/` directory, pass the database DSN as a flag:

```bash
go run ./scripts/bootstrap/cmd -database-url postgres://user:pass@localhost:5432/finance
```

…or as an environment variable; the flag wins when both are set:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/finance go run ./scripts/bootstrap/cmd
```

Supplying neither fails with an error.

## Parameters

| Flag | Environment variable | Required | Meaning |
|---|---|---|---|
| `-database-url` | `DATABASE_URL` | yes | `postgres://` DSN of the target database. |

## Notes

- Does **not** run migrations — bring the schema up first (`go run ./cmd/server`, or `go run ./cmd/migrate up`), so the `users` table exists.
- **Idempotent**: an account whose username already exists is skipped, so re-running is safe.

## Testing requirement for this directory

Deferred — see [`scripts/README.md`](../README.md#testing-requirement-for-this-directory).
