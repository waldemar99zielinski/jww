# scripts/bootstrap

One-command local-development bootstrap. It applies migrations, creates the
standard `julcia`, `wozi`, and `baby` accounts and PLN/USD/EUR currencies, then
adds:

- deterministic exchange rates for **every day** from two calendar months ago
  through today (no external rate API is required);
- 5,000 varied income and expense transactions spread over the same period;
- 75 PDF/JPEG attachments uploaded to the configured S3-compatible store.

Transactions vary by user, category, currency, source, amount, name, and date.
The fixture assets in `assets/` are embedded into the command, so the command
works regardless of its current working directory.

## Usage

Start the local Postgres and MinIO stack, then run from `server/` (the values in
`server/.env` are loaded by the same configuration mechanism as other tools):

```bash
go run ./scripts/bootstrap/cmd
```

Or provide values explicitly:

```bash
DATABASE_URL=postgres://finance:finance@localhost:5432/finance \
STORAGE_BUCKET=jww STORAGE_REGION=us-east-1 \
STORAGE_ENDPOINT=http://localhost:9000 STORAGE_FORCE_PATH_STYLE=true \
STORAGE_ACCESS_KEY_ID=jww STORAGE_SECRET_ACCESS_KEY=jww-development \
go run ./scripts/bootstrap/cmd
```

| Flag | Environment variable | Default |
|---|---|---:|
| `-database-url` | `DATABASE_URL` | required |
| `-transactions` | `BOOTSTRAP_TRANSACTIONS` | 5000 |
| `-attachments` | `BOOTSTRAP_ATTACHMENTS` | 75 |
| `-random-seed` | `BOOTSTRAP_RANDOM_SEED` | 20261025 |

Set attachments to `0` to seed without object storage. Re-running is safe:
rates are upserted, transactions are recognized by their development marker,
and an existing seeded attachment is not uploaded again. To regenerate dates
relative to a later day, recreate the disposable local database/volume first.

The accounts use their username as their password. This command is for local,
disposable data only and must not be run against production.
