import { describe, expect, it } from 'vitest'
import { formatShortDate, formatTime, startOfMonth, toISODate, toISODateTime, toLocalISODate } from '@/lib/dates'

describe('toISODate', () => {
  it('formats a date as YYYY-MM-DD', () => {
    expect(toISODate(new Date('2026-03-05T12:34:56Z'))).toBe('2026-03-05')
  })
})

describe('toLocalISODate', () => {
  it('formats a date as YYYY-MM-DD using local calendar fields', () => {
    // Constructed from local parts, so the result is the same regardless of the runner's timezone.
    expect(toLocalISODate(new Date(2026, 2, 5, 1, 2, 3))).toBe('2026-03-05')
  })

  it('zero-pads single-digit months and days', () => {
    expect(toLocalISODate(new Date(2026, 0, 9))).toBe('2026-01-09')
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

describe('toISODateTime', () => {
  it('combines local date and time into an RFC3339 instant', () => {
    const parsed = new Date(toISODateTime('2026-03-05', '14:30'))
    expect(parsed.getFullYear()).toBe(2026)
    expect(parsed.getMonth()).toBe(2)
    expect(parsed.getDate()).toBe(5)
    expect(parsed.getHours()).toBe(14)
    expect(parsed.getMinutes()).toBe(30)
  })

  it('falls back to midnight for an empty time', () => {
    const parsed = new Date(toISODateTime('2026-03-05', ''))
    expect(parsed.getHours()).toBe(0)
    expect(parsed.getMinutes()).toBe(0)
  })
})
