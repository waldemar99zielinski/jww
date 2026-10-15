import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { SelectInput } from '@/components/ui/SelectInput'

describe('SelectInput', () => {
  it('calls onChange with the selected option value', async () => {
    const onChange = vi.fn()
    render(
      <SelectInput aria-label="Currency" value="EUR" onChange={onChange}>
        <option value="EUR">EUR</option>
        <option value="USD">USD</option>
      </SelectInput>,
    )

    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Currency' }), 'USD')

    expect(onChange).toHaveBeenCalled()
  })

  it('renders a trailing icon when given', () => {
    render(
      <SelectInput aria-label="Currency" value="EUR" onChange={() => {}} icon={<svg data-testid="chevron" />}>
        <option value="EUR">EUR</option>
      </SelectInput>,
    )
    expect(screen.getByTestId('chevron')).toBeInTheDocument()
  })
})
