const formatter = new Intl.NumberFormat('pl-PL', { style: 'currency', currency: 'PLN' })

export function formatAmount(amount: number): string {
  return formatter.format(amount)
}

export function formatSignedAmount(amount: number, type: 'income' | 'expense'): string {
  const sign = type === 'income' ? '+' : '−'
  return `${sign}${formatter.format(Math.abs(amount))}`
}
