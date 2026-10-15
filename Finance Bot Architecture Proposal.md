# Personal Finance Bot — Architecture Proposal

**Status: restarting from the data model.** Earlier revisions of this document were built around a Google Sheets/Excel source of truth. That's gone. New direction:

- **Storage:** PostgreSQL — not a spreadsheet, not Google Sheets.
- **Primary interface:** the Telegram Mini App (full-screen — see the design mockup) for adding/browsing.
- **Chat:** stays for quick natural-language capture and conversational Q&A, same as before.
- Everything from the old revisions about Google Drive/Sheets API, `.xlsx` conversion, `openpyxl`/`gspread`, and the spreadsheet-specific formula tricks (`amount_display`, `SUMIFS` mirroring, etc.) is void and removed.

Working one layer at a time. This revision covers only the data model. Everything else (API shape, queue/worker design, bot commands, Mini App wiring) comes after the schema is settled.

---

## 1. Data model — draft

Translating the concepts that were already validated in the spreadsheet version (categories, recurring templates, goals, multi-currency with a historical rate) into a normalized relational schema, with a few genuine improvements a real database makes possible (proper foreign keys, a real audit trail, one `transactions` table instead of parallel Income/Expense sheets).

**This draft assumes defaults where the schema shape genuinely forks on a decision — see §2 for the specific questions those defaults answer.**

### Reference / lookup tables

**`users`**
| column | type | notes |
|---|---|---|
| id | uuid, PK | |
| telegram_user_id | bigint, unique | identifies who's talking, for free, via Telegram |
| display_name | text | |
| created_at | timestamptz | |

Replaces `Config!People`. One row per household member who can log things.

**`currencies`**
| column | type | notes |
|---|---|---|
| code | text, PK | `'PLN'`, `'EUR'` |
| name | text | |
| is_base | boolean | exactly one row true — PLN, the reporting currency everything converts to |

**`categories`**
| column | type | notes |
|---|---|---|
| id | uuid, PK | |
| type | enum(`income`,`expense`) | |
| name | text | |
| active | boolean | |
| sort_order | smallint | |

One table, not two sheets — `type` discriminates. Unlike the spreadsheet's dropdown (a suggestion, not enforced), every `transactions.category_id` is a real foreign key here: a category typo is now structurally impossible.

**`exchange_rates`**
| column | type | notes |
|---|---|---|
| id | bigint, PK | |
| currency_code | text, FK → currencies | |
| rate_to_base | numeric | PLN per 1 unit of `currency_code` |
| recorded_at | timestamptz | |

Append-only log rather than the spreadsheet's single "current rate" cell — "the rate right now" is just the latest row per currency, and "what was the rate on 6 Jan" is a real, answerable query instead of relying on a manually-overridden per-transaction snapshot. (Exact shape pending §2 Q3.)

### Transactional core

**`transactions`**
| column | type | notes |
|---|---|---|
| id | uuid, PK | |
| type | enum(`income`,`expense`) | |
| occurred_on | date | the real-world transaction date, can be backdated |
| amount | numeric(12,2) | in `currency_code`, always positive — sign is implied by `type` |
| currency_code | text, FK → currencies | |
| exchange_rate_to_base | numeric | snapshotted at insert time from `exchange_rates` — immutable once written, so a later rate update never silently rewrites a past transaction's reported value |
| amount_base | numeric(12,2) | `amount * exchange_rate_to_base`, stored (not recomputed live) for fast aggregation |
| category_id | uuid, FK → categories | must match `type` |
| description | text | |
| user_id | uuid, FK → users | who logged it |
| source | enum(`manual`,`recurring`) | how the transaction was created |
| recurring_template_id | uuid, FK → recurring_templates, nullable | set when generated from a template |
| goal_id | uuid, FK → goals, nullable | set when this expense exists because a goal was funded |
| created_at | timestamptz | when the row was actually inserted (audit, "undo last") |
| deleted_at | timestamptz, nullable | soft-delete — pending §2 Q2 |

One table replaces both the Income and Expenses sheets. `type` + `source` + the two nullable FKs (`recurring_template_id`, `goal_id`) carry everything the spreadsheet needed two sheets, a `goal_id` column, and a `recurring_id` column to express.

**`recurring_templates`**
| column | type | notes |
|---|---|---|
| id | uuid, PK | |
| type | enum(`income`,`expense`) | |
| category_id | uuid, FK → categories | |
| description | text | |
| amount | numeric(12,2) | |
| currency_code | text, FK → currencies | |
| frequency | enum(`monthly`,`weekly`,`yearly`,`one_time`) | |
| start_date | date | |
| next_due_date | date | rolled forward after each generated transaction |
| end_date | date, nullable | |
| active | boolean | |
| user_id | uuid, FK → users | |
| last_generated_date | date, nullable | idempotency guard |
| notes | text | |

**`goals`**
| column | type | notes |
|---|---|---|
| id | uuid, PK | |
| category_id | uuid, FK → categories | restricted to `type = 'expense'` |
| name | text | |
| description | text | |
| target_amount | numeric(12,2) | |
| current_amount | numeric(12,2) | default 0 |
| currency_code | text, FK → currencies | |
| target_date | date, nullable | |
| moved_to_expenses | boolean | default false |
| transaction_id | uuid, FK → transactions, nullable | the row created once funded |
| user_id | uuid, FK → users | |
| created_at | timestamptz | |
| notes | text | |

`achieved` is deliberately **not a stored column** — it's `current_amount >= target_amount`, either a Postgres generated column or computed in every query. No formula to keep in sync, no stale cache.

### Settings

**`settings`**
Single-row (or small key/value) table: `starting_balance`, `starting_balance_date`, `timezone`. `default_currency`/display-currency placement depends on §2 Q4.

---

## 2. Open questions — these change the table list above

Asked in chat, not reproduced here to avoid drift — answers get folded back into §1 once in. The four biggest forks: unified `transactions` vs. separate income/expense tables, soft-delete vs. hard delete, a historical exchange-rate log vs. a single current-rate column, and whether Display Currency is a global setting or per-user.

---

## 3. Not yet designed

Everything downstream of the schema — the API/ORM layer, the queue/worker split, bot commands, Mini App ↔ backend wiring, classification (Jev/Claude) — is intentionally not in this revision. Comes after §1 is locked.
