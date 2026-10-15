# utils

**Clean Architecture layer: Frameworks & Drivers (shared helpers).** Outer-layer utilities usable by any layer that may import them.

**Role.** Dependency-light, reusable helpers shared across layers.

## Packages

- [`configuration/`](configuration/README.md) — generic env-var struct binder (`Load[T]()`).
- [`logger/`](logger/README.md) — structured `log/slog` wrapper with component labels and per-request context.

## Rules

- A util package may import `internal/domain` and other utils, but **never** `internal/infrastructure`, `internal/adapter`, or `cmd`.
- Keep utils generic — a util that names a specific resource belongs in that resource's layer instead.
- `configuration` must stay stdlib-only.

## Testing requirement for this directory

No tests of its own — a parent directory with no code. See `configuration/`
and `logger/`'s own READMEs, and `server/TESTING.md`.
