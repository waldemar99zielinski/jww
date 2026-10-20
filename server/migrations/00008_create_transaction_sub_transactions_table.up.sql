-- Line items under a transaction ("sub-transactions"): the products on a
-- receipt, or the split parts of a single expense. A transaction owns zero or
-- more rows; the list is additive, with no ordinal or count column, so identity
-- is the UUID alone and ordering is by creation time then id.
--
-- `label` is free text assigned by the user or an extraction run (e.g.
-- "dairy") and is deliberately not a transaction_categories FK — per-item
-- categorisation is out of scope. `amount` is the same fixed-point
-- numeric(12,2) as transactions.amount (domain.Money is an int64 of cents).
-- Discounts, coupons, and adjustments are legitimate line items with a
-- negative amount, so the amount CHECK rejects zero only, not negative
-- values — the parent transaction keeps its own amount > 0 constraint,
-- enforced after summing the items.
--
-- The name/amount invariants are app-enforced by
-- domain.TransactionSubTransaction.Validate and repeated here as CHECKs so an
-- invalid row cannot be written outside the application.

CREATE TABLE transaction_sub_transactions (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id bigint NOT NULL,
    name           text NOT NULL,
    label          text NOT NULL DEFAULT '',
    amount         numeric(12,2) NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT transaction_sub_transactions_transaction_fk
        FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE,
    CONSTRAINT transaction_sub_transactions_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT transaction_sub_transactions_amount_nonzero CHECK (amount <> 0)
);

-- The lookup every "list this transaction's line items in a stable order"
-- query makes. Unlike transaction_attachments there is no draft state, so the
-- index is not partial.
CREATE INDEX transaction_sub_transactions_transaction_idx
    ON transaction_sub_transactions (transaction_id, created_at, id);
