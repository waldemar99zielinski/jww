import { describe, expect, it } from 'vitest'
import { isLowConfidence, PREFILL_CONFIDENCE_THRESHOLD } from './prefillConfidence'

describe('isLowConfidence', () => {
  it('uses 80 as the review threshold', () => {
    expect(PREFILL_CONFIDENCE_THRESHOLD).toBe(80)
    expect(isLowConfidence({ confidence: 79 })).toBe(true)
    expect(isLowConfidence({ confidence: 80 })).toBe(false)
    expect(isLowConfidence({ confidence: 100 })).toBe(false)
  })

  it('does not flag an unscored value', () => {
    expect(isLowConfidence({})).toBe(false)
    expect(isLowConfidence(undefined)).toBe(false)
  })
})
