import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import type { CurrencyDto } from '@/api/types'
import { AmountHero } from './AmountHero'

const currencies: CurrencyDto[] = [
  { code: 'EUR', name: 'Euro', is_base: true },
  { code: 'USD', name: 'US Dollar', is_base: false },
]

/** Stateful harness so the input actually reflects what the sanitizer/normalizer return. */
function ControlledAmount({ initial = '' }: { initial?: string }) {
  const [value, setValue] = useState(initial)
  return <AmountHero amount={value} currencyCode="EUR" onAmountChange={setValue} />
}

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

  it('sanitizes typed input to digits and commas before calling onAmountChange', async () => {
    const user = userEvent.setup()
    const onAmountChange = vi.fn()
    render(<AmountHero amount="" currencyCode="EUR" onAmountChange={onAmountChange} />)

    await user.type(screen.getByLabelText('Amount'), 'a1b')

    expect(onAmountChange).toHaveBeenCalledWith('1')
  })

  it('drops dots and keeps only a single comma with at most two decimals', async () => {
    const user = userEvent.setup()
    render(<ControlledAmount />)

    const input = screen.getByLabelText('Amount')
    await user.type(input, '12.345,678')

    // The dot never appears; the first comma survives and the fraction is clamped to two digits.
    expect(input).toHaveValue('12345,67')
  })

  it('pads to two decimals on blur, and leaves an empty field empty', async () => {
    const user = userEvent.setup()
    render(<ControlledAmount />)
    const input = screen.getByLabelText('Amount')

    await user.type(input, '5,5')
    await user.tab()
    expect(input).toHaveValue('5,50')

    await user.clear(input)
    await user.type(input, '5')
    await user.tab()
    expect(input).toHaveValue('5,00')

    await user.clear(input)
    await user.tab()
    expect(input).toHaveValue('')
  })

  it('renders the derived amount as read-only text with a note when readOnly is set', () => {
    const onAmountChange = vi.fn()
    render(
      <AmountHero amount="12,48" currencyCode="EUR" onAmountChange={onAmountChange} readOnly note="= sum of 2 items" />,
    )

    expect(screen.getByText('12,48')).toBeInTheDocument()
    expect(screen.getByText('= sum of 2 items')).toBeInTheDocument()
    // Read-only means there is no input to type into at all.
    expect(screen.queryByLabelText('Amount')).not.toBeInTheDocument()
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(onAmountChange).not.toHaveBeenCalled()
  })

  it('keeps the editable input when readOnly is not set', async () => {
    const user = userEvent.setup()
    const onAmountChange = vi.fn()
    render(<AmountHero amount="" currencyCode="EUR" onAmountChange={onAmountChange} />)

    const input = screen.getByLabelText('Amount')
    await user.type(input, '5')

    expect(onAmountChange).toHaveBeenCalledWith('5')
  })
})
