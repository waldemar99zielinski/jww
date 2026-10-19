import { describe, expect, it } from 'vitest'
import {
  amountToCents,
  decimalStringToAmountDisplay,
  normalizeAmountInput,
  normalizeSignedAmountInput,
  sanitizeAmountInput,
  sanitizeSignedAmountInput,
} from './amount'

describe('sanitizeAmountInput', () => {
  it('keeps digits and a single comma, dropping dots and other characters', () => {
    expect(sanitizeAmountInput('1a2.3b4')).toBe('1234')
    expect(sanitizeAmountInput('12,34')).toBe('12,34')
  })

  it('collapses extra commas, keeping only the first as the separator', () => {
    expect(sanitizeAmountInput('1,2,3')).toBe('1,23')
    expect(sanitizeAmountInput(',,')).toBe(',')
  })

  it('clamps the fraction to two digits', () => {
    expect(sanitizeAmountInput('12,345')).toBe('12,34')
  })
})

describe('signed line-item amount input', () => {
  it('preserves one leading minus for discounts', () => {
    expect(sanitizeSignedAmountInput('-1a2,345')).toBe('-12,34')
    expect(sanitizeSignedAmountInput('-')).toBe('-')
    expect(sanitizeSignedAmountInput('12,34')).toBe('12,34')
  })

  it('normalizes negative values without losing the sign', () => {
    expect(normalizeSignedAmountInput('-5')).toBe('-5,00')
    expect(normalizeSignedAmountInput('-5,5')).toBe('-5,50')
    expect(normalizeSignedAmountInput('5,5')).toBe('5,50')
  })

  it('converts a signed display value to negative cents', () => {
    expect(amountToCents('-1,00')).toBe(-100)
  })
})

describe('normalizeAmountInput', () => {
  it('pads to exactly two decimals', () => {
    expect(normalizeAmountInput('5')).toBe('5,00')
    expect(normalizeAmountInput('5,5')).toBe('5,50')
    expect(normalizeAmountInput('5,50')).toBe('5,50')
    expect(normalizeAmountInput('5,')).toBe('5,00')
  })

  it('drops redundant leading zeros and keeps an empty field empty', () => {
    expect(normalizeAmountInput('007')).toBe('7,00')
    expect(normalizeAmountInput('')).toBe('')
    expect(normalizeAmountInput('0,00')).toBe('0,00')
  })
})

describe('amountToCents', () => {
  it('converts the display string to whole minor units', () => {
    expect(amountToCents('42,50')).toBe(4250)
    expect(amountToCents('42,5')).toBe(4250)
    expect(amountToCents('42')).toBe(4200)
    expect(amountToCents(',50')).toBe(50)
  })

  it('treats an empty field as zero', () => {
    expect(amountToCents('')).toBe(0)
  })
})

describe('decimalStringToAmountDisplay', () => {
  it('reformats a server decimal string into the comma display form', () => {
    expect(decimalStringToAmountDisplay('24.90')).toBe('24,90')
    expect(decimalStringToAmountDisplay('24.9')).toBe('24,90')
    expect(decimalStringToAmountDisplay('24')).toBe('24,00')
  })

  it('passes an already-formatted display value through unchanged', () => {
    expect(decimalStringToAmountDisplay('24,90')).toBe('24,90')
  })

  it('accepts the bare JSON number the Go Money type marshals', () => {
    expect(decimalStringToAmountDisplay(24.9)).toBe('24,90')
    expect(decimalStringToAmountDisplay(42.5)).toBe('42,50')
  })
})
