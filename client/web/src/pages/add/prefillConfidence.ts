import type { PrefillField } from '@/api/types'

/**
 * Extracted values scored below this threshold are outlined for review. Keep this a single
 * constant so the form highlight and any future confidence UI stay in sync.
 */
export const PREFILL_CONFIDENCE_THRESHOLD = 80

/** An absent score means the extractor did not rate the value, so there is nothing to flag. */
export function isLowConfidence(field: Pick<PrefillField<unknown>, 'confidence'> | undefined): boolean {
  return field?.confidence !== undefined && field.confidence < PREFILL_CONFIDENCE_THRESHOLD
}
