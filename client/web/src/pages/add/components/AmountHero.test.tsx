import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { CurrencyDto } from '@/api/types'
import { AmountHero } from './AmountHero'

const currencies: CurrencyDto[] = [
  { code: 'EUR', name: 'Euro', is_base: true },
  { code: 'USD', name: 'US Dollar', is_base: false },
]

describe('AmountHero', () => {
  it('shows the currency code as plain text when there is only one currency', () => {
    render(
      <AmountHero amount="10" currencyCode="EUR" onAmountChange={vi.fn()} currencies={[currencies[0]]} onCurrencyChange={vi.fn()} />,
    )

    expect(screen.queryByRole('combobox', { name: 'Currency' })).not.toBeInTheDocument()
    expect(screen.getByText('EUR')).toBeInTheDocument()
  })

  it('renders a currency select when multiple currencies are available and calls onCurrencyChange', async () => {
    const user = userEvent.setup()
    const onCurrencyChange = vi.fn()
    render(<AmountHero amount="10" currencyCode="EUR" onAmountChange={vi.fn()} currencies={currencies} onCurrencyChange={onCurrencyChange} />)

    await user.selectOptions(screen.getByRole('combobox', { name: 'Currency' }), 'USD')

    expect(onCurrencyChange).toHaveBeenCalledWith('USD')
  })

  it('sanitizes typed input to digits, commas, and dots before calling onAmountChange', async () => {
    const user = userEvent.setup()
    const onAmountChange = vi.fn()
    render(<AmountHero amount="" currencyCode="EUR" onAmountChange={onAmountChange} />)

    await user.type(screen.getByLabelText('Amount'), 'a1b')

    expect(onAmountChange).toHaveBeenCalledWith('1')
  })
})
