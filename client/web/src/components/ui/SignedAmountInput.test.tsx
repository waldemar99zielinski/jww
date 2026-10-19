import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'
import { SignedAmountInput } from './SignedAmountInput'

/** Stateful harness, because the parent owns the display string. */
function Controlled({ initial = '' }: { initial?: string }) {
  const [value, setValue] = useState(initial)
  return <SignedAmountInput value={value} onChange={setValue} label="Amount" />
}

describe('SignedAmountInput', () => {
  it('toggles a typed amount negative and back without needing a minus key', async () => {
    const user = userEvent.setup()
    render(<Controlled />)

    const input = screen.getByLabelText('Amount')
    await user.type(input, '8,50')

    const negative = screen.getByRole('button', { name: 'Make amount negative' })
    await user.click(negative)

    expect(input).toHaveValue('-8,50')
    expect(screen.getByRole('button', { name: 'Make amount positive' })).toHaveAttribute('aria-pressed', 'true')

    await user.click(screen.getByRole('button', { name: 'Make amount positive' }))
    expect(input).toHaveValue('8,50')
  })

  it('starts a negative amount from an empty field', async () => {
    const user = userEvent.setup()
    render(<Controlled />)

    await user.click(screen.getByRole('button', { name: 'Make amount negative' }))
    const input = screen.getByLabelText('Amount')
    expect(input).toHaveValue('-')

    await user.type(input, '3')
    await user.tab()
    expect(input).toHaveValue('-3,00')
  })
})
