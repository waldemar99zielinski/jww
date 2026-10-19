import { describe, expect, it } from 'vitest'
import { formatAmount, formatSignedAmount } from '@/lib/currency'

describe('formatAmount', () => {
  it('formats a positive amount as currency', () => {
    expect(formatAmount(10)).toContain('10')
  })

  it('formats zero', () => {
    expect(formatAmount(0)).toMatch(/0/)
  })

  it('formats a negative amount differently from its positive counterpart', () => {
    expect(formatAmount(-5)).not.toBe(formatAmount(5))
  })

  it('uses the supplied currency instead of the PLN fallback', () => {
    expect(formatAmount(10, 'EUR')).toContain('€')
    expect(formatAmount(10, 'USD')).toContain('USD')
    expect(formatAmount(10, 'EUR')).not.toBe(formatAmount(10, 'PLN'))
  })
})

describe('formatSignedAmount', () => {
  it('prefixes income with a plus sign and the unsigned formatted amount', () => {
    expect(formatSignedAmount(10, 'income')).toBe(`+${formatAmount(10)}`)
  })

  it('prefixes expense with a minus sign and the unsigned formatted amount', () => {
    expect(formatSignedAmount(10, 'expense')).toBe(`−${formatAmount(10)}`)
  })

  it('always shows the absolute value regardless of the input sign', () => {
    expect(formatSignedAmount(-10, 'expense')).toBe(formatSignedAmount(10, 'expense'))
    expect(formatSignedAmount(-10, 'income')).toBe(formatSignedAmount(10, 'income'))
  })
})
