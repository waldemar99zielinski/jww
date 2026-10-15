const NON_BREAKING_SPACE = String.fromCharCode(0xa0)

/**
 * React Testing Library's `getByText`/`findByText` normalize the *DOM side*'s whitespace
 * (collapsing runs, including non-breaking spaces, to a single ' ') but never touch the query
 * string you pass in. `lib/currency.ts`'s `Intl.NumberFormat` output embeds a real non-breaking
 * space before the currency symbol (e.g. "42,50 zl"), so comparing a rendered amount against a
 * raw `formatAmount(...)`/`formatSignedAmount(...)` result never matches — normalize the expected
 * string the same way first.
 */
export function normalizeText(text: string): string {
  return text.split(NON_BREAKING_SPACE).join(' ')
}
