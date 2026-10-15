# scripts/signup

**Clean Architecture layer: Frameworks & Drivers (composition root; outermost).** Standalone tool — not part of the server binary.

**Role.** Create a user straight in the database, using the same `gateway` repository and Argon2id `security` hasher the server uses, so the account can log in through `/v1/auth/login` like any other. Useful for bootstrapping a first account locally, seeding a dev database, or recovering access when the normal signup path is unavailable.

## Files

- `client.go` — the reusable `Client`: builds the server's `user.Service` over a pgx pool and exposes `Create`/`GetByUsername`. `scripts/bootstrap` reuses it.
- `cmd/main.go` — the command: resolves its parameters, configures the logger, opens the pgx pool, and calls `Client.Create`.

## Usage

From the `server/` directory, pass values as command-line flags:

```bash
go run ./scripts/signup/cmd \
  -database-url postgres://user:pass@localhost:5432/finance \
  -username alice -password s3cret
```

…or as environment variables; a flag wins when both are set:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/finance \
  SIGNUP_USERNAME=alice SIGNUP_PASSWORD=s3cret go run ./scripts/signup/cmd
```

Supplying neither source for a required value fails with an error.

## Parameters

| Flag | Environment variable | Required | Meaning |
|---|---|---|---|
| `-database-url` | `DATABASE_URL` | yes | `postgres://` DSN of the target database. |
| `-username` | `SIGNUP_USERNAME` | yes | Username of the account to create (unique). |
| `-password` | `SIGNUP_PASSWORD` | yes | Plaintext password, hashed with Argon2id before insert. |
| `-display-name` | `SIGNUP_DISPLAY_NAME` | no | Display name; defaults to the username. |

## Notes

- Does **not** run migrations — bring the schema up first (`go run ./cmd/server`, or `goose`), so the `users` table exists.
- Fails on a duplicate username (`users_username_uq`), leaving the database unchanged.
- Reuses the server's adapters rather than issuing its own SQL, so the hashed row stays compatible with login.

## Testing requirement for this directory

Deferred — see [`scripts/README.md`](../README.md#testing-requirement-for-this-directory).
