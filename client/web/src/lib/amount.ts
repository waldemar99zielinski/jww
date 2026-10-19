/**
 * Amount input is a *display* string, not the wire value: digits with an
 * optional comma and at most two fractional digits (e.g. "42,50"). The comma
 * and its decimals exist only for the user; what the API receives is always
 * whole minor units ("cents"), see `amountToCents`.
 */

/**
 * Keeps digits and a single comma, dropping everything else (in particular the
 * dot, which is never a valid separator here), and clamps the fraction to two
 * digits. Runs on every keystroke.
 */
export function sanitizeAmountInput(raw: string): string {
  const digitsAndCommas = raw.replace(/[^\d,]/g, '')
  const commaIndex = digitsAndCommas.indexOf(',')
  if (commaIndex === -1) return digitsAndCommas
  const intPart = digitsAndCommas.slice(0, commaIndex)
  const fracPart = digitsAndCommas.slice(commaIndex + 1).replace(/,/g, '').slice(0, 2)
  return `${intPart},${fracPart}`
}

/**
 * Same display rules as `sanitizeAmountInput`, but preserves one leading minus
 * so discount/adjustment line items can be entered. A lone "-" is retained
 * while the user is still typing.
 */
export function sanitizeSignedAmountInput(raw: string): string {
  const negative = raw.trimStart().startsWith('-')
  const unsigned = sanitizeAmountInput(raw.replace(/-/g, ''))
  if (!negative) return unsigned
  return unsigned === '' ? '-' : `-${unsigned}`
}

/** Normalizes a signed line-item amount, preserving the leading minus. */
export function normalizeSignedAmountInput(display: string): string {
  if (display === '') return ''
  const negative = display.startsWith('-')
  const normalized = normalizeAmountInput(negative ? display.slice(1) : display)
  return negative ? `-${normalized}` : normalized
}

/**
 * Pads a display string to exactly two decimals so it reads as money:
 * "5" -> "5,00", "5,5" -> "5,50", "5,50" stays put. An empty field stays
 * empty so the placeholder can show. Runs when the field loses focus.
 */
export function normalizeAmountInput(display: string): string {
  if (display === '') return ''
  const [rawInt = '', rawFrac = ''] = display.split(',')
  const intPart = rawInt.replace(/^0+(?=\d)/, '') || '0'
  const fracPart = rawFrac.padEnd(2, '0')
  return `${intPart},${fracPart}`
}

/**
 * Converts a display string to whole minor units (cents): "42,50" and "42,5"
 * both become 4250. Tolerates an un-normalized value (a keystroke before blur)
 * and treats a missing integer part as zero (",50" -> 50).
 */
export function amountToCents(display: string): number {
  if (display === '') return 0
  const [rawInt = '', rawFrac = ''] = display.split(',')
  const intPart = rawInt === '' ? '0' : rawInt
  const fracPart = (rawFrac + '00').slice(0, 2)
  return Number(`${intPart}${fracPart}`)
}

/** Converts whole cents to the comma display form: 4127 -> "41,27". */
export function centsToAmountDisplay(cents: number): string {
  return decimalStringToAmountDisplay(cents / 100)
}

/**
 * Reformats an extraction decimal amount ("24.90") into the comma display form ("24,90") so
 * prefilled/seeded amounts match what the input accepts. A value already in display form passes
 * through unchanged.
 *
 * Accepts a number too: the Go `Money` type marshals as a bare JSON number (e.g. `12.34`), so a
 * real API response can hand us `42.5` even though the DTO's TS type says `string`.
 */
export function decimalStringToAmountDisplay(value: string | number): string {
  return normalizeAmountInput(String(value).replace('.', ','))
}
