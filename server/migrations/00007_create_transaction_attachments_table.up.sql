-- Attachment metadata. The bytes live in private S3-compatible object
-- storage; this table is the only authoritative mapping from an attachment
-- (and, once linked, its transaction) to that object.
--
-- A NULL transaction_id is the entire "unattached draft" state: a file can be
-- uploaded before its transaction exists, then linked exactly once. There is
-- no ordinal, no sequence number, and no attachment-count column — identity is
-- the UUID alone. Linking is a one-way NULL -> transaction_id transition
-- enforced by conditional updates and row locks in the repository, not by a
-- database state machine.
--
-- storage_version is the exact S3 VersionID of the stored object; deleting it
-- (rather than a plain DeleteObject, which only writes a delete marker on a
-- versioned bucket) is what makes permanent deletion real.

CREATE TABLE transaction_attachments (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id     bigint NULL,
    original_filename  text NOT NULL,
    media_type         text NOT NULL,
    byte_size          bigint NOT NULL,
    storage_key        text NOT NULL,
    storage_version    text NOT NULL DEFAULT '',
    etag               text NOT NULL DEFAULT '',
    created_at         timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT transaction_attachments_transaction_fk
        FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE,
    CONSTRAINT transaction_attachments_filename_not_blank CHECK (btrim(original_filename) <> ''),
    CONSTRAINT transaction_attachments_filename_length CHECK (char_length(original_filename) <= 255),
    CONSTRAINT transaction_attachments_media_type_not_blank CHECK (btrim(media_type) <> ''),
    CONSTRAINT transaction_attachments_byte_size_nonnegative CHECK (byte_size >= 0),
    CONSTRAINT transaction_attachments_storage_key_uq UNIQUE (storage_key)
);

-- The lookup every "list this transaction's attachments in a stable order"
-- query makes. Partial on linked rows only; drafts are addressed by primary key.
CREATE INDEX transaction_attachments_transaction_idx
    ON transaction_attachments (transaction_id, created_at, id)
    WHERE transaction_id IS NOT NULL;

