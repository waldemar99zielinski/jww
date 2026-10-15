import { describe, expect, it } from 'vitest'
import { formatShortDate, formatTime, startOfMonth, toISODate } from '@/lib/dates'

describe('toISODate', () => {
  it('formats a date as YYYY-MM-DD', () => {
    expect(toISODate(new Date('2026-03-05T12:34:56Z'))).toBe('2026-03-05')
  })
})

describe('startOfMonth', () => {
  it('returns the first day of the given date\'s month', () => {
    const result = startOfMonth(new Date(2026, 9, 15))
    expect(result.getDate()).toBe(1)
    expect(result.getMonth()).toBe(9)
    expect(result.getFullYear()).toBe(2026)
  })
})

describe('formatTime', () => {
  it('formats an ISO datetime as HH:MM in 24h form', () => {
    expect(formatTime('2026-10-10T09:05:00Z')).toMatch(/^\d{2}:\d{2}$/)
  })
})

describe('formatShortDate', () => {
  it('formats an ISO date as "D Mon"', () => {
    expect(formatShortDate('2026-01-02')).toBe('2 Jan')
  })
})
