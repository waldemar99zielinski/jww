# utils/configuration

**Role.** A single generic, reflection-based binder that populates a config struct from environment variables, driven entirely by struct tags.

## API

```go
func Load[T any]() (T, error)
```

- `T` must be a struct; every exported field with an `env` tag is processed.
- Absent + `bind:"required"` → collected as a missing var; `Load` fails once listing **all** missing required vars.
- Absent + `bind:"optional"` → uses `default`, else leaves the zero value.
- Untagged fields are left untouched.
- Supported field types: `string`, all `int` widths, `bool`.

Tag contract:

```go
type Config struct {
    Host string `env:"HTTP_HOST" bind:"required"`
    Port int    `env:"HTTP_PORT" bind:"optional" default:"8000"`
}
```

## How a module defines config

1. In the owning package: `type Config struct { ... }` with `env`/`bind`/`default` tags.
2. `func LoadConfig() (Config, error) { return configuration.Load[Config]() }`.
3. Derived values (e.g. comma-split lists, `host:port`) are **methods** on the struct, not tags.
4. Aggregate the config in `cmd/server/config.go`; document the vars in `.env.example`.

## Rules / avoid

- No slice/map/nested-struct fields expecting them to bind.
- No direct `os.Getenv` elsewhere — always go through a `LoadConfig`.
- Never give a security-sensitive value a default; use `bind:"required"`.
- Keep this package stdlib-only (`fmt`, `os`, `reflect`, `strconv`, `strings`).

## Testing requirement for this directory

**Location:** alongside the source, not a `tests/` subdirectory — these are
white-box (`package configuration`), reaching this package's unexported
`bind`/reflection internals. See `server/TESTING.md#conventions`.

White-box (`package configuration`), stdlib `testing` (no I/O, no need for
testify) — see `configuration_test.go`. Every bind behavior needs a case:
required-and-present, required-and-missing (collected, not fail-fast),
optional-with-default, optional-without-default, each supported field type,
and an unsupported field type's error.
