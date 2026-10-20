# cmd

**Role.** Parent directory for executables. It contains **no `.go` files and no Go package of its own** — just `server/` and `migrate/`.

## Rules

- Do not add `.go` files directly under `cmd/`.
- Each binary lives in its own `cmd/<name>/` package (`package main`) and follows the same composition-root pattern as [`cmd/server`](server/README.md).
- A binary may import anything internal; nothing imports a `cmd` package.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. See
[`cmd/server`](server/README.md#testing-requirement-for-this-directory).
