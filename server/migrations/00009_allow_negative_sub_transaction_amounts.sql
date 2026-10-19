-- +goose Up
-- Discounts, coupons, and adjustments are legitimate line items with a
-- negative amount. Zero still carries no meaning, so the positive-only CHECK
-- becomes a non-zero CHECK. The parent transaction keeps its own amount > 0
-- constraint, which is enforced after summing the items.
ALTER TABLE transaction_sub_transactions
    DROP CONSTRAINT transaction_sub_transactions_amount_positive;

ALTER TABLE transaction_sub_transactions
    ADD CONSTRAINT transaction_sub_transactions_amount_nonzero CHECK (amount <> 0);

-- +goose Down
ALTER TABLE transaction_sub_transactions
    DROP CONSTRAINT transaction_sub_transactions_amount_nonzero;

ALTER TABLE transaction_sub_transactions
    ADD CONSTRAINT transaction_sub_transactions_amount_positive CHECK (amount > 0);
