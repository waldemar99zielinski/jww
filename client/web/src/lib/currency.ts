const formatters = new Map<string, Intl.NumberFormat>()

function formatterFor(currencyCode: string): Intl.NumberFormat {
  const normalized = currencyCode.trim().toUpperCase() || 'PLN'
  const cached = formatters.get(normalized)
  if (cached) return cached
  const formatter = new Intl.NumberFormat('pl-PL', {
    style: 'currency',
    currency: normalized,
  })
  formatters.set(normalized, formatter)
  return formatter
}

export function formatAmount(amount: number, currencyCode = 'PLN'): string {
  return formatterFor(currencyCode).format(amount)
}

export function formatSignedAmount(
  amount: number,
  type: 'income' | 'expense',
  currencyCode = 'PLN',
): string {
  const sign = type === 'income' ? '+' : '−'
  return `${sign}${formatAmount(Math.abs(amount), currencyCode)}`
}
