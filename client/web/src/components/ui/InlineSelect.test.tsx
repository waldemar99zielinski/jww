import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { InlineSelect } from '@/components/ui/InlineSelect'

describe('InlineSelect', () => {
  it('calls onChange with the selected option value', async () => {
    const onChange = vi.fn()
    render(
      <InlineSelect aria-label="Period" value="month" onChange={onChange}>
        <option value="month">Month</option>
        <option value="year">Year</option>
      </InlineSelect>,
    )

    await userEvent.selectOptions(screen.getByRole('combobox', { name: 'Period' }), 'year')

    expect(onChange).toHaveBeenCalled()
  })
})
