/** DTOs mirroring server/internal/adapter/schema/*.go — the HTTP wire shapes. */

export type TransactionType = 'income' | 'expense'
export type TransactionSource = 'manual' | 'recurring'
export type SummaryPeriod = 'day' | 'week' | 'month' | 'year'

export interface UserDto {
  id: string
  username: string
  display_name: string
  created_at: string
}

/** One enrolled biometric (Face ID/Touch ID/...) credential. */
export interface WebAuthnCredentialDto {
  id: string
  name: string
  created_at: string
  last_used_at: string | null
}

export interface CurrencyDto {
  code: string
  name: string
  is_base: boolean
}

export interface TransactionCategoryDto {
  id: string
  type: TransactionType
  name: string
  active: boolean
  sort_order: number
}

export interface TransactionAttachmentDto {
  id: string
  transaction_id: number | null
  original_filename: string
  media_type: string
  byte_size: number
  created_at: string
}

/** A value the reader proposed for one form field, with optional provenance. */
export interface PrefillField<T = string> {
  value: T
  /** 0..100 when the extractor scores the value; absent means "unscored". */
  confidence?: number
  /** Free-form provenance from the server, e.g. "page 1", "line 12". */
  source?: string
}

/** Mirrors the server's attachment-extraction response. Every field is optional by design. */
export interface TransactionPrefillDto {
  /** Legacy first attachment id, retained by the server for single-file clients. */
  attachment_id?: string
  /** Ordered images/pages that produced this one combined extraction result. */
  attachment_ids: string[]
  type?: PrefillField<TransactionType>
  /** Integer cents, matching TransactionDto.amount. */
  amount?: PrefillField<number>
  currency_code?: PrefillField<string>
  occurred_on?: PrefillField<string>
  /** Purchase time from the attachment as HH:MM, when a time is visible. */
  occurred_at?: PrefillField<string>
  name?: PrefillField<string>
  description?: PrefillField<string>
  /** A category *name* to match against the loaded list — never an id. */
  category_hint?: PrefillField<string>
  /**
   * Proposed line items, with amounts in integer cents (negative for discounts). A
   * suggestion has no ids yet — it is not persisted until the user accepts it.
   */
  sub_transactions?: { name: string; label: string; amount: number; confidence?: number }[]
  warnings?: string[]
  duration_ms?: number
}

/** One persisted line item under a transaction. Amount is whole cents, like `TransactionDto.amount`. */
export interface SubTransactionDto {
  id: string
  name: string
  label: string
  amount: number
}

export interface TransactionDto {
  id: number
  type: TransactionType
  /** Whole minor units (cents). */
  amount: number
  currency_code: string
  category_id: string
  category_name: string
  user_id: string
  source: TransactionSource
  name: string
  created_at: string
  description: string
  attachments: TransactionAttachmentDto[]
  /** Always present and non-null, matching `attachments`; empty when the transaction has no items. */
  sub_transactions: SubTransactionDto[]
}

/**
 * Attachment shape returned by `GET /transactions/{id}` only. Adds a short-lived capability token
 * for safe inline preview (server-minted for raster images and PDF; `null` for everything else).
 * List/create responses keep the lightweight `TransactionAttachmentDto` and never mint grants.
 */
export interface TransactionAttachmentDetailDto extends TransactionAttachmentDto {
  preview_grant: string | null
}

/** `GET /transactions/{id}` response: a transaction whose attachments carry preview grants. */
export interface TransactionDetailsDto extends Omit<TransactionDto, 'attachments'> {
  attachments: TransactionAttachmentDetailDto[]
}

export interface CurrencyStatsDto {
  currency_code: string
  /** Whole minor units (cents). */
  total_income: number
  /** Whole minor units (cents). */
  total_expense: number
  /** Whole minor units (cents). */
  net: number
}

export interface SummaryDto {
  period: SummaryPeriod | null
  period_start: string | null
  period_end: string
  currency_code: string
  /** Whole minor units (cents). */
  total_income: number
  /** Whole minor units (cents). */
  total_expense: number
  /** Whole minor units (cents). */
  net: number
  by_currency: Record<string, CurrencyStatsDto>
  transaction_count: number
}
